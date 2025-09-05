package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
)

// GameUser 游戏用户结构
type GameUser struct {
	ID         int64     `orm:"pk;auto" json:"id"`
	PlayerId   string    `orm:"size(100);unique" json:"playerId"`
	Data       string    `orm:"type(longtext)" json:"data"`
	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime"`
	// 解析后的数据
	PlayerInfo map[string]interface{} `orm:"-" json:"playerInfo"`
	BanStatus  *UserBanRecord         `orm:"-" json:"banStatus,omitempty"`
}

// UserBanRecord 用户封禁记录
type UserBanRecord struct {
	ID           string     `orm:"pk;size(64)" json:"id"`
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
	CreateTime   time.Time  `orm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime   time.Time  `orm:"auto_now;type(datetime)" json:"updateTime"`
}

// UserStats 用户统计信息
type UserStats struct {
	PlayerId         string                 `json:"playerId"`
	TotalLogins      int64                  `json:"totalLogins"`
	LastLoginTime    *time.Time             `json:"lastLoginTime"`
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
	var whereClause string
	var params []interface{}

	if keyword != "" {
		whereClause = "WHERE player_id LIKE ?"
		params = append(params, "%"+keyword+"%")
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
	querySql := fmt.Sprintf("SELECT * FROM %s %s ORDER BY create_time DESC LIMIT ? OFFSET ?", tableName, whereClause)
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

		// 获取封禁状态
		if status == "banned" || status == "" {
			banRecord, _ := GetActiveUserBan(appId, users[i].PlayerId)
			if banRecord != nil {
				users[i].BanStatus = banRecord
			}
		}
	}

	// 如果筛选封禁用户，需要过滤结果
	if status == "banned" {
		var bannedUsers []GameUser
		for _, user := range users {
			if user.BanStatus != nil {
				bannedUsers = append(bannedUsers, user)
			}
		}
		return bannedUsers, int64(len(bannedUsers)), nil
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

	sql := fmt.Sprintf("UPDATE %s SET data = ?, update_time = NOW() WHERE player_id = ?", tableName)
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

	_, err := o.Insert(record)
	if err != nil {
		logs.Error("创建封禁记录失败:", err)
		return err
	}

	return nil
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

	// 更新封禁记录
	now := time.Now()
	banRecord.IsActive = false
	banRecord.UnbanAdminId = &adminId
	banRecord.UnbanTime = &now
	banRecord.UnbanReason = unbanReason
	banRecord.UpdateTime = now

	_, err = o.Update(banRecord, "IsActive", "UnbanAdminId", "UnbanTime", "UnbanReason", "UpdateTime")
	if err != nil {
		logs.Error("更新封禁记录失败:", err)
		return err
	}

	return nil
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
		record.IsActive = false
		record.UpdateTime = time.Now()
		o.Update(&record, "IsActive", "UpdateTime")
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
	sql += " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	params = append(params, pageSize, (page-1)*pageSize)

	var users []*GameUser
	_, err = o.Raw(sql, params...).QueryRows(&users)

	return users, total, err
}

// GetLeaderboardList 获取排行榜列表
func GetLeaderboardList(appId string, page, pageSize int, leaderboardName string) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("leaderboard_%s", appId)

	sql := fmt.Sprintf("SELECT * FROM %s ORDER BY score DESC LIMIT ? OFFSET ?", tableName)
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

// GetAllMailList 获取邮件列表
func GetAllMailList(appId string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("mail_%s", appId)

	sql := fmt.Sprintf("SELECT * FROM %s ORDER BY create_time DESC LIMIT ? OFFSET ?", tableName)
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

// SendMail 发送邮件给特定用户
func SendMail(appId, playerId, title, content string, attachments string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("mail_%s", appId)

	sql := fmt.Sprintf("INSERT INTO %s (player_id, title, content, attachments, create_time, is_read, is_claimed) VALUES (?, ?, ?, ?, NOW(), 0, 0)", tableName)
	_, err := o.Raw(sql, playerId, title, content, attachments).Exec()

	return err
}

// SendBroadcastMail 发送广播邮件
func SendBroadcastMail(appId, title, content string, attachments string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("mail_%s", appId)

	// 获取所有用户
	userTableName := fmt.Sprintf("user_%s", appId)
	sql := fmt.Sprintf("SELECT DISTINCT player_id FROM %s", userTableName)

	var playerIds []string
	_, err := o.Raw(sql).QueryRows(&playerIds)
	if err != nil {
		return err
	}

	// 给每个用户发送邮件
	for _, playerId := range playerIds {
		mailSql := fmt.Sprintf("INSERT INTO %s (player_id, title, content, attachments, create_time, is_read, is_claimed) VALUES (?, ?, ?, ?, NOW(), 0, 0)", tableName)
		_, err = o.Raw(mailSql, playerId, title, content, attachments).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetConfigList 获取配置列表 (别名)
func GetConfigList(appId string, page, pageSize int, keyword string) ([]*GameConfig, int64, error) {
	return GetAllGameConfigs(page, pageSize, appId, keyword)
}

// SetConfig 设置配置 (别名)
func SetConfig(appId, key, value string) error {
	config := &GameConfig{
		AppId:       appId,
		ConfigKey:   key,
		ConfigValue: value,
		ConfigType:  "string",
		Status:      1,
		IsPublic:    1,
	}
	return AddGameConfig(config)
}

// DeleteConfig 删除配置
func DeleteConfig(appId, key string) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("game_configs").Filter("app_id", appId).Filter("config_key", key).Delete()
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
	sql = fmt.Sprintf("DELETE FROM %s WHERE player_id = ?", mailTable)
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
		return nil, err
	}
	stats.RegistrationTime = user.CreateTime

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
		SUM(CASE WHEN is_read = 1 THEN 1 ELSE 0 END) as total_read,
		SUM(CASE WHEN is_received = 1 THEN 1 ELSE 0 END) as total_claimed,
		SUM(CASE WHEN is_read = 0 THEN 1 ELSE 0 END) as unread_count
		FROM %s WHERE player_id = ?
	`, mailTable)
	o.Raw(sql, playerId).QueryRow(&mailStats)
	stats.MailStats = mailStats

	// 获取封禁历史
	var banHistory []UserBanRecord
	o.QueryTable("user_ban_records").
		Filter("app_id", appId).
		Filter("player_id", playerId).
		OrderBy("-create_time").
		All(&banHistory)
	stats.BanHistory = banHistory

	return stats, nil
}

// GetUserRegistrationStats 获取用户注册统计
func GetUserRegistrationStats(appId string, days int) ([]RegistrationStats, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	sql := fmt.Sprintf(`
		SELECT DATE(create_time) as date, COUNT(*) as count
		FROM %s 
		WHERE create_time >= DATE_SUB(NOW(), INTERVAL ? DAY)
		GROUP BY DATE(create_time)
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
