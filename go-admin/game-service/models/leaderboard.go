package models

import (
	"fmt"
	"time"

	"game-service/utils"

	"github.com/beego/beego/v2/client/orm"
)

// getLeaderboardTableName 获取排行榜表名
func getLeaderboardTableName(appId string) string {
	return utils.GetLeaderboardTableName(appId)
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

// SubmitScore 提交分数到排行榜
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

	// 如果成功，更新排名
	if err == nil {
		err = updateLeaderboardRanks(o, tableName, leaderboardName)
	}

	return err
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

// GetLeaderboard 获取排行榜
func GetLeaderboard(appId, leaderboardName string, limit int) ([]Leaderboard, error) {
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

// GetUserRank 获取用户在排行榜中的排名
func GetUserRank(appId, userId, leaderboardName string) (int, int64, error) {
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

// ResetLeaderboard 重置排行榜
func ResetLeaderboard(appId, leaderboardName string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	sql := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE leaderboard_name = ?
	`, tableName)

	_, err := o.Raw(sql, leaderboardName).Exec()
	return err
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

// CreateLeaderboardTable 创建排行榜表
func CreateLeaderboardTable(appId string) error {
	o := orm.NewOrm()
	tableName := getLeaderboardTableName(appId)

	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			leaderboard_name VARCHAR(100) NOT NULL COMMENT '排行榜名称',
			player_id VARCHAR(100) NOT NULL COMMENT '玩家ID',
			player_name VARCHAR(100) DEFAULT '' COMMENT '玩家昵称',
			score BIGINT DEFAULT 0 COMMENT '分数',
			extra_data TEXT COMMENT '额外数据JSON',
			rank INT DEFAULT 0 COMMENT '排名',
			season VARCHAR(50) DEFAULT 'default' COMMENT '赛季',
			category VARCHAR(50) DEFAULT 'general' COMMENT '分类',
			is_active BOOLEAN DEFAULT TRUE COMMENT '是否活跃',
			last_update_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '最后更新时间',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_leaderboard_player (leaderboard_name, player_id),
			INDEX idx_leaderboard_score (leaderboard_name, score DESC),
			INDEX idx_leaderboard_rank (leaderboard_name, rank),
			INDEX idx_player_id (player_id),
			INDEX idx_season_category (season, category),
			INDEX idx_last_update (last_update_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排行榜表'
	`, tableName)

	_, err := o.Raw(sql).Exec()
	if err != nil {
		return fmt.Errorf("创建排行榜表失败: %v", err)
	}

	return nil
}
