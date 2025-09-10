package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/go-redis/redis/v8"
)

// getLeaderboardTableName 获取排行榜表名
func getLeaderboardTableName(appId string) string {
	return utils.GetLeaderboardTableName(appId)
}

// LeaderboardConfig 排行榜配置结构
type LeaderboardConfig struct {
	Id              int64     `json:"id"`
	AppId           string    `json:"appId"`
	LeaderboardType string    `json:"leaderboardType"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ScoreType       string    `json:"scoreType"`
	MaxRank         int       `json:"maxRank"`
	Enabled         bool      `json:"enabled"`
	Category        string    `json:"category"`
	ResetType       string    `json:"resetType"`
	ResetValue      int       `json:"resetValue"`
	ResetTime       time.Time `json:"resetTime"`
	UpdateStrategy  int       `json:"updateStrategy"`
	Sort            int       `json:"sort"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// getLeaderboardConfig 获取排行榜配置
func getLeaderboardConfig(appId, leaderboardName string) (*LeaderboardConfig, error) {
	o := orm.NewOrm()

	var result []orm.Params
	sql := `SELECT id, app_id, leaderboard_type, name, description, score_type, max_rank, enabled, category, reset_type, reset_value, reset_time, update_strategy, sort, created_at, updated_at FROM leaderboard_config WHERE app_id = ? AND leaderboard_type = ?`
	_, err := o.Raw(sql, appId, leaderboardName).Values(&result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("leaderboard config not found for app_id: %s, leaderboard_type: %s", appId, leaderboardName)
	}

	data := result[0]
	config := &LeaderboardConfig{}

	// 解析各字段
	if id, ok := data["id"].(int64); ok {
		config.Id = id
	}
	if appId, ok := data["app_id"].(string); ok {
		config.AppId = appId
	}
	if leaderboardType, ok := data["leaderboard_type"].(string); ok {
		config.LeaderboardType = leaderboardType
	}
	if name, ok := data["name"].(string); ok {
		config.Name = name
	}
	if description, ok := data["description"].(string); ok {
		config.Description = description
	}
	if scoreType, ok := data["score_type"].(string); ok {
		config.ScoreType = scoreType
	}
	if maxRank, ok := data["max_rank"].(int); ok {
		config.MaxRank = maxRank
	} else if maxRank, ok := data["max_rank"].(int64); ok {
		config.MaxRank = int(maxRank)
	}
	if enabled, ok := data["enabled"].(bool); ok {
		config.Enabled = enabled
	}
	if category, ok := data["category"].(string); ok {
		config.Category = category
	}
	if resetType, ok := data["reset_type"].(string); ok {
		config.ResetType = resetType
	}
	if resetValue, ok := data["reset_value"].(int); ok {
		config.ResetValue = resetValue
	} else if resetValue, ok := data["reset_value"].(int64); ok {
		config.ResetValue = int(resetValue)
	}
	if resetTime, ok := data["reset_time"].(time.Time); ok {
		config.ResetTime = resetTime
	}
	if updateStrategy, ok := data["update_strategy"].(int); ok {
		config.UpdateStrategy = updateStrategy
	} else if updateStrategy, ok := data["update_strategy"].(int64); ok {
		config.UpdateStrategy = int(updateStrategy)
	}
	if sort, ok := data["sort"].(int); ok {
		config.Sort = sort
	} else if sort, ok := data["sort"].(int64); ok {
		config.Sort = int(sort)
	}
	if createdAt, ok := data["created_at"].(time.Time); ok {
		config.CreatedAt = createdAt
	}
	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		config.UpdatedAt = updatedAt
	}

	return config, nil
}

// getLeaderboardConfigId 根据appId和leaderboardName获取type
func getLeaderboardConfigId(appId, leaderboardName string) (int64, error) {
	config, err := getLeaderboardConfig(appId, leaderboardName)
	if err != nil {
		return 0, err
	}
	return config.Id, nil
}

// getLeaderboardRedisKey 获取排行榜Redis键名
func getLeaderboardRedisKey(appId, leaderboardName string) string {
	return fmt.Sprintf("leaderboard:%s:%s", appId, leaderboardName)
}

// LeaderboardRedisEntry Redis排行榜条目结构
type LeaderboardRedisEntry struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName,omitempty"`
	Score      int64  `json:"score"`
	ExtraData  string `json:"extraData,omitempty"`
	UpdateTime int64  `json:"updateTime"`
}

// Leaderboard 排行榜条目模型 - 对应数据库设计的leaderboard_[appid]表
type Leaderboard struct {
	Id        int64                  `orm:"auto" json:"id"`
	Type      string                 `orm:"size(50)" json:"type"`
	UserId    string                 `orm:"size(100);column(player_id)" json:"user_id"`
	Score     int64                  `orm:"default(0)" json:"score"`
	ExtraData string                 `orm:"type(text);column(extra_data)" json:"extra_data"`
	UserInfo  map[string]interface{} `orm:"-" json:"userInfo,omitempty"` // 用户信息，不存储到数据库
	CreatedAt string                 `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt string                 `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// GetTableName 获取动态表名
func (l *Leaderboard) GetTableName(appId string) string {
	return getLeaderboardTableName(appId)
}

// calculateNextResetTime 计算下次重置时间
func calculateNextResetTime(resetType string, resetValue int) *time.Time {
	now := time.Now()
	var nextReset time.Time

	switch resetType {
	case "daily":
		nextReset = now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	case "weekly":
		// 获取本周一的开始时间，然后加一周
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 将周日从0改为7
		}
		daysToMonday := weekday - 1
		startOfWeek := now.AddDate(0, 0, -daysToMonday).Truncate(24 * time.Hour)
		nextReset = startOfWeek.Add(7 * 24 * time.Hour)
	case "monthly":
		// 下个月的第一天
		year, month, _ := now.Date()
		nextReset = time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location())
	case "custom":
		if resetValue > 0 {
			nextReset = now.Add(time.Duration(resetValue) * time.Hour)
		} else {
			return nil
		}
	default:
		return nil
	}

	return &nextReset
}

// checkAndResetLeaderboard 检查并重置排行榜
func checkAndResetLeaderboard(appId, leaderboardName string, config *LeaderboardConfig) error {
	if config.ResetType == "permanent" || config.ResetTime.IsZero() {
		return nil
	}

	now := time.Now()
	if now.After(config.ResetTime) {
		logs.Info("排行榜需要重置: appId=%s, type=%s, resetTime=%v", appId, leaderboardName, config.ResetTime)

		// 清空排行榜数据
		err := ResetLeaderboard(appId, leaderboardName)
		if err != nil {
			return fmt.Errorf("重置排行榜失败: %v", err)
		}

		// 计算下次重置时间
		nextResetTime := calculateNextResetTime(config.ResetType, config.ResetValue)
		if nextResetTime != nil {
			// 更新配置中的重置时间
			err = updateLeaderboardResetTime(appId, leaderboardName, *nextResetTime)
			if err != nil {
				return fmt.Errorf("更新重置时间失败: %v", err)
			}
			logs.Info("排行榜重置完成，下次重置时间: %v", *nextResetTime)
		}
	}

	return nil
}

// updateLeaderboardResetTime 更新排行榜重置时间
func updateLeaderboardResetTime(appId, leaderboardName string, resetTime time.Time) error {
	o := orm.NewOrm()

	sql := `UPDATE leaderboard_config SET reset_time = ?, updated_at = NOW() WHERE app_id = ? AND leaderboard_type = ?`
	_, err := o.Raw(sql, resetTime, appId, leaderboardName).Exec()
	if err != nil {
		logs.Error("更新排行榜重置时间失败: %v", err)
		return err
	}

	return nil
}

// SubmitScore 提交分数到排行榜（支持Redis缓存，包含完整的JS逻辑）
func SubmitScore(appId, userId, leaderboardName string, score int64, extraData string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 1. 验证用户是否存在
	user, err := GetUserByPlayerId(appId, userId)
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %v", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 2. 获取排行榜配置
	config, err := getLeaderboardConfig(appId, leaderboardName)
	if err != nil {
		return fmt.Errorf("获取排行榜配置失败: %v", err)
	}

	// 3. 验证更新策略
	if config.UpdateStrategy < 0 || config.UpdateStrategy > 2 {
		return fmt.Errorf("更新策略异常")
	}

	// 4. 检查是否需要重置排行榜
	err = checkAndResetLeaderboard(appId, leaderboardName, config)
	if err != nil {
		return err
	}

	// 5. 查找现有记录
	var existingRecord []orm.Params
	findSQL := fmt.Sprintf(`SELECT id, score FROM %s WHERE type = ? AND player_id = ?`, tableName)
	_, err = o.Raw(findSQL, leaderboardName, userId).Values(&existingRecord)
	if err != nil {
		return fmt.Errorf("查询现有记录失败: %v", err)
	}

	// 6. 使用安全的插入/更新逻辑，避免重复记录
	var oldScore int64 = 0
	var shouldUpdate bool = true

	// 如果存在记录，获取旧分数并根据策略计算是否需要更新
	if len(existingRecord) > 0 {
		// 安全的类型转换
		if scoreVal, ok := existingRecord[0]["score"]; ok {
			switch v := scoreVal.(type) {
			case int64:
				oldScore = v
			case int:
				oldScore = int64(v)
			case int32:
				oldScore = int64(v)
			case string:
				if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
					oldScore = parsed
				} else {
					return fmt.Errorf("无效的分数格式: %s", v)
				}
			case []byte:
				if parsed, err := strconv.ParseInt(string(v), 10, 64); err == nil {
					oldScore = parsed
				} else {
					return fmt.Errorf("无效的分数格式: %s", string(v))
				}
			default:
				return fmt.Errorf("不支持的分数类型: %T", v)
			}
		}

		// 根据更新策略决定是否更新
		switch config.UpdateStrategy {
		case 0: // 历史最高值
			shouldUpdate = oldScore <= score
		case 1: // 最近记录
			shouldUpdate = true
		case 2: // 历史总和
			shouldUpdate = true
		}
	}

	// 使用 INSERT ... ON DUPLICATE KEY UPDATE 确保不会产生重复记录
	if shouldUpdate {
		var upsertSQL string
		var args []interface{}

		switch config.UpdateStrategy {
		case 0: // 历史最高值 - 只有新分数更高时才更新
			upsertSQL = fmt.Sprintf(`
				INSERT INTO %s (type, player_id, score, extra_data, created_at, updated_at)
				VALUES (?, ?, ?, ?, NOW(), NOW())
				ON DUPLICATE KEY UPDATE
					score = CASE WHEN VALUES(score) > score THEN VALUES(score) ELSE score END,
					extra_data = CASE WHEN VALUES(score) > score THEN VALUES(extra_data) ELSE extra_data END,
					updated_at = NOW()
			`, tableName)
			args = []interface{}{leaderboardName, userId, score, extraData}

		case 1: // 最近记录 - 总是更新
			upsertSQL = fmt.Sprintf(`
				INSERT INTO %s (type, player_id, score, extra_data, created_at, updated_at)
				VALUES (?, ?, ?, ?, NOW(), NOW())
				ON DUPLICATE KEY UPDATE
					score = VALUES(score),
					extra_data = VALUES(extra_data),
					updated_at = NOW()
			`, tableName)
			args = []interface{}{leaderboardName, userId, score, extraData}

		case 2: // 历史总和 - 累加分数
			upsertSQL = fmt.Sprintf(`
				INSERT INTO %s (type, player_id, score, extra_data, created_at, updated_at)
				VALUES (?, ?, ?, ?, NOW(), NOW())
				ON DUPLICATE KEY UPDATE
					score = score + VALUES(score),
					extra_data = VALUES(extra_data),
					updated_at = NOW()
			`, tableName)
			args = []interface{}{leaderboardName, userId, score, extraData}
		}

		_, err = o.Raw(upsertSQL, args...).Exec()
		if err != nil {
			return fmt.Errorf("更新分数记录失败: %v", err)
		}
	}

	// 7. 同步到Redis（如果Redis可用）
	if RedisClient != nil {
		err = syncScoreToRedis(appId, userId, leaderboardName, score, extraData)
		if err != nil {
			// Redis错误不影响主流程，记录日志即可
			logs.Warn("Redis同步失败: %v", err)
		}
	}

	return nil
}

// syncScoreToRedis 同步分数到Redis
func syncScoreToRedis(appId, userId, leaderboardName string, score int64, extraData string) error {
	ctx := RedisClient.Context()

	// 排行榜有序集合的key（用于存储分数排名）
	scoreKey := getLeaderboardRedisKey(appId, leaderboardName)
	// 用户详情哈希表的key（用于存储额外数据）
	detailKey := getLeaderboardRedisKey(appId, leaderboardName) + ":details"

	// 1. 更新有序集合中的分数（使用用户ID作为member，分数作为score）
	member := &redis.Z{
		Score:  float64(score),
		Member: userId,
	}

	err := RedisClient.ZAdd(ctx, scoreKey, member).Err()
	if err != nil {
		return err
	}

	// 2. 更新用户详情（存储额外数据和更新时间）
	userDetails := map[string]interface{}{
		"extra_data":  extraData,
		"update_time": time.Now().Unix(),
		"score":       score, // 冗余存储，方便查询时减少Redis操作
	}

	// 序列化用户详情为JSON
	detailsJSON, err := json.Marshal(userDetails)
	if err != nil {
		return err
	}

	err = RedisClient.HSet(ctx, detailKey, userId, string(detailsJSON)).Err()
	if err != nil {
		return err
	}

	// 3. 设置过期时间（24小时）
	RedisClient.Expire(ctx, scoreKey, 24*time.Hour)
	RedisClient.Expire(ctx, detailKey, 24*time.Hour)

	return nil
}

func UpdateScore(appId, userId, leaderboardName string, score int64, extraData string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardName)
	if err != nil {
		return err
	}

	// 更新现有记录
	sql := fmt.Sprintf(`
		UPDATE %s 
		SET score = ?, extra_data = ?, updated_at = NOW()
		WHERE type = ? AND player_id = ?
	`, tableName)

	result, err := o.Raw(sql, score, extraData, configId, userId).Exec()
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

// GetLeaderboard 获取排行榜（优先从Redis读取，包含重置检查）
func GetLeaderboard(appId, leaderboardName string, limit int) ([]Leaderboard, error) {
	// 1. 获取排行榜配置并检查重置
	config, err := getLeaderboardConfig(appId, leaderboardName)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜配置失败: %v", err)
	}

	// 2. 检查是否需要重置排行榜
	err = checkAndResetLeaderboard(appId, leaderboardName, config)
	if err != nil {
		return nil, err
	}

	// 3. 尝试从Redis获取
	if RedisClient != nil {
		leaderboards, err := getLeaderboardFromRedis(appId, leaderboardName, limit)
		if err == nil && len(leaderboards) > 0 {
			return leaderboards, nil
		}
		// Redis失败或没有数据，继续从数据库读取
	}

	// 4. 从数据库获取
	return getLeaderboardFromDBWithConfig(appId, leaderboardName, limit, config)
}

// getLeaderboardFromRedis 从Redis获取排行榜
func getLeaderboardFromRedis(appId, leaderboardName string, limit int) ([]Leaderboard, error) {
	ctx := RedisClient.Context()

	// 排行榜有序集合的key
	scoreKey := getLeaderboardRedisKey(appId, leaderboardName)
	// 用户详情哈希表的key
	detailKey := getLeaderboardRedisKey(appId, leaderboardName) + ":details"

	// 获取Redis有序集合的前N个成员（按分数降序）
	results, err := RedisClient.ZRevRangeWithScores(ctx, scoreKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var leaderboards []Leaderboard
	for _, result := range results {
		userId := result.Member.(string)
		score := int64(result.Score)

		// 获取该用户的详情
		userDetail, err := RedisClient.HGet(ctx, detailKey, userId).Result()

		var extraData string
		var updateTime int64

		if err == nil && userDetail != "" {
			// 解析用户详情JSON
			var detailMap map[string]interface{}
			if err := json.Unmarshal([]byte(userDetail), &detailMap); err == nil {
				if ed, ok := detailMap["extra_data"].(string); ok {
					extraData = ed
				}
				if ut, ok := detailMap["update_time"].(float64); ok {
					updateTime = int64(ut)
				}
			}
		}

		if updateTime == 0 {
			updateTime = time.Now().Unix()
		}

		lb := Leaderboard{
			Type:      leaderboardName,
			UserId:    userId,
			Score:     score,
			ExtraData: extraData,
			UpdatedAt: time.Unix(updateTime, 0).Format("2006-01-02 15:04:05"),
		}

		leaderboards = append(leaderboards, lb)
	}

	return leaderboards, nil
}

// getLeaderboardFromDBWithConfig 从数据库获取排行榜（带配置）
func getLeaderboardFromDBWithConfig(appId, leaderboardName string, limit int, config *LeaderboardConfig) ([]Leaderboard, error) {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 确定排序方式
	sortOrder := "DESC"
	if config.Sort == 0 {
		sortOrder = "ASC"
	}

	// 查询排行榜数据
	sql := fmt.Sprintf(`
		SELECT 
			id,
			type,
			player_id as user_id,
			score,
			extra_data,
			created_at as create_time,
			updated_at as update_time
		FROM %s
		WHERE type = ?
		ORDER BY score %s, created_at ASC
		LIMIT ?
	`, tableName, sortOrder)

	var results []orm.Params
	_, err := o.Raw(sql, leaderboardName, limit).Values(&results)
	if err != nil {
		return nil, err
	}

	// 获取用户信息映射
	userInfoMap, err := getUserInfoMapForLeaderboard(appId, results)
	if err != nil {
		logs.Warn("获取用户信息失败: %v", err)
		userInfoMap = make(map[string]map[string]interface{})
	}

	// 转换为Leaderboard结构
	var leaderboards []Leaderboard
	for _, result := range results {
		lb := Leaderboard{}

		// 安全的ID类型转换
		if idVal, ok := result["id"]; ok {
			switch v := idVal.(type) {
			case int64:
				lb.Id = v
			case int:
				lb.Id = int64(v)
			case int32:
				lb.Id = int64(v)
			case string:
				if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
					lb.Id = parsed
				}
			}
		}
		if configIdVal, ok := result["type"].(string); ok {
			lb.Type = configIdVal
		}
		if userId, ok := result["user_id"].(string); ok {
			lb.UserId = userId
			// 添加用户信息
			if userInfo, exists := userInfoMap[userId]; exists {
				lb.UserInfo = userInfo
			} else {
				lb.UserInfo = map[string]interface{}{}
			}
		}
		// 安全的分数类型转换
		if scoreVal, ok := result["score"]; ok {
			switch v := scoreVal.(type) {
			case int64:
				lb.Score = v
			case int:
				lb.Score = int64(v)
			case int32:
				lb.Score = int64(v)
			case string:
				if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
					lb.Score = parsed
				}
			}
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

// getUserInfoMapForLeaderboard 获取排行榜用户信息映射
func getUserInfoMapForLeaderboard(appId string, leaderboardResults []orm.Params) (map[string]map[string]interface{}, error) {
	userInfoMap := make(map[string]map[string]interface{})

	// 收集所有用户ID
	userIds := make([]string, 0)
	for _, result := range leaderboardResults {
		if userId, ok := result["user_id"].(string); ok && userId != "" {
			userIds = append(userIds, userId)
		}
	}

	if len(userIds) == 0 {
		return userInfoMap, nil
	}

	// 去重
	uniqueUserIds := make([]string, 0)
	seenIds := make(map[string]bool)
	for _, id := range userIds {
		if !seenIds[id] {
			uniqueUserIds = append(uniqueUserIds, id)
			seenIds[id] = true
		}
	}

	// 批量获取用户信息
	o := orm.NewOrm()
	userTableName := utils.GetUserTableName(appId)

	// 构建IN查询
	placeholders := strings.Repeat("?,", len(uniqueUserIds))
	placeholders = placeholders[:len(placeholders)-1] // 去掉最后的逗号

	sql := fmt.Sprintf(`
		SELECT player_id, nickname, avatar 
		FROM %s 
		WHERE player_id IN (%s)
	`, userTableName, placeholders)

	args := make([]interface{}, len(uniqueUserIds))
	for i, id := range uniqueUserIds {
		args[i] = id
	}

	var userResults []orm.Params
	_, err := o.Raw(sql, args...).Values(&userResults)
	if err != nil {
		return nil, err
	}

	// 构建用户信息映射
	for _, userResult := range userResults {
		if playerId, ok := userResult["player_id"].(string); ok {
			userInfo := make(map[string]interface{})
			if nickname, ok := userResult["nickname"].(string); ok {
				userInfo["nickName"] = nickname
			} else {
				userInfo["nickName"] = ""
			}
			if avatar, ok := userResult["avatar"].(string); ok {
				userInfo["avatarUrl"] = avatar
			} else {
				userInfo["avatarUrl"] = ""
			}
			userInfoMap[playerId] = userInfo
		}
	}

	return userInfoMap, nil
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

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardName)
	if err != nil {
		return 0, 0, err
	}

	// 先获取用户分数
	userSQL := fmt.Sprintf(`
		SELECT score FROM %s 
		WHERE type = ? AND player_id = ?
	`, tableName)

	var userResult []orm.Params
	_, err = o.Raw(userSQL, configId, userId).Values(&userResult)
	if err != nil {
		return 0, 0, err
	}

	if len(userResult) == 0 {
		return 0, 0, nil // 用户不在排行榜中
	}

	// 安全的类型转换
	var userScore int64
	if scoreVal, ok := userResult[0]["score"]; ok {
		switch v := scoreVal.(type) {
		case int64:
			userScore = v
		case int:
			userScore = int64(v)
		case int32:
			userScore = int64(v)
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				userScore = parsed
			} else {
				return 0, 0, fmt.Errorf("无效的分数格式: %s", v)
			}
		case []byte:
			if parsed, err := strconv.ParseInt(string(v), 10, 64); err == nil {
				userScore = parsed
			} else {
				return 0, 0, fmt.Errorf("无效的分数格式: %s", string(v))
			}
		default:
			return 0, 0, fmt.Errorf("不支持的分数类型: %T", v)
		}
	}

	// 计算排名：统计分数比该用户高的人数 + 1
	rankSQL := fmt.Sprintf(`
		SELECT COUNT(*) + 1 as rank FROM %s 
		WHERE type = ? AND score > ?
	`, tableName)

	var result []orm.Params
	_, err = o.Raw(rankSQL, configId, userScore).Values(&result)
	if err != nil {
		return 0, 0, err
	}

	if len(result) == 0 {
		return 0, 0, nil
	}

	data := result[0]
	var rank int64

	// 安全的类型转换
	if rankVal, ok := data["rank"]; ok {
		switch v := rankVal.(type) {
		case int64:
			rank = v
		case int:
			rank = int64(v)
		case int32:
			rank = int64(v)
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				rank = parsed
			} else {
				return 0, 0, fmt.Errorf("无效的排名格式: %s", v)
			}
		case []byte:
			if parsed, err := strconv.ParseInt(string(v), 10, 64); err == nil {
				rank = parsed
			} else {
				return 0, 0, fmt.Errorf("无效的排名格式: %s", string(v))
			}
		default:
			return 0, 0, fmt.Errorf("不支持的排名类型: %T", v)
		}
	}

	return int(rank), userScore, nil
}

// ResetLeaderboard 重置排行榜（同时清理Redis）
func ResetLeaderboard(appId, leaderboardName string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardName)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE type = ?
	`, tableName)

	_, err = o.Raw(sql, configId).Exec()
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
	whereClause := "WHERE "
	args := []interface{}{}

	if leaderboardName != "" {
		// 获取排行榜配置ID
		configId, err := getLeaderboardConfigId(appId, leaderboardName)
		if err != nil {
			return nil, 0, err
		}
		whereClause += " type = ?"
		args = append(args, configId)
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
			type,
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
		if configIdVal, ok := result["type"].(string); ok {
			lb.Type = configIdVal
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
// 注意：由于数据库表中没有rank字段，这个函数暂时不执行任何操作
// 排名是在查询时动态计算的
func updateLeaderboardRanks(o orm.Ormer, tableName, leaderboardName string) error {
	// 排名是动态计算的，不需要在数据库中存储
	// 这里保留函数签名以保持兼容性，但不执行任何操作
	return nil
}
