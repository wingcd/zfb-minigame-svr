package models

import (
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
	OpenId    string    `orm:"size(100);unique" json:"openId"`
	PlayerId  string    `orm:"size(100);unique" json:"playerId"`
	Data      string    `orm:"type(longtext)" json:"data"`
	Banned    bool      `orm:"default(false)" json:"banned"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"createdAt"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"updatedAt"`
	// 解析后的数据
	PlayerInfo map[string]interface{} `orm:"-" json:"playerInfo"`
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
}

// GetTableName 动态获取用户表名
func (u *GameUser) GetTableName(appId string) string {
	return fmt.Sprintf("user_%s", appId)
}

// GetAllGameUsers 获取游戏用户列表（分页、搜索）
func GetAllGameUsers(appId string, page, pageSize int, status string, playerId, openId string) ([]GameUser, int64, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	// 构建查询条件
	var whereConditions []string
	var params []interface{}

	if playerId != "" {
		whereConditions = append(whereConditions, "player_id LIKE ?")
		params = append(params, "%"+playerId+"%")
	}

	if openId != "" {
		whereConditions = append(whereConditions, "open_id LIKE ?")
		params = append(params, "%"+openId+"%")
	}

	// 根据状态筛选
	if status == "banned" {
		whereConditions = append(whereConditions, "banned = true")
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
	tableName := fmt.Sprintf("user_%s", appId)

	// 计算封禁结束时间
	var banExpire *time.Time
	if banType == "temporary" && banHours > 0 {
		expireTime := time.Now().Add(time.Duration(banHours) * time.Hour)
		banExpire = &expireTime
	}

	// 直接更新用户表的封禁状态
	var sql string
	var err error
	if banExpire != nil {
		sql = fmt.Sprintf("UPDATE %s SET banned = true, ban_reason = ?, ban_expire = ?, updated_at = NOW() WHERE player_id = ?", tableName)
		_, err = o.Raw(sql, banReason, banExpire, playerId).Exec()
	} else {
		sql = fmt.Sprintf("UPDATE %s SET banned = true, ban_reason = ?, ban_expire = NULL, updated_at = NOW() WHERE player_id = ?", tableName)
		_, err = o.Raw(sql, banReason, playerId).Exec()
	}
	if err != nil {
		logs.Error("更新用户封禁状态失败:", err)
		return err
	}

	return nil
}

// UnbanGameUser 解封游戏用户
func UnbanGameUser(appId, playerId string, adminId int64, unbanReason string) error {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	// 直接更新用户表的封禁状态
	sql := fmt.Sprintf("UPDATE %s SET banned = false, ban_reason = NULL, ban_expire = NULL, updated_at = NOW() WHERE player_id = ?", tableName)
	_, err := o.Raw(sql, playerId).Exec()
	if err != nil {
		logs.Error("更新用户解封状态失败:", err)
		return err
	}

	return nil
}

// CheckUserBanStatus 检查用户封禁状态（并自动解封过期的临时封禁）
func CheckUserBanStatus(appId, playerId string) (bool, error) {
	o := orm.NewOrm()
	tableName := fmt.Sprintf("user_%s", appId)

	var banned bool
	var banExpire *time.Time

	sql := fmt.Sprintf("SELECT banned, ban_expire FROM %s WHERE player_id = ?", tableName)
	err := o.Raw(sql, playerId).QueryRow(&banned, &banExpire)
	if err != nil {
		if err == orm.ErrNoRows {
			return false, nil
		}
		logs.Error("查询用户封禁状态失败:", err)
		return false, err
	}

	// 如果用户被封禁，检查是否过期
	if banned && banExpire != nil && banExpire.Before(time.Now()) {
		// 自动解封过期的临时封禁
		sql = fmt.Sprintf("UPDATE %s SET banned = false, ban_reason = NULL, ban_expire = NULL, updated_at = NOW() WHERE player_id = ?", tableName)
		_, err = o.Raw(sql, playerId).Exec()
		if err != nil {
			logs.Error("自动解封过期用户失败:", err)
		}
		return false, nil
	}

	return banned, nil
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

	// 封禁历史功能已移除，用户表中的banned字段已足够

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

// GetUserList 获取用户列表（别名）
func GetUserList(page, pageSize int, status, appId, playerId, openId string) ([]GameUser, int64, error) {
	return GetAllGameUsers(appId, page, pageSize, status, playerId, openId)
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
