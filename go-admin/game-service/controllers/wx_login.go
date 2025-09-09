package controllers

import (
	"encoding/json"
	"game-service/models"
	"game-service/utils"
	"time"
)

// WxLoginRequest 微信登录请求结构
type WxLoginRequest struct {
	Code     string `json:"code"`     // 微信登录授权码
	AppId    string `json:"appId"`    // 应用ID
	Nickname string `json:"nickname"` // 用户昵称（可选）
	Avatar   string `json:"avatar"`   // 用户头像（可选）
}

// WxUserInfo 微信用户信息结构
type WxUserInfo struct {
	OpenId    string `json:"openId"`
	UnionId   string `json:"unionId"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Gender    int    `json:"gender"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	Language  string `json:"language"`
}

// WxLogin 微信登录方法（对齐zy-sdk/user.ts）
func (c *UserDataController) WxLogin() {
	var req WxLoginRequest
	
	// 解析请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(c.Ctx, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 参数校验
	if req.Code == "" {
		utils.ErrorResponse(c.Ctx, 4001, "微信授权码不能为空", nil)
		return
	}

	if req.AppId == "" {
		utils.ErrorResponse(c.Ctx, 4001, "应用ID不能为空", nil)
		return
	}

	// 模拟微信登录验证（实际项目中需要调用微信API验证code）
	// 这里为了演示，生成一个模拟的openId
	openId := utils.GenerateOpenID(req.Code, req.AppId)
	
	// 构造微信用户信息
	wxUserInfo := WxUserInfo{
		OpenId:   openId,
		UnionId:  "",
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   0,
		City:     "",
		Province: "",
		Country:  "",
		Language: "zh_CN",
	}

	// 生成用户ID（基于openId）
	userId := utils.GenerateUserIDFromOpenID(openId, req.AppId)

	// 尝试获取用户数据，如果不存在则创建新用户
	userData, err := models.GetUserDataWithKey(req.AppId, userId, "userInfo")
	if err != nil {
		utils.ErrorResponse(c.Ctx, 5001, "获取用户数据失败: "+err.Error(), nil)
		return
	}

	var userInfo map[string]interface{}
	isNewUser := false
	
	if userData == "" {
		// 新用户，创建用户信息
		isNewUser = true
		userInfo = map[string]interface{}{
			"userId":     userId,
			"openId":     openId,
			"nickname":   wxUserInfo.Nickname,
			"avatar":     wxUserInfo.Avatar,
			"appId":      req.AppId,
			"loginType":  "wechat",
			"loginCount": 1,
			"lastLogin":  time.Now().Format("2006-01-02 15:04:05"),
			"createdAt":  time.Now().Format("2006-01-02 15:04:05"),
		}
		
		// 保存用户信息
		userInfoBytes, _ := json.Marshal(userInfo)
		err = models.SaveUserDataWithKey(req.AppId, userId, "userInfo", string(userInfoBytes))
		if err != nil {
			utils.ErrorResponse(c.Ctx, 5001, "创建用户失败: "+err.Error(), nil)
			return
		}

		// 保存微信用户详细信息
		wxInfoBytes, _ := json.Marshal(wxUserInfo)
		err = models.SaveUserDataWithKey(req.AppId, userId, "wxUserInfo", string(wxInfoBytes))
		if err != nil {
			utils.ErrorResponse(c.Ctx, 5001, "保存微信用户信息失败: "+err.Error(), nil)
			return
		}
	} else {
		// 现有用户，更新登录信息
		err = json.Unmarshal([]byte(userData), &userInfo)
		if err != nil {
			utils.ErrorResponse(c.Ctx, 5001, "用户数据解析失败: "+err.Error(), nil)
			return
		}
		
		// 更新登录信息和用户资料
		loginCount := 1
		if count, ok := userInfo["loginCount"].(float64); ok {
			loginCount = int(count) + 1
		}
		
		userInfo["loginCount"] = loginCount
		userInfo["lastLogin"] = time.Now().Format("2006-01-02 15:04:05")
		
		// 更新昵称和头像（如果提供）
		if req.Nickname != "" {
			userInfo["nickname"] = req.Nickname
			wxUserInfo.Nickname = req.Nickname
		}
		if req.Avatar != "" {
			userInfo["avatar"] = req.Avatar
			wxUserInfo.Avatar = req.Avatar
		}
		
		// 保存更新后的用户信息
		userInfoBytes, _ := json.Marshal(userInfo)
		err = models.SaveUserDataWithKey(req.AppId, userId, "userInfo", string(userInfoBytes))
		if err != nil {
			utils.ErrorResponse(c.Ctx, 5001, "更新用户信息失败: "+err.Error(), nil)
			return
		}

		// 更新微信用户信息
		wxInfoBytes, _ := json.Marshal(wxUserInfo)
		err = models.SaveUserDataWithKey(req.AppId, userId, "wxUserInfo", string(wxInfoBytes))
		if err != nil {
			utils.ErrorResponse(c.Ctx, 5001, "更新微信用户信息失败: "+err.Error(), nil)
			return
		}
	}

	// 生成会话token
	sessionToken := utils.GenerateSessionToken(userId, req.AppId)

	// 保存会话信息
	sessionInfo := map[string]interface{}{
		"userId":    userId,
		"openId":    openId,
		"appId":     req.AppId,
		"loginType": "wechat",
		"loginTime": time.Now().Format("2006-01-02 15:04:05"),
		"ip":        c.Ctx.Input.IP(),
	}
	sessionBytes, _ := json.Marshal(sessionInfo)
	models.SaveUserDataWithKey(req.AppId, userId, "session_"+sessionToken, string(sessionBytes))

	// 返回登录成功结果
	result := map[string]interface{}{
		"userId":     userId,
		"openId":     openId,
		"token":      sessionToken,
		"userInfo":   userInfo,
		"wxUserInfo": wxUserInfo,
		"isNewUser":  isNewUser,
		"loginTime":  time.Now().Format("2006-01-02 15:04:05"),
	}

	utils.SuccessResponse(c.Ctx, "微信登录成功", result)
}
