package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"
	"strconv"

	"github.com/beego/beego/v2/server/web"
)

type ApplicationController struct {
	web.Controller
}

// GetApplications 获取应用列表
func (c *ApplicationController) GetApplications() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取分页参数
	pageStr := c.GetString("page", "1")
	pageSizeStr := c.GetString("pageSize", "10")
	keyword := c.GetString("keyword", "")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 获取应用列表
	applications, total, err := models.GetApplicationList(page, pageSize, keyword)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取应用列表失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"applications": applications,
		"total":        total,
		"page":         page,
		"pageSize":     pageSize,
	}

	utils.SuccessResponse(&c.Controller, "获取成功", result)
}

// CreateApplication 创建应用
func (c *ApplicationController) CreateApplication() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 解析JSON请求体
	var request struct {
		AppName       string `json:"appName"`
		Platform      string `json:"platform"`
		ChannelAppId  string `json:"channelAppId"`
		ChannelAppKey string `json:"channelAppKey"`
		Description   string `json:"description"`
		Status        string `json:"status"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		utils.ErrorResponse(&c.Controller, 1001, "请求参数格式错误", nil)
		return
	}

	if request.AppName == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用名称不能为空", nil)
		return
	}

	// 设置状态
	status := 1
	if request.Status == "inactive" {
		status = 0
	}

	// 创建应用
	application := &models.Application{
		AppName:       request.AppName,
		Platform:      request.Platform,
		ChannelAppId:  request.ChannelAppId,
		ChannelAppKey: request.ChannelAppKey,
		Description:   request.Description,
		Status:        status,
	}

	err := application.Insert()
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "创建应用失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "创建应用", "创建应用: "+request.AppName)

	result := map[string]interface{}{
		"id":        application.Id,
		"appId":     application.AppId,
		"appSecret": application.AppSecret,
		"appName":   application.AppName,
	}

	utils.SuccessResponse(&c.Controller, "创建成功", result)
}

// GetApplication 获取应用详情
func (c *ApplicationController) GetApplication() {
	// JWT验证
	if utils.ValidateJWT(c.Ctx) == nil {
		return
	}

	// 获取参数
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID格式错误", nil)
		return
	}

	// 获取应用详情
	application := &models.Application{}
	err = application.GetById(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "获取应用详情失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "获取成功", application)
}

// UpdateApplication 更新应用
func (c *ApplicationController) UpdateApplication() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID格式错误", nil)
		return
	}

	appName := c.GetString("appName")
	description := c.GetString("description")
	statusStr := c.GetString("status")

	if appName == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用名称不能为空", nil)
		return
	}

	status := 1
	if statusStr != "" {
		status, err = strconv.Atoi(statusStr)
		if err != nil {
			utils.ErrorResponse(&c.Controller, 1002, "状态格式错误", nil)
			return
		}
	}

	// 更新应用
	application := &models.Application{
		BaseModel:   models.BaseModel{Id: id},
		AppName:     appName,
		Description: description,
		Status:      status,
	}
	err = application.Update("app_name", "description", "status")
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新应用失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "更新应用", "更新应用: "+appName)

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// DeleteApplication 删除应用
func (c *ApplicationController) DeleteApplication() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID格式错误", nil)
		return
	}

	// 获取应用信息用于日志
	application := &models.Application{}
	err = application.GetById(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "应用不存在", nil)
		return
	}

	// 删除应用
	err = models.DeleteApplication(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "删除应用失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "删除应用", "删除应用: "+application.AppName)

	utils.SuccessResponse(&c.Controller, "删除成功", nil)
}

// ResetAppSecret 重置应用密钥
func (c *ApplicationController) ResetAppSecret() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 获取参数
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID格式错误", nil)
		return
	}

	// 重置密钥
	application := &models.Application{}
	err = application.GetById(id)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "应用不存在", nil)
		return
	}

	// 生成新密钥
	application.AppSecret = utils.GenerateAppSecret()
	err = application.Update("app_secret")
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "重置密钥失败: "+err.Error(), nil)
		return
	}
	newSecret := application.AppSecret

	// 记录操作日志
	utils.LogOperation(claims.UserID, "重置密钥", "重置应用密钥，应用ID: "+idStr)

	result := map[string]interface{}{
		"appSecret": newSecret,
	}

	utils.SuccessResponse(&c.Controller, "重置成功", result)
}
