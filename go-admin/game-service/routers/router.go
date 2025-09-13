package routers

import (
	"game-service/controllers"
	"game-service/controllers/login"
	"game-service/middlewares"

	"github.com/beego/beego/v2/server/web"
)

func init() {
	// 注册CORS中间件 - 在所有路由之前处理
	web.InsertFilter("/*", web.BeforeRouter, middlewares.CORSMiddleware)

	// 注册认证中间件 - 跳过健康检查等公开接口
	web.InsertFilter("/*", web.BeforeExec, middlewares.SignAuthMiddleware)

	// 注册日志中间件
	web.InsertFilter("/*", web.BeforeExec, middlewares.LogMiddleware)

	// 注册限流中间件
	web.InsertFilter("/*", web.BeforeExec, middlewares.RateLimitMiddleware)

	// 健康检查
	web.Router("/health", &controllers.HealthController{}, "post:Health")

	// 心跳接口
	web.Router("/heartbeat", &controllers.HealthController{}, "post:Heartbeat")

	// ===== zy-sdk对齐接口 =====
	// 登录接口（重构到login包）
	web.Router("/user/login", &login.CommonLoginController{}, "post:CommonLogin")
	web.Router("/user/login/wx", &login.WechatLoginController{}, "post:WxLogin")
	web.Router("/user/login/alipay", &login.AlipayLoginController{}, "post:AlipayLogin")
	web.Router("/user/login/douyin", &login.DouyinLoginController{}, "post:DouyinLogin")
	web.Router("/user/login/qq", &login.QQLoginController{}, "post:QQLogin")
	web.Router("/user/login/yalla", &login.YallaLoginController{}, "post:YallaLogin")

	// 用户数据接口（对齐zy-sdk/user.ts）
	web.Router("/user/getData", &controllers.UserController{}, "post:GetData")
	web.Router("/user/saveData", &controllers.UserController{}, "post:SaveData")
	web.Router("/user/saveUserInfo", &controllers.UserController{}, "post:SaveUserInfo")

	// 排行榜接口（对齐zy-sdk/leaderboard.ts）
	web.Router("/leaderboard/commit", &controllers.LeaderboardController{}, "post:CommitScore")
	web.Router("/leaderboard/queryTopRank", &controllers.LeaderboardController{}, "post:QueryTopRank")

	// 计数器接口（对齐zy-sdk/counter.ts）
	web.Router("/counter/increment", &controllers.CounterController{}, "post:IncrementCounter")
	web.Router("/counter/get", &controllers.CounterController{}, "post:GetCounter")

	// 邮件接口（对齐zy-sdk/mail.ts）
	web.Router("/mail/getUserMails", &controllers.MailController{}, "post:GetUserMails")
	web.Router("/mail/updateStatus", &controllers.MailController{}, "post:UpdateMailStatus")

	// ===== 向后兼容接口（保留原有接口）=====
	// 用户数据接口
	web.Router("/saveData", &controllers.UserController{}, "post:SaveData")
	web.Router("/getData", &controllers.UserController{}, "post:GetData")

	// 计数器接口
	web.Router("/getCounter", &controllers.CounterController{}, "post:GetCounter")
	web.Router("/incrementCounter", &controllers.CounterController{}, "post:IncrementCounter")
	web.Router("/decrementCounter", &controllers.CounterController{}, "post:DecrementCounter")
	web.Router("/setCounter", &controllers.CounterController{}, "post:SetCounter")
	web.Router("/resetCounter", &controllers.CounterController{}, "post:ResetCounter")
	web.Router("/getAllCounters", &controllers.CounterController{}, "post:GetAllCounters")

	// 邮件接口
	web.Router("/readMail", &controllers.MailController{}, "post:ReadMail")
	web.Router("/claimRewards", &controllers.MailController{}, "post:ClaimRewards")
	web.Router("/deleteMail", &controllers.MailController{}, "post:DeleteMail")

	// 配置接口
	web.Router("/getConfig", &controllers.ConfigController{}, "post:GetConfig")
	web.Router("/setConfig", &controllers.ConfigController{}, "post:SetConfig")
	web.Router("/getConfigsByVersion", &controllers.ConfigController{}, "post:GetConfigsByVersion")
	web.Router("/getAllConfigs", &controllers.ConfigController{}, "post:GetAllConfigs")
	web.Router("/deleteConfig", &controllers.ConfigController{}, "post:DeleteConfig")

	// 表管理接口已移除
}
