package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"
)

type SystemController struct {
	web.Controller
}

// GetSystemConfig 获取系统配置
func (c *SystemController) GetSystemConfig() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取系统配置
	systemConfig, err := models.GetSystemConfig()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取系统配置失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", systemConfig)
}

// UpdateSystemConfig 更新系统配置
func (c *SystemController) UpdateSystemConfig() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 解析JSON请求数据
	var requestData struct {
		SiteName           string `json:"siteName"`
		SiteUrl            string `json:"siteUrl"`
		SiteLogo           string `json:"siteLogo"`
		SiteDescription    string `json:"siteDescription"`
		SiteKeywords       string `json:"siteKeywords"`
		AdminEmail         string `json:"adminEmail"`
		EnableRegister     bool   `json:"enableRegister"`
		EnableEmailVerify  bool   `json:"enableEmailVerify"`
		EnableCaptcha      bool   `json:"enableCaptcha"`
		JwtSecret          string `json:"jwtSecret"`
		JwtExpireHours     int    `json:"jwtExpireHours"`
		EnableCache        bool   `json:"enableCache"`
		CacheExpireMinutes int    `json:"cacheExpireMinutes"`
		LogLevel           string `json:"logLevel"`
		LogRetentionDays   int    `json:"logRetentionDays"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &requestData); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	siteName := requestData.SiteName
	siteUrl := requestData.SiteUrl
	siteLogo := requestData.SiteLogo
	siteDescription := requestData.SiteDescription
	siteKeywords := requestData.SiteKeywords
	adminEmail := requestData.AdminEmail
	enableRegister := requestData.EnableRegister
	enableEmailVerify := requestData.EnableEmailVerify
	enableCaptcha := requestData.EnableCaptcha
	jwtSecret := requestData.JwtSecret
	jwtExpireHours := requestData.JwtExpireHours
	enableCache := requestData.EnableCache
	cacheExpireMinutes := requestData.CacheExpireMinutes
	logLevel := requestData.LogLevel
	logRetentionDays := requestData.LogRetentionDays

	if siteName == "" {
		utils.ErrorResponse(&c.Controller, 1002, "站点名称不能为空", nil)
		return
	}

	// 设置默认值
	if jwtExpireHours <= 0 {
		jwtExpireHours = 24
	}
	if cacheExpireMinutes <= 0 {
		cacheExpireMinutes = 30
	}
	if logRetentionDays <= 0 {
		logRetentionDays = 30
	}
	if logLevel == "" {
		logLevel = "info"
	}

	// 创建SystemConfig结构
	config := &models.SystemConfig{}

	// 直接设置字段
	config.SiteName = siteName
	config.SiteUrl = siteUrl
	config.SiteLogo = siteLogo
	config.SiteDescription = siteDescription
	config.SiteKeywords = siteKeywords
	config.AdminEmail = adminEmail
	config.EnableRegister = enableRegister
	config.EnableEmailVerify = enableEmailVerify
	config.EnableCaptcha = enableCaptcha
	config.JwtSecret = jwtSecret
	config.JwtExpireHours = jwtExpireHours
	config.EnableCache = enableCache
	config.CacheExpireMinutes = cacheExpireMinutes
	config.LogLevel = logLevel
	config.LogRetentionDays = logRetentionDays

	err := models.UpdateSystemConfig(config)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新系统配置失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "更新系统配置", "更新系统配置")

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// GetSystemStatus 获取系统状态
func (c *SystemController) GetSystemStatus() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取系统状态
	systemStatus, err := models.GetSystemStatus()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取系统状态失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", systemStatus)
}

// ClearCache 清理缓存
func (c *SystemController) ClearCache() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取缓存类型
	cacheType := c.GetString("cacheType", "all") // all, user, app, config

	// 清理缓存
	err := models.ClearCache()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "清理缓存失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "清理缓存", "清理"+cacheType+"缓存")

	utils.SuccessResponse(&c.Controller, "清理成功", nil)
}

// GetCacheStats 获取缓存统计
func (c *SystemController) GetCacheStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取缓存统计
	cacheStats, err := models.GetCacheStats()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取缓存统计失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", cacheStats)
}

// CleanLogs 清理日志
func (c *SystemController) CleanLogs() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	daysStr := c.GetString("days", "30")
	logType := c.GetString("logType", "all") // all, operation, error, access

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	// 清理日志
	err = models.CleanLogs(days)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "清理日志失败: "+err.Error(), nil)
		return
	}
	cleanedCount := 0 // 假设清理了0条记录

	// 记录操作日志
	utils.LogOperation(claims.UserID, "清理日志", "清理"+strconv.Itoa(days)+"天前的"+logType+"日志")

	result := map[string]interface{}{
		"cleanedCount": cleanedCount,
		"days":         days,
		"logType":      logType,
	}

	utils.SuccessResponse(&c.Controller, "清理成功", result)
}

// BackupData 备份数据
func (c *SystemController) BackupData() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数 (暂时不使用)
	_ = c.GetString("backupType", "full") // full, data, structure
	_ = c.GetStrings("appIds")

	// 创建备份
	backupFile, err := models.CreateBackup()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "创建备份失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "创建备份", "创建备份")

	result := map[string]interface{}{
		"backupFile": backupFile,
		"createdAt":  time.Now().Format("2006-01-02 15:04:05"),
	}

	utils.SuccessResponse(&c.Controller, "备份成功", result)
}

// GetBackupList 获取备份列表
func (c *SystemController) GetBackupList() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 获取备份列表
	backups, err := models.GetBackupList()
	total := len(backups)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取备份列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"backups":  backups,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// RestoreBackup 恢复备份
func (c *SystemController) RestoreBackup() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	backupId := c.GetString("backupId")
	if backupId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "备份ID不能为空", nil)
		return
	}

	// 恢复备份
	err := models.RestoreBackup(backupId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "恢复备份失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "恢复备份", "恢复备份ID: "+backupId)

	utils.SuccessResponse(&c.Controller, "恢复成功", nil)
}

// DeleteBackup 删除备份
func (c *SystemController) DeleteBackup() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取备份ID
	backupIdStr := c.Ctx.Input.Param(":id")
	if backupIdStr == "" {
		utils.ErrorResponse(&c.Controller, 1002, "备份ID不能为空", nil)
		return
	}

	_, err := strconv.Atoi(backupIdStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "备份ID格式错误", nil)
		return
	}

	// 删除备份
	err = models.DeleteBackup(backupIdStr)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "删除备份失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "删除备份", "删除备份ID: "+backupIdStr)

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}

// GetServerInfo 获取服务器信息
func (c *SystemController) GetServerInfo() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取服务器信息
	serverInfo, err := models.GetServerInfo()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取服务器信息失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", serverInfo)
}

// GetDatabaseInfo 获取数据库信息
func (c *SystemController) GetDatabaseInfo() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取数据库信息
	dbInfo, err := models.GetDatabaseInfo()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取数据库信息失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", dbInfo)
}

// OptimizeDatabase 优化数据库
func (c *SystemController) OptimizeDatabase() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 优化数据库
	result, err := models.OptimizeDatabase()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "优化数据库失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "优化数据库", "执行数据库优化")

	utils.SuccessResponse(&c.Controller, "优化成功", result)
}
