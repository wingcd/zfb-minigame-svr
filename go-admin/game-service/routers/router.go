package routers

import (
	"game-service/controllers"
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
	web.Router("/health", &controllers.HealthController{}, "get:Health")

	// 心跳接口
	web.Router("/heartbeat", &controllers.HealthController{}, "get:Heartbeat")

	// ===== zy-sdk对齐接口 =====
	// 用户数据接口（对齐zy-sdk/user.ts）
	web.Router("/user/login", &controllers.UserDataController{}, "post:Login")
	web.Router("/user/login/wx", &controllers.UserDataController{}, "post:WxLogin")
	web.Router("/user/getData", &controllers.UserDataController{}, "post:GetData")
	web.Router("/user/saveData", &controllers.UserDataController{}, "post:SaveData")
	web.Router("/user/saveUserInfo", &controllers.UserDataController{}, "post:SaveUserInfo")
	web.Router("/user/deleteData", &controllers.UserDataController{}, "post:DeleteData")

	// 排行榜接口（对齐zy-sdk/leaderboard.ts）
	web.Router("/leaderboard/commit", &controllers.LeaderboardController{}, "post:CommitScore")
	web.Router("/leaderboard/queryTopRank", &controllers.LeaderboardController{}, "post:QueryTopRank")

	// 计数器接口（对齐zy-sdk/counter.ts）
	web.Router("/counter/increment", &controllers.CounterController{}, "post:IncrementCounter")
	web.Router("/counter/get", &controllers.CounterController{}, "get:GetCounter")

	// 邮件接口（对齐zy-sdk/mail.ts）
	web.Router("/mail/getUserMails", &controllers.MailController{}, "get:GetUserMails")
	web.Router("/mail/updateStatus", &controllers.MailController{}, "post:UpdateMailStatus")

	// ===== 向后兼容接口（保留原有接口）=====
	// 用户数据接口
	web.Router("/saveData", &controllers.UserDataController{}, "post:SaveData")
	web.Router("/getData", &controllers.UserDataController{}, "post:GetData")
	web.Router("/deleteData", &controllers.UserDataController{}, "post:DeleteData")

	// 排行榜接口
	web.Router("/submitScore", &controllers.LeaderboardController{}, "post:SubmitScore")
	web.Router("/getLeaderboard", &controllers.LeaderboardController{}, "post:GetLeaderboard")
	web.Router("/getUserRank", &controllers.LeaderboardController{}, "post:GetUserRank")
	web.Router("/resetLeaderboard", &controllers.LeaderboardController{}, "post:ResetLeaderboard")

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
