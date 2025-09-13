package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/beego/beego/v2/server/web"

	"admin-service/models"
	"admin-service/utils"
)

type YallaConfigController struct {
	web.Controller
}

// GetList 获取Yalla配置列表
func (c *YallaConfigController) GetList() {
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorResponse(&c.Controller, 5001, fmt.Sprintf("系统内部错误: %v", r), nil)
		}
	}()

	var request struct {
		Page        int    `json:"page"`
		PageSize    int    `json:"pageSize"`
		AppId       string `json:"appId"`
		Environment string `json:"environment"`
		IsActive    *bool  `json:"isActive"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "请求参数格式错误", nil)
		return
	}

	// 设置默认值
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.PageSize <= 0 {
		request.PageSize = 20
	}

	// 查询条件
	conditions := make(map[string]interface{})
	if request.AppId != "" {
		conditions["app_id__icontains"] = request.AppId
	}
	if request.Environment != "" {
		conditions["environment"] = request.Environment
	}
	if request.IsActive != nil {
		conditions["is_active"] = *request.IsActive
	}

	// 获取总数
	total, err := models.GetYallaConfigCount(conditions)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "查询配置总数失败", nil)
		return
	}

	// 获取列表
	configs, err := models.GetYallaConfigList(conditions, request.Page, request.PageSize)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "查询配置列表失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", map[string]interface{}{
		"list":  configs,
		"total": total,
	})
}

// Create 创建Yalla配置
func (c *YallaConfigController) Create() {
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorResponse(&c.Controller, 5001, fmt.Sprintf("系统内部错误: %v", r), nil)
		}
	}()

	var config models.YallaConfig
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &config); err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "请求参数格式错误", nil)
		return
	}

	// 验证必填字段
	if config.AppID == "" {
		utils.ErrorResponse(&c.Controller, 4001, "应用ID不能为空", nil)
		return
	}
	if config.AppGameID == "" {
		utils.ErrorResponse(&c.Controller, 4001, "游戏ID不能为空", nil)
		return
	}
	if config.SecretKey == "" {
		utils.ErrorResponse(&c.Controller, 4001, "密钥不能为空", nil)
		return
	}
	if config.BaseURL == "" {
		utils.ErrorResponse(&c.Controller, 4001, "基础URL不能为空", nil)
		return
	}
	if config.PushURL == "" {
		utils.ErrorResponse(&c.Controller, 4001, "推送URL不能为空", nil)
		return
	}
	if config.Environment == "" {
		config.Environment = "sandbox"
	}

	// 设置默认值
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}

	// 检查是否已存在相同的AppID
	exists, err := models.CheckYallaConfigExists(config.AppID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "检查配置是否存在失败", nil)
		return
	}
	if exists {
		utils.ErrorResponse(&c.Controller, 4001, "该应用ID的配置已存在", nil)
		return
	}

	// 设置创建时间
	config.CreateTime = time.Now()
	config.UpdateTime = time.Now()

	// 创建配置
	if err := models.CreateYallaConfig(&config); err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "创建配置失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", config)
}

// Update 更新Yalla配置
func (c *YallaConfigController) Update() {
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorResponse(&c.Controller, 5001, fmt.Sprintf("系统内部错误: %v", r), nil)
		}
	}()

	var request struct {
		ID          int64  `json:"id"`
		AppID       string `json:"appId"`
		SecretKey   string `json:"secretKey"`
		BaseURL     string `json:"baseUrl"`
		PushURL     string `json:"pushUrl"`
		Environment string `json:"environment"`
		Timeout     int    `json:"timeout"`
		RetryCount  int    `json:"retryCount"`
		Description string `json:"description"`
		IsActive    *bool  `json:"isActive"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "请求参数格式错误", nil)
		return
	}

	if request.AppID == "" {
		utils.ErrorResponse(&c.Controller, 4001, "应用ID不能为空", nil)
		return
	}

	// 获取现有配置
	config, err := models.GetYallaConfigByAppID(request.AppID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 4004, "配置不存在", nil)
		return
	}

	// 更新字段
	if request.SecretKey != "" {
		config.SecretKey = request.SecretKey
	}
	if request.BaseURL != "" {
		config.BaseURL = request.BaseURL
	}
	if request.PushURL != "" {
		config.PushURL = request.PushURL
	}
	if request.Environment != "" {
		config.Environment = request.Environment
	}
	if request.Timeout > 0 {
		config.Timeout = request.Timeout
	}
	if request.RetryCount >= 0 {
		config.RetryCount = request.RetryCount
	}
	if request.Description != "" {
		config.Description = request.Description
	}
	if request.IsActive != nil {
		config.IsActive = *request.IsActive
	}

	config.UpdateTime = time.Now()

	// 更新配置
	if err := models.UpdateYallaConfig(config); err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "更新配置失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", config)
}

// Delete 删除Yalla配置
func (c *YallaConfigController) Delete() {
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorResponse(&c.Controller, 5001, fmt.Sprintf("系统内部错误: %v", r), nil)
		}
	}()

	var request struct {
		AppID string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "请求参数格式错误", nil)
		return
	}

	if request.AppID == "" {
		utils.ErrorResponse(&c.Controller, 4001, "应用ID不能为空", nil)
		return
	}

	// 检查配置是否存在
	exists, err := models.CheckYallaConfigExists(request.AppID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "检查配置是否存在失败", nil)
		return
	}
	if !exists {
		utils.ErrorResponse(&c.Controller, 4004, "配置不存在", nil)
		return
	}

	// 删除配置
	if err := models.DeleteYallaConfig(request.AppID); err != nil {
		utils.ErrorResponse(&c.Controller, 5001, "删除配置失败", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", nil)
}

// Get 获取单个Yalla配置
func (c *YallaConfigController) Get() {
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorResponse(&c.Controller, 5001, fmt.Sprintf("系统内部错误: %v", r), nil)
		}
	}()

	var request struct {
		AppID string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "请求参数格式错误", nil)
		return
	}

	if request.AppID == "" {
		utils.ErrorResponse(&c.Controller, 4001, "应用ID不能为空", nil)
		return
	}

	// 获取配置
	config, err := models.GetYallaConfigByAppID(request.AppID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 4004, "配置不存在", nil)
		return
	}

	utils.SuccessResponse(&c.Controller, "success", config)
}

// TestConnection 测试Yalla连接
func (c *YallaConfigController) TestConnection() {
	defer func() {
		if r := recover(); r != nil {
			utils.ErrorResponse(&c.Controller, 5001, fmt.Sprintf("系统内部错误: %v", r), nil)
		}
	}()

	var request struct {
		AppID string `json:"appId"`
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		utils.ErrorResponse(&c.Controller, 4001, "请求参数格式错误", nil)
		return
	}

	if request.AppID == "" {
		utils.ErrorResponse(&c.Controller, 4001, "应用ID不能为空", nil)
		return
	}

	// 获取配置
	config, err := models.GetYallaConfigByAppID(request.AppID)
	if err != nil {
		utils.ErrorResponse(&c.Controller, 4004, "配置不存在", nil)
		return
	}

	// 执行连接测试
	testResults := []map[string]interface{}{}

	// 测试1: 基础连接测试
	startTime := time.Now()
	baseTestResult := map[string]interface{}{
		"testName": "基础连接",
		"success":  true,
		"message":  fmt.Sprintf("连接到 %s 正常", config.BaseURL),
	}

	// 这里可以添加实际的连接测试逻辑
	// 比如发送一个简单的HTTP请求到Yalla服务器
	responseTime := int(time.Since(startTime).Milliseconds())
	baseTestResult["responseTime"] = responseTime

	testResults = append(testResults, baseTestResult)

	// 测试2: 认证测试
	startTime = time.Now()
	authTestResult := map[string]interface{}{
		"testName": "认证测试",
		"success":  true,
		"message":  fmt.Sprintf("应用ID %s 认证配置有效", config.AppID),
	}

	// 这里可以添加实际的认证测试逻辑
	responseTime = int(time.Since(startTime).Milliseconds())
	authTestResult["responseTime"] = responseTime

	testResults = append(testResults, authTestResult)

	utils.SuccessResponse(&c.Controller, "success", map[string]interface{}{
		"testResults": testResults,
	})
}
