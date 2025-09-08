package controllers

import (
	"admin-service/models"
	"admin-service/utils"
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
)

type ApplicationController struct {
	web.Controller
}

// GetApplications 获取应用列表（对齐云函数getAppList接口）
func (c *ApplicationController) GetApplications() {
	var req struct {
		Page     int    `json:"page"`
		PageSize int    `json:"pageSize"`
		Keyword  string `json:"keyword"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	// 获取应用列表
	applications, total, err := models.GetApplicationList(req.Page, req.PageSize, req.Keyword)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "获取应用列表失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 创建安全的应用列表响应，排除敏感字段
	safeAppList := make([]map[string]interface{}, len(applications))
	for i, app := range applications {
		safeAppList[i] = map[string]interface{}{
			"id":           app.ID,
			"appId":        app.AppId,
			"appName":      app.AppName,
			"description":  app.Description,
			"channelAppId": app.ChannelAppId,
			// ChannelAppKey 敏感信息，不返回给客户端
			"category":      app.Category,
			"platform":      app.Platform,
			"status":        app.Status,
			"version":       app.Version,
			"minVersion":    app.MinVersion,
			"settings":      app.Settings,
			"userCount":     app.UserCount,
			"scoreCount":    app.ScoreCount,
			"dailyActive":   app.DailyActive,
			"monthlyActive": app.MonthlyActive,
			"createdBy":     app.CreatedBy,
			"createdAt":     app.CreatedAt,
			"updatedAt":     app.UpdatedAt,
		}
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"list":     safeAppList,
			"total":    total,
			"page":     req.Page,
			"pageSize": req.PageSize,
		},
	}
	c.ServeJSON()
}

// CreateApplication 创建应用（对齐云函数createApp接口）
func (c *ApplicationController) CreateApplication() {
	var req struct {
		AppName       string `json:"appName"`
		Platform      string `json:"platform"`
		ChannelAppId  string `json:"channelAppId"`
		ChannelAppKey string `json:"channelAppKey"`
		Description   string `json:"description"`
		Status        string `json:"status"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.AppName == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "应用名称不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 创建应用
	application := &models.Application{
		AppName:       req.AppName,
		Platform:      req.Platform,
		ChannelAppId:  req.ChannelAppId,
		ChannelAppKey: req.ChannelAppKey,
		Description:   req.Description,
		Status:        "active", // 默认状态为活跃
	}

	err := application.Insert()
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "创建应用失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "CREATE", "APP", map[string]interface{}{
		"appId":   application.AppId,
		"appName": application.AppName,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "创建成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"id":      application.ID,
			"appId":   application.AppId,
			"appName": application.AppName,
		},
	}
	c.ServeJSON()
}

// GetApplication 获取应用详情（对齐云函数getApp接口）
func (c *ApplicationController) GetApplication() {
	var req struct {
		AppId string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "应用ID不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取应用详情
	application := &models.Application{}
	err := application.GetByAppId(req.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "应用不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 创建安全的响应数据，排除敏感字段
	safeAppData := map[string]interface{}{
		"id":           application.ID,
		"appId":        application.AppId,
		"appName":      application.AppName,
		"description":  application.Description,
		"channelAppId": application.ChannelAppId,
		// ChannelAppKey 敏感信息，不返回给客户端
		"category":      application.Category,
		"platform":      application.Platform,
		"status":        application.Status,
		"version":       application.Version,
		"minVersion":    application.MinVersion,
		"settings":      application.Settings,
		"userCount":     application.UserCount,
		"scoreCount":    application.ScoreCount,
		"dailyActive":   application.DailyActive,
		"monthlyActive": application.MonthlyActive,
		"createdBy":     application.CreatedBy,
		"createdAt":     application.CreatedAt,
		"updatedAt":     application.UpdatedAt,
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      safeAppData,
	}
	c.ServeJSON()
}

// UpdateApplication 更新应用（对齐云函数updateApp接口）
func (c *ApplicationController) UpdateApplication() {
	// JWT验证
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	// 解析JSON请求体
	var request struct {
		AppId         string `json:"appId"` // 支持appId参数
		AppName       string `json:"appName"`
		Platform      string `json:"platform"`
		ChannelAppId  string `json:"channelAppId"`  // 添加渠道相关字段
		ChannelAppKey string `json:"channelAppKey"` // 添加渠道相关字段
		Description   string `json:"description"`
		Status        string `json:"status"` // 状态
	}

	// 解析JSON请求体
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 如果appId为空，尝试从URL参数获取
	if request.AppId == "" {
		request.AppId = c.Ctx.Input.Param(":id")
	}

	if request.AppId == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用ID不能为空", nil)
		return
	}

	if request.AppName == "" {
		utils.ErrorResponse(&c.Controller, 1002, "应用名称不能为空", nil)
		return
	}

	// 获取原应用信息
	application := &models.Application{}
	err := application.GetByAppId(request.AppId)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "应用不存在: "+err.Error(), nil)
		return
	}

	// 更新字段
	fieldsToUpdate := []string{"app_name", "description"}

	application.AppName = request.AppName
	application.Description = request.Description
	if request.Platform != "" {
		application.Platform = request.Platform
		fieldsToUpdate = append(fieldsToUpdate, "platform")
	}
	if request.ChannelAppKey != "" {
		application.ChannelAppKey = request.ChannelAppKey
		fieldsToUpdate = append(fieldsToUpdate, "channel_app_key")
	}
	if request.ChannelAppId != "" {
		application.ChannelAppId = request.ChannelAppId
		fieldsToUpdate = append(fieldsToUpdate, "channel_app_id")
	}
	if request.Status != "" {
		application.Status = request.Status
		fieldsToUpdate = append(fieldsToUpdate, "status")
	}

	// 执行更新
	err = application.Update(fieldsToUpdate...)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 1003, "更新应用失败: "+err.Error(), nil)
		return
	}

	// 记录操作日志
	utils.LogOperation(claims.UserID, "更新应用", "更新应用: "+request.AppName)

	utils.SuccessResponse(&c.Controller, "更新成功", nil)
}

// DeleteApplication 删除应用（对齐云函数deleteApp接口）
func (c *ApplicationController) DeleteApplication() {
	// JWT验证并获取用户信息
	claims := utils.ValidateJWT(c.Ctx)
	if claims == nil {
		return
	}

	var req struct {
		AppId string `json:"appId"`
		Force bool   `json:"force"` // 是否强制删除（硬删除）
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "参数解析失败",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	if req.AppId == "" {
		c.Data["json"] = map[string]interface{}{
			"code":      4001,
			"msg":       "应用ID不能为空",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 获取应用信息用于验证
	application := &models.Application{}
	err := application.GetByAppId(req.AppId)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      4004,
			"msg":       "应用不存在",
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查是否为超级管理员
	isSuperAdmin, err := c.isSuperAdmin(claims.RoleID)
	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "权限验证失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 检查应用状态
	isAppActive := application.Status == "active"

	var deleteType string
	// 根据用户权限、应用状态和请求参数决定删除方式
	if isSuperAdmin && (req.Force || !isAppActive) {
		// 超级管理员在以下情况可以进行硬删除：
		// 1. 明确指定force=true
		// 2. 应用状态为非活跃（inactive或pending）
		err = models.HardDeleteApplication(application.ID)
		deleteType = "HARD_DELETE"
	} else {
		// 普通管理员或删除活跃状态应用时进行软删除
		err = models.DeleteApplication(application.ID)
		deleteType = "SOFT_DELETE"
	}

	if err != nil {
		c.Data["json"] = map[string]interface{}{
			"code":      5001,
			"msg":       "删除应用失败: " + err.Error(),
			"timestamp": utils.UnixMilli(),
			"data":      nil,
		}
		c.ServeJSON()
		return
	}

	// 记录操作日志
	models.LogAdminOperation(claims.UserID, claims.Username, deleteType, "APP", map[string]interface{}{
		"deletedAppId": req.AppId,
		"appName":      application.AppName,
		"force":        req.Force,
		"isSuperAdmin": isSuperAdmin,
		"isAppActive":  isAppActive,
		"appStatus":    application.Status,
	})

	var message string
	if deleteType == "HARD_DELETE" {
		if req.Force {
			message = "应用已彻底删除（强制删除，包含所有数据表）"
		} else {
			message = "应用已彻底删除（应用非活跃状态，自动硬删除）"
		}
	} else {
		message = "应用已停用（软删除）"
	}

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       message,
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"deleteType": deleteType,
			"appId":      req.AppId,
		},
	}
	c.ServeJSON()
}

// isSuperAdmin 检查用户是否为超级管理员
func (c *ApplicationController) isSuperAdmin(roleID int64) (bool, error) {
	role := &models.AdminRole{}
	err := role.GetById(roleID)
	if err != nil {
		return false, err
	}
	return role.RoleCode == "super_admin", nil
}
