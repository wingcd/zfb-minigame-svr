package login

import (
	"fmt"
	"game-service/models"
	"game-service/utils"
	"game-service/yalla/services"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// YallaLoginController Yalla登录控制器
type YallaLoginController struct {
	BaseLoginController
}

// YallaLoginRequest Yalla登录请求结构
type YallaLoginRequest struct {
	AppId     string `json:"appId"`     // 应用ID
	SdkUserId string `json:"sdkUserId"` // Yalla SDK用户ID (作为openId使用)
	Timestamp int64  `json:"timestamp"` // 时间戳
	Ver       string `json:"ver"`       // 版本号
	Sign      string `json:"sign"`      // 签名
}

// YallaLogin Yalla登录接口
func (c *YallaLoginController) YallaLogin() {
	var req YallaLoginRequest

	// 解析请求参数
	if err := c.parseRequest(&req); err != nil {
		ret := c.createErrorResponse(4001, "参数解析失败: "+err.Error())
		c.sendResponse(ret)
		return
	}

	// 验证基础参数
	if req.AppId == "" {
		ret := c.createErrorResponse(4001, "appId不能为空")
		c.sendResponse(ret)
		return
	}

	if req.SdkUserId == "" {
		ret := c.createErrorResponse(4001, "sdkUserId不能为空")
		c.sendResponse(ret)
		return
	}

	// 验证应用是否存在
	if !c.validateApp(req.AppId) {
		ret := c.createErrorResponse(4004, "appId不存在或已禁用")
		c.sendResponse(ret)
		return
	}

	// 验证Yalla SDK用户ID (可选，如果需要与Yalla服务器验证)
	if err := c.validateYallaUser(req.AppId, req.SdkUserId); err != nil {
		ret := c.createErrorResponse(4005, "Yalla用户验证失败: "+err.Error())
		c.sendResponse(ret)
		return
	}

	// 处理登录逻辑，将sdkUserId作为openId处理
	loginData, err := c.processYallaLogin(req.AppId, req.SdkUserId)
	if err != nil {
		ret := c.createErrorResponse(5001, err.Error())
		c.sendResponse(ret)
		return
	}

	// 保存用户token到redis
	if err := models.SaveUserStatusToRedis(req.AppId, loginData.PlayerId, loginData.Token); err != nil {
		logs.Warning("保存用户token到redis失败:", err)
	}

	ret := c.createSuccessResponse(loginData)
	c.sendResponse(ret)
}

// validateApp 验证应用是否存在且有效
func (c *YallaLoginController) validateApp(appId string) bool {
	app := &models.Application{}
	err := app.GetByAppId(appId)
	if err != nil {
		logs.Error("获取应用配置失败:", err)
		return false
	}

	if app.Status != "active" {
		logs.Warning("应用已被禁用:", appId)
		return false
	}

	return true
}

// validateYallaUser 验证Yalla用户 (可选实现)
func (c *YallaLoginController) validateYallaUser(appId, sdkUserId string) error {
	// 如果配置了Yalla服务，可以进行用户验证
	// 这里可以调用Yalla服务来验证sdkUserId的有效性

	// 获取Yalla配置
	yallaConfig, err := web.AppConfig.String("yalla::app_id")
	if err != nil || yallaConfig == "" {
		// 如果没有配置Yalla服务，跳过验证
		logs.Info("Yalla服务未配置，跳过用户验证")
		return nil
	}

	// 创建Yalla服务实例进行验证
	yallaService, err := services.NewYallaService(yallaConfig)
	if err != nil {
		logs.Warning("创建Yalla服务失败，跳过用户验证:", err)
		return nil // 不阻断登录流程
	}

	// 验证SDK用户ID (这里可以根据实际需求实现)
	_, err = yallaService.GetUserInfo(sdkUserId)
	if err != nil {
		logs.Warning("Yalla用户验证失败:", err)
		// 根据业务需求决定是否阻断登录
		// return err
		return nil // 暂时不阻断登录流程
	}

	logs.Info("Yalla用户验证成功:", sdkUserId)
	return nil
}

// processYallaLogin 处理Yalla登录逻辑
func (c *YallaLoginController) processYallaLogin(appId, sdkUserId string) (*LoginData, error) {
	// 将sdkUserId作为openId来查找或创建用户
	user, isNew, err := c.findOrCreateUser(appId, sdkUserId, "")
	if err != nil {
		return nil, fmt.Errorf("处理用户数据失败: %v", err)
	}

	// 生成token
	token := generateToken(appId, user.PlayerId)

	// 更新最后登录时间
	if err := c.updateLastLoginTime(user.PlayerId); err != nil {
		logs.Warning("更新最后登录时间失败:", err)
	}

	return &LoginData{
		Token:    token,
		PlayerId: user.PlayerId,
		IsNew:    isNew,
		OpenId:   sdkUserId, // 将sdkUserId作为openId返回
		UnionId:  "",        // Yalla可能没有unionId概念
		Data:     user.Data,
	}, nil
}

// findOrCreateUser 查找或创建用户
func (c *YallaLoginController) findOrCreateUser(appId, openId, unionId string) (*models.User, bool, error) {
	// 尝试根据openId查找用户
	user, err := models.GetUserByOpenId(appId, openId)
	if err == nil && user != nil {
		// 用户已存在
		return user, false, nil
	}

	// 用户不存在，创建新用户
	playerId := utils.GeneratePlayerId()

	newUser := &models.User{
		AppId:    appId,
		PlayerId: playerId,
		OpenId:   openId,
		Data:     "{}",
	}

	if err := models.CreateUser(appId, newUser); err != nil {
		return nil, false, fmt.Errorf("创建用户失败: %v", err)
	}

	logs.Info("创建新Yalla用户成功, appId:", appId, "playerId:", playerId, "sdkUserId:", openId)
	return newUser, true, nil
}

// updateLastLoginTime 更新最后登录时间
func (c *YallaLoginController) updateLastLoginTime(playerId string) error {
	// 这里可以添加更新最后登录时间的逻辑
	// 暂时返回nil，后续可以扩展
	return nil
}
