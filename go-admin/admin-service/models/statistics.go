package models

import (
	"fmt"
	"strings"
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
	CreatedAt   time.Time `json:"createdAt"`
}

// UserGrowth 用户增长数据
type UserGrowth struct {
	Date       string `json:"date"`
	NewUsers   int64  `json:"newUsers"`
	TotalUsers int64  `json:"totalUsers"`
}

// PlatformDistribution 平台分布数据
type PlatformDistribution struct {
	Platform string `json:"platform"`
	Name     string `json:"name"`
	Count    int64  `json:"count"`
	Value    int64  `json:"value"` // 用于图表显示
}

// GetDashboardStats 获取仪表板统计数据
func GetDashboardStats() (*DashboardStats, error) {
	o := orm.NewOrm()
	stats := &DashboardStats{}

	// 获取总用户数（统计所有应用的用户）
	var totalUsers int64 = 0

	// 获取所有激活的应用
	var apps []Application
	_, err := o.QueryTable("apps").Filter("status", "active").All(&apps)
	if err == nil {
		// 遍历所有应用，统计各应用的用户数据
		for _, app := range apps {
			// 清理应用ID，确保表名安全
			cleanAppId := strings.ReplaceAll(app.AppId, "-", "_")
			cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
			userTableName := fmt.Sprintf("user_%s", cleanAppId)

			// 检查用户表是否存在
			exists, err := checkTableExists(userTableName)
			if err != nil || !exists {
				continue // 跳过不存在的表
			}

			// 统计该应用的用户数
			var appUsers int64
			sql := fmt.Sprintf("SELECT COUNT(*) FROM %s", userTableName)
			err = o.Raw(sql).QueryRow(&appUsers)
			if err == nil {
				totalUsers += appUsers
			}
		}
	}
	stats.TotalUsers = totalUsers

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
	var apps []Application

	// 获取所有激活的应用
	_, err := o.QueryTable("apps").Filter("status", "active").OrderBy("-created_at").Limit(limit).All(&apps)
	if err != nil {
		return nil, err
	}

	var topApps []TopApp
	for _, app := range apps {
		// 清理应用ID，确保表名安全
		cleanAppId := strings.ReplaceAll(app.AppId, "-", "_")
		cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
		userTableName := fmt.Sprintf("user_%s", cleanAppId)

		// 统计该应用的用户数
		var userCount int64 = 0
		exists, err := checkTableExists(userTableName)
		if err == nil && exists {
			sql := fmt.Sprintf("SELECT COUNT(*) FROM %s", userTableName)
			o.Raw(sql).QueryRow(&userCount)
		}

		topApps = append(topApps, TopApp{
			AppId:       app.AppId,
			AppName:     app.AppName,
			UserCount:   userCount,
			AccessCount: 0, // 访问次数暂时设为0，如果需要可以后续添加统计
		})
	}

	return topApps, nil
}

// GetRecentActivity 获取最近活动
func GetRecentActivity(limit int) ([]RecentActivity, error) {
	o := orm.NewOrm()
	var activities []RecentActivity

	sql := `
		SELECT id, user_id, username, action, resource as description, created_at as createdAt
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
	o := orm.NewOrm()
	var growth []UserGrowth

	// 获取所有激活的应用
	var apps []Application
	_, err := o.QueryTable("apps").Filter("status", "active").All(&apps)
	if err != nil {
		return growth, err
	}

	// 根据时间范围生成数据
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		var newUsers int64 = 0
		var totalUsers int64 = 0

		// 遍历所有应用，统计各应用的用户数据
		for _, app := range apps {
			// 清理应用ID，确保表名安全
			cleanAppId := strings.ReplaceAll(app.AppId, "-", "_")
			cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
			userTableName := fmt.Sprintf("user_%s", cleanAppId)

			// 检查用户表是否存在
			exists, err := checkTableExists(userTableName)
			if err != nil || !exists {
				continue // 跳过不存在的表
			}

			// 统计当日新增用户数
			var dailyNew int64
			sql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(created_at) = ?", userTableName)
			err = o.Raw(sql, dateStr).QueryRow(&dailyNew)
			if err == nil {
				newUsers += dailyNew
			}

			// 统计累计用户数（截至当日）
			var dailyTotal int64
			sql = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE DATE(created_at) <= ?", userTableName)
			err = o.Raw(sql, dateStr).QueryRow(&dailyTotal)
			if err == nil {
				totalUsers += dailyTotal
			}
		}

		growth = append(growth, UserGrowth{
			Date:       dateStr,
			NewUsers:   newUsers,
			TotalUsers: totalUsers,
		})
	}

	return growth, nil
}

// GetPlatformDistribution 获取平台分布统计
func GetPlatformDistribution() ([]PlatformDistribution, error) {
	o := orm.NewOrm()
	var distribution []PlatformDistribution

	sql := `
		SELECT 
			platform,
			COUNT(*) as count
		FROM apps 
		WHERE status = 'active' AND platform != ''
		GROUP BY platform
		ORDER BY count DESC
	`

	type platformResult struct {
		Platform string `json:"platform"`
		Count    int64  `json:"count"`
	}

	var results []platformResult
	_, err := o.Raw(sql).QueryRows(&results)
	if err != nil {
		return distribution, err
	}

	// 转换为前端需要的格式，添加中文名称
	platformNames := map[string]string{
		"wechat":  "微信小程序",
		"alipay":  "支付宝小程序",
		"douyin":  "抖音小程序",
		"baidu":   "百度小程序",
		"ios":     "iOS应用",
		"android": "Android应用",
	}

	for _, result := range results {
		name := platformNames[result.Platform]
		if name == "" {
			name = result.Platform // 如果没有匹配的中文名，使用原始值
		}

		distribution = append(distribution, PlatformDistribution{
			Platform: result.Platform,
			Name:     name,
			Count:    result.Count,
			Value:    result.Count, // ECharts需要value字段
		})
	}

	return distribution, nil
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
	// 获取基础统计数据
	stats, err := GetLeaderboardStatsByAppId(appId)
	if err != nil {
		return nil, err
	}

	// 保持向后兼容性，添加别名
	if _, ok := stats["totalPlayers"]; ok {
		stats["active"] = stats["total"] // 活跃排行榜数等于总数
	}

	if todaySubmissions, ok := stats["todaySubmissions"]; ok {
		stats["todaySubmits"] = todaySubmissions // 兼容旧字段名
	}

	// 获取排行榜列表，包含玩家详细信息
	leaderboards, _, err := GetLeaderboardList(appId, 1, 100, "")
	if err != nil {
		return stats, err
	}

	var leaderboardData []map[string]interface{}
	for _, lb := range leaderboards {
		// 获取每个排行榜的前10名数据，包含玩家详细信息
		topPlayers, _, err := GetLeaderboardData(appId, lb.LeaderboardType, 1, 10)
		if err != nil {
			continue
		}

		leaderboardInfo := map[string]interface{}{
			"leaderboardType": lb.LeaderboardType,
			"name":            lb.Name,
			"description":     lb.Description,
			"scoreType":       lb.ScoreType,
			"maxRank":         lb.MaxRank,
			"enabled":         lb.Enabled,
			"category":        lb.Category,
			"resetType":       lb.ResetType,
			"topPlayers":      topPlayers,
		}
		leaderboardData = append(leaderboardData, leaderboardInfo)
	}

	stats["leaderboards"] = leaderboardData

	return stats, nil
}
