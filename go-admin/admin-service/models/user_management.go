package models

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// GameUser 游戏用户结构
type GameUser struct {
	ID        int64     `orm:"pk;auto" json:"id"`
	PlayerId  string    `orm:"size(100);unique" json:"playerId"`
	Data      string    `orm:"type(longtext)" json:"data"`
	Banned    bool      `orm:"default(false)" json:"banned"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"createdAt"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"updatedAt"`
	// 解析后的数据
	PlayerInfo map[string]interface{} `orm:"-" json:"playerInfo"`
	BanStatus  *UserBanRecord         `orm:"-" json:"banStatus,omitempty"`
}

// UserBanRecord 用户封禁记录
type UserBanRecord struct {
	ID           string     `orm:"pk;size(64);column(id)" json:"id"`
	AppId        string     `orm:"size(100)" json:"appId"`
	PlayerId     string     `orm:"size(100)" json:"playerId"`
	AdminId      int64      `json:"adminId"`
	BanType      string     `orm:"size(20)" json:"banType"` // temporary, permanent
	BanReason    string     `orm:"type(text)" json:"banReason"`
	BanStartTime time.Time  `orm:"auto_now_add;type(datetime)" json:"banStartTime"`
	BanEndTime   *time.Time `orm:"null;type(datetime)" json:"banEndTime"`
	IsActive     bool       `orm:"default(true)" json:"isActive"`
	UnbanAdminId *int64     `orm:"null" json:"unbanAdminId"`
	UnbanTime    *time.Time `orm:"null;type(datetime)" json:"unbanTime"`
	UnbanReason  string     `orm:"type(text)" json:"unbanReason"`
	CreatedAt    time.Time  `orm:"auto_now_add;type(datetime)" json:"createdAt"`
	UpdatedAt    time.Time  `orm:"auto_now;type(datetime)" json:"updatedAt"`
}

func (u *UserBanRecord) TableName() string {
	return "user_ban_records"
}

// UserStats 用户统计信息
type UserStats struct {
	PlayerId         string                 `json:"playerId"`
	TotalLogins      int64                  `json:"totalLogins"`
	LastLoginTime    *time.Time             `json:"lastLoginAt"`
	RegistrationTime time.Time              `json:"registrationTime"`
	LeaderboardStats []LeaderboardUserStats `json:"leaderboardStats"`
	MailStats        MailUserStats          `json:"mailStats"`
	CounterStats     []CounterUserStats     `json:"counterStats"`
	BanHistory       []UserBanRecord        `json:"banHistory"`
}

type LeaderboardUserStats struct {
	LeaderboardId string `json:"leaderboardId"`
	BestScore     int64  `json:"bestScore"`
	CurrentRank   int    `json:"currentRank"`
	TotalSubmits  int64  `json:"totalSubmits"`
}

type MailUserStats struct {
	TotalReceived int64 `json:"totalReceived"`
	TotalRead     int64 `json:"totalRead"`
	TotalClaimed  int64 `json:"totalClaimed"`
	UnreadCount   int64 `json:"unreadCount"`
}

type CounterUserStats struct {
	CounterKey string `json:"counterKey"`
	Value      int64  `json:"value"`
}

type RegistrationStats struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// GetUserDataCount 获取用户数据统计
func GetUserDataCount(appId string) (int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)
	var count int64
	err := o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).QueryRow(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func init() {
	orm.RegisterModel(new(UserBanRecord))
}

// GetTableName 动态获取用户表名
func (u *GameUser) GetTableName(appId string) string {
	return fmt.Sprintf("user_%s", appId)
}

// GetAllGameUsers 获取游戏用户列表（分页、搜索）
func GetAllGameUsers(appId string, page, pageSize int, keyword, status string) ([]GameUser, int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	// 构建查询条件
	var whereConditions []string
	var params []interface{}

	if keyword != "" {
		whereConditions = append(whereConditions, "player_id LIKE ?")
		params = append(params, "%"+keyword+"%")
	}

	// 根据状态筛选
	if status == "banned" {
		whereConditions = append(whereConditions, "banned = true")
	} else if status == "normal" {
		whereConditions = append(whereConditions, "banned = false")
	}

	var whereClause string
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 获取总数
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", tableName, whereClause)
	var total int64
	err := o.Raw(countSql, params...).QueryRow(&total)
	if err != nil {
		logs.Error("获取用户总数失败:", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	querySql := fmt.Sprintf("SELECT * FROM %s %s ORDER BY created_at DESC LIMIT ? OFFSET ?", tableName, whereClause)
	params = append(params, pageSize, offset)

	var users []GameUser
	_, err = o.Raw(querySql, params...).QueryRows(&users)
	if err != nil {
		logs.Error("查询用户列表失败:", err)
		return nil, 0, err
	}

	// 解析用户数据并获取封禁状态
	for i := range users {
		// 解析JSON数据
		if users[i].Data != "" {
			var playerInfo map[string]interface{}
			if err := json.Unmarshal([]byte(users[i].Data), &playerInfo); err == nil {
				users[i].PlayerInfo = playerInfo
			}
		}

		// 获取详细的封禁状态（如果需要）
		if users[i].Banned {
			banRecord, _ := GetActiveUserBan(appId, users[i].PlayerId)
			if banRecord != nil {
				users[i].BanStatus = banRecord
			}
		}
	}

	return users, total, nil
}

// GetGameUserDetail 获取游戏用户详细信息
func GetGameUserDetail(appId, playerId string) (*GameUser, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	var user GameUser
	sql := fmt.Sprintf("SELECT * FROM %s WHERE player_id = ?", tableName)
	err := o.Raw(sql, playerId).QueryRow(&user)
	if err != nil {
		if err == orm.ErrNoRows {
			logs.Info(fmt.Sprintf("用户不存在: appId=%s, playerId=%s", appId, playerId))
			return nil, fmt.Errorf("用户不存在")
		}
		logs.Error("获取用户详情失败:", err)
		return nil, err
	}

	// 解析JSON数据
	if user.Data != "" {
		var playerInfo map[string]interface{}
		if err := json.Unmarshal([]byte(user.Data), &playerInfo); err == nil {
			user.PlayerInfo = playerInfo
		}
	}

	// 获取封禁状态
	banRecord, _ := GetActiveUserBan(appId, playerId)
	if banRecord != nil {
		user.BanStatus = banRecord
	}

	return &user, nil
}

// UpdateGameUserData 更新游戏用户数据
func UpdateGameUserData(appId, playerId, data string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	sql := fmt.Sprintf("UPDATE %s SET data = ?, updated_at = NOW() WHERE player_id = ?", tableName)
	_, err := o.Raw(sql, data, playerId).Exec()
	if err != nil {
		logs.Error("更新用户数据失败:", err)
		return err
	}

	return nil
}

// BanGameUser 封禁游戏用户
func BanGameUser(appId, playerId string, adminId int64, banType, banReason string, banHours int) error {
	o := orm.NewOrm()

	// 先检查是否已有活跃的封禁记录
	existing, _ := GetActiveUserBan(appId, playerId)
	if existing != nil {
		return fmt.Errorf("用户已被封禁")
	}

	// 开启事务
	tx, err := o.Begin()
	if err != nil {
		return err
	}

	// 创建封禁记录
	record := &UserBanRecord{
		ID:           generateID(),
		AppId:        appId,
		PlayerId:     playerId,
		AdminId:      adminId,
		BanType:      banType,
		BanReason:    banReason,
		BanStartTime: time.Now(),
		IsActive:     true,
	}

	// 设置封禁结束时间
	if banType == "temporary" && banHours > 0 {
		endTime := time.Now().Add(time.Duration(banHours) * time.Hour)
		record.BanEndTime = &endTime
	}

	_, err = tx.Insert(record)
	if err != nil {
		tx.Rollback()
		logs.Error("创建封禁记录失败:", err)
		return err
	}

	// 更新用户表的banned字段
	tableName := fmt.Sprintf("user_%s", appId)
	sql := fmt.Sprintf("UPDATE %s SET banned = true, updated_at = NOW() WHERE player_id = ?", tableName)
	_, err = tx.Raw(sql, playerId).Exec()
	if err != nil {
		tx.Rollback()
		logs.Error("更新用户封禁状态失败:", err)
		return err
	}

	return tx.Commit()
}

// UnbanGameUser 解封游戏用户
func UnbanGameUser(appId, playerId string, adminId int64, unbanReason string) error {
	o := orm.NewOrm()

	// 查找活跃的封禁记录
	banRecord, err := GetActiveUserBan(appId, playerId)
	if err != nil {
		return err
	}
	if banRecord == nil {
		return fmt.Errorf("用户未被封禁")
	}

	// 开启事务
	tx, err := o.Begin()
	if err != nil {
		return err
	}

	// 更新封禁记录
	now := time.Now()
	banRecord.IsActive = false
	banRecord.UnbanAdminId = &adminId
	banRecord.UnbanTime = &now
	banRecord.UnbanReason = unbanReason
	banRecord.UpdatedAt = now

	_, err = tx.Update(banRecord, "IsActive", "UnbanAdminId", "UnbanTime", "UnbanReason", "UpdatedAt")
	if err != nil {
		tx.Rollback()
		logs.Error("更新封禁记录失败:", err)
		return err
	}

	// 更新用户表的banned字段
	tableName := fmt.Sprintf("user_%s", appId)
	sql := fmt.Sprintf("UPDATE %s SET banned = false, updated_at = NOW() WHERE player_id = ?", tableName)
	_, err = tx.Raw(sql, playerId).Exec()
	if err != nil {
		tx.Rollback()
		logs.Error("更新用户解封状态失败:", err)
		return err
	}

	return tx.Commit()
}

// GetActiveUserBan 获取用户的活跃封禁记录
func GetActiveUserBan(appId, playerId string) (*UserBanRecord, error) {
	o := orm.NewOrm()

	var record UserBanRecord
	err := o.QueryTable("user_ban_records").
		Filter("app_id", appId).
		Filter("player_id", playerId).
		Filter("is_active", true).
		One(&record)

	if err == orm.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logs.Error("查询封禁记录失败:", err)
		return nil, err
	}

	// 检查临时封禁是否已过期
	if record.BanType == "temporary" && record.BanEndTime != nil && record.BanEndTime.Before(time.Now()) {
		// 自动解封
		tx, err := o.Begin()
		if err == nil {
			record.IsActive = false
			record.UpdatedAt = time.Now()
			tx.Update(&record, "IsActive", "UpdatedAt")

			// 更新用户表的banned字段
			tableName := fmt.Sprintf("user_%s", appId)
			sql := fmt.Sprintf("UPDATE %s SET banned = false, updated_at = NOW() WHERE player_id = ?", tableName)
			tx.Raw(sql, playerId).Exec()

			tx.Commit()
		}
		return nil, nil
	}

	return &record, nil
}

// GetUserDataList 获取用户数据列表
func GetUserDataList(appId string, page, pageSize int, keyword string) ([]*GameUser, int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	sql := fmt.Sprintf("SELECT * FROM %s", tableName)
	params := []interface{}{}

	if keyword != "" {
		sql += " WHERE player_id LIKE ?"
		params = append(params, "%"+keyword+"%")
	}

	// 计算总数
	countSql := strings.Replace(sql, "SELECT *", "SELECT COUNT(*)", 1)
	var total int64
	err := o.Raw(countSql, params...).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	sql += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	params = append(params, pageSize, (page-1)*pageSize)

	var users []*GameUser
	_, err = o.Raw(sql, params...).QueryRows(&users)

	return users, total, err
}

// GetLeaderboardList 函数已移至 leaderboard.go 模块

// GetCounterList 获取计数器列表
func GetCounterList(appId string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("counter_%s", appId)

	sql := fmt.Sprintf("SELECT * FROM %s ORDER BY id DESC LIMIT ? OFFSET ?", tableName)
	params := []interface{}{pageSize, (page - 1) * pageSize}

	var results []orm.Params
	_, err := o.Raw(sql, params...).Values(&results)
	if err != nil {
		return nil, 0, err
	}

	// 计算总数
	countSql := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var total int64
	err = o.Raw(countSql).QueryRow(&total)
	if err != nil {
		return nil, 0, err
	}

	// 转换结果
	var list []map[string]interface{}
	for _, result := range results {
		item := make(map[string]interface{})
		for k, v := range result {
			item[k] = v
		}
		list = append(list, item)
	}

	return list, total, err
}

// GetConfigList 获取配置列表 (别名)
func GetConfigList(appId string, page, pageSize int, keyword string) ([]*GameConfig, int64, error) {
	return GetAllGameConfigs(page, pageSize, appId, keyword)
}

// SetConfig 设置配置 (别名)
func SetConfig(appId, key, value string, configType, version string) error {
	config := &GameConfig{
		AppID:       appId,
		ConfigKey:   key,
		ConfigValue: value,
		ConfigType:  configType,
		Version:     version,
	}
	return AddGameConfig(config)
}

// DeleteConfig 删除配置
func DeleteConfig(appId, key string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("game_configs").Filter("appId", appId).Filter("configKey", key).Delete()
	return err
}

// DeleteGameUser 删除游戏用户（危险操作）
func DeleteGameUser(appId, playerId string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	// 开启事务
	tx, err := o.Begin()
	if err != nil {
		return err
	}

	// 删除用户数据
	sql := fmt.Sprintf("DELETE FROM %s WHERE player_id = ?", tableName)
	_, err = tx.Raw(sql, playerId).Exec()
	if err != nil {
		tx.Rollback()
		logs.Error("删除用户数据失败:", err)
		return err
	}

	// 删除相关的排行榜数据
	leaderboardTable := fmt.Sprintf("leaderboard_%s", appId)
	sql = fmt.Sprintf("DELETE FROM %s WHERE player_id = ?", leaderboardTable)
	tx.Raw(sql, playerId).Exec()

	// 删除相关的邮件数据
	mailTable := fmt.Sprintf("mail_%s", appId)
	sql = fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", mailTable)
	tx.Raw(sql, playerId).Exec()

	// 删除封禁记录
	tx.QueryTable("user_ban_records").Filter("app_id", appId).Filter("player_id", playerId).Delete()

	return tx.Commit()
}

// GetGameUserStats 获取游戏用户统计信息
func GetGameUserStats(appId, playerId string) (*UserStats, error) {
	o := orm.NewOrm()

	stats := &UserStats{
		PlayerId: playerId,
	}

	// 获取用户基本信息
	user, err := GetGameUserDetail(appId, playerId)
	if err != nil {
		if err.Error() == "用户不存在" {
			// 返回一个空的统计信息，表示用户不存在
			return &UserStats{
				PlayerId:         playerId,
				LeaderboardStats: []LeaderboardUserStats{},
				RegistrationTime: time.Time{},
			}, nil
		}
		return nil, err
	}
	stats.RegistrationTime = user.CreatedAt

	// 获取排行榜统计
	leaderboardTable := fmt.Sprintf("leaderboard_%s", appId)
	var leaderboardStats []LeaderboardUserStats
	sql := fmt.Sprintf(`
		SELECT leaderboard_id, MAX(score) as best_score, COUNT(*) as total_submits
		FROM %s WHERE player_id = ? GROUP BY leaderboard_id
	`, leaderboardTable)
	o.Raw(sql, playerId).QueryRows(&leaderboardStats)
	stats.LeaderboardStats = leaderboardStats

	// 获取邮件统计
	mailTable := fmt.Sprintf("mail_%s", appId)
	var mailStats MailUserStats
	sql = fmt.Sprintf(`
		SELECT 
		COUNT(*) as total_received,
		SUM(CASE WHEN status >= 1 THEN 1 ELSE 0 END) as total_read,
		SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as total_claimed,
		SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as unread_count
		FROM %s WHERE user_id = ?
	`, mailTable)
	o.Raw(sql, playerId).QueryRow(&mailStats)
	stats.MailStats = mailStats

	// 获取封禁历史
	var banHistory []UserBanRecord
	o.QueryTable("user_ban_records").
		Filter("app_id", appId).
		Filter("player_id", playerId).
		OrderBy("-created_at").
		All(&banHistory)
	stats.BanHistory = banHistory

	return stats, nil
}

// GetUserRegistrationStats 获取用户注册统计
func GetUserRegistrationStats(appId string, days int) ([]RegistrationStats, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	sql := fmt.Sprintf(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM %s 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`, tableName)

	var stats []RegistrationStats
	_, err := o.Raw(sql, days).QueryRows(&stats)
	if err != nil {
		logs.Error("获取注册统计失败:", err)
		return nil, err
	}

	return stats, nil
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// MigrateUserTableAddBannedField 为用户表添加banned字段的迁移函数
func MigrateUserTableAddBannedField(appId string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	// 检查字段是否已存在
	checkSql := fmt.Sprintf("SHOW COLUMNS FROM %s LIKE 'banned'", tableName)
	var result []orm.Params
	_, err := o.Raw(checkSql).Values(&result)
	if err != nil {
		logs.Error("检查字段存在性失败:", err)
		return err
	}

	// 如果字段已存在，直接返回
	if len(result) > 0 {
		logs.Info("字段 banned 已存在于表 %s", tableName)
		return nil
	}

	// 添加banned字段
	alterSql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN banned BOOLEAN NOT NULL DEFAULT FALSE", tableName)
	_, err = o.Raw(alterSql).Exec()
	if err != nil {
		logs.Error("添加banned字段失败:", err)
		return err
	}

	// 根据现有的封禁记录更新banned字段
	updateSql := fmt.Sprintf(`
		UPDATE %s u 
		SET banned = TRUE 
		WHERE EXISTS (
			SELECT 1 FROM user_ban_records ubr 
			WHERE ubr.app_id = ? AND ubr.player_id = u.player_id AND ubr.is_active = TRUE
		)
	`, tableName)
	_, err = o.Raw(updateSql, appId).Exec()
	if err != nil {
		logs.Error("更新现有用户封禁状态失败:", err)
		return err
	}

	logs.Info("成功为表 %s 添加 banned 字段并更新现有数据", tableName)
	return nil
}

// GetUserList 获取用户列表（别名）
func GetUserList(page, pageSize int, keyword, status, appId string) ([]GameUser, int64, error) {
	return GetAllGameUsers(appId, page, pageSize, keyword, status)
}

// BanUser 封禁用户（别名）
func BanUser(appId, playerId string, adminId int64, banType, banReason string, banHours int) error {
	return BanGameUser(appId, playerId, adminId, banType, banReason, banHours)
}

// UnbanUser 解封用户（别名）
func UnbanUser(appId, playerId string, adminId int64, unbanReason string) error {
	return UnbanGameUser(appId, playerId, adminId, unbanReason)
}

// DeleteUser 删除用户（别名）
func DeleteUser(appId, playerId string) error {
	return DeleteGameUser(appId, playerId)
}

// GetUserDetail 获取用户详情（别名）
func GetUserDetail(appId, playerId string) (*GameUser, error) {
	return GetGameUserDetail(appId, playerId)
}

// SetUserDetail 设置用户详情（别名）
func SetUserDetail(appId, playerId, data string) error {
	return UpdateGameUserData(appId, playerId, data)
}

// GetUserStats 获取用户统计（别名）
func GetUserStats(appId, playerId string) (*UserStats, error) {
	return GetGameUserStats(appId, playerId)
}

// AppUserStats 应用用户统计信息（对齐云函数格式）
type AppUserStats struct {
	Total         int64                  `json:"total"`
	NewToday      int64                  `json:"newToday"`
	ActiveToday   int64                  `json:"activeToday"`
	Banned        int64                  `json:"banned"`
	WeeklyActive  int64                  `json:"weeklyActive"`
	MonthlyActive int64                  `json:"monthlyActive"`
	Growth        AppUserStatsGrowth     `json:"growth"`
	Comparison    AppUserStatsComparison `json:"comparison"`
}

type AppUserStatsGrowth struct {
	NewTodayGrowth    float64 `json:"newTodayGrowth"`
	ActiveTodayGrowth float64 `json:"activeTodayGrowth"`
}

type AppUserStatsComparison struct {
	NewYesterday    int64 `json:"newYesterday"`
	ActiveYesterday int64 `json:"activeYesterday"`
}

// GetAppUserStats 获取应用用户统计（使用原生SQL实现）
func GetAppUserStats(appId string) (*AppUserStats, error) {
	o := orm.NewOrm()

	// 设置表名
	tableName := fmt.Sprintf("user_%s", appId)

	// 检查表是否存在
	exists, err := checkTableExists(tableName)
	if err != nil || !exists {
		return nil, fmt.Errorf("应用不存在或用户表不存在")
	}

	stats := &AppUserStats{}

	// 时间范围计算
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	yesterdayStart := todayStart.Add(-24 * time.Hour)
	yesterdayEnd := todayStart
	weekStart := todayStart.Add(-time.Duration(int(now.Weekday())) * 24 * time.Hour)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 总用户数 - 使用原生SQL
	var total int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).QueryRow(&total)
	if err != nil {
		return nil, fmt.Errorf("查询总用户数失败: %v", err)
	}
	stats.Total = total

	// 今日新增用户 - 使用原生SQL
	var newToday int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE created_at >= ? AND created_at < ?", tableName),
		todayStart, todayEnd).QueryRow(&newToday)
	if err != nil {
		return nil, fmt.Errorf("查询今日新增用户失败: %v", err)
	}
	stats.NewToday = newToday

	// 今日活跃用户（今天有更新的用户）- 使用原生SQL
	var activeToday int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE updated_at >= ? AND updated_at < ?", tableName),
		todayStart, todayEnd).QueryRow(&activeToday)
	if err != nil {
		// 如果表中没有 updated_at 字段，使用 created_at
		activeToday = newToday
	}
	stats.ActiveToday = activeToday

	// 封禁用户数 - 使用原生SQL
	var banned int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE banned = 1", tableName)).QueryRow(&banned)
	if err != nil {
		// 如果表中没有 banned 字段，则跳过此统计
		stats.Banned = 0
	} else {
		stats.Banned = banned
	}

	// 本周活跃用户 - 使用原生SQL
	var weeklyActive int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE created_at >= ?", tableName),
		weekStart).QueryRow(&weeklyActive)
	if err != nil {
		return nil, fmt.Errorf("查询本周活跃用户失败: %v", err)
	}
	stats.WeeklyActive = weeklyActive

	// 本月活跃用户 - 使用原生SQL
	var monthlyActive int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE created_at >= ?", tableName),
		monthStart).QueryRow(&monthlyActive)
	if err != nil {
		return nil, fmt.Errorf("查询本月活跃用户失败: %v", err)
	}
	stats.MonthlyActive = monthlyActive

	// 昨日新增用户（用于计算增长率）- 使用原生SQL
	var newYesterday int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE created_at >= ? AND created_at < ?", tableName),
		yesterdayStart, yesterdayEnd).QueryRow(&newYesterday)
	if err != nil {
		return nil, fmt.Errorf("查询昨日新增用户失败: %v", err)
	}
	stats.Comparison.NewYesterday = newYesterday

	// 昨日活跃用户（用于计算增长率）- 使用原生SQL
	var activeYesterday int64
	err = o.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE updated_at >= ? AND updated_at < ?", tableName),
		yesterdayStart, yesterdayEnd).QueryRow(&activeYesterday)
	if err != nil {
		// 如果表中没有 updated_at 字段，使用 created_at
		activeYesterday = newYesterday
	}
	stats.Comparison.ActiveYesterday = activeYesterday

	// 计算增长率
	if newYesterday > 0 {
		stats.Growth.NewTodayGrowth = float64(newToday-newYesterday) / float64(newYesterday) * 100
	} else {
		stats.Growth.NewTodayGrowth = 0
	}

	if activeYesterday > 0 {
		stats.Growth.ActiveTodayGrowth = float64(activeToday-activeYesterday) / float64(activeYesterday) * 100
	} else {
		stats.Growth.ActiveTodayGrowth = 0
	}

	// 保留两位小数
	stats.Growth.NewTodayGrowth = math.Round(stats.Growth.NewTodayGrowth*100) / 100
	stats.Growth.ActiveTodayGrowth = math.Round(stats.Growth.ActiveTodayGrowth*100) / 100

	return stats, nil
}

// checkTableExists 检查表是否存在（使用ORM实现）
func checkTableExists(tableName string) (bool, error) {
	o := orm.NewOrm()

	// 使用 ORM 的 Raw 方法检查表是否存在
	var count int64
	err := o.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ? AND table_schema = DATABASE()", tableName).QueryRow(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
