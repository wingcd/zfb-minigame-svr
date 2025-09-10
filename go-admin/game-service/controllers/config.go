package controllers

import (
	"encoding/json"
	"game-service/models"
	"game-service/utils"

	"github.com/beego/beego/v2/server/web"
)

type ConfigController struct {
	web.Controller
}

// GetConfigRequest 获取配置请求
type GetConfigRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
	ConfigKey string `json:"configKey"`
}

// SetConfigRequest 设置配置请求
type SetConfigRequest struct {
	AppId       string `json:"appId"`
	PlayerId    string `json:"playerId"`
	Token       string `json:"token"`
	Timestamp   int64  `json:"timestamp"`
	Ver         string `json:"ver"`
	Sign        string `json:"sign"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// GetConfigsByVersionRequest 获取版本配置请求
type GetConfigsByVersionRequest struct {
	Version string `json:"version"`
}

// DeleteConfigRequest 删除配置请求
type DeleteConfigRequest struct {
	ConfigKey string `json:"configKey"`
}

// parseRequest 解析请求参数
func (c *ConfigController) parseRequest(req interface{}) error {
	return json.Unmarshal(c.Ctx.Input.RequestBody, req)
}

// GetConfig 获取配置
func (c *ConfigController) GetConfig() {
	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req GetConfigRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	if req.ConfigKey == "" {
		utils.ErrorResponse(c.Ctx, 1002, "configKey参数不能为空", nil)
		return
	}

	configKey := req.ConfigKey

	// 获取配置
	configValue, err := models.GetConfig(appId, configKey)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取配置失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"configKey":   configKey,
		"configValue": configValue,
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}

// SetConfig 设置配置
func (c *ConfigController) SetConfig() {
	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req SetConfigRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	if req.ConfigKey == "" || req.ConfigValue == "" {
		utils.ErrorResponse(c.Ctx, 1002, "configKey和configValue参数不能为空", nil)
		return
	}

	configKey := req.ConfigKey
	configValue := req.ConfigValue
	version := req.Version
	description := req.Description

	// 设置配置
	err := models.SetConfig(appId, configKey, configValue, version, description)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "设置配置失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"configKey":   configKey,
		"configValue": configValue,
		"version":     version,
		"description": description,
	}

	utils.SuccessResponse(c.Ctx, "设置成功", result)
}

// GetConfigsByVersion 获取版本配置
func (c *ConfigController) GetConfigsByVersion() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req GetConfigsByVersionRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	version := req.Version

	// 获取配置
	configs, err := models.GetConfigsByVersion(appId, version)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取配置失败: "+err.Error(), nil)
		return
	}

	result := map[string]interface{}{
		"version": version,
		"configs": configs,
	}

	utils.SuccessResponse(c.Ctx, "获取成功", result)
}

// GetAllConfigs 获取所有配置
func (c *ConfigController) GetAllConfigs() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 获取所有配置
	configs, err := models.GetAllConfigs(appId)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "获取配置失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "获取成功", configs)
}

// DeleteConfig 删除配置
func (c *ConfigController) DeleteConfig() {
	// 从中间件获取应用ID
	appId := c.Ctx.Input.GetData("app_id").(string)

	// 解析请求参数
	var req DeleteConfigRequest
	if err := c.parseRequest(&req); err != nil {
		utils.ErrorResponse(c.Ctx, 1002, "参数解析失败: "+err.Error(), nil)
		return
	}

	if req.ConfigKey == "" {
		utils.ErrorResponse(c.Ctx, 1002, "configKey参数不能为空", nil)
		return
	}

	configKey := req.ConfigKey

	// 删除配置
	err := models.DeleteConfig(appId, configKey)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "删除配置失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "删除成功", nil)
}
