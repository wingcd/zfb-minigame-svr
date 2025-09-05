package utils

import (
	"encoding/json"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

// ResponseCode 响应码定义
const (
	CodeSuccess         = 200 // 成功
	CodeError           = 500 // 服务器错误
	CodeInvalidParam    = 400 // 参数错误
	CodeUnauthorized    = 401 // 未授权
	CodeForbidden       = 403 // 禁止访问
	CodeNotFound        = 404 // 未找到
	CodeConflict        = 409 // 数据冲突
	CodeTooManyRequests = 429 // 请求过多
)

// ResponseMessage 响应消息定义
var ResponseMessage = map[int]string{
	CodeSuccess:         "操作成功",
	CodeError:           "服务器内部错误",
	CodeInvalidParam:    "参数错误",
	CodeUnauthorized:    "未授权访问",
	CodeForbidden:       "禁止访问",
	CodeNotFound:        "资源未找到",
	CodeConflict:        "数据冲突",
	CodeTooManyRequests: "请求过于频繁",
}

// APIResponse API响应结构
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Pages    int         `json:"pages"`
}

// Success 成功响应
func Success(ctx *web.Controller, data interface{}) {
	response := APIResponse{
		Code: CodeSuccess,
		Msg:  ResponseMessage[CodeSuccess],
		Data: data,
	}
	ctx.Data["json"] = response
	ctx.ServeJSON()
}

// Error 错误响应
func Error(ctx *web.Controller, code int, message ...string) {
	msg := ResponseMessage[code]
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := APIResponse{
		Code: code,
		Msg:  msg,
	}
	ctx.Data["json"] = response
	ctx.ServeJSON()
}

// ErrorResponse 错误响应（用于web.Controller，对齐云函数格式）
func ErrorResponse(ctx *web.Controller, code int, message string, data interface{}) {
	response := APIResponse{
		Code: code,
		Msg:  message,
		Data: data,
	}
	ctx.Data["json"] = response
	ctx.ServeJSON()
}

// SuccessResponse 成功响应（用于web.Controller，对齐云函数格式）
func SuccessResponse(ctx *web.Controller, message string, data interface{}) {
	response := APIResponse{
		Code: CodeSuccess,
		Msg:  message,
		Data: data,
	}
	ctx.Data["json"] = response
	ctx.ServeJSON()
}

// ErrorResponseContext 错误响应（用于context.Context）
func ErrorResponseContext(ctx *context.Context, code int, message string, data interface{}) {
	response := APIResponse{
		Code: code,
		Msg:  message,
		Data: data,
	}
	ctx.Output.JSON(response, false, false)
}

// SuccessResponseContext 成功响应（用于context.Context）
func SuccessResponseContext(ctx *context.Context, message string, data interface{}) {
	response := APIResponse{
		Code: CodeSuccess,
		Msg:  message,
		Data: data,
	}
	ctx.Output.JSON(response, false, false)
}

// PageSuccess 分页成功响应
func PageSuccess(ctx *web.Controller, list interface{}, total int64, page, pageSize int) {
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}

	pageData := PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	}

	Success(ctx, pageData)
}

// GetClientIP 获取客户端IP
func GetClientIP(ctx *web.Controller) string {
	// 首先检查 X-Forwarded-For 头
	ip := ctx.Ctx.Request.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}

	// 检查 X-Real-IP 头
	ip = ctx.Ctx.Request.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 返回远程地址
	return ctx.Ctx.Request.RemoteAddr
}

// ParseJSON 解析JSON参数
func ParseJSON(ctx *web.Controller, v interface{}) error {
	return json.Unmarshal(ctx.Ctx.Input.RequestBody, v)
}

// GetStringParam 获取字符串参数
func GetStringParam(ctx *web.Controller, key string, defaultValue ...string) string {
	value := ctx.GetString(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// GetIntParam 获取整型参数
func GetIntParam(ctx *web.Controller, key string, defaultValue ...int) int {
	value, err := ctx.GetInt(key)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// GetInt64Param 获取长整型参数
func GetInt64Param(ctx *web.Controller, key string, defaultValue ...int64) int64 {
	value, err := ctx.GetInt64(key)
	if err != nil && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// ValidateRequired 验证必填参数
func ValidateRequired(ctx *web.Controller, params map[string]interface{}) bool {
	for key, value := range params {
		if value == nil || value == "" {
			Error(ctx, CodeInvalidParam, "参数 "+key+" 不能为空")
			return false
		}
	}
	return true
}

// InSlice 检查值是否在切片中
func InSlice(item string, slice []string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// InIntSlice 检查整数值是否在切片中
func InIntSlice(item int, slice []int) bool {
	for _, s := range slice {
		if item == s {
			return true
		}
	}
	return false
}
