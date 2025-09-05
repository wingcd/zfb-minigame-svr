package utils

import (
	"encoding/json"

	"github.com/beego/beego/v2/server/web/context"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(ctx *context.Context, message string, data interface{}) {
	response := Response{
		Code:    200,
		Message: message,
		Data:    data,
	}

	ctx.Output.Header("Content-Type", "application/json")
	ctx.Output.SetStatus(200)
	
	if err := json.NewEncoder(ctx.Output.Context.ResponseWriter).Encode(response); err != nil {
		ctx.Output.SetStatus(500)
		ctx.WriteString(`{"code":500,"message":"Internal server error"}`)
	}
}

// ErrorResponse 错误响应
func ErrorResponse(ctx *context.Context, code int, message string, data interface{}) {
	response := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}

	ctx.Output.Header("Content-Type", "application/json")
	ctx.Output.SetStatus(code)
	
	if err := json.NewEncoder(ctx.Output.Context.ResponseWriter).Encode(response); err != nil {
		ctx.Output.SetStatus(500)
		ctx.WriteString(`{"code":500,"message":"Internal server error"}`)
	}
} 