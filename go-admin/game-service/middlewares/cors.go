package middlewares

import (
	"github.com/beego/beego/v2/server/web/context"
)

// CORSMiddleware 处理跨域请求的中间件
func CORSMiddleware(ctx *context.Context) {
	// 设置 CORS 头
	ctx.Output.Header("Access-Control-Allow-Origin", "*")
	ctx.Output.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Output.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, App-Id, User-Id, Sign, Timestamp")
	ctx.Output.Header("Access-Control-Allow-Credentials", "true")
	ctx.Output.Header("Access-Control-Max-Age", "86400") // 24小时预检缓存

	// 如果是 OPTIONS 预检请求，直接返回 200
	if ctx.Input.Method() == "OPTIONS" {
		ctx.Output.SetStatus(200)
		ctx.Output.Body([]byte(""))
		return
	}
}
