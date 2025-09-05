package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"strconv"

	"github.com/beego/beego/v2/server/web"
)

type GameDataController struct {
	web.Controller
}

// GetUserDataList 获取用户数据列表
func (c *GameDataController) GetUserDataList() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")
	userId := c.GetString("userId", "")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 获取用户数据列表
	userDataList, total, err := models.GetUserDataList(appId, page, pageSize, userId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取用户数据列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"userDataList": userDataList,
		"total":        total,
		"page":         page,
		"pageSize":     pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetLeaderboardList 获取排行榜列表
func (c *GameDataController) GetLeaderboardList() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	leaderboardName := c.GetString("leaderboardName", "")
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 获取排行榜列表
	leaderboardList, total, err := models.GetLeaderboardList(appId, page, pageSize, leaderboardName)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取排行榜列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"leaderboardList": leaderboardList,
		"total":           total,
		"page":            page,
		"pageSize":        pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetCounterList 获取计数器列表
func (c *GameDataController) GetCounterList() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 获取计数器列表
	counterList, total, err := models.GetCounterList(appId, page, pageSize)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取计数器列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"list":     counterList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetMailList 获取邮件列表
func (c *GameDataController) GetMailList() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 获取邮件列表
	mailList, total, err := models.GetAllMailList(appId, page, pageSize)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取邮件列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"mailList": mailList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// SendMail 发送邮件
func (c *GameDataController) SendMail() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	userId := c.GetString("userId")
	title := c.GetString("title")
	content := c.GetString("content")
	rewards := c.GetString("rewards")

	if appId == "" || userId == "" || title == "" || content == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID、用户ID、标题和内容不能为空", nil)
		return
	}

	// 发送邮件
	err := models.SendMail(appId, userId, title, content, rewards)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "发送邮件失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "发送邮件", "向用户 "+userId+" 发送邮件: "+title)

	utils.SuccessResponse(&c.Controller, "发送成功", nil)
}

// SendBroadcastMail 发送广播邮件
func (c *GameDataController) SendBroadcastMail() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	title := c.GetString("title")
	content := c.GetString("content")
	rewards := c.GetString("rewards")
	userIds := c.GetStrings("userIds")

	if appId == "" || title == "" || content == "" || len(userIds) == 0 {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID、标题、内容和用户列表不能为空", nil)
		return
	}

	// 发送广播邮件
	err := models.SendBroadcastMail(appId, title, content, rewards)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "发送广播邮件失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "发送广播邮件", "发送广播邮件: "+title)

	utils.SuccessResponse(&c.Controller, "发送成功", nil)
}

// GetConfigList 获取配置列表
func (c *GameDataController) GetConfigList() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")
	keyword := c.GetString("keyword", "")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 获取配置列表
	configList, total, err := models.GetConfigList(appId, page, pageSize, keyword)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取配置列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"configList": configList,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// UpdateConfig 更新配置
func (c *GameDataController) UpdateConfig() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	configKey := c.GetString("configKey")
	configValue := c.GetString("configValue")

	if appId == "" || configKey == "" || configValue == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID、配置键和配置值不能为空", nil)
		return
	}

	// 更新配置
	err := models.SetConfig(appId, configKey, configValue)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新配置失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "更新配置", "更新配置: "+configKey)

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// DeleteConfig 删除配置
func (c *GameDataController) DeleteConfig() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	configKey := c.GetString("configKey")

	if appId == "" || configKey == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID和配置键不能为空", nil)
		return
	}

	// 删除配置
	err := models.DeleteConfig(appId, configKey)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "删除配置失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "删除配置", "删除配置: "+configKey)

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}
