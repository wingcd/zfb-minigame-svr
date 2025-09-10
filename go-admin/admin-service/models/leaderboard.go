package models

import (
	"encoding/json"
	"fmt"
	"time"

	"admin-service/utils"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// LeaderboardConfig 排行榜配置结构（管理表）
type LeaderboardConfig struct {
	BaseModel
	AppId            string    `orm:"size(100)" json:"appId"`
	LeaderboardType  string    `orm:"size(100);column(leaderboard_type)" json:"leaderboardType"`
	Name             string    `orm:"size(200)" json:"name"`
	Description      string    `orm:"type(text)" json:"description"`
	ScoreType        string    `orm:"size(20);default(higher_better)" json:"scoreType"` // higher_better, lower_better
	MaxRank          int       `orm:"default(1000);column(max_rank)" json:"maxRank"`
	Enabled          bool      `orm:"default(true)" json:"enabled"`
	Category         string    `orm:"size(100)" json:"category"`
	ResetType        string    `orm:"size(50);default(permanent)" json:"resetType"` // permanent, daily, weekly, monthly, custom
	ResetValue       int       `orm:"default(0);column(reset_value)" json:"resetValue"`
	ResetTime        time.Time `orm:"null;type(datetime);column(reset_time)" json:"resetTime"`
	UpdateStrategy   int       `orm:"default(0);column(update_strategy)" json:"updateStrategy"` // 0=最高分, 1=最新分, 2=累计分
	Sort             int       `orm:"default(1)" json:"sort"`                                   // 0=升序, 1=降序
	ScoreCount       int       `orm:"default(0);column(score_count)" json:"scoreCount"`
	ParticipantCount int       `orm:"default(0);column(participant_count)" json:"participantCount"`
	LastResetTime    time.Time `orm:"null;type(datetime);column(last_reset_time)" json:"lastResetTime"`
	CreatedBy        string    `orm:"size(100);column(created_by)" json:"createdBy"`
}

// Leaderboard 排行榜数据结构
type Leaderboard struct {
	Id        int64  `orm:"auto" json:"id"`
	Type      string `orm:"size(50)" json:"type"`
	UserId    string `orm:"size(100);column(player_id)" json:"userId"`
	Score     int64  `orm:"default(0)" json:"score"`
	ExtraData string `orm:"type(text);column(extra_data)" json:"extraData"`
	CreatedAt string `orm:"auto_now_add;type(datetime);column(created_at)" json:"createdAt"`
	UpdatedAt string `orm:"auto_now;type(datetime);column(updated_at)" json:"updatedAt"`
}

// GetLeaderboardCount 获取排行榜数量统计
func GetLeaderboardCount(appId string) (int64, error) {
	o := orm.NewOrm()
	count, err := o.QueryTable("leaderboard_config").Filter("appId", appId).Count()
	return count, err
}

// GetLeaderboardList 获取排行榜配置列表
func GetLeaderboardList(appId string, page, pageSize int, leaderboardName string) ([]*LeaderboardConfig, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("leaderboard_config").Filter("appId", appId)

	// 添加名称筛选
	if leaderboardName != "" {
		qs = qs.Filter("name__icontains", leaderboardName)
	}

	total, _ := qs.Count()

	var leaderboards []*LeaderboardConfig
	offset := (page - 1) * pageSize
	_, err := qs.OrderBy("-id").Limit(pageSize, offset).All(&leaderboards)

	return leaderboards, total, err
}

// GetTableName 获取动态表名
func (l *Leaderboard) GetTableName(appId string) string {
	cleanAppId := utils.CleanAppId(appId)
	return fmt.Sprintf("leaderboard_%s", cleanAppId)
}

// TableName 获取配置表名
func (l *LeaderboardConfig) TableName() string {
	return "leaderboard_config"
}

func init() {
	orm.RegisterModel(new(LeaderboardConfig))
	orm.RegisterModel(new(Leaderboard))
}

// TableName 获取表名
func (l *Leaderboard) TableName() string {
	return "leaderboard_config"
}

// CreateLeaderboardConfig 创建排行榜配置
func CreateLeaderboardConfig(config *LeaderboardConfig) error {
	o := orm.NewOrm()

	// 检查是否已存在
	exist := o.QueryTable("leaderboard_config").
		Filter("app_id", config.AppId).
		Filter("leaderboard_type", config.LeaderboardType).
		Exist()

	if exist {
		return fmt.Errorf("排行榜已存在")
	}

	_, err := o.Insert(config)
	if err != nil {
		return err
	}

	// 创建动态排行榜表（如果不存在）
	return createLeaderboardTable(config.AppId)
}

// UpdateLeaderboard 更新排行榜配置
func UpdateLeaderboard(appId, leaderboardType string, fields map[string]interface{}) error {
	o := orm.NewOrm()

	// 添加更新时间
	fields["updatedAt"] = time.Now()

	qs := o.QueryTable("leaderboard_config").
		Filter("appId", appId).
		Filter("leaderboard_type", leaderboardType)

	_, err := qs.Update(fields)
	return err
}

func getLeaderboardConfigId(appId, leaderboardType string) (uint64, error) {
	o := orm.NewOrm()
	var result []orm.Params
	sql := `SELECT id FROM leaderboard_config WHERE app_id = ? AND leaderboard_type = ?`
	_, err := o.Raw(sql, appId, leaderboardType).Values(&result)
	if err != nil {
		return 0, err
	}
	return result[0]["id"].(uint64), nil
}

// DeleteLeaderboard 删除排行榜
func DeleteLeaderboard(appId, leaderboardType string) error {
	o := orm.NewOrm()

	// 先获取排行榜配置ID（在删除配置之前）
	configId, err := getLeaderboardConfigId(appId, leaderboardType)
	if err != nil {
		return err
	}

	// 删除动态表中的排行榜数据
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE type = ?", tableName)
	_, err = o.Raw(deleteSQL, configId).Exec()
	if err != nil {
		return err
	}

	// 最后删除排行榜配置
	_, err = o.QueryTable("leaderboard_config").
		Filter("app_id", appId).
		Filter("leaderboard_type", leaderboardType).
		Delete()

	return err
}

// GetLeaderboardData 获取排行榜数据
func GetLeaderboardData(appId, leaderboardType string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardType)
	if err != nil {
		return nil, 0, err
	}

	// 使用动态表
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	// 获取总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE type = ?", tableName)
	var total int64
	err = o.Raw(countSQL, configId).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	querySQL := fmt.Sprintf(`
		SELECT id, type, player_id, score, extra_data, created_at as createdAt, updated_at as updatedAt 
		FROM %s 
		WHERE type = ? 
		ORDER BY score DESC, created_at ASC 
		LIMIT ? OFFSET ?
	`, tableName)

	var results []orm.Params
	_, err = o.Raw(querySQL, configId, pageSize, offset).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 转换为map格式并添加排名，同时获取玩家详细信息
	var result []map[string]interface{}
	userTableName := fmt.Sprintf("user_%s", appId)

	for i, row := range results {
		var playerId string
		if pid, ok := row["playerId"].(string); ok {
			playerId = pid
		} else {
			playerId = fmt.Sprintf("%v", row["playerId"])
		}

		// 获取玩家详细信息
		playerInfo := map[string]interface{}{
			"playerId": playerId,
			"token":    "",
			"nickname": "",
			"avatar":   "",
			"data":     map[string]interface{}{},
			"level":    0,
			"exp":      0,
			"coin":     0,
			"diamond":  0,
			"vipLevel": 0,
		}

		// 从用户表中获取玩家数据
		var userData []orm.Params
		userSQL := fmt.Sprintf("SELECT data FROM %s WHERE player_id = ?", userTableName)
		_, err = o.Raw(userSQL, playerId).Values(&userData)
		if err == nil && len(userData) > 0 {
			if dataStr, ok := userData[0]["data"].(string); ok && dataStr != "" {
				var parsedData map[string]interface{}
				if json.Unmarshal([]byte(dataStr), &parsedData) == nil {
					// 解析并设置玩家信息字段
					if token, exists := parsedData["token"]; exists {
						playerInfo["token"] = token
					}
					if nickname, exists := parsedData["nickname"]; exists {
						playerInfo["nickname"] = nickname
					}
					if avatar, exists := parsedData["avatar"]; exists {
						playerInfo["avatar"] = avatar
					}
					if level, exists := parsedData["level"]; exists {
						playerInfo["level"] = level
					}
					if exp, exists := parsedData["exp"]; exists {
						playerInfo["exp"] = exp
					}
					if coin, exists := parsedData["coin"]; exists {
						playerInfo["coin"] = coin
					}
					if diamond, exists := parsedData["diamond"]; exists {
						playerInfo["diamond"] = diamond
					}
					if vipLevel, exists := parsedData["vipLevel"]; exists {
						playerInfo["vipLevel"] = vipLevel
					}
					// 保存完整的游戏数据
					playerInfo["data"] = parsedData
				}
			}
		}

		item := map[string]interface{}{
			"rank":      offset + i + 1,
			"playerId":  playerId,
			"score":     row["score"],
			"extraData": row["extraData"],
			"createdAt": row["createdAt"],
			"updatedAt": row["updatedAt"],
			// 添加玩家详细信息
			"token":    playerInfo["token"],
			"nickname": playerInfo["nickname"],
			"avatar":   playerInfo["avatar"],
			"data":     playerInfo["data"],
			"level":    playerInfo["level"],
			"exp":      playerInfo["exp"],
			"coin":     playerInfo["coin"],
			"diamond":  playerInfo["diamond"],
			"vipLevel": playerInfo["vipLevel"],
		}
		result = append(result, item)
	}

	return result, total, nil
}

// UpdateLeaderboardScore 更新排行榜分数
func UpdateLeaderboardScore(appId, leaderboardType, playerId string, score int64) error {
	o := orm.NewOrm()

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardType)
	if err != nil {
		return err
	}

	// 使用动态表
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	// 检查记录是否存在
	var existingId int64
	checkSQL := fmt.Sprintf("SELECT id FROM %s WHERE type = ? AND player_id = ?", tableName)
	err = o.Raw(checkSQL, configId, playerId).QueryRow(&existingId)

	if err == orm.ErrNoRows {
		// 插入新记录
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (type, type, player_id, score, created_at, updated_at) 
			VALUES (?, ?, ?, ?, NOW(), NOW())
		`, tableName)
		_, err = o.Raw(insertSQL, configId, leaderboardType, playerId, score).Exec()
	} else if err == nil {
		// 更新现有记录
		updateSQL := fmt.Sprintf(`
			UPDATE %s SET score = ?, updated_at = NOW() 
			WHERE type = ? AND player_id = ?
		`, tableName)
		_, err = o.Raw(updateSQL, score, configId, playerId).Exec()
	}

	return err
}

// DeleteLeaderboardScore 删除排行榜分数
func DeleteLeaderboardScore(appId, leaderboardType, playerId string) error {
	o := orm.NewOrm()

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardType)
	if err != nil {
		return err
	}

	// 使用动态表
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE type = ? AND player_id = ?", tableName)
	_, err = o.Raw(deleteSQL, configId, playerId).Exec()

	return err
}

// CommitLeaderboardScore 提交排行榜分数
func CommitLeaderboardScore(appId, leaderboardType, playerId string, score int64) error {
	o := orm.NewOrm()

	// 检查排行榜是否存在且启用
	var leaderboard LeaderboardConfig
	err := o.QueryTable("leaderboard_config").
		Filter("appId", appId).
		Filter("leaderboard_type", leaderboardType).
		Filter("enabled", true).
		One(&leaderboard)

	if err != nil {
		if err == orm.ErrNoRows {
			return fmt.Errorf("排行榜不存在或已禁用")
		}
		return err
	}

	// 使用动态表
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	// 查找现有记录
	var existingScore int64
	checkSQL := fmt.Sprintf("SELECT score FROM %s WHERE type = ? AND player_id = ?", tableName)
	err = o.Raw(checkSQL, leaderboard.ID, playerId).QueryRow(&existingScore)

	if err == orm.ErrNoRows {
		// 检查是否超过最大条目数
		if leaderboard.MaxRank > 0 {
			countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE type = ?", tableName)
			var count int64
			o.Raw(countSQL, leaderboard.ID).QueryRow(&count)

			if count >= int64(leaderboard.MaxRank) {
				// 删除最低分记录
				lowestSQL := fmt.Sprintf("SELECT score FROM %s WHERE type = ? ORDER BY score ASC, created_at DESC LIMIT 1", tableName)
				var lowestScore int64
				err = o.Raw(lowestSQL, leaderboard.ID).QueryRow(&lowestScore)

				if err == nil && score > lowestScore {
					deleteLowestSQL := fmt.Sprintf("DELETE FROM %s WHERE type = ? ORDER BY score ASC, created_at DESC LIMIT 1", tableName)
					o.Raw(deleteLowestSQL, leaderboard.ID).Exec()
				} else if err == nil {
					return fmt.Errorf("分数太低，无法进入排行榜")
				}
			}
		}

		// 创建新记录
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (type, type, player_id, score, created_at, updated_at) 
			VALUES (?, ?, ?, ?, NOW(), NOW())
		`, tableName)
		_, err = o.Raw(insertSQL, leaderboard.ID, leaderboardType, playerId, score).Exec()
		return err
	} else if err != nil {
		return err
	}

	// 检查分数类型决定是否更新
	shouldUpdate := false
	if leaderboard.ScoreType == "higher_better" && score > existingScore {
		shouldUpdate = true
	} else if leaderboard.ScoreType == "lower_better" && score < existingScore {
		shouldUpdate = true
	}

	if shouldUpdate {
		updateSQL := fmt.Sprintf(`
			UPDATE %s SET score = ?, updated_at = NOW() 
			WHERE type = ? AND player_id = ?
		`, tableName)
		_, err = o.Raw(updateSQL, score, leaderboard.ID, playerId).Exec()
	}

	return err
}

// QueryLeaderboardScore 查询排行榜分数
func QueryLeaderboardScore(appId, leaderboardType, playerId string) (int64, int, error) {
	o := orm.NewOrm()

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardType)
	if err != nil {
		return 0, 0, err
	}

	// 使用动态表
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	// 获取用户分数
	var userScore int64
	scoreSQL := fmt.Sprintf("SELECT score FROM %s WHERE type = ? AND player_id = ?", tableName)
	err = o.Raw(scoreSQL, configId, playerId).QueryRow(&userScore)

	if err != nil {
		if err == orm.ErrNoRows {
			return 0, 0, nil // 用户没有分数记录
		}
		return 0, 0, err
	}

	// 计算排名
	rankSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE type = ? AND score > ?", tableName)
	var rank int64
	err = o.Raw(rankSQL, configId, userScore).QueryRow(&rank)

	if err != nil {
		return userScore, 0, err
	}

	return userScore, int(rank) + 1, nil
}

// FixLeaderboardUserInfo 修复排行榜用户信息（暂时保留兼容性）
func FixLeaderboardUserInfo(appId, leaderboardType string) (int64, error) {
	logs.Info("排行榜已迁移到动态表，用户信息修复功能已不需要")
	return 0, nil
}

// GetLeaderboardStatsByAppId 获取应用的排行榜统计信息
func GetLeaderboardStatsByAppId(appId string) (map[string]interface{}, error) {
	o := orm.NewOrm()
	stats := make(map[string]interface{})

	// 获取排行榜总数
	total, err := GetLeaderboardCount(appId)
	if err != nil {
		return stats, err
	}
	stats["total"] = total

	// 如果没有排行榜，返回默认值
	if total == 0 {
		stats["totalPlayers"] = int64(0)
		stats["highestScore"] = int64(0)
		stats["averageScore"] = float64(0)
		stats["todaySubmissions"] = int64(0)
		return stats, nil
	}

	// 获取动态表名
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	// 检查表是否存在
	checkSQL := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	var exists string
	err = o.Raw(checkSQL).QueryRow(&exists)
	if err == orm.ErrNoRows {
		// 表不存在，返回默认值
		stats["totalPlayers"] = int64(0)
		stats["highestScore"] = int64(0)
		stats["averageScore"] = float64(0)
		stats["todaySubmissions"] = int64(0)
		return stats, nil
	}

	// 获取总玩家数（去重）
	var totalPlayers int64
	playerCountSQL := fmt.Sprintf("SELECT COUNT(DISTINCT player_id) FROM %s", tableName)
	err = o.Raw(playerCountSQL).QueryRow(&totalPlayers)
	if err != nil {
		totalPlayers = 0
	}
	stats["totalPlayers"] = totalPlayers

	// 获取最高分数
	var highestScore int64
	highestScoreSQL := fmt.Sprintf("SELECT IFNULL(MAX(score), 0) FROM %s", tableName)
	err = o.Raw(highestScoreSQL).QueryRow(&highestScore)
	if err != nil {
		highestScore = 0
	}
	stats["highestScore"] = highestScore

	// 获取平均分数
	var averageScore float64
	averageScoreSQL := fmt.Sprintf("SELECT IFNULL(AVG(score), 0) FROM %s", tableName)
	err = o.Raw(averageScoreSQL).QueryRow(&averageScore)
	if err != nil {
		averageScore = 0
	}
	stats["averageScore"] = averageScore

	// 获取今日提交数
	var todaySubmissions int64
	todaySubmissionsSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(updated_at) = CURDATE()", tableName)
	err = o.Raw(todaySubmissionsSQL).QueryRow(&todaySubmissions)
	if err != nil {
		todaySubmissions = 0
	}
	stats["todaySubmissions"] = todaySubmissions

	return stats, nil
}

// createLeaderboardTable 创建排行榜数据表
func createLeaderboardTable(appId string) error {
	o := orm.NewOrm()

	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	// 检查表是否存在
	checkSQL := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	var exists string
	err := o.Raw(checkSQL).QueryRow(&exists)

	if err == orm.ErrNoRows {
		// 表不存在，创建表
		createSQL := fmt.Sprintf(`
			CREATE TABLE %s (
				id BIGINT AUTO_INCREMENT PRIMARY KEY,
				type BIGINT NOT NULL COMMENT '排行榜配置ID，关联leaderboard_config.id',
				type VARCHAR(50) NOT NULL COMMENT '排行榜类型',
				player_id VARCHAR(100) NOT NULL,
				score BIGINT DEFAULT 0,
				extra_data TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				UNIQUE KEY uk_leaderboard_user (type, player_id),
				KEY idx_leaderboard_score (type, score DESC),
				KEY idx_leaderboard_type (type),
				KEY idx_updated_at (updated_at),
				KEY idx_type (type)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
		`, tableName)

		_, err = o.Raw(createSQL).Exec()
		return err
	}

	return nil
}

// UpdateLeaderboardScoreWithExtra 更新排行榜分数（支持额外数据）
func UpdateLeaderboardScoreWithExtra(appId, leaderboardType, playerId string, score int64, extraDataJson string) error {
	o := orm.NewOrm()

	// 检查应用是否存在
	var app Application
	err := o.QueryTable("application").Filter("appId", appId).One(&app)
	if err != nil {
		return fmt.Errorf("应用不存在")
	}

	// 检查排行榜配置
	var config LeaderboardConfig
	err = o.QueryTable("leaderboard_config").Filter("appId", appId).Filter("leaderboard_type", leaderboardType).One(&config)
	if err != nil {
		return fmt.Errorf("排行榜配置不存在")
	}

	if !config.Enabled {
		return fmt.Errorf("排行榜已禁用")
	}

	// 获取动态表名
	leaderboardData := &Leaderboard{}
	tableName := leaderboardData.GetTableName(appId)

	// 获取排行榜配置ID
	configId, err := getLeaderboardConfigId(appId, leaderboardType)
	if err != nil {
		return err
	}

	// 查找现有记录 - 使用原生SQL查询
	var existingRecord struct {
		ID        int64  `json:"id"`
		ConfigId  int64  `json:"type"`
		Type      string `json:"type"`
		PlayerId  string `json:"player_id"`
		Score     int64  `json:"score"`
		ExtraData string `json:"extra_data"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	err = o.Raw(fmt.Sprintf("SELECT id, type, type, player_id, score, extra_data, created_at, updated_at FROM %s WHERE type = ? AND player_id = ?", tableName), configId, playerId).QueryRow(&existingRecord)

	if err == orm.ErrNoRows {
		// 新增记录 - 使用原生SQL插入
		_, err = o.Raw(fmt.Sprintf("INSERT INTO %s (type, type, player_id, score, extra_data) VALUES (?, ?, ?, ?, ?)", tableName), configId, leaderboardType, playerId, score, extraDataJson).Exec()
		if err != nil {
			return fmt.Errorf("插入排行榜数据失败: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("查询排行榜数据失败: %v", err)
	} else {
		// 更新记录
		shouldUpdate := false

		switch config.UpdateStrategy {
		case 0: // 最高分
			if config.ScoreType == "higher_better" && score > existingRecord.Score {
				shouldUpdate = true
			} else if config.ScoreType == "lower_better" && score < existingRecord.Score {
				shouldUpdate = true
			}
		case 1: // 最新分
			shouldUpdate = true
		case 2: // 累计分
			score = score + existingRecord.Score
			shouldUpdate = true
		}

		if shouldUpdate {
			// 使用原生SQL更新
			_, err = o.Raw(fmt.Sprintf("UPDATE %s SET score = ?, extra_data = ?, updated_at = NOW() WHERE id = ?", tableName), score, extraDataJson, existingRecord.ID).Exec()
			if err != nil {
				return fmt.Errorf("更新排行榜数据失败: %v", err)
			}
		}
	}

	return nil
}
