package controllers

import (
	"encoding/json"
	"fmt"
	"game-service/models"
	"game-service/utils"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
)

// UserController 用户控制器 - 统一管理所有用户相关接口
type UserController struct {
	web.Controller
}

// 基础请求结构
type BaseRequest struct {
	AppId     string `json:"appId"`
	PlayerId  string `json:"playerId"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Ver       string `json:"ver"`
	Sign      string `json:"sign"`
}

// 统一响应结构
type ResponseCommon struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// =============================================================================
// 登录相关接口
// =============================================================================

// LoginRequest 普通登录请求结构
type LoginRequest struct {
	Code      string `json:"code"`      // 授权码
	AppId     string `json:"appId"`     // 应用ID
	Timestamp int64  `json:"timestamp"` // 时间戳
	Ver       string `json:"ver"`       // 版本号
	Sign      string `json:"sign"`      // 签名
}

// WxLoginRequest 微信登录请求结构
type WxLoginRequest struct {
	Code      string `json:"code"`      // 微信code
	AppId     string `json:"appId"`     // 应用ID
	Timestamp int64  `json:"timestamp"` // 时间戳
	Ver       string `json:"ver"`       // 版本号
	Sign      string `json:"sign"`      // 签名
}

// LoginData 登录响应数据结构
type LoginData struct {
	Token    string      `json:"token"`    // 会话令牌
	PlayerId string      `json:"playerId"` // 玩家ID
	IsNew    bool        `json:"isNew"`    // 是否为新用户
	OpenId   string      `json:"openId"`   // 开放平台用户ID
	UnionId  string      `json:"unionid"`  // 微信unionid（仅微信登录）
	Data     interface{} `json:"data"`     // 用户数据
}

// Login 普通登录接口
func (c *UserController) Login() {
	var req LoginRequest
	ret := ResponseCommon{
		Code:      0,
		Msg:       "success",
		Timestamp: time.Now().UnixMilli(),
	}

	// 解析请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		ret.Code = 4001
		ret.Msg = "参数解析失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 参数校验
	if req.AppId == "" {
		ret.Code = 4001
		ret.Msg = "参数[appId]错误"
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	if req.Code == "" {
		ret.Code = 4001
		ret.Msg = "参数[code]错误"
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 验证应用是否存在
	if !c.validateApp(req.AppId) {
		ret.Code = 4004
		ret.Msg = "appId不存在"
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 处理登录逻辑
	openId := req.Code // 简单登录，code就是openId
	loginData, err := c.processLogin(req.AppId, openId, "")
	if err != nil {
		ret.Code = 5001
		ret.Msg = err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	ret.Data = loginData
	c.Ctx.Output.JSON(ret, false, false)
}

// WxLogin 微信登录接口
func (c *UserController) WxLogin() {
	var req WxLoginRequest
	ret := ResponseCommon{
		Code:      0,
		Msg:       "success",
		Timestamp: time.Now().UnixMilli(),
	}

	// 解析请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		ret.Code = 4001
		ret.Msg = "参数解析失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 参数校验
	if req.AppId == "" {
		ret.Code = 4001
		ret.Msg = "参数[appId]错误"
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	if req.Code == "" {
		ret.Code = 4001
		ret.Msg = "参数[code]错误"
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 获取应用配置
	appConfig, err := c.getAppConfig(req.AppId)
	if err != nil {
		ret.Code = 4004
		ret.Msg = "appId不存在或配置错误"
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 调用微信API获取openId
	wxResp, err := c.callWxAPI(appConfig.ChannelAppId, appConfig.ChannelAppKey, req.Code)
	if err != nil {
		ret.Code = 4004
		ret.Msg = "微信登录失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 处理登录逻辑
	loginData, err := c.processLogin(req.AppId, wxResp.OpenId, wxResp.UnionId)
	if err != nil {
		ret.Code = 5001
		ret.Msg = err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	ret.Data = loginData
	c.Ctx.Output.JSON(ret, false, false)
}

// =============================================================================
// 数据相关接口
// =============================================================================

// GetData 获取用户数据接口
func (c *UserController) GetData() {
	var req BaseRequest
	ret := ResponseCommon{
		Code:      0,
		Msg:       "success",
		Timestamp: time.Now().UnixMilli(),
	}

	// 解析请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		ret.Code = 4001
		ret.Msg = "参数解析失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 从中间件获取已验证的appId和playerId
	appId := c.Ctx.Input.GetData("app_id").(string)
	playerId := req.PlayerId

	// 获取用户数据
	userData, err := models.GetUserData(appId, playerId)
	if err != nil {
		ret.Code = 5001
		ret.Msg = "获取数据失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	ret.Data = userData
	c.Ctx.Output.JSON(ret, false, false)
}

// SaveDataRequest 保存数据请求结构
type SaveDataRequest struct {
	BaseRequest
	Data interface{} `json:"data"`
}

// SaveData 保存用户数据接口
func (c *UserController) SaveData() {
	var req SaveDataRequest
	ret := ResponseCommon{
		Code:      0,
		Msg:       "success",
		Timestamp: time.Now().UnixMilli(),
	}

	// 解析请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		ret.Code = 4001
		ret.Msg = "参数解析失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)
	playerId := req.PlayerId

	// 保存数据
	var data string
	if req.Data != nil {
		data = req.Data.(string)
	}
	err := models.SaveUserData(appId, playerId, data)
	if err != nil {
		ret.Code = 5001
		ret.Msg = "保存数据失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	c.Ctx.Output.JSON(ret, false, false)
}

// SaveUserInfoRequest 保存用户信息请求结构
type SaveUserInfoRequest struct {
	BaseRequest
	UserInfo string `json:"userInfo"` // JSON字符串格式的用户信息
}

// UserInfo 用户信息结构 - 对应TypeScript接口
type UserInfo struct {
	NickName         string `json:"nickName"`
	AvatarUrl        string `json:"avatarUrl"`
	Gender           int    `json:"gender"`
	Province         string `json:"province"`
	City             string `json:"city"`
	Level            int    `json:"level"`
	Exp              int64  `json:"exp"`
	LastLoginTime    string `json:"lastLoginTime"`
	LastLogoutTime   string `json:"lastLogoutTime"`
	LastLoginIp      string `json:"lastLoginIp"`
	LastLogoutIp     string `json:"lastLogoutIp"`
	LastLoginDevice  string `json:"lastLoginDevice"`
	LastLogoutDevice string `json:"lastLogoutDevice"`
}

// SaveUserInfo 保存用户信息接口
func (c *UserController) SaveUserInfo() {
	var req SaveUserInfoRequest
	ret := ResponseCommon{
		Code:      0,
		Msg:       "success",
		Timestamp: time.Now().UnixMilli(),
	}

	// 解析请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		ret.Code = 4001
		ret.Msg = "参数解析失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 从中间件获取已验证的appId
	appId := c.Ctx.Input.GetData("app_id").(string)
	playerId := req.PlayerId

	// 解析用户信息
	var userInfo UserInfo
	if err := json.Unmarshal([]byte(req.UserInfo), &userInfo); err != nil {
		ret.Code = 4001
		ret.Msg = "用户信息格式错误: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 获取现有用户
	user, err := models.GetUserByPlayerId(appId, playerId)
	if err != nil {
		ret.Code = 5001
		ret.Msg = "获取用户失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	if user == nil {
		ret.Code = 4004
		ret.Msg = "用户不存在"
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	// 更新用户信息
	if userInfo.NickName != "" {
		user.Nickname = userInfo.NickName
	}
	if userInfo.AvatarUrl != "" {
		user.Avatar = userInfo.AvatarUrl
	}
	if userInfo.Level > 0 {
		user.Level = userInfo.Level
	}
	if userInfo.Exp > 0 {
		user.Exp = userInfo.Exp
	}

	// 保存更新
	err = models.UpdateUser(appId, user)
	if err != nil {
		ret.Code = 5001
		ret.Msg = "更新用户信息失败: " + err.Error()
		c.Ctx.Output.JSON(ret, false, false)
		return
	}

	c.Ctx.Output.JSON(ret, false, false)
}

// =============================================================================
// 辅助方法
// =============================================================================

// validateApp 验证应用是否存在
func (c *UserController) validateApp(appId string) bool {
	application := &models.Application{}
	err := application.GetByAppId(appId)
	return err == nil && application.Status == "active"
}

// processLogin 处理登录逻辑
func (c *UserController) processLogin(appId, openId, unionId string) (*LoginData, error) {
	// 查询现有用户
	existingUser, err := models.GetUserByOpenId(appId, openId)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 生成token
	token := utils.GenerateUUID()
	loginData := &LoginData{
		Token:   token,
		OpenId:  openId,
		UnionId: unionId,
	}

	if existingUser == nil {
		// 新用户
		playerId := utils.GeneratePlayerId()

		newUser := &models.User{
			OpenId:        openId,
			PlayerId:      playerId,
			Token:         token,
			Data:          "{}",
			Level:         1,
			LoginCount:    1,
			LastLoginTime: time.Now(),
			RegisterTime:  time.Now(),
		}

		err = models.CreateUser(appId, newUser)
		if err != nil {
			return nil, fmt.Errorf("创建用户失败: %v", err)
		}

		loginData.PlayerId = playerId
		loginData.IsNew = true
		loginData.Data = map[string]interface{}{}

		logs.Info("新用户登录成功: appId=%s, openId=%s, playerId=%s", appId, openId, playerId)
	} else {
		// 现有用户
		existingUser.Token = token
		existingUser.LoginCount++
		existingUser.LastLoginTime = time.Now()

		err = models.UpdateUser(appId, existingUser)
		if err != nil {
			return nil, fmt.Errorf("更新用户失败: %v", err)
		}

		// 解析用户数据
		var userData map[string]interface{}
		if existingUser.Data != "" && existingUser.Data != "{}" {
			err = json.Unmarshal([]byte(existingUser.Data), &userData)
			if err != nil {
				userData = map[string]interface{}{}
			}
		} else {
			userData = map[string]interface{}{}
		}

		loginData.PlayerId = existingUser.PlayerId
		loginData.IsNew = false
		loginData.Data = userData

		logs.Info("用户登录成功: appId=%s, openId=%s, playerId=%s", appId, openId, existingUser.PlayerId)
	}

	return loginData, nil
}

// AppConfig 应用配置结构
type AppConfig struct {
	ChannelAppId  string
	ChannelAppKey string
}

// getAppConfig 获取应用配置
func (c *UserController) getAppConfig(appId string) (*AppConfig, error) {
	application := &models.Application{}
	err := application.GetByAppId(appId)
	if err != nil {
		return nil, fmt.Errorf("应用不存在: %v", err)
	}

	if application.Status != "active" {
		return nil, fmt.Errorf("应用状态异常: %s", application.Status)
	}

	if application.ChannelAppId == "" || application.ChannelAppKey == "" {
		return nil, fmt.Errorf("应用微信配置不完整")
	}

	return &AppConfig{
		ChannelAppId:  application.ChannelAppId,
		ChannelAppKey: application.ChannelAppKey,
	}, nil
}

// WxAPIResponse 微信API响应结构
type WxAPIResponse struct {
	OpenId     string `json:"openid"`
	UnionId    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// callWxAPI 调用微信API
func (c *UserController) callWxAPI(appId, appSecret, code string) (*WxAPIResponse, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appId, appSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求微信API失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取微信API响应失败: %v", err)
	}

	var wxResp WxAPIResponse
	err = json.Unmarshal(body, &wxResp)
	if err != nil {
		return nil, fmt.Errorf("解析微信API响应失败: %v", err)
	}

	if wxResp.ErrCode != 0 {
		return nil, fmt.Errorf("微信API错误: %s", wxResp.ErrMsg)
	}

	return &wxResp, nil
}
