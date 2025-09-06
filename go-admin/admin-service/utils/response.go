package utils

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/beego/beego/v2/server/web"
)

// APIResponse 统一API响应结构
type APIResponse struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ResponseCode 响应码常量
const (
	// 成功
	CodeSuccess = 0

	// 客户端错误 4xxx
	CodeBadRequest      = 4001 // 参数错误
	CodeResourceExists  = 4002 // 资源已存在
	CodeUnauthorized    = 4003 // 未授权
	CodeNotFound        = 4004 // 资源不存在
	CodeForbidden       = 4005 // 权限不足
	CodeValidationError = 4006 // 数据验证错误

	// 服务器错误 5xxx
	CodeServerError   = 5001 // 服务器内部错误
	CodeDatabaseError = 5002 // 数据库错误
	CodeCacheError    = 5003 // 缓存错误
)

// ResponseMessage 响应消息映射
var ResponseMessage = map[int]string{
	CodeSuccess:         "success",
	CodeBadRequest:      "参数错误",
	CodeResourceExists:  "资源已存在",
	CodeUnauthorized:    "未授权",
	CodeNotFound:        "资源不存在",
	CodeForbidden:       "权限不足",
	CodeValidationError: "数据验证错误",
	CodeServerError:     "服务器内部错误",
	CodeDatabaseError:   "数据库错误",
	CodeCacheError:      "缓存错误",
}

// NewResponse 创建新的响应
func NewResponse(code int, data interface{}) *APIResponse {
	msg := ResponseMessage[code]
	if msg == "" {
		msg = "未知错误"
	}

	return &APIResponse{
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().UnixNano() / 1e6, // 毫秒时间戳
		Data:      data,
	}
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Code:      CodeSuccess,
		Msg:       "success",
		Timestamp: time.Now().UnixNano() / 1e6,
		Data:      data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, customMsg ...string) *APIResponse {
	msg := ResponseMessage[code]
	if len(customMsg) > 0 && customMsg[0] != "" {
		msg = customMsg[0]
	}
	if msg == "" {
		msg = "未知错误"
	}

	return &APIResponse{
		Code:      code,
		Msg:       msg,
		Timestamp: time.Now().UnixNano() / 1e6,
		Data:      nil,
	}
}

// NewListResponse 创建列表响应
func NewListResponse(list interface{}, total int64, page, pageSize int) *APIResponse {
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	data := map[string]interface{}{
		"list":       list,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	}

	return NewSuccessResponse(data)
}

// SendResponse 发送响应
func SendResponse(c *web.Controller, response *APIResponse) {
	c.Data["json"] = response
	c.ServeJSON()
}

// SendSuccess 发送成功响应
func SendSuccess(c *web.Controller, data interface{}) {
	SendResponse(c, NewSuccessResponse(data))
}

// SendError 发送错误响应
func SendError(c *web.Controller, code int, customMsg ...string) {
	SendResponse(c, NewErrorResponse(code, customMsg...))
}

// SendList 发送列表响应
func SendList(c *web.Controller, list interface{}, total int64, page, pageSize int) {
	SendResponse(c, NewListResponse(list, total, page, pageSize))
}

// ParseJSONRequest 解析JSON请求
func ParseJSONRequest(c *web.Controller, v interface{}) error {
	return json.Unmarshal(c.Ctx.Input.RequestBody, v)
}

// ValidateAndSendError 验证参数并发送错误响应
func ValidateAndSendError(c *web.Controller, condition bool, msg string) bool {
	if condition {
		SendError(c, CodeBadRequest, msg)
		return true
	}
	return false
}

// SuccessResponse 兼容性成功响应方法
func SuccessResponse(c *web.Controller, message string, data interface{}) {
	response := &APIResponse{
		Code:      CodeSuccess,
		Msg:       "success", // 云函数兼容格式
		Timestamp: time.Now().UnixNano() / 1e6,
		Data:      data,
	}
	c.Data["json"] = response
	c.ServeJSON()
}

// ErrorResponse 兼容性错误响应方法
func ErrorResponse(c *web.Controller, code int, message string, data interface{}) {
	response := &APIResponse{
		Code:      code,
		Msg:       message,
		Timestamp: time.Now().UnixNano() / 1e6,
		Data:      data,
	}
	c.Data["json"] = response
	c.ServeJSON()
}

// 兼容性常量和函数
const (
	CodeInvalidParam = CodeBadRequest
	CodeError        = CodeServerError
	CodeConflict     = CodeResourceExists
)

// ParseJSON 解析JSON请求 - 兼容性函数
func ParseJSON(c *web.Controller, v interface{}) error {
	return ParseJSONRequest(c, v)
}

// Error 发送错误响应 - 兼容性函数
func Error(c *web.Controller, code int, message string) {
	SendError(c, code, message)
}

// Success 发送成功响应 - 兼容性函数
func Success(c *web.Controller, data interface{}) {
	SendSuccess(c, data)
}

// PageSuccess 发送分页成功响应 - 兼容性函数
func PageSuccess(c *web.Controller, list interface{}, total int64, page, pageSize int) {
	SendList(c, list, total, page, pageSize)
}

// ValidateRequired 验证必填参数 - 兼容性函数
func ValidateRequired(c *web.Controller, fields map[string]interface{}) bool {
	for name, value := range fields {
		if value == nil || value == "" {
			SendError(c, CodeBadRequest, "参数"+name+"不能为空")
			return false
		}
	}
	return true
}

// GetIntParam 获取整型参数 - 兼容性函数
func GetIntParam(c *web.Controller, key string, defaultValue ...int) int {
	str := c.GetString(key)
	if str == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return val
}

// GetStringParam 获取字符串参数 - 兼容性函数
func GetStringParam(c *web.Controller, key string, defaultValue ...string) string {
	str := c.GetString(key)
	if str == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return str
}

// GetClientIP 获取客户端IP - 兼容性函数
func GetClientIP(c *web.Controller) string {
	ip := c.Ctx.Input.Header("X-Forwarded-For")
	if ip == "" {
		ip = c.Ctx.Input.Header("X-Real-Ip")
	}
	if ip == "" {
		ip = c.Ctx.Request.RemoteAddr
	}
	return ip
}

// CloudResponse 云函数兼容响应格式
func CloudResponse(c *web.Controller, code int, message string, data interface{}) {
	response := &APIResponse{
		Code:      code,
		Msg:       message,
		Timestamp: time.Now().UnixNano() / 1e6,
		Data:      data,
	}
	c.Data["json"] = response
	c.ServeJSON()
}

// CloudSuccess 云函数兼容成功响应
func CloudSuccess(c *web.Controller, data interface{}) {
	CloudResponse(c, 0, "success", data)
}

// CloudError 云函数兼容错误响应
func CloudError(c *web.Controller, code int, message string) {
	CloudResponse(c, code, message, nil)
}
