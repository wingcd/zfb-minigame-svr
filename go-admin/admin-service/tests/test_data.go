package tests

import (
	"admin-service/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// TestData 测试数据管理器
type TestData struct {
	CreatedApps  []string
	CreatedUsers []string
	CreatedBans  []string
}

// NewTestData 创建测试数据管理器
func NewTestData() *TestData {
	return &TestData{
		CreatedApps:  make([]string, 0),
		CreatedUsers: make([]string, 0),
		CreatedBans:  make([]string, 0),
	}
}

// Cleanup 清理所有测试数据
func (td *TestData) Cleanup() error {
	// 清理封禁记录
	for _, banId := range td.CreatedBans {
		CleanupBanRecord("", banId)
	}

	// 清理用户
	for _, userId := range td.CreatedUsers {
		// 这里需要解析userId来获取appId和playerId
		// 格式: appId:playerId
		parts := splitUserID(userId)
		if len(parts) == 2 {
			CleanupTestUser(parts[0], parts[1])
		}
	}

	// 清理应用
	for _, appId := range td.CreatedApps {
		CleanupTestApp(appId)
	}

	return nil
}

// CreateTestApp 创建测试应用
func CreateTestApp(appId string) error {
	o := orm.NewOrm()

	// 检查应用是否已存在
	var count int64
	err := o.Raw("SELECT COUNT(*) FROM apps WHERE appId = ?", appId).QueryRow(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 应用已存在
	}

	// 创建应用（简化版）
	appName := fmt.Sprintf("测试应用_%s", appId)
	description := fmt.Sprintf("这是用于测试的应用: %s", appId)

	// 插入应用数据
	sql := "INSERT INTO apps (appId, app_name, description, status, createdAt, updatedAt) VALUES (?, ?, ?, ?, NOW(), NOW())"
	_, err = o.Raw(sql, appId, appName, description, 1).Exec()
	if err != nil {
		logs.Error("创建测试应用失败:", err)
		return err
	}

	// 创建对应的用户表
	userTableName := fmt.Sprintf("user_%s", appId)
	createUserTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			playerId VARCHAR(100) UNIQUE NOT NULL,
			data LONGTEXT,
			createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
			updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_playerId (playerId),
			INDEX idx_createdAt (createdAt)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, userTableName)

	_, err = o.Raw(createUserTableSQL).Exec()
	if err != nil {
		logs.Error("创建用户表失败:", err)
		return err
	}

	// 创建排行榜表
	leaderboardTableName := fmt.Sprintf("leaderboard_%s", appId)
	createLeaderboardTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			leaderboard_id VARCHAR(100) NOT NULL,
			playerId VARCHAR(100) NOT NULL,
			score BIGINT DEFAULT 0,
			extraData TEXT,
			createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
			updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_leaderboard_score (leaderboard_id, score DESC),
			INDEX idx_player (playerId),
			UNIQUE KEY uk_leaderboard_player (leaderboard_id, playerId)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, leaderboardTableName)

	_, err = o.Raw(createLeaderboardTableSQL).Exec()
	if err != nil {
		logs.Error("创建排行榜表失败:", err)
		return err
	}

	// 创建邮件表
	mailTableName := fmt.Sprintf("mail_%s", appId)
	createMailTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			playerId VARCHAR(100) NOT NULL,
			title VARCHAR(200) NOT NULL,
			content TEXT,
			attachments TEXT,
			is_read TINYINT DEFAULT 0,
			is_claimed TINYINT DEFAULT 0,
			expire_time DATETIME,
			createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
			updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_player (playerId),
			INDEX idx_createdAt (createdAt)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, mailTableName)

	_, err = o.Raw(createMailTableSQL).Exec()
	if err != nil {
		logs.Error("创建邮件表失败:", err)
		return err
	}

	// 创建计数器表
	counterTableName := fmt.Sprintf("counter_%s", appId)
	createCounterTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			counterKey VARCHAR(100) NOT NULL,
			counter_value BIGINT DEFAULT 0,
			description TEXT,
			createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
			updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_counterKey (counterKey)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`, counterTableName)

	_, err = o.Raw(createCounterTableSQL).Exec()
	if err != nil {
		logs.Error("创建计数器表失败:", err)
		return err
	}

	return nil
}

// CreateTestUser 创建测试用户
func CreateTestUser(appId, playerId string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	// 检查用户是否已存在
	var count int64
	err := o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE playerId = ?", tableName), playerId).QueryRow(&count)
	if err != nil {
		// 表可能不存在，先创建应用
		if err := CreateTestApp(appId); err != nil {
			return err
		}
	}

	if count > 0 {
		return nil // 用户已存在
	}

	// 创建测试用户数据
	userData := map[string]interface{}{
		"level":        1,
		"score":        100,
		"coins":        50,
		"nickname":     fmt.Sprintf("测试玩家_%s", playerId),
		"lastLogin":    time.Now().Format("2006-01-02 15:04:05"),
		"totalGames":   5,
		"winRate":      0.6,
		"achievements": []string{"新手上路", "初级玩家"},
		"inventory": map[string]interface{}{
			"items": []string{"sword", "shield"},
			"equipment": map[string]string{
				"weapon": "iron_sword",
				"armor":  "leather_armor",
			},
		},
	}

	userDataJSON, err := json.Marshal(userData)
	if err != nil {
		return err
	}

	// 插入用户数据
	sql := fmt.Sprintf("INSERT INTO %s (playerId, data, createdAt, updatedAt) VALUES (?, ?, NOW(), NOW())", tableName)
	_, err = o.Raw(sql, playerId, string(userDataJSON)).Exec()
	if err != nil {
		logs.Error("创建测试用户失败:", err)
		return err
	}

	// 创建一些测试的排行榜数据
	leaderboardTable := fmt.Sprintf("leaderboard_%s", appId)
	leaderboards := []struct {
		LeaderboardId string
		Score         int64
		ExtraData     string
	}{
		{"daily_score", 1000, `{"date": "2023-10-01"}`},
		{"weekly_score", 5000, `{"week": "2023-W40"}`},
		{"level_ranking", int64(userData["level"].(int)), `{"experience": 1500}`},
	}

	for _, lb := range leaderboards {
		sql := fmt.Sprintf("INSERT INTO %s (leaderboard_id, playerId, score, extraData, createdAt, updatedAt) VALUES (?, ?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE score = VALUES(score), updatedAt = NOW()", leaderboardTable)
		o.Raw(sql, lb.LeaderboardId, playerId, lb.Score, lb.ExtraData).Exec()
	}

	// 创建一些测试邮件
	mailTable := fmt.Sprintf("mail_%s", appId)
	mails := []struct {
		Title       string
		Content     string
		Attachments string
		IsRead      int
		IsClaimed   int
	}{
		{"欢迎邮件", "欢迎来到游戏世界！", `{"coins": 100, "items": ["welcome_gift"]}`, 1, 1},
		{"每日奖励", "这是您的每日登录奖励", `{"coins": 50, "experience": 100}`, 0, 0},
		{"系统通知", "游戏将在今晚进行维护", "", 0, 0},
	}

	for _, mail := range mails {
		sql := fmt.Sprintf("INSERT INTO %s (playerId, title, content, attachments, is_read, is_claimed, createdAt, updatedAt) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())", mailTable)
		o.Raw(sql, playerId, mail.Title, mail.Content, mail.Attachments, mail.IsRead, mail.IsClaimed).Exec()
	}

	return nil
}

// BanTestUser 封禁测试用户
func BanTestUser(appId, playerId string) error {
	return models.BanUser(appId, playerId, 1, "temporary", "测试封禁", 24)
}

// CleanupTestApp 清理测试应用
func CleanupTestApp(appId string) error {
	o := orm.NewOrm()

	// 删除相关表
	tables := []string{
		fmt.Sprintf("user_%s", appId),
		fmt.Sprintf("leaderboard_%s", appId),
		fmt.Sprintf("mail_%s", appId),
		fmt.Sprintf("counter_%s", appId),
	}

	for _, table := range tables {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		_, err := o.Raw(sql).Exec()
		if err != nil {
			logs.Error("删除表失败:", table, err)
		}
	}

	// 删除应用记录
	_, err := o.Raw("DELETE FROM apps WHERE appId = ?", appId).Exec()
	if err != nil {
		logs.Error("删除应用记录失败:", err)
		return err
	}

	// 删除相关配置
	_, err = o.Raw("DELETE FROM game_configs WHERE appId = ?", appId).Exec()
	if err != nil {
		logs.Error("删除应用配置失败:", err)
	}

	return nil
}

// CleanupTestUser 清理测试用户
func CleanupTestUser(appId, playerId string) error {
	return models.DeleteUser(appId, playerId)
}

// CleanupBanRecord 清理封禁记录
func CleanupBanRecord(appId, playerId string) error {
	o := orm.NewOrm()

	if appId != "" && playerId != "" {
		// 根据appId和playerId删除
		_, err := o.Raw("DELETE FROM user_ban_records WHERE appId = ? AND playerId = ?", appId, playerId).Exec()
		return err
	} else if playerId != "" {
		// 根据ID删除（playerId在这里作为记录ID）
		_, err := o.Raw("DELETE FROM user_ban_records WHERE id = ?", playerId).Exec()
		return err
	}

	return nil
}

// CreateTestGameConfigs 创建测试游戏配置
func CreateTestGameConfigs(appId string) error {
	configs := []struct {
		Key   string
		Value string
		Type  string
		Desc  string
	}{
		{"max_level", "100", "number", "最大等级"},
		{"daily_reward", `{"coins": 100, "experience": 50}`, "json", "每日奖励"},
		{"game_version", "1.0.0", "string", "游戏版本"},
		{"maintenance_mode", "false", "boolean", "维护模式"},
		{"server_name", "测试服务器", "string", "服务器名称"},
		{"exp_multiplier", "1.5", "number", "经验倍数"},
		{"shop_items", `[{"id": "sword", "price": 100}, {"id": "shield", "price": 80}]`, "json", "商店物品"},
	}

	for _, config := range configs {
		gameConfig := &models.GameConfig{
			AppId:       appId,
			ConfigKey:   config.Key,
			ConfigValue: config.Value,
			ConfigType:  config.Type,
			Description: config.Desc,
			Status:      1,
			IsPublic:    1,
		}

		models.AddGameConfig(gameConfig)
	}

	return nil
}

// CreateTestCounters 创建测试计数器
func CreateTestCounters(appId string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("counter_%s", appId)

	counters := []struct {
		Key         string
		Value       int64
		Description string
	}{
		{"total_players", 1000, "总玩家数"},
		{"online_players", 150, "在线玩家数"},
		{"daily_active_users", 500, "日活跃用户"},
		{"total_games_played", 10000, "总游戏场次"},
		{"server_restarts", 5, "服务器重启次数"},
		{"bug_reports", 23, "Bug报告数"},
		{"feature_requests", 45, "功能请求数"},
	}

	for _, counter := range counters {
		sql := fmt.Sprintf("INSERT INTO %s (counterKey, counter_value, description, createdAt, updatedAt) VALUES (?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE counter_value = VALUES(counter_value), updatedAt = NOW()", tableName)
		_, err := o.Raw(sql, counter.Key, counter.Value, counter.Description).Exec()
		if err != nil {
			logs.Error("创建测试计数器失败:", err)
			return err
		}
	}

	return nil
}

// CreateTestLeaderboards 创建测试排行榜数据
func CreateTestLeaderboards(appId string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("leaderboard_%s", appId)

	// 创建多个测试玩家的排行榜数据
	players := []struct {
		PlayerId string
		Score    int64
	}{
		{"test_top_player_1", 10000},
		{"test_top_player_2", 9500},
		{"test_top_player_3", 9000},
		{"test_top_player_4", 8500},
		{"test_top_player_5", 8000},
	}

	leaderboardIds := []string{"daily_score", "weekly_score", "monthly_score", "all_time_score"}

	for _, lb := range leaderboardIds {
		for i, player := range players {
			// 根据排行榜类型调整分数
			score := player.Score
			switch lb {
			case "weekly_score":
				score = score * 7
			case "monthly_score":
				score = score * 30
			case "all_time_score":
				score = score * 365
			}

			// 添加一些随机性
			score += int64(i * 100)

			extraData := fmt.Sprintf(`{"rank": %d, "leaderboard": "%s"}`, i+1, lb)

			sql := fmt.Sprintf("INSERT INTO %s (leaderboard_id, playerId, score, extraData, createdAt, updatedAt) VALUES (?, ?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE score = VALUES(score), updatedAt = NOW()", tableName)
			_, err := o.Raw(sql, lb, player.PlayerId, score, extraData).Exec()
			if err != nil {
				logs.Error("创建测试排行榜数据失败:", err)
				return err
			}
		}
	}

	return nil
}

// CreateTestMails 创建测试邮件
func CreateTestMails(appId, playerId string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("mail_%s", appId)

	mails := []struct {
		Title       string
		Content     string
		Attachments string
		IsRead      int
		IsClaimed   int
		ExpireTime  *time.Time
	}{
		{
			"新手礼包",
			"恭喜您成为我们的新玩家！这里是您的新手礼包，包含金币和道具。",
			`{"coins": 1000, "items": [{"id": "beginner_sword", "count": 1}, {"id": "health_potion", "count": 5}]}`,
			0, 0, nil,
		},
		{
			"每日签到奖励",
			"感谢您的每日登录！这是您今天的签到奖励。",
			`{"coins": 100, "experience": 50, "items": [{"id": "daily_chest", "count": 1}]}`,
			1, 1, nil,
		},
		{
			"活动通知",
			"双倍经验活动现在开始！在活动期间，所有获得的经验都将翻倍。活动时间：今天 14:00 - 18:00",
			"",
			1, 0, nil,
		},
		{
			"系统维护通知",
			"系统将在今晚 2:00 - 4:00 进行维护，维护期间无法登录游戏。维护完成后将发放补偿奖励。",
			`{"coins": 200, "items": [{"id": "maintenance_gift", "count": 1}]}`,
			0, 0, nil,
		},
		{
			"限时礼包",
			"限时特惠礼包现已上线！包含稀有装备和大量资源，限时24小时。",
			`{"coins": 500, "items": [{"id": "rare_equipment_box", "count": 1}, {"id": "resource_pack", "count": 3}]}`,
			0, 0, timePtr(time.Now().Add(24 * time.Hour)),
		},
	}

	for _, mail := range mails {
		var expireTimeStr interface{}
		if mail.ExpireTime != nil {
			expireTimeStr = mail.ExpireTime.Format("2006-01-02 15:04:05")
		}

		sql := fmt.Sprintf("INSERT INTO %s (playerId, title, content, attachments, is_read, is_claimed, expire_time, createdAt, updatedAt) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())", tableName)
		_, err := o.Raw(sql, playerId, mail.Title, mail.Content, mail.Attachments, mail.IsRead, mail.IsClaimed, expireTimeStr).Exec()
		if err != nil {
			logs.Error("创建测试邮件失败:", err)
			return err
		}
	}

	return nil
}

// CreateFullTestEnvironment 创建完整的测试环境
func CreateFullTestEnvironment(appId string) error {
	// 创建应用
	if err := CreateTestApp(appId); err != nil {
		return fmt.Errorf("创建测试应用失败: %v", err)
	}

	// 创建测试用户
	testUsers := []string{
		"test_player_001", "test_player_002", "test_player_003",
		"test_player_ban", "test_player_unban", "test_player_delete",
		"test_player_stats", "test_top_player_1", "test_top_player_2",
		"test_top_player_3", "test_top_player_4", "test_top_player_5",
	}

	for _, playerId := range testUsers {
		if err := CreateTestUser(appId, playerId); err != nil {
			logs.Error("创建测试用户失败:", playerId, err)
		}
	}

	// 创建测试配置
	if err := CreateTestGameConfigs(appId); err != nil {
		logs.Error("创建测试配置失败:", err)
	}

	// 创建测试计数器
	if err := CreateTestCounters(appId); err != nil {
		logs.Error("创建测试计数器失败:", err)
	}

	// 创建测试排行榜数据
	if err := CreateTestLeaderboards(appId); err != nil {
		logs.Error("创建测试排行榜数据失败:", err)
	}

	// 为主要测试用户创建邮件
	mainTestUsers := []string{"test_player_001", "test_player_002", "test_player_003"}
	for _, playerId := range mainTestUsers {
		if err := CreateTestMails(appId, playerId); err != nil {
			logs.Error("创建测试邮件失败:", playerId, err)
		}
	}

	return nil
}

// 辅助函数
func timePtr(t time.Time) *time.Time {
	return &t
}

func splitUserID(userID string) []string {
	// 简单的分割函数，实际实现可能需要更复杂的逻辑
	for i, c := range userID {
		if c == ':' {
			return []string{userID[:i], userID[i+1:]}
		}
	}
	return []string{userID}
}
