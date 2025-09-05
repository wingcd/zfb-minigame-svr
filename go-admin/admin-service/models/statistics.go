package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// DashboardStats 仪表板统计数据
type DashboardStats struct {
	TotalUsers        int64 `json:"totalUsers"`
	TotalApplications int64 `json:"totalApplications"`
	TotalAdmins       int64 `json:"totalAdmins"`
	TodayOperations   int64 `json:"todayOperations"`
}

// TopApp 热门应用
type TopApp struct {
	AppId       string `json:"appId"`
	AppName     string `json:"appName"`
	UserCount   int64  `json:"userCount"`
	AccessCount int64  `json:"accessCount"`
}

// UserActivity 用户活动数据
type UserActivity struct {
	Date       string `json:"date"`
	UserCount  int64  `json:"userCount"`
	LoginCount int64  `json:"loginCount"`
}

// DataTrend 数据趋势
type DataTrend struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
	Type  string `json:"type"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	Version      string `json:"version"`
	StartTime    string `json:"startTime"`
	DatabaseSize string `json:"databaseSize"`
	CacheStatus  string `json:"cacheStatus"`
}

// RecentActivity 最近活动
type RecentActivity struct {
	Id          int64     `json:"id"`
	UserId      int64     `json:"userId"`
	Username    string    `json:"username"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreateTime  time.Time `json:"createTime"`
}

// UserGrowth 用户增长数据
type UserGrowth struct {
	Date       string `json:"date"`
	NewUsers   int64  `json:"newUsers"`
	TotalUsers int64  `json:"totalUsers"`
}

// GetDashboardStats 获取仪表板统计数据
func GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 获取总用户数（这里返回模拟数据，实际应该统计所有应用的用户）
	stats.TotalUsers = 0

	// 获取总应用数
	totalApps, _ := GetTotalApplications()
	stats.TotalApplications = totalApps

	// 获取总管理员数
	totalAdmins, _ := GetTotalAdmins()
	stats.TotalAdmins = totalAdmins

	// 获取今日操作数
	today := time.Now().Format("2006-01-02")
	todayOps, _ := GetOperationsByDate(today)
	stats.TodayOperations = todayOps

	return stats, nil
}

// GetTopApps 获取热门应用
func GetTopApps(limit int) ([]TopApp, error) {
	o := orm.NewOrm()
	var apps []TopApp

	// 这里应该根据实际业务逻辑统计热门应用
	// 目前返回应用列表作为示例
	sql := `
		SELECT app_id, app_name, 0 as user_count, 0 as access_count 
		FROM applications 
		WHERE status = 1 
		ORDER BY created_at DESC 
		LIMIT ?
	`
	_, err := o.Raw(sql, limit).QueryRows(&apps)

	return apps, err
}

// GetRecentActivity 获取最近活动
func GetRecentActivity(limit int) ([]RecentActivity, error) {
	o := orm.NewOrm()
	var activities []RecentActivity

	sql := `
		SELECT id, user_id, username, action, resource as description, created_at as create_time
		FROM admin_operation_logs 
		ORDER BY created_at DESC 
		LIMIT ?
	`
	_, err := o.Raw(sql, limit).QueryRows(&activities)

	return activities, err
}

// GetUserActivity 获取用户活动统计
func GetUserActivity(appId string, days int) ([]UserActivity, error) {
	// 返回模拟数据，实际应该根据应用统计用户活动
	var activities []UserActivity

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		activities = append(activities, UserActivity{
			Date:       date,
			UserCount:  0,
			LoginCount: 0,
		})
	}

	return activities, nil
}

// GetDataTrends 获取数据趋势
func GetDataTrends(appId, trendType string, days int) ([]DataTrend, error) {
	var trends []DataTrend

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		trends = append(trends, DataTrend{
			Date:  date,
			Value: 0,
			Type:  trendType,
		})
	}

	return trends, nil
}

// GetUserGrowth 获取用户增长数据
func GetUserGrowth(appId string, days int) ([]UserGrowth, error) {
	var growth []UserGrowth

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		growth = append(growth, UserGrowth{
			Date:       date,
			NewUsers:   0,
			TotalUsers: 0,
		})
	}

	return growth, nil
}

// ExportData 导出数据
func ExportData(dataType string, filters map[string]interface{}) (string, error) {
	// 这里应该实现数据导出逻辑
	// 目前返回模拟的文件路径
	return "/tmp/export_" + dataType + ".csv", nil
}

// GetSystemInfo 获取系统信息
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		Version:      "1.0.0",
		StartTime:    time.Now().Format("2006-01-02 15:04:05"),
		DatabaseSize: "未知",
		CacheStatus:  "正常",
	}

	return info, nil
}

// GetAppStats 获取应用统计
func GetAppStats(appId string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取用户数量
	userCount, _ := GetUserDataCount(appId)
	stats["userCount"] = userCount

	// 获取排行榜数量
	leaderboardCount, _ := GetLeaderboardCount(appId)
	stats["leaderboardCount"] = leaderboardCount

	// 获取计数器数量
	counterCount, _ := GetCounterCount(appId)
	stats["counterCount"] = counterCount

	// 获取邮件数量
	mailCount, _ := GetMailCount(appId)
	stats["mailCount"] = mailCount

	// 获取配置数量
	configCount, _ := GetConfigCount(appId)
	stats["configCount"] = configCount

	return stats, nil
}

// GetLeaderboardStats 获取排行榜统计
func GetLeaderboardStats(appId string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 获取排行榜总数
	total, _ := GetLeaderboardCount(appId)
	stats["total"] = total

	// 获取活跃排行榜数（模拟数据）
	stats["active"] = total

	// 获取今日提交数（模拟数据）
	stats["todaySubmits"] = int64(0)

	return stats, nil
}
