package middlewares

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"game-service/models"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
)

// SignAuthMiddleware API签名验证中间件
func SignAuthMiddleware(ctx *context.Context) {
	// 跳过健康检查等公开接口
	skipPaths := []string{
		"/health",
		"/ping",
	}

	requestPath := ctx.Request.URL.Path
	for _, path := range skipPaths {
		if strings.HasSuffix(requestPath, path) {
			return
		}
	}

	// 获取必要的请求头
	appId := ctx.Input.Header("App-Id")
	timestamp := ctx.Input.Header("Timestamp")
	sign := ctx.Input.Header("Sign")

	if appId == "" || timestamp == "" || sign == "" {
		responseError(ctx, 1001, "缺少必要的请求头参数")
		return
	}

	// 验证时间戳
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		responseError(ctx, 1001, "时间戳格式错误")
		return
	}

	// 检查时间戳是否在有效期内（5分钟）
	if time.Now().Unix()-ts > 300 || ts-time.Now().Unix() > 300 {
		responseError(ctx, 1001, "请求时间戳过期")
		return
	}

	// 获取应用信息
	app := &models.Application{AppId: appId}
	err = app.GetByAppId()
	if err != nil {
		responseError(ctx, 1001, "应用不存在")
		return
	}

	if app.Status != 1 {
		responseError(ctx, 1001, "应用已被禁用")
		return
	}

	// 获取请求体
	var requestBody map[string]interface{}
	if ctx.Input.RequestBody != nil && len(ctx.Input.RequestBody) > 0 {
		err = json.Unmarshal(ctx.Input.RequestBody, &requestBody)
		if err != nil {
			responseError(ctx, 1002, "请求体格式错误")
			return
		}
	} else {
		requestBody = make(map[string]interface{})
	}

	// 验证签名
	expectedSign := generateSign(requestBody, ts, app.AppSecret)
	if sign != expectedSign {
		responseError(ctx, 1001, "签名验证失败")
		return
	}

	// 将应用信息存储到上下文中
	ctx.Input.SetData("app_id", appId)
	ctx.Input.SetData("app_secret", app.AppSecret)
}

// CORSMiddleware CORS中间件
func CORSMiddleware(ctx *context.Context) {
	origin := ctx.Input.Header("Origin")
	if origin != "" {
		ctx.Output.Header("Access-Control-Allow-Origin", origin)
	} else {
		ctx.Output.Header("Access-Control-Allow-Origin", "*")
	}

	ctx.Output.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	ctx.Output.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With,App-Id,Token,Timestamp,Sign")
	ctx.Output.Header("Access-Control-Allow-Credentials", "true")
	ctx.Output.Header("Access-Control-Max-Age", "86400")

	// 处理预检请求
	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.SetStatus(200)
		return
	}
}

// LogMiddleware 日志中间件
func LogMiddleware(ctx *context.Context) {
	// 记录请求日志
	appId := ctx.Input.GetData("app_id")
	if appId != nil {
		logs.Info("Game API Request: App[%s] %s %s from %s", appId, ctx.Input.Method(), ctx.Request.URL.Path, ctx.Input.IP())
	} else {
		logs.Info("Request: %s %s from %s", ctx.Input.Method(), ctx.Request.URL.Path, ctx.Input.IP())
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(ctx *context.Context) {
	// 简单的IP+AppId限流（实际项目中可以使用Redis实现更复杂的限流）
	clientIP := ctx.Input.IP()
	appId := ctx.Input.GetData("app_id")

	key := fmt.Sprintf("%s:%v", clientIP, appId)
	logs.Debug("Rate limit check for key: %s", key)

	// 这里可以实现基于Redis的限流逻辑
	// 为了简化，暂时不实现具体限流逻辑
}

// generateSign 生成API签名
func generateSign(params map[string]interface{}, timestamp int64, appSecret string) string {
	// 将参数按键名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signParts []string
	for _, k := range keys {
		v := params[k]
		if v != nil {
			signParts = append(signParts, fmt.Sprintf("%s=%v", k, v))
		}
	}

	// 添加时间戳和密钥
	signStr := strings.Join(signParts, "&")
	if signStr != "" {
		signStr += "&"
	}
	signStr += fmt.Sprintf("timestamp=%d&key=%s", timestamp, appSecret)

	// MD5加密
	h := md5.New()
	h.Write([]byte(signStr))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// responseError 返回错误响应
func responseError(ctx *context.Context, code int, message string) {
	response := models.ErrorResponse(code, message)

	ctx.Output.Header("Content-Type", "application/json")
	ctx.Output.SetStatus(200) // 游戏SDK通常返回200，在响应体中标识错误

	jsonData, _ := json.Marshal(response)
	ctx.Output.Body(jsonData)
}
