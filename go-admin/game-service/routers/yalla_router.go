package routers

import (
	"game-service/yalla/controllers"

	"github.com/beego/beego/v2/server/web"
)

func init() {
	// Yalla API路由
	yallaNamespace := web.NewNamespace("/api/yalla",
		// 用户认证
		web.NSRouter("/auth", &controllers.YallaController{}, "post:Auth"),

		// 用户信息
		web.NSRouter("/user/info", &controllers.YallaController{}, "get:GetUserInfo"),
		web.NSRouter("/user/binding", &controllers.YallaController{}, "get:GetUserBinding"),

		// 奖励系统
		web.NSRouter("/reward/send", &controllers.YallaController{}, "post:SendReward"),

		// 数据同步
		web.NSRouter("/data/sync", &controllers.YallaController{}, "post:SyncGameData"),

		// 事件上报
		web.NSRouter("/event/report", &controllers.YallaController{}, "post:ReportEvent"),

		// 配置管理
		web.NSRouter("/config", &controllers.YallaController{}, "get:GetConfig;put:UpdateConfig"),

		// 日志查询
		web.NSRouter("/logs", &controllers.YallaController{}, "get:GetCallLogs"),
	)

	// 注册命名空间
	web.AddNamespace(yallaNamespace)
}
