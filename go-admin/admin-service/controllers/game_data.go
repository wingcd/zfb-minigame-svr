package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
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

	// 解析JSON请求参数 - 对齐云函数格式
	type GetUserDataListRequest struct {
		AppId    string `json:"appId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		UserId   string `json:"userId,omitempty"`
		PlayerId string `json:"playerId,omitempty"` // 兼容参数
	}

	var req GetUserDataListRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数[appId]错误", nil)
		return
	}

	// 参数校验 - 对齐云函数错误码
	if req.AppId == "" {
		utils.CloudResponse(&c.Controller, 4001, "参数[appId]错误", nil)
		return
	}

	// 设置默认值 - 对齐云函数逻辑
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 兼容userId和playerId
	userId := req.UserId
	if userId == "" {
		userId = req.PlayerId
	}

	// 获取用户数据列表
	userDataList, total, err := models.GetUserDataList(req.AppId, req.Page, req.PageSize, userId)
	if err != nil {
		utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		return
	}

	// 构建响应数据 - 对齐云函数格式
	data := map[string]interface{}{
		"list":     userDataList,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}

	utils.CloudResponse(&c.Controller, 0, "success", data)
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
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type SendMailRequest struct {
		AppId       string `json:"appId"`
		PlayerId    string `json:"playerId"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		Attachments string `json:"attachments,omitempty"`
	}

	var req SendMailRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数错误", nil)
		return
	}

	// 参数校验 - 对齐云函数错误码
	if req.AppId == "" {
		utils.CloudResponse(&c.Controller, 4001, "应用ID不能为空", nil)
		return
	}

	if req.PlayerId == "" {
		utils.CloudResponse(&c.Controller, 4001, "用户ID不能为空", nil)
		return
	}

	if req.Title == "" {
		utils.CloudResponse(&c.Controller, 4001, "邮件标题不能为空", nil)
		return
	}

	if req.Content == "" {
		utils.CloudResponse(&c.Controller, 4001, "邮件内容不能为空", nil)
		return
	}

	// 发送邮件
	err := models.SendMail(req.AppId, req.PlayerId, req.Title, req.Content, req.Attachments)
	if err != nil {
		utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		return
	}

	utils.CloudResponse(&c.Controller, 0, "success", map[string]interface{}{})
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
	expireDayInt, _ := c.GetInt("expireDay", 7)

	if appId == "" || title == "" || content == "" || len(userIds) == 0 {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID、标题、内容和用户列表不能为空", nil)
		return
	}

	// 发送广播邮件
	err := models.SendBroadcastMail(appId, title, content, rewards, expireDayInt)
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
	configType := c.GetString("configType")
	version := c.GetString("version")

	if appId == "" || configKey == "" || configValue == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID、配置键和配置值不能为空", nil)
		return
	}

	// 更新配置
	err := models.SetConfig(appId, configKey, configValue, configType, version)
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
