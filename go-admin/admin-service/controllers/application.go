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

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"list":     applications,
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
		Status        int    `json:"status"`
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
		ChannelAppKey: req.ChannelAppId,
		AppSecret:     req.ChannelAppKey,
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
			"id":      application.Id,
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

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "获取成功",
		"timestamp": utils.UnixMilli(),
		"data":      application,
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
		AppId         string      `json:"appId"` // 支持appId参数
		AppName       string      `json:"appName"`
		Platform      string      `json:"platform"`
		ChannelAppId  string      `json:"channelAppId"`  // 添加渠道相关字段
		ChannelAppKey string      `json:"channelAppKey"` // 添加渠道相关字段
		Description   string      `json:"description"`
		Status        interface{} `json:"status"` // 改为interface{}以支持int和string
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
	application.AppName = request.AppName
	application.Description = request.Description
	if request.Platform != "" {
		application.Platform = request.Platform
	}
	if request.ChannelAppKey != "" {
		application.ChannelAppKey = request.ChannelAppKey
	}
	if request.ChannelAppId != "" {
		application.ChannelAppId = request.ChannelAppId
	}

	// 处理状态字段 - 支持int和string类型
	if request.Status != nil {
		switch v := request.Status.(type) {
		case string:
			if v == "active" || v == "1" {
				application.Status = "active"
			} else if v == "inactive" || v == "0" {
				application.Status = "inactive"
			} else if v == "pending" {
				application.Status = "pending"
			} else {
				utils.ErrorResponse(&c.Controller, 1002, "状态格式错误", nil)
				return
			}
		case float64: // JSON中的数字会被解析为float64
			if v == 1 {
				application.Status = "active"
			} else if v == 0 {
				application.Status = "inactive"
			} else {
				utils.ErrorResponse(&c.Controller, 1002, "状态格式错误", nil)
				return
			}
		case int:
			if v == 1 {
				application.Status = "active"
			} else if v == 0 {
				application.Status = "inactive"
			} else {
				utils.ErrorResponse(&c.Controller, 1002, "状态格式错误", nil)
				return
			}
		default:
			utils.ErrorResponse(&c.Controller, 1002, "状态格式错误", nil)
			return
		}
	}

	// 执行更新
	err = application.Update("appName", "description", "status", "platform", "channelAppId", "channelAppKey")
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

	// 删除应用
	err = models.DeleteApplication(application.Id)
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
	models.LogAdminOperation(0, "SYSTEM", "DELETE", "APP", map[string]interface{}{
		"deletedAppId": req.AppId,
		"appName":      application.AppName,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "删除成功",
		"timestamp": utils.UnixMilli(),
		"data":      map[string]interface{}{},
	}
	c.ServeJSON()
}

// ResetAppSecret 重置应用密钥（对齐云函数resetAppSecret接口）
func (c *ApplicationController) ResetAppSecret() {
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

	// 获取应用信息
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

	// 记录操作日志
	models.LogAdminOperation(0, "SYSTEM", "RESET_SECRET", "APP", map[string]interface{}{
		"appId":   req.AppId,
		"appName": application.AppName,
	})

	c.Data["json"] = map[string]interface{}{
		"code":      0,
		"msg":       "重置成功",
		"timestamp": utils.UnixMilli(),
		"data": map[string]interface{}{
			"appSecret": application.AppSecret,
		},
	}
	c.ServeJSON()
}
