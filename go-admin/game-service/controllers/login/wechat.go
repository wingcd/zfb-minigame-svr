package login

import (
	"encoding/json"
	"fmt"
	"game-service/models"
	"game-service/utils"
	"io/ioutil"
	"net/http"

	"github.com/beego/beego/v2/core/logs"
)

// WechatLoginController 微信登录控制器
type WechatLoginController struct {
	BaseLoginController
}

// WxLoginRequest 微信登录请求结构
type WxLoginRequest struct {
	Code      string `json:"code"`      // 微信code
	AppId     string `json:"appId"`     // 应用ID
	Timestamp int64  `json:"timestamp"` // 时间戳
	Ver       string `json:"ver"`       // 版本号
	Sign      string `json:"sign"`      // 签名
}

// WxAPIResponse 微信API响应结构
type WxAPIResponse struct {
	SessionKey string `json:"session_key"`
	OpenId     string `json:"openid"`
	UnionId    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WxLogin 微信登录接口
func (c *WechatLoginController) WxLogin() {
	var req WxLoginRequest

	// 解析请求参数
	if err := c.parseRequest(&req); err != nil {
		ret := c.createErrorResponse(4001, "参数解析失败: "+err.Error())
		c.sendResponse(ret)
		return
	}

	// 验证基础参数
	if errResp := c.validateBasicParams(req.AppId, req.Code); errResp != nil {
		c.sendResponse(*errResp)
		return
	}

	// 获取应用配置
	appConfig, err := c.getAppConfig(req.AppId)
	if err != nil {
		ret := c.createErrorResponse(4004, "appId不存在或配置错误")
		c.sendResponse(ret)
		return
	}

	// 调用微信API获取openId
	wxResp, err := c.callWxAPI(appConfig.ChannelAppId, appConfig.ChannelAppKey, req.Code)
	if err != nil {
		ret := c.createErrorResponse(4004, "微信登录失败: "+err.Error())
		c.sendResponse(ret)
		return
	}

	// 处理登录逻辑
	loginData, err := c.processLogin(req.AppId, wxResp.OpenId, wxResp.UnionId)
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

// getAppConfig 获取应用配置
func (c *WechatLoginController) getAppConfig(appId string) (*models.Application, error) {
	app := &models.Application{}
	err := app.GetByAppId(appId)
	if err != nil {
		logs.Error("获取应用配置失败:", err)
		return nil, err
	}
	return app, nil
}

// callWxAPI 调用微信API获取用户信息
func (c *WechatLoginController) callWxAPI(appId, appSecret, code string) (*WxAPIResponse, error) {
	// 构建微信API URL
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appId, appSecret, code)

	// 发起HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		logs.Error("调用微信API失败:", err)
		return nil, fmt.Errorf("网络请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("读取微信API响应失败:", err)
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应JSON
	var wxResp WxAPIResponse
	if err := json.Unmarshal(body, &wxResp); err != nil {
		logs.Error("解析微信API响应失败:", err, "响应内容:", string(body))
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查微信API错误
	if wxResp.ErrCode != 0 {
		logs.Error("微信API返回错误:", wxResp.ErrCode, wxResp.ErrMsg)
		return nil, fmt.Errorf("微信API错误: %s (code: %d)", wxResp.ErrMsg, wxResp.ErrCode)
	}

	// 检查必要字段
	if wxResp.OpenId == "" {
		logs.Error("微信API未返回openid")
		return nil, fmt.Errorf("微信API未返回有效的openid")
	}

	logs.Info("微信登录成功, openId:", wxResp.OpenId, "unionId:", wxResp.UnionId)
	return &wxResp, nil
}

// processLogin 处理登录逻辑 (复用CommonLoginController的逻辑)
func (c *WechatLoginController) processLogin(appId, openId, unionId string) (*LoginData, error) {
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

	return &LoginData{
		Token:    token,
		PlayerId: user.PlayerId,
		IsNew:    isNew,
		OpenId:   openId,
		UnionId:  unionId,
		Data:     user.Data,
	}, nil
}

// findOrCreateUser 查找或创建用户
func (c *WechatLoginController) findOrCreateUser(appId, openId, unionId string) (*models.User, bool, error) {
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
func (c *WechatLoginController) updateLastLoginTime(playerId string) error {
	// 这里可以添加更新最后登录时间的逻辑
	// 暂时返回nil，后续可以扩展
	return nil
}
