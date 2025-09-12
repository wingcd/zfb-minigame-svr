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

	var skipTokenPath = []string{
		"/user/login",
		"/user/login/wx",
		"/user/login/alipay",
		"/user/login/douyin",
		"/user/login/qq",
	}

	requestPath := ctx.Request.URL.Path
	for _, path := range skipPaths {
		if strings.HasSuffix(requestPath, path) {
			return
		}
	}

	// 获取请求体
	var requestBody map[string]interface{}
	if len(ctx.Input.RequestBody) > 0 {
		err := json.Unmarshal(ctx.Input.RequestBody, &requestBody)
		if err != nil {
			responseError(ctx, 1002, "请求体格式错误")
			return
		}
	} else {
		responseError(ctx, 1001, "请求体不能为空")
		return
	}

	// 从请求体获取必要参数
	appId, ok := requestBody["appId"].(string)
	if !ok || appId == "" {
		responseError(ctx, 1001, "缺少appId参数")
		return
	}

	timestampVal, ok := requestBody["timestamp"]
	if !ok {
		responseError(ctx, 1001, "缺少timestamp参数")
		return
	}

	var skipToken = false
	for _, path := range skipTokenPath {
		if strings.HasSuffix(requestPath, path) {
			skipToken = true
			break
		}
	}

	if !skipToken {
		// 从数据库查询用户token是否有效
		var token string
		tokenVal, ok := requestBody["token"]
		if ok {
			playerId, ok := requestBody["playerId"]
			if !ok {
				responseError(ctx, 1001, "缺少playerId参数")
				return
			}

			token = tokenVal.(string)
			userToken, err := models.GetUserStatusFromRedis(appId, playerId.(string))
			if err != nil && userToken != token {
				responseError(ctx, 1001, "token无效")
				return
			}
		}
	}

	// 处理timestamp类型转换（可能是number或string）
	var ts int64
	switch v := timestampVal.(type) {
	case float64:
		ts = int64(v)
	case int64:
		ts = v
	case string:
		var err error
		ts, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			responseError(ctx, 1001, "时间戳格式错误")
			return
		}
	default:
		responseError(ctx, 1001, "时间戳格式错误")
		return
	}
	requestBody["timestamp"] = strconv.FormatInt(ts, 10)

	sign, ok := requestBody["sign"].(string)
	if !ok || sign == "" {
		responseError(ctx, 1001, "缺少sign参数")
		return
	}

	// 检查时间戳是否在有效期内（5分钟）
	if time.Now().Unix()-ts > 300 || ts-time.Now().Unix() > 300 {
		responseError(ctx, 1001, "请求时间戳过期")
		return
	}

	// 获取应用信息
	app := &models.Application{AppId: appId}
	err := app.GetByAppId(appId)
	if err != nil {
		responseError(ctx, 1001, "应用不存在")
		return
	}

	if app.Status != "active" {
		responseError(ctx, 1001, "应用已被禁用")
		return
	}

	// 验证签名
	expectedSign := generateSign(requestBody)
	if sign != expectedSign {
		responseError(ctx, 1001, "签名验证失败")
		return
	}

	// 将应用信息存储到上下文中
	ctx.Input.SetData("app_id", appId)
	ctx.Input.SetData("appSecret", app.ChannelAppKey)

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
func generateSign(params map[string]interface{}) string {
	// 将参数按键名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for _, k := range keys {
		if k == "sign" || k == "ver" {
			continue
		}
		v := params[k]
		if v != nil && v != "" && v != 0 {
			// 如果v不是字符串，需要转为字符串
			if _, ok := v.(string); !ok {
				v = fmt.Sprintf("%v", v)
			}
			if v == "0" {
				continue
			}
			signStr.WriteString(fmt.Sprintf("%s%s", k, v))
		}
	}
	// MD5加密
	h := md5.New()
	h.Write([]byte(signStr.String()))
	// println(signStr.String())
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
