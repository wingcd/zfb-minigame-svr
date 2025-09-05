package middlewares

import (
	"github.com/beego/beego/v2/server/web/context"
)

// CORSMiddleware CORS中间件
func CORSMiddleware(ctx *context.Context) {
	origin := ctx.Input.Header("Origin")
	if origin != "" {
		ctx.Output.Header("Access-Control-Allow-Origin", origin)
	} else {
		ctx.Output.Header("Access-Control-Allow-Origin", "*")
	}

	ctx.Output.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	ctx.Output.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With")
	ctx.Output.Header("Access-Control-Allow-Credentials", "true")
	ctx.Output.Header("Access-Control-Max-Age", "86400")

	// 处理预检请求
	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.SetStatus(200)
		ctx.Output.Body([]byte("OK"))
		return
	}
}
