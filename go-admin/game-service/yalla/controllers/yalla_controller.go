package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/beego/beego/v2/server/web"

	"game-service/yalla/models"
	"game-service/yalla/services"
)

// YallaController Yalla控制器
type YallaController struct {
	web.Controller
}

// BaseResponse 基础响应结构
type BaseResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func (c *YallaController) SuccessResponse(data interface{}) {
	c.Data["json"] = BaseResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	}
	c.ServeJSON()
}

// ErrorResponse 错误响应
func (c *YallaController) ErrorResponse(code int, message string) {
	c.Data["json"] = BaseResponse{
		Code:    code,
		Message: message,
	}
	c.ServeJSON()
}

// Auth 用户认证
func (c *YallaController) Auth() {
	appID := c.GetString("app_id")
	userID := c.GetString("user_id")
	authToken := c.GetString("auth_token")

	if appID == "" || userID == "" || authToken == "" {
		c.ErrorResponse(400, "missing required parameters")
		return
	}

	// 创建Yalla服务
	yallaService, err := services.NewYallaService(appID)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("create yalla service failed: %v", err))
		return
	}

	// 用户认证
	userInfo, err := yallaService.AuthenticateUser(userID, authToken)
	if err != nil {
		c.ErrorResponse(401, fmt.Sprintf("authentication failed: %v", err))
		return
	}

	c.SuccessResponse(userInfo)
}

// GetUserInfo 获取用户信息
func (c *YallaController) GetUserInfo() {
	appID := c.GetString("app_id")
	yallaUserID := c.GetString("yalla_user_id")

	if appID == "" || yallaUserID == "" {
		c.ErrorResponse(400, "missing required parameters")
		return
	}

	// 创建Yalla服务
	yallaService, err := services.NewYallaService(appID)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("create yalla service failed: %v", err))
		return
	}

	// 获取用户信息
	userInfo, err := yallaService.GetUserInfo(yallaUserID)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("get user info failed: %v", err))
		return
	}

	c.SuccessResponse(userInfo)
}

// SendReward 发放奖励
func (c *YallaController) SendReward() {
	appID := c.GetString("app_id")
	yallaUserID := c.GetString("yalla_user_id")
	rewardType := c.GetString("reward_type")
	rewardAmountStr := c.GetString("reward_amount")
	description := c.GetString("description")

	if appID == "" || yallaUserID == "" || rewardType == "" || rewardAmountStr == "" {
		c.ErrorResponse(400, "missing required parameters")
		return
	}

	rewardAmount, err := strconv.ParseInt(rewardAmountStr, 10, 64)
	if err != nil {
		c.ErrorResponse(400, "invalid reward_amount")
		return
	}

	// 解析奖励数据
	var rewardData map[string]interface{}
	rewardDataStr := c.GetString("reward_data")
	if rewardDataStr != "" {
		err = json.Unmarshal([]byte(rewardDataStr), &rewardData)
		if err != nil {
			c.ErrorResponse(400, "invalid reward_data format")
			return
		}
	}

	// 创建Yalla服务
	yallaService, err := services.NewYallaService(appID)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("create yalla service failed: %v", err))
		return
	}

	// 发放奖励
	result, err := yallaService.SendReward(yallaUserID, rewardType, rewardAmount, rewardData, description)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("send reward failed: %v", err))
		return
	}

	c.SuccessResponse(result)
}

// SyncGameData 同步游戏数据
func (c *YallaController) SyncGameData() {
	appID := c.GetString("app_id")
	yallaUserID := c.GetString("yalla_user_id")
	dataType := c.GetString("data_type")
	gameDataStr := c.GetString("game_data")

	if appID == "" || yallaUserID == "" || dataType == "" || gameDataStr == "" {
		c.ErrorResponse(400, "missing required parameters")
		return
	}

	// 解析游戏数据
	var gameData map[string]interface{}
	err := json.Unmarshal([]byte(gameDataStr), &gameData)
	if err != nil {
		c.ErrorResponse(400, "invalid game_data format")
		return
	}

	// 创建Yalla服务
	yallaService, err := services.NewYallaService(appID)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("create yalla service failed: %v", err))
		return
	}

	// 同步游戏数据
	result, err := yallaService.SyncGameData(yallaUserID, dataType, gameData)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("sync game data failed: %v", err))
		return
	}

	c.SuccessResponse(result)
}

// ReportEvent 上报事件
func (c *YallaController) ReportEvent() {
	appID := c.GetString("app_id")
	yallaUserID := c.GetString("yalla_user_id")
	eventType := c.GetString("event_type")
	eventDataStr := c.GetString("event_data")

	if appID == "" || yallaUserID == "" || eventType == "" {
		c.ErrorResponse(400, "missing required parameters")
		return
	}

	// 解析事件数据
	var eventData map[string]interface{}
	if eventDataStr != "" {
		err := json.Unmarshal([]byte(eventDataStr), &eventData)
		if err != nil {
			c.ErrorResponse(400, "invalid event_data format")
			return
		}
	}

	// 创建Yalla服务
	yallaService, err := services.NewYallaService(appID)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("create yalla service failed: %v", err))
		return
	}

	// 上报事件
	result, err := yallaService.ReportEvent(yallaUserID, eventType, eventData)
	if err != nil {
		c.ErrorResponse(500, fmt.Sprintf("report event failed: %v", err))
		return
	}

	c.SuccessResponse(result)
}

// GetUserBinding 获取用户绑定信息
func (c *YallaController) GetUserBinding() {
	appID := c.GetString("app_id")
	gameUserID := c.GetString("game_user_id")

	if appID == "" || gameUserID == "" {
		c.ErrorResponse(400, "missing required parameters")
		return
	}

	// 获取用户绑定信息
	binding, err := services.GetUserBinding(appID, gameUserID)
	if err != nil {
		c.ErrorResponse(404, fmt.Sprintf("user binding not found: %v", err))
		return
	}

	c.SuccessResponse(binding)
}

// Config 配置管理

// GetConfig 获取配置
func (c *YallaController) GetConfig() {
	appID := c.GetString("app_id")

	if appID == "" {
		c.ErrorResponse(400, "missing app_id parameter")
		return
	}

	config, err := services.GetYallaConfig(appID)
	if err != nil {
		c.ErrorResponse(404, fmt.Sprintf("config not found: %v", err))
		return
	}

	// 隐藏敏感信息
	config.SecretKey = "***"

	c.SuccessResponse(config)
}

// UpdateConfig 更新配置
func (c *YallaController) UpdateConfig() {
	var config models.YallaConfig
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &config)
	if err != nil {
		c.ErrorResponse(400, "invalid request body")
		return
	}

	if config.AppID == "" {
		c.ErrorResponse(400, "missing app_id")
		return
	}

	// TODO: 实现配置更新逻辑
	c.ErrorResponse(501, "update config not implemented")
}

// GetCallLogs 获取调用日志
func (c *YallaController) GetCallLogs() {
	appID := c.GetString("app_id")
	pageStr := c.GetString("page", "1")
	limitStr := c.GetString("limit", "20")

	if appID == "" {
		c.ErrorResponse(400, "missing app_id parameter")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// TODO: 实现获取调用日志逻辑
	c.ErrorResponse(501, "get call logs not implemented")
}
