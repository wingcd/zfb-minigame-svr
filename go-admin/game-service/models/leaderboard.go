package models

import (
	"encoding/json"
	"fmt"
	"time"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/go-redis/redis/v8"
)

// getLeaderboardTableName 获取排行榜表名
func getLeaderboardTableName(appId string) string {
	return utils.GetLeaderboardTableName(appId)
}

// getLeaderboardRedisKey 获取排行榜Redis键名
func getLeaderboardRedisKey(appId, leaderboardName string) string {
	return fmt.Sprintf("leaderboard:%s:%s", appId, leaderboardName)
}

// getUserRankRedisKey 获取用户排名Redis键名
func getUserRankRedisKey(appId, leaderboardName, userId string) string {
	return fmt.Sprintf("rank:%s:%s:%s", appId, leaderboardName, userId)
}

// LeaderboardRedisEntry Redis排行榜条目结构
type LeaderboardRedisEntry struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName,omitempty"`
	Score      int64  `json:"score"`
	ExtraData  string `json:"extraData,omitempty"`
	UpdateTime int64  `json:"updateTime"`
}

// Leaderboard 排行榜模型 - 兼容旧版本API
type Leaderboard struct {
	Id              int64  `orm:"auto" json:"id"`
	LeaderboardName string `orm:"size(100)" json:"leaderboard_name"`
	UserId          string `orm:"size(100)" json:"user_id"`
	Score           int64  `orm:"default(0)" json:"score"`
	ExtraData       string `orm:"type(text)" json:"extra_data"`
	CreatedAt       string `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt       string `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// LeaderboardEntry 排行榜条目模型 - 对应数据库设计的leaderboard_[appid]表
type LeaderboardEntry struct {
	ID              int64     `orm:"pk;auto" json:"id"`
	AppId           string    `orm:"-" json:"appId"`                                                // 应用ID（仅用于逻辑，不存储到数据库）
	LeaderboardName string    `orm:"size(100);column(leaderboard_name)" json:"leaderboardName"`     // 排行榜名称
	PlayerID        string    `orm:"size(100);column(player_id)" json:"playerId"`                   // 玩家ID - 修复JSON字段名
	PlayerName      string    `orm:"size(100);column(player_name)" json:"playerName"`               // 玩家昵称
	Score           int64     `orm:"default(0)" json:"score"`                                       // 分数
	ExtraData       string    `orm:"type(text);column(extra_data)" json:"extraData"`                // 额外数据（JSON格式）
	Rank            int       `orm:"default(0)" json:"rank"`                                        // 排名
	Season          string    `orm:"size(50);default(default)" json:"season"`                       // 赛季
	Category        string    `orm:"size(50);default(general)" json:"category"`                     // 分类
	IsActive        bool      `orm:"default(true);column(is_active)" json:"isActive"`               // 是否活跃
	LastUpdateTime  time.Time `orm:"type(datetime);column(last_update_time)" json:"lastUpdateTime"` // 最后更新时间
	CreatedAt       time.Time `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt       time.Time `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// GetTableName 获取动态表名
func (l *Leaderboard) GetTableName(appId string) string {
	return getLeaderboardTableName(appId)
}

func (le *LeaderboardEntry) GetTableName(appId string) string {
	return getLeaderboardTableName(appId)
}

// SubmitScore 提交分数到排行榜（支持Redis缓存）
func SubmitScore(appId, userId, leaderboardName string, score int64, extraData string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 使用 ON DUPLICATE KEY UPDATE 进行 upsert 操作
	sql := fmt.Sprintf(`
		INSERT INTO %s (leaderboard_name, player_id, player_name, score, extra_data, last_update_time, created_at, updated_at)
		VALUES (?, ?, '', ?, ?, NOW(), NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			score = VALUES(score),
			extra_data = VALUES(extra_data),
			last_update_time = NOW(),
			updated_at = NOW()
	`, tableName)

	_, err := o.Raw(sql, leaderboardName, userId, score, extraData).Exec()
	if err != nil {
		return err
	}

	// 更新数据库排名
	err = updateLeaderboardRanks(o, tableName, leaderboardName)
	if err != nil {
		return err
	}

	// 同步到Redis（如果Redis可用）
	if RedisClient != nil {
		err = syncScoreToRedis(appId, userId, leaderboardName, score, extraData)
		if err != nil {
			// Redis错误不影响主流程，记录日志即可
			fmt.Printf("Redis同步失败: %v\n", err)
		}
	}

	return nil
}

// syncScoreToRedis 同步分数到Redis
func syncScoreToRedis(appId, userId, leaderboardName string, score int64, extraData string) error {
	redisKey := getLeaderboardRedisKey(appId, leaderboardName)

	// 创建Redis排行榜条目
	entry := LeaderboardRedisEntry{
		PlayerID:   userId,
		Score:      score,
		ExtraData:  extraData,
		UpdateTime: time.Now().Unix(),
	}

	// 序列化到JSON
	entryData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// 使用有序集合存储排行榜（分数作为score，用户数据作为member）
	member := &redis.Z{
		Score:  float64(score),
		Member: string(entryData),
	}

	// 添加到Redis有序集合
	err = RedisClient.ZAdd(RedisClient.Context(), redisKey, member).Err()
	if err != nil {
		return err
	}

	// 设置过期时间（24小时）
	RedisClient.Expire(RedisClient.Context(), redisKey, 24*time.Hour)

	return nil
}

func UpdateScore(appId, userId, leaderboardName string, score int64, extraData string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 更新现有记录
	sql := fmt.Sprintf(`
		UPDATE %s 
		SET score = ?, extra_data = ?, last_update_time = NOW(), updated_at = NOW()
		WHERE leaderboard_name = ? AND player_id = ?
	`, tableName)

	result, err := o.Raw(sql, score, extraData, leaderboardName, userId).Exec()
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("排行榜记录不存在")
	}

	// 更新排名
	return updateLeaderboardRanks(o, tableName, leaderboardName)
}

// GetLeaderboard 获取排行榜（优先从Redis读取）
func GetLeaderboard(appId, leaderboardName string, limit int) ([]Leaderboard, error) {
	// 尝试从Redis获取
	if RedisClient != nil {
		leaderboards, err := getLeaderboardFromRedis(appId, leaderboardName, limit)
		if err == nil && len(leaderboards) > 0 {
			return leaderboards, nil
		}
		// Redis失败或没有数据，继续从数据库读取
	}

	// 从数据库获取
	return getLeaderboardFromDB(appId, leaderboardName, limit)
}

// getLeaderboardFromRedis 从Redis获取排行榜
func getLeaderboardFromRedis(appId, leaderboardName string, limit int) ([]Leaderboard, error) {
	redisKey := getLeaderboardRedisKey(appId, leaderboardName)

	// 获取Redis有序集合的前N个成员（按分数降序）
	results, err := RedisClient.ZRevRangeWithScores(RedisClient.Context(), redisKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var leaderboards []Leaderboard
	for _, result := range results {
		// 解析member数据
		var entry LeaderboardRedisEntry
		err := json.Unmarshal([]byte(result.Member.(string)), &entry)
		if err != nil {
			continue // 跳过格式错误的数据
		}

		lb := Leaderboard{
			LeaderboardName: leaderboardName,
			UserId:          entry.PlayerID,
			Score:           entry.Score,
			ExtraData:       entry.ExtraData,
			UpdatedAt:       time.Unix(entry.UpdateTime, 0).Format("2006-01-02 15:04:05"),
		}

		leaderboards = append(leaderboards, lb)
	}

	return leaderboards, nil
}

// getLeaderboardFromDB 从数据库获取排行榜
func getLeaderboardFromDB(appId, leaderboardName string, limit int) ([]Leaderboard, error) {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 查询排行榜数据
	sql := fmt.Sprintf(`
		SELECT 
			id,
			leaderboard_name,
			player_id as user_id,
			score,
			extra_data,
			created_at as create_time,
			updated_at as update_time
		FROM %s
		WHERE leaderboard_name = ? AND is_active = 1
		ORDER BY score DESC, created_at ASC
		LIMIT ?
	`, tableName)

	var results []orm.Params
	_, err := o.Raw(sql, leaderboardName, limit).Values(&results)
	if err != nil {
		return nil, err
	}

	// 转换为Leaderboard结构
	var leaderboards []Leaderboard
	for _, result := range results {
		lb := Leaderboard{}

		if id, ok := result["id"].(int64); ok {
			lb.Id = id
		}
		if lbName, ok := result["leaderboard_name"].(string); ok {
			lb.LeaderboardName = lbName
		}
		if userId, ok := result["user_id"].(string); ok {
			lb.UserId = userId
		}
		if score, ok := result["score"].(int64); ok {
			lb.Score = score
		}
		if extraData, ok := result["extra_data"].(string); ok {
			lb.ExtraData = extraData
		}
		if createTime, ok := result["create_time"].(time.Time); ok {
			lb.CreatedAt = createTime.Format("2006-01-02 15:04:05")
		}
		if updateTime, ok := result["update_time"].(time.Time); ok {
			lb.UpdatedAt = updateTime.Format("2006-01-02 15:04:05")
		}

		leaderboards = append(leaderboards, lb)
	}

	return leaderboards, nil
}

// GetUserRank 获取用户在排行榜中的排名（优先从Redis读取）
func GetUserRank(appId, userId, leaderboardName string) (int, int64, error) {
	// 尝试从Redis获取
	if RedisClient != nil {
		rank, score, err := getUserRankFromRedis(appId, userId, leaderboardName)
		if err == nil && rank > 0 {
			return rank, score, nil
		}
		// Redis失败或没有数据，继续从数据库读取
	}

	// 从数据库获取
	return getUserRankFromDB(appId, userId, leaderboardName)
}

// getUserRankFromRedis 从Redis获取用户排名
func getUserRankFromRedis(appId, userId, leaderboardName string) (int, int64, error) {
	redisKey := getLeaderboardRedisKey(appId, leaderboardName)

	// 查找用户在有序集合中的排名（从0开始，需要+1）
	rank, err := RedisClient.ZRevRank(RedisClient.Context(), redisKey, userId).Result()
	if err == redis.Nil {
		return 0, 0, nil // 用户不在排行榜中
	}
	if err != nil {
		return 0, 0, err
	}

	// 获取用户分数
	score, err := RedisClient.ZScore(RedisClient.Context(), redisKey, userId).Result()
	if err != nil {
		return 0, 0, err
	}

	return int(rank) + 1, int64(score), nil
}

// getUserRankFromDB 从数据库获取用户排名
func getUserRankFromDB(appId, userId, leaderboardName string) (int, int64, error) {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 获取用户分数和排名
	sql := fmt.Sprintf(`
		SELECT score, rank FROM %s 
		WHERE leaderboard_name = ? AND player_id = ? AND is_active = 1
	`, tableName)

	var result []orm.Params
	_, err := o.Raw(sql, leaderboardName, userId).Values(&result)
	if err != nil {
		return 0, 0, err
	}

	if len(result) == 0 {
		return 0, 0, nil // 用户不在排行榜中
	}

	data := result[0]
	var rank int64
	var score int64

	if r, ok := data["rank"].(int64); ok {
		rank = r
	}
	if s, ok := data["score"].(int64); ok {
		score = s
	}

	return int(rank), score, nil
}

// ResetLeaderboard 重置排行榜（同时清理Redis）
func ResetLeaderboard(appId, leaderboardName string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	sql := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE leaderboard_name = ?
	`, tableName)

	_, err := o.Raw(sql, leaderboardName).Exec()
	if err != nil {
		return err
	}

	// 同时清理Redis数据
	if RedisClient != nil {
		redisKey := getLeaderboardRedisKey(appId, leaderboardName)
		RedisClient.Del(RedisClient.Context(), redisKey)
	}

	return nil
}

// GetLeaderboardList 获取排行榜列表（管理后台使用）
func GetLeaderboardList(appId string, page, pageSize int, leaderboardName string) ([]Leaderboard, int64, error) {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 构建查询条件
	whereClause := "WHERE is_active = 1"
	args := []interface{}{}

	if leaderboardName != "" {
		whereClause += " AND leaderboard_name = ?"
		args = append(args, leaderboardName)
	}

	// 查询总数
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s %s
	`, tableName, whereClause)

	var total int64
	err := o.Raw(countSQL, args...).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	offset := (page - 1) * pageSize
	dataSQL := fmt.Sprintf(`
		SELECT 
			id,
			leaderboard_name,
			player_id as user_id,
			score,
			extra_data,
			created_at as create_time,
			updated_at as update_time
		FROM %s %s
		ORDER BY score DESC, created_at ASC
		LIMIT ? OFFSET ?
	`, tableName, whereClause)

	var results []orm.Params
	_, err = o.Raw(dataSQL, append(args, pageSize, offset)...).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 转换为Leaderboard结构
	var leaderboards []Leaderboard
	for _, result := range results {
		lb := Leaderboard{}

		if id, ok := result["id"].(int64); ok {
			lb.Id = id
		}
		if lbName, ok := result["leaderboard_name"].(string); ok {
			lb.LeaderboardName = lbName
		}
		if userId, ok := result["user_id"].(string); ok {
			lb.UserId = userId
		}
		if score, ok := result["score"].(int64); ok {
			lb.Score = score
		}
		if extraData, ok := result["extra_data"].(string); ok {
			lb.ExtraData = extraData
		}
		if createTime, ok := result["create_time"].(time.Time); ok {
			lb.CreatedAt = createTime.Format("2006-01-02 15:04:05")
		}
		if updateTime, ok := result["update_time"].(time.Time); ok {
			lb.UpdatedAt = updateTime.Format("2006-01-02 15:04:05")
		}

		leaderboards = append(leaderboards, lb)
	}

	return leaderboards, total, nil
}

// updateLeaderboardRanks 更新排行榜排名
func updateLeaderboardRanks(o orm.Ormer, tableName, leaderboardName string) error {
	// 使用窗口函数更新排名
	sql := fmt.Sprintf(`
		UPDATE %s t1
		JOIN (
			SELECT id, 
				   ROW_NUMBER() OVER (ORDER BY score DESC, created_at ASC) as new_rank
			FROM %s 
			WHERE leaderboard_name = ? AND is_active = 1
		) t2 ON t1.id = t2.id
		SET t1.rank = t2.new_rank
		WHERE t1.leaderboard_name = ?
	`, tableName, tableName)

	_, err := o.Raw(sql, leaderboardName, leaderboardName).Exec()
	return err
}
