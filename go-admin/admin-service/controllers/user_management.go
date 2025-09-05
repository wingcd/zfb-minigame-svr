package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"strconv"

	"github.com/beego/beego/v2/server/web"
)

type UserManagementController struct {
	web.Controller
}

// GetAllUsers 获取用户列表（分页、搜索）
func (c *UserManagementController) GetAllUsers() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	appId := c.GetString("appId")
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")
	keyword := c.GetString("keyword", "")
	status := c.GetString("status", "")

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

	// 获取用户列表
	userList, total, err := models.GetAllGameUsers(appId, page, pageSize, keyword, status)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取用户列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"userList": userList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// GetUserDetail 获取用户详细信息
func (c *UserManagementController) GetUserDetail() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	appId := c.GetString("appId")
	playerId := c.GetString("playerId")

	if appId == "" || playerId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID和玩家ID不能为空", nil)
		return
	}

	// 获取用户详细信息
	userDetail, err := models.GetGameUserDetail(appId, playerId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取用户详情失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", userDetail)
}

// UpdateUserData 更新用户游戏数据
func (c *UserManagementController) UpdateUserData() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	var req struct {
		AppId    string          `json:"appId"`
		PlayerId string          `json:"playerId"`
		Data     json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "请求参数解析失败", nil)
		return
	}

	if req.AppId == "" || req.PlayerId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID和玩家ID不能为空", nil)
		return
	}

	// 更新用户数据
	err := models.UpdateGameUserData(req.AppId, req.PlayerId, string(req.Data))
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新用户数据失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// BanUser 封禁用户
func (c *UserManagementController) BanUser() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	var req struct {
		AppId     string `json:"appId"`
		PlayerId  string `json:"playerId"`
		BanType   string `json:"banType"` // temporary, permanent
		BanReason string `json:"banReason"`
		BanHours  int    `json:"banHours"` // 封禁时长（小时），永久封禁时为0
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "请求参数解析失败", nil)
		return
	}

	if req.AppId == "" || req.PlayerId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID、玩家ID不能为空", nil)
		return
	}

	// 获取管理员信息
	adminInfo := utils.GetJWTUserInfo(c.Ctx)
	if adminInfo == nil {
		utils.ErrorResponse(&c.Controller, 4001, "未登录", nil)
		return
	}

	// 创建封禁记录
	err := models.BanGameUser(req.AppId, req.PlayerId, adminInfo.ID, req.BanType, req.BanReason, req.BanHours)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "封禁用户失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "封禁成功", nil)
}

// UnbanUser 解封用户
func (c *UserManagementController) UnbanUser() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	var req struct {
		AppId       string `json:"appId"`
		PlayerId    string `json:"playerId"`
		UnbanReason string `json:"unbanReason"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "请求参数解析失败", nil)
		return
	}

	if req.AppId == "" || req.PlayerId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID、玩家ID不能为空", nil)
		return
	}

	// 获取管理员信息
	adminInfo := utils.GetJWTUserInfo(c.Ctx)
	if adminInfo == nil {
		utils.ErrorResponse(&c.Controller, 4001, "未登录", nil)
		return
	}

	// 解封用户
	err := models.UnbanGameUser(req.AppId, req.PlayerId, adminInfo.ID, req.UnbanReason)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "解封用户失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "解封成功", nil)
}

// DeleteUser 删除用户（危险操作）
func (c *UserManagementController) DeleteUser() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	appId := c.GetString("appId")
	playerId := c.GetString("playerId")

	if appId == "" || playerId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID和玩家ID不能为空", nil)
		return
	}

	// 获取管理员信息并检查权限
	adminInfo := utils.GetJWTUserInfo(c.Ctx)
	if adminInfo == nil {
		utils.ErrorResponse(&c.Controller, 4001, "未登录", nil)
		return
	}

	// 只有超级管理员才能删除用户
	if adminInfo.Role != "super_admin" {
		utils.ErrorResponse(&c.Controller, 4003, "权限不足，只有超级管理员才能删除用户", nil)
		return
	}

	// 删除用户数据
	err := models.DeleteGameUser(appId, playerId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "删除用户失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}

// GetUserStats 获取用户统计信息
func (c *UserManagementController) GetUserStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	appId := c.GetString("appId")
	playerId := c.GetString("playerId")

	if appId == "" || playerId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID和玩家ID不能为空", nil)
		return
	}

	// 获取用户统计信息
	stats, err := models.GetGameUserStats(appId, playerId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取用户统计失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", stats)
}

// GetUserRegistrationStats 获取用户注册统计
func (c *UserManagementController) GetUserRegistrationStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	appId := c.GetString("appId")
	days := c.GetString("days", "7")

	if appId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	daysInt, err := strconv.Atoi(days)
	if err != nil || daysInt <= 0 {
		daysInt = 7
	}

	// 获取注册统计
	stats, err := models.GetUserRegistrationStats(appId, daysInt)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取注册统计失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", stats)
}
