package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"
)

type StatisticsController struct {
	web.Controller
}

// GetDashboard 获取仪表盘数据
func (c *StatisticsController) GetDashboard() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取总体统计
	totalApps, err := models.GetTotalApplications()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取应用总数失败: "+err.Error(), nil)
		return
	}

	totalAdmins, err := models.GetTotalAdmins()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取管理员总数失败: "+err.Error(), nil)
		return
	}

	// 获取今日统计
	today := time.Now().Format("2006-01-02")
	todayOperations, err := models.GetOperationsByDate(today)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取今日操作数失败: "+err.Error(), nil)
		return
	}

	// 获取应用状态统计
	activeApps, err := models.GetApplicationsByStatus(1)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取活跃应用数失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"totalApps":       totalApps,
		"totalAdmins":     totalAdmins,
		"activeApps":      activeApps,
		"todayOperations": todayOperations,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetApplicationStats 获取应用统计
func (c *StatisticsController) GetApplicationStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	// 获取用户数据统计
	userDataCount, err := models.GetUserDataCount(appId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取用户数据统计失败: "+err.Error(), nil)
		return
	}

	// 获取排行榜统计
	leaderboardCount, err := models.GetLeaderboardCount(appId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取排行榜统计失败: "+err.Error(), nil)
		return
	}

	// 获取计数器统计
	counterCount, err := models.GetCounterCount(appId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取计数器统计失败: "+err.Error(), nil)
		return
	}

	// 获取邮件统计
	mailCount, err := models.GetMailCount(appId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取邮件统计失败: "+err.Error(), nil)
		return
	}

	// 获取配置统计
	configCount, err := models.GetConfigCount(appId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取配置统计失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"userDataCount":    userDataCount,
		"leaderboardCount": leaderboardCount,
		"counterCount":     counterCount,
		"mailCount":        mailCount,
		"configCount":      configCount,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetOperationLogs 获取操作日志
func (c *StatisticsController) GetOperationLogs() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "20")
	adminId := c.GetString("adminId", "")
	action := c.GetString("action", "")
	startDate := c.GetString("startDate", "")
	endDate := c.GetString("endDate", "")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 获取操作日志
	logs, total, err := models.GetOperationLogs(page, pageSize, adminId, action, startDate, endDate)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取操作日志失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"logs":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetUserActivity 获取用户活跃度统计
func (c *StatisticsController) GetUserActivity() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	days := c.GetString("days", "7")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	dayCount, err := strconv.Atoi(days)
	if err != nil || dayCount <= 0 || dayCount > 30 {
		dayCount = 7
	}

	// 获取用户活跃度数据
	activityData, err := models.GetUserActivity(appId, dayCount)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取用户活跃度失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", activityData)
}

// GetDataTrends 获取数据趋势
func (c *StatisticsController) GetDataTrends() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	dataType := c.GetString("dataType", "userData") // userData, leaderboard, counter, mail
	days := c.GetString("days", "7")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	dayCount, err := strconv.Atoi(days)
	if err != nil || dayCount <= 0 || dayCount > 30 {
		dayCount = 7
	}

	// 获取数据趋势
	trendData, err := models.GetDataTrends(appId, dataType, dayCount)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取数据趋势失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"dataType": dataType,
		"days":     dayCount,
		"trends":   trendData,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// ExportData 导出数据
func (c *StatisticsController) ExportData() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	dataType := c.GetString("dataType")    // userData, leaderboard, counter, mail, config
	format := c.GetString("format", "csv") // csv, excel

	if appId == "" || dataType == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID和数据类型不能为空", nil)
		return
	}

	// 导出数据
	filters := map[string]interface{}{
		"appId":    appId,
		"dataType": dataType,
		"format":   format,
	}
	filePath, err := models.ExportData(dataType, filters)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "导出数据失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "导出数据", "导出应用 "+appId+" 的 "+dataType+" 数据")

	result := map[string]interface{}{
		"filePath": filePath,
		"dataType": dataType,
		"format":   format,
	}

	utils.SuccessResponse(&c.Controller, "导出成功", result)
}

// GetSystemInfo 获取系统信息
func (c *StatisticsController) GetSystemInfo() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取系统信息
	systemInfo, err := models.GetSystemInfo()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取系统信息失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", systemInfo)
}
