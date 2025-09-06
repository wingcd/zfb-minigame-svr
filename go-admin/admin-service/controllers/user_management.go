package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"fmt"

	"github.com/beego/beego/v2/server/web"
)

type UserManagementController struct {
	web.Controller
}

// validateApp 验证应用是否存在的辅助函数
func (c *UserManagementController) validateApp(appId string) bool {
	if appId == "" {
		utils.CloudResponse(&c.Controller, 4001, "参数[appId]错误", nil)
		return false
	}

	app := &models.Application{}
	if err := app.GetByAppId(appId); err != nil {
		utils.CloudResponse(&c.Controller, 4004, "应用不存在", nil)
		return false
	}
	return true
}

// GetAllUsers 获取用户列表（分页、搜索）
func (c *UserManagementController) GetAllUsers() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type GetAllUsersRequest struct {
		AppId    string `json:"appId"`
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		PlayerId string `json:"playerId,omitempty"`
		OpenId   string `json:"openId,omitempty"`
	}

	var req GetAllUsersRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数[appId]错误", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	// 设置默认值 - 对齐云函数逻辑
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 获取用户列表
	userList, total, err := models.GetAllGameUsers(req.AppId, req.Page, req.PageSize, req.PlayerId, "")
	if err != nil {
		utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		return
	}

	// 构建响应数据 - 对齐云函数格式
	data := map[string]interface{}{
		"list":     userList,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}

	utils.CloudResponse(&c.Controller, 0, "success", data)
}

// GetUserDetail 获取用户详细信息
func (c *UserManagementController) GetUserDetail() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type GetUserDetailRequest struct {
		AppId    string `json:"appId"`
		OpenId   string `json:"openId"`
		PlayerId string `json:"playerId,omitempty"` // 兼容旧参数
	}

	var req GetUserDetailRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数解析错误", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	// 优先使用playerId，其次使用openId
	playerId := req.PlayerId
	if playerId == "" {
		playerId = req.OpenId
	}

	if playerId == "" {
		utils.CloudResponse(&c.Controller, 4001, "用户openId不能为空", nil)
		return
	}

	// 获取用户详细信息
	userDetail, err := models.GetGameUserDetail(req.AppId, playerId)
	if err != nil {
		utils.CloudResponse(&c.Controller, 4004, "用户不存在", nil)
		return
	}

	// 构建响应数据 - 对齐云函数格式
	data := map[string]interface{}{
		"baseInfo": map[string]interface{}{
			"id":       userDetail.ID,
			"openId":   playerId,
			"playerId": playerId,
		},
		"userData": userDetail.Data,
		"userInfo": userDetail,
		"gameStats": map[string]interface{}{
			"totalScores":  0,
			"bestScore":    0,
			"avgScore":     0,
			"lastPlayTime": nil,
			"playDays":     0,
		},
	}

	utils.CloudResponse(&c.Controller, 0, "success", data)
}

// UpdateUserData 更新用户游戏数据
func (c *UserManagementController) UpdateUserData() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type UpdateUserDataRequest struct {
		AppId     string      `json:"appId"`
		OpenId    string      `json:"openId"`
		PlayerId  string      `json:"playerId,omitempty"` // 兼容旧参数
		UserData  interface{} `json:"userData"`
		NickName  string      `json:"nickName,omitempty"`
		Banned    *bool       `json:"banned,omitempty"`
		BanReason string      `json:"banReason,omitempty"`
	}

	var req UpdateUserDataRequest
	fmt.Printf("DEBUG UpdateUserData: RequestBody length: %d, content: %s\n", len(c.Ctx.Input.RequestBody), string(c.Ctx.Input.RequestBody))
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		fmt.Printf("DEBUG UpdateUserData: JSON unmarshal error: %v\n", err)
		utils.CloudResponse(&c.Controller, 4005, "无效的JSON数据格式", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	// 优先使用playerId，其次使用openId
	playerId := req.PlayerId
	if playerId == "" {
		playerId = req.OpenId
	}

	if playerId == "" {
		utils.CloudResponse(&c.Controller, 4001, "用户openId不能为空", nil)
		return
	}

	// 将UserData转换为JSON字符串
	var userDataJSON string
	if req.UserData != nil {
		if userData, err := json.Marshal(req.UserData); err != nil {
			utils.CloudResponse(&c.Controller, 4005, "无效的JSON数据格式", nil)
			return
		} else {
			userDataJSON = string(userData)
		}
	}

	// 更新用户数据
	err := models.UpdateGameUserData(req.AppId, playerId, userDataJSON)
	if err != nil {
		utils.CloudResponse(&c.Controller, 4004, "用户不存在", nil)
		return
	}

	utils.CloudResponse(&c.Controller, 0, "更新成功", map[string]interface{}{})
}

// BanUser 封禁用户
func (c *UserManagementController) BanUser() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type BanUserRequest struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
		Reason   string `json:"reason"`   // 封禁原因
		Duration int    `json:"duration"` // 封禁时长（小时），0表示永久封禁
	}

	var req BanUserRequest
	fmt.Printf("DEBUG BanUser: RequestBody length: %d, content: %s\n", len(c.Ctx.Input.RequestBody), string(c.Ctx.Input.RequestBody))
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		fmt.Printf("DEBUG BanUser: JSON unmarshal error: %v\n", err)
		utils.CloudResponse(&c.Controller, 4001, "参数[appId]错误", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	if req.PlayerId == "" {
		utils.CloudResponse(&c.Controller, 4001, "参数[playerId]错误", nil)
		return
	}

	// 设置默认封禁原因
	if req.Reason == "" {
		req.Reason = "违规行为"
	}

	// 设置默认封禁原因 - 对齐云函数逻辑
	if req.Reason == "" {
		req.Reason = "违规行为"
	}

	// 获取管理员信息
	adminInfo := utils.GetJWTUserInfo(c.Ctx)
	if adminInfo == nil {
		utils.CloudResponse(&c.Controller, 4003, "权限不足", nil)
		return
	}

	// 确定封禁类型
	banType := "temporary"
	if req.Duration <= 0 {
		banType = "permanent"
	}

	// 创建封禁记录
	err := models.BanGameUser(req.AppId, req.PlayerId, adminInfo.ID, banType, req.Reason, req.Duration)
	if err != nil {
		if err.Error() == "user already banned" {
			utils.CloudResponse(&c.Controller, 4003, "用户已被封禁", nil)
		} else if err.Error() == "user not found" {
			utils.CloudResponse(&c.Controller, 4004, "用户不存在", nil)
		} else {
			utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		}
		return
	}

	// 构建响应数据 - 对齐云函数格式
	data := map[string]interface{}{
		"playerId":  req.PlayerId,
		"banReason": req.Reason,
		"permanent": req.Duration == 0,
	}

	utils.CloudResponse(&c.Controller, 0, "封禁成功", data)
}

// UnbanUser 解封用户
func (c *UserManagementController) UnbanUser() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type UnbanUserRequest struct {
		AppId       string `json:"appId"`
		PlayerId    string `json:"playerId"`
		UnbanReason string `json:"unbanReason,omitempty"`
	}

	var req UnbanUserRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数解析错误", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	if req.PlayerId == "" {
		utils.CloudResponse(&c.Controller, 4001, "参数[playerId]错误", nil)
		return
	}

	// 设置默认解封原因
	if req.UnbanReason == "" {
		req.UnbanReason = "管理员操作"
	}

	// 获取管理员信息
	adminInfo := utils.GetJWTUserInfo(c.Ctx)
	if adminInfo == nil {
		utils.CloudResponse(&c.Controller, 4003, "权限不足", nil)
		return
	}

	// 解封用户
	err := models.UnbanGameUser(req.AppId, req.PlayerId, adminInfo.ID, req.UnbanReason)
	if err != nil {
		utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		return
	}

	utils.CloudResponse(&c.Controller, 0, "success", map[string]interface{}{})
}

// DeleteUser 删除用户（危险操作）
func (c *UserManagementController) DeleteUser() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type DeleteUserRequest struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
	}

	var req DeleteUserRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数解析错误", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	if req.PlayerId == "" {
		utils.CloudResponse(&c.Controller, 4001, "参数[playerId]错误", nil)
		return
	}

	// 获取管理员信息并检查权限
	adminInfo := utils.GetJWTUserInfo(c.Ctx)
	if adminInfo == nil {
		utils.CloudResponse(&c.Controller, 4003, "权限不足", nil)
		return
	}

	// 只有超级管理员才能删除用户
	if adminInfo.Role != "super_admin" {
		utils.CloudResponse(&c.Controller, 4005, "权限不足", nil)
		return
	}

	// 删除用户数据
	err := models.DeleteGameUser(req.AppId, req.PlayerId)
	if err != nil {
		utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		return
	}

	utils.CloudResponse(&c.Controller, 0, "success", map[string]interface{}{})
}

// GetUserStats 获取用户统计信息
func (c *UserManagementController) GetUserStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type GetUserStatsRequest struct {
		AppId    string `json:"appId"`
		PlayerId string `json:"playerId"`
	}

	var req GetUserStatsRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数解析错误", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	// 参数校验
	if req.PlayerId == "" {
		utils.CloudResponse(&c.Controller, 4001, "参数[playerId]错误", nil)
		return
	}

	// 获取用户统计信息
	stats, err := models.GetGameUserStats(req.AppId, req.PlayerId)
	if err != nil {
		utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		return
	}

	utils.CloudResponse(&c.Controller, 0, "success", stats)
}

// GetUserRegistrationStats 获取用户注册统计
func (c *UserManagementController) GetUserRegistrationStats() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 解析JSON请求参数 - 对齐云函数格式
	type GetRegistrationStatsRequest struct {
		AppId string `json:"appId"`
		Days  int    `json:"days,omitempty"`
	}

	var req GetRegistrationStatsRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.CloudResponse(&c.Controller, 4001, "参数解析错误", nil)
		return
	}

	// 应用验证
	if !c.validateApp(req.AppId) {
		return
	}

	// 设置默认值
	if req.Days <= 0 {
		utils.CloudResponse(&c.Controller, 4001, "参数[appId]错误", nil)
		return
	}

	// 设置默认值
	if req.Days <= 0 {
		req.Days = 7
	}

	// 获取注册统计
	stats, err := models.GetUserRegistrationStats(req.AppId, req.Days)
	if err != nil {
		utils.CloudResponse(&c.Controller, 5001, err.Error(), nil)
		return
	}

	utils.CloudResponse(&c.Controller, 0, "success", stats)
}
