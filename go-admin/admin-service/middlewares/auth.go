package middlewares

import (
	"admin-service/utils"
	"encoding/json"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(ctx *context.Context) {
	// 跳过登录和健康检查等公开接口
	skipPaths := []string{
		"/admin/login",
		"/admin/verifyToken",
		"/health",
		"/ping",
		"/install",
	}

	requestPath := ctx.Request.URL.Path
	for _, path := range skipPaths {
		if strings.HasSuffix(requestPath, path) {
			return
		}
	}

	// 获取Authorization头
	authHeader := ctx.Input.Header("Authorization")
	if authHeader == "" {
		responseError(ctx, 401, "缺少认证信息")
		ctx.Abort(401, "")
		return
	}

	// 检查Bearer格式
	if !strings.HasPrefix(authHeader, "Bearer ") {
		responseError(ctx, 401, "认证格式错误")
		ctx.Abort(401, "")
		return
	}

	// 提取token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		responseError(ctx, 401, "Token不能为空")
		ctx.Abort(401, "")
		return
	}

	// 验证token (统一使用utils.ParseJWT)
	claims, err := utils.ParseJWT(token)
	if err != nil {
		responseError(ctx, 401, "Token验证失败: "+err.Error())
		ctx.Abort(401, "")
		return
	}

	// 将用户信息存储到上下文中
	if claims != nil {
		ctx.Input.SetData("playerId", claims.UserID)
		ctx.Input.SetData("username", claims.Username)
		ctx.Input.SetData("playerId", claims.RoleID)
		ctx.Input.SetData("role", claims.Role)
	}
}

// LogMiddleware 日志中间件
func LogMiddleware(ctx *context.Context) {
	// 记录请求日志
	logs.Info("Request: %s %s from %s", ctx.Input.Method(), ctx.Request.URL.Path, ctx.Input.IP())

	// 如果是管理员操作，记录操作日志
	if ctx.Input.GetData("playerId") != nil {
		userId := ctx.Input.GetData("playerId").(int64)
		username := ctx.Input.GetData("username").(string)

		// 记录到操作日志表（这里简化处理，实际可以异步记录）
		logs.Info("Admin Operation: User[%d:%s] %s %s", userId, username, ctx.Input.Method(), ctx.Request.URL.Path)
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(ctx *context.Context) {
	// 简单的IP限流（实际项目中可以使用Redis实现更复杂的限流）
	clientIP := ctx.Input.IP()

	// 这里可以实现基于IP的限流逻辑
	// 为了简化，暂时不实现具体限流逻辑
	logs.Debug("Rate limit check for IP: %s", clientIP)
}

// responseError 返回错误响应
func responseError(ctx *context.Context, code int, message string) {
	// 统一使用4001错误码表示认证失败
	errorCode := 4001
	if code == 401 {
		errorCode = 4001
	} else {
		errorCode = code
	}

	response := map[string]interface{}{
		"code": errorCode,
		"msg":  message,
		"data": nil,
	}

	ctx.Output.Header("Content-Type", "application/json")
	ctx.Output.SetStatus(401) // HTTP状态码保持401

	jsonData, _ := json.Marshal(response)
	ctx.Output.Body(jsonData)
}
