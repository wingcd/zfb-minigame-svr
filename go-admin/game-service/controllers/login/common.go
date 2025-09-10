package login

import (
	"fmt"
	"game-service/models"
	"game-service/utils"

	"github.com/beego/beego/v2/core/logs"
)

// CommonLoginController 通用登录控制器
type CommonLoginController struct {
	BaseLoginController
}

// CommonLoginRequest 通用登录请求结构
type CommonLoginRequest struct {
	AppId     string `json:"appId"`     // 应用ID
	OpenId    string `json:"openId"`    // 用户唯一标识
	UnionId   string `json:"unionId"`   // 用户在开放平台的唯一标识（暂不使用）
	Timestamp int64  `json:"timestamp"` // 时间戳
	Ver       string `json:"ver"`       // 版本号
	Sign      string `json:"sign"`      // 签名
	Code      string `json:"code"`      // 授权码
}

// CommonLogin 通用登录接口
func (c *CommonLoginController) CommonLogin() {
	var req CommonLoginRequest

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

	if req.Code == "" {
		ret := c.createErrorResponse(4001, "code不能为空")
		c.sendResponse(ret)
		return
	}

	// 模拟openId
	req.OpenId = req.Code

	// 处理登录逻辑
	loginData, err := c.processCommonLogin(req.AppId, req.OpenId, req.UnionId)
	if err != nil {
		ret := c.createErrorResponse(5001, err.Error())
		c.sendResponse(ret)
		return
	}

	ret := c.createSuccessResponse(loginData)
	c.sendResponse(ret)
}

// processCommonLogin 处理通用登录逻辑
func (c *CommonLoginController) processCommonLogin(appId, openId, unionId string) (*LoginData, error) {
	// 查找或创建用户
	user, isNew, err := c.findOrCreateUser(appId, openId, unionId)
	if err != nil {
		return nil, fmt.Errorf("处理用户数据失败: %v", err)
	}

	// 生成token
	token := generateToken(appId, user.PlayerId)

	// 更新最后登录时间
	if err := c.updateLastLoginTime(user.PlayerId); err != nil {
		logs.Warning("更新最后登录时间失败:", err)
	}

	// 保存用户token到redis
	if err := models.SaveUserStatusToRedis(appId, user.PlayerId, token); err != nil {
		logs.Warning("保存用户token到redis失败:", err)
	}

	return &LoginData{
		Token:    token,
		PlayerId: user.PlayerId,
		IsNew:    isNew,
		OpenId:   openId,
		UnionId:  unionId,
	}, nil
}

// findOrCreateUser 查找或创建用户
func (c *CommonLoginController) findOrCreateUser(appId, openId, unionId string) (*models.User, bool, error) {
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

	logs.Info("创建新用户成功, appId:", appId, "playerId:", playerId, "openId:", openId)
	return newUser, true, nil
}

// updateLastLoginTime 更新最后登录时间
func (c *CommonLoginController) updateLastLoginTime(playerId string) error {
	// 这里可以添加更新最后登录时间的逻辑
	// 暂时返回nil，后续可以扩展
	return nil
}
