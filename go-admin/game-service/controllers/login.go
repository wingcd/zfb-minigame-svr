package controllers

import (
	"encoding/json"
	"game-service/models"
	"game-service/utils"
	"time"
)

// LoginRequest 用户登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	AppId    string `json:"appId"`
}

// LoginResponse 用户登录响应结构
type LoginResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// Login 用户登录方法（对齐zy-sdk/user.ts）
func (c *UserDataController) Login() {
	var req LoginRequest

	// 解析请求参数
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		utils.ErrorResponse(c.Ctx, 4001, "参数解析失败: "+err.Error(), nil)
		return
	}

	// 参数校验
	if req.Username == "" {
		utils.ErrorResponse(c.Ctx, 4001, "用户名不能为空", nil)
		return
	}

	if req.Password == "" {
		utils.ErrorResponse(c.Ctx, 4001, "密码不能为空", nil)
		return
	}

	if req.AppId == "" {
		utils.ErrorResponse(c.Ctx, 4001, "应用ID不能为空", nil)
		return
	}

	// 生成用户ID（基于用户名和应用ID）
	userId := utils.GenerateUserID(req.Username, req.AppId)

	// 验证用户密码（简单实现，实际项目中应该有用户表）
	// 这里为了兼容游戏服务的轻量级特性，暂时使用简单的密码验证
	passwordHash := utils.HashPassword(req.Password)

	// 尝试获取用户数据，如果不存在则创建新用户
	userData, err := models.GetUserDataWithKey(req.AppId, userId, "userInfo")
	if err != nil {
		utils.ErrorResponse(c.Ctx, 5001, "获取用户数据失败: "+err.Error(), nil)
		return
	}

	var userInfo map[string]interface{}

	if userData == "" {
		// 新用户，创建用户信息
		userInfo = map[string]interface{}{
			"userId":     userId,
			"username":   req.Username,
			"password":   passwordHash,
			"appId":      req.AppId,
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
	} else {
		// 现有用户，验证密码
		err = json.Unmarshal([]byte(userData), &userInfo)
		if err != nil {
			utils.ErrorResponse(c.Ctx, 5001, "用户数据解析失败: "+err.Error(), nil)
			return
		}

		// 验证密码
		if storedPassword, ok := userInfo["password"].(string); !ok || storedPassword != passwordHash {
			utils.ErrorResponse(c.Ctx, 4001, "用户名或密码错误", nil)
			return
		}

		// 更新登录信息
		loginCount := 1
		if count, ok := userInfo["loginCount"].(float64); ok {
			loginCount = int(count) + 1
		}

		userInfo["loginCount"] = loginCount
		userInfo["lastLogin"] = time.Now().Format("2006-01-02 15:04:05")

		// 保存更新后的用户信息
		userInfoBytes, _ := json.Marshal(userInfo)
		err = models.SaveUserDataWithKey(req.AppId, userId, "userInfo", string(userInfoBytes))
		if err != nil {
			utils.ErrorResponse(c.Ctx, 5001, "更新用户信息失败: "+err.Error(), nil)
			return
		}
	}

	// 生成会话token（简单实现）
	sessionToken := utils.GenerateSessionToken(userId, req.AppId)

	// 保存会话信息
	sessionInfo := map[string]interface{}{
		"userId":    userId,
		"appId":     req.AppId,
		"loginTime": time.Now().Format("2006-01-02 15:04:05"),
		"ip":        c.Ctx.Input.IP(),
	}
	sessionBytes, _ := json.Marshal(sessionInfo)
	models.SaveUserDataWithKey(req.AppId, userId, "session_"+sessionToken, string(sessionBytes))

	// 返回登录成功结果
	result := map[string]interface{}{
		"userId":    userId,
		"username":  req.Username,
		"token":     sessionToken,
		"userInfo":  userInfo,
		"loginTime": time.Now().Format("2006-01-02 15:04:05"),
	}

	utils.SuccessResponse(c.Ctx, "登录成功", result)
}
