package controllers

import (
	"game-service/models"
	"game-service/utils"

	"github.com/beego/beego/v2/server/web"
)

type ConfigController struct {
	web.Controller
}

// GetConfig 获取配置
func (c *ConfigController) GetConfig() {
	// 验证签名
	appId, _, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	configKey := c.GetString("configKey")
	if configKey == "" {
		utils.ErrorResponse(c.Ctx, 1002, "configKey参数不能为空", nil)
		return
	}

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
	// 验证签名
	appId, _, err := utils.ValidateSignature(c.Ctx.Request)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1001, "签名验证失败: "+err.Error(), nil)
		return
	}

	// 获取参数
	configKey := c.GetString("configKey")
	configValue := c.GetString("configValue")
	version := c.GetString("version")
	description := c.GetString("description")

	if configKey == "" || configValue == "" {
		utils.ErrorResponse(c.Ctx, 1002, "configKey和configValue参数不能为空", nil)
		return
	}

	// 设置配置
	err = models.SetConfig(appId, configKey, configValue, version, description)
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

	// 获取参数
	version := c.GetString("version", "")

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

	// 获取参数
	configKey := c.GetString("configKey")
	if configKey == "" {
		utils.ErrorResponse(c.Ctx, 1002, "configKey参数不能为空", nil)
		return
	}

	// 删除配置
	err := models.DeleteConfig(appId, configKey)
	if err != nil {
		utils.ErrorResponse(c.Ctx, 1003, "删除配置失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c.Ctx, "删除成功", nil)
}
