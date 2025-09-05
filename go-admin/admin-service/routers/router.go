package routers

import (
	"admin-service/controllers"
	"admin-service/middlewares"

	"github.com/beego/beego/v2/server/web"
)

func init() {
	// 注册CORS中间件 - 在所有路由之前处理
	web.InsertFilter("/*", web.BeforeRouter, middlewares.CORSMiddleware)
	// 安装相关路由
	web.Router("/install", &controllers.InstallController{}, "get:ShowInstallPage")
	web.Router("/install/status", &controllers.InstallController{}, "get:CheckStatus")
	web.Router("/install/init", &controllers.InstallController{}, "post:InitSystem")
	web.Router("/install/auto", &controllers.InstallController{}, "post:AutoInstall")
	web.Router("/install/manual", &controllers.InstallController{}, "post:ManualInstall")
	web.Router("/install/test", &controllers.InstallController{}, "post:TestConnection")
	web.Router("/install/uninstall", &controllers.InstallController{}, "post:Uninstall")

	// 健康检查
	web.Router("/health", &controllers.HealthController{}, "get:Health")

	// 基本认证模块
	web.Router("/admin/login", &controllers.AuthController{}, "post:AdminLogin")
	web.Router("/admin/verifyToken", &controllers.AuthController{}, "post:VerifyToken")

	// 基本管理员管理模块
	web.Router("/admin/create", &controllers.AdminController{}, "post:CreateAdmin")
	web.Router("/admin/init", &controllers.AdminController{}, "post:InitAdmin")
	web.Router("/admin/getList", &controllers.AdminController{}, "post:GetAdmins")
	web.Router("/admin/update", &controllers.AdminController{}, "post:UpdateAdmin")
	web.Router("/admin/delete", &controllers.AdminController{}, "post:DeleteAdmin")
	web.Router("/admin/resetPwd", &controllers.AdminController{}, "post:ResetPassword")

	// 基本应用管理模块
	web.Router("/app/getAll", &controllers.ApplicationController{}, "post:GetApplications")
	web.Router("/app/create", &controllers.ApplicationController{}, "post:CreateApplication")
	web.Router("/app/update", &controllers.ApplicationController{}, "post:UpdateApplication")
	web.Router("/app/delete", &controllers.ApplicationController{}, "post:DeleteApplication")
	web.Router("/app/init", &controllers.ApplicationController{}, "post:CreateApplication")
	web.Router("/app/query", &controllers.ApplicationController{}, "post:GetApplication")
	web.Router("/app/getDetail", &controllers.ApplicationController{}, "post:GetApplication")

	// 用户管理模块
	web.Router("/user/getAll", &controllers.UserController{}, "post:GetAllUsers")
	web.Router("/user/ban", &controllers.UserController{}, "post:BanUser")
	web.Router("/user/unban", &controllers.UserController{}, "post:UnbanUser")
	web.Router("/user/delete", &controllers.UserController{}, "post:DeleteUser")
	web.Router("/user/getDetail", &controllers.UserController{}, "post:GetUserDetail")
	web.Router("/user/setDetail", &controllers.UserController{}, "post:SetUserDetail")
	web.Router("/user/getStats", &controllers.UserController{}, "post:GetUserStats")
	// 排行榜管理模块
	web.Router("/leaderboard/getAll", &controllers.LeaderboardController{}, "post:GetAllLeaderboards")
	web.Router("/leaderboard/create", &controllers.LeaderboardController{}, "post:CreateLeaderboard")
	web.Router("/leaderboard/update", &controllers.LeaderboardController{}, "post:UpdateLeaderboard")
	web.Router("/leaderboard/delete", &controllers.LeaderboardController{}, "post:DeleteLeaderboard")
	web.Router("/leaderboard/getData", &controllers.LeaderboardController{}, "post:GetLeaderboardData")
	web.Router("/leaderboard/updateScore", &controllers.LeaderboardController{}, "post:UpdateLeaderboardScore")
	web.Router("/leaderboard/deleteScore", &controllers.LeaderboardController{}, "post:DeleteLeaderboardScore")
	// 计数器管理模块
	web.Router("/counter/getList", &controllers.CounterController{}, "post:GetCounterList")
	web.Router("/counter/create", &controllers.CounterController{}, "post:CreateCounter")
	web.Router("/counter/update", &controllers.CounterController{}, "post:UpdateCounter")
	web.Router("/counter/delete", &controllers.CounterController{}, "post:DeleteCounter")
	web.Router("/counter/getAllStats", &controllers.CounterController{}, "post:GetAllCounterStats")
	// 统计模块
	web.Router("/stat/dashboard", &controllers.StatsController{}, "post:GetDashboardStats")
	web.Router("/stat/getTopApps", &controllers.StatsController{}, "post:GetTopApps")
	web.Router("/stat/getRecentActivity", &controllers.StatsController{}, "post:GetRecentActivity")
	web.Router("/stat/getUserGrowth", &controllers.StatsController{}, "post:GetUserGrowth")
	web.Router("/stat/getAppStats", &controllers.StatsController{}, "post:GetAppStats")
	web.Router("/stat/getLeaderboardStats", &controllers.StatsController{}, "post:GetLeaderboardStats")
	// 邮件管理模块
	web.Router("/mail/getAll", &controllers.MailController{}, "post:GetAllMails")
	web.Router("/mail/create", &controllers.MailController{}, "post:CreateMail")
	web.Router("/mail/update", &controllers.MailController{}, "post:UpdateMail")
	web.Router("/mail/delete", &controllers.MailController{}, "post:DeleteMail")
	web.Router("/mail/send", &controllers.MailController{}, "post:SendMail")
	web.Router("/mail/getStats", &controllers.MailController{}, "post:GetMailStats")
	web.Router("/mail/getUserMails", &controllers.MailController{}, "post:GetUserMails")
	web.Router("/mail/initSystem", &controllers.MailController{}, "post:InitMailSystem")
	// 游戏配置模块
	web.Router("/gameConfig/getList", &controllers.GameConfigController{}, "post:GetGameConfigList")
	web.Router("/gameConfig/create", &controllers.GameConfigController{}, "post:CreateGameConfig")
	web.Router("/gameConfig/update", &controllers.GameConfigController{}, "post:UpdateGameConfig")
	web.Router("/gameConfig/delete", &controllers.GameConfigController{}, "post:DeleteGameConfig")
	web.Router("/gameConfig/get", &controllers.GameConfigController{}, "post:GetGameConfig")

	apiNamespace := web.NewNamespace("/api",
		// 认证相关
		web.NSRouter("/auth/login", &controllers.AuthController{}, "post:Login"),
		web.NSRouter("/auth/logout", &controllers.AuthController{}, "post:Logout"),
		web.NSRouter("/auth/profile", &controllers.AuthController{}, "get:GetProfile"),
		web.NSRouter("/auth/profile", &controllers.AuthController{}, "put:UpdateProfile"),
		web.NSRouter("/auth/password", &controllers.AuthController{}, "put:ChangePassword"),

		// 应用管理
		web.NSRouter("/applications", &controllers.ApplicationController{}, "get:GetApplications"),
		web.NSRouter("/applications", &controllers.ApplicationController{}, "post:CreateApplication"),
		web.NSRouter("/applications/:id", &controllers.ApplicationController{}, "get:GetApplication"),
		web.NSRouter("/applications/:id", &controllers.ApplicationController{}, "put:UpdateApplication"),
		web.NSRouter("/applications/:id", &controllers.ApplicationController{}, "delete:DeleteApplication"),
		web.NSRouter("/applications/:id/reset-secret", &controllers.ApplicationController{}, "post:ResetAppSecret"),

		// 管理员管理
		web.NSRouter("/admins", &controllers.AdminController{}, "get:GetAdmins"),
		web.NSRouter("/admins", &controllers.AdminController{}, "post:CreateAdmin"),
		web.NSRouter("/admins/:id", &controllers.AdminController{}, "get:GetAdmin"),
		web.NSRouter("/admins/:id", &controllers.AdminController{}, "put:UpdateAdmin"),
		web.NSRouter("/admins/:id", &controllers.AdminController{}, "delete:DeleteAdmin"),
		web.NSRouter("/admins/:id/reset-password", &controllers.AdminController{}, "post:ResetPassword"),

		// 游戏数据管理
		web.NSRouter("/game-data/user-data", &controllers.GameDataController{}, "get:GetUserDataList"),
		web.NSRouter("/game-data/leaderboard", &controllers.GameDataController{}, "get:GetLeaderboardList"),

		// 用户管理模块
		web.NSRouter("/user-management/users", &controllers.UserManagementController{}, "get:GetAllUsers"),
		web.NSRouter("/user-management/user/detail", &controllers.UserManagementController{}, "get:GetUserDetail"),
		web.NSRouter("/user-management/user/data", &controllers.UserManagementController{}, "put:UpdateUserData"),
		web.NSRouter("/user-management/user/ban", &controllers.UserManagementController{}, "post:BanUser"),
		web.NSRouter("/user-management/user/unban", &controllers.UserManagementController{}, "post:UnbanUser"),
		web.NSRouter("/user-management/user/delete", &controllers.UserManagementController{}, "delete:DeleteUser"),
		web.NSRouter("/user-management/user/stats", &controllers.UserManagementController{}, "get:GetUserStats"),
		web.NSRouter("/user-management/stats/registration", &controllers.UserManagementController{}, "get:GetUserRegistrationStats"),
		web.NSRouter("/game-data/counter", &controllers.GameDataController{}, "get:GetCounterList"),
		web.NSRouter("/game-data/mail", &controllers.GameDataController{}, "get:GetMailList"),
		web.NSRouter("/game-data/mail", &controllers.GameDataController{}, "post:SendMail"),
		web.NSRouter("/game-data/mail/broadcast", &controllers.GameDataController{}, "post:SendBroadcastMail"),
		web.NSRouter("/game-data/config", &controllers.GameDataController{}, "get:GetConfigList"),
		web.NSRouter("/game-data/config", &controllers.GameDataController{}, "put:UpdateConfig"),
		web.NSRouter("/game-data/config", &controllers.GameDataController{}, "delete:DeleteConfig"),

		// 统计分析
		web.NSRouter("/statistics/dashboard", &controllers.StatisticsController{}, "get:GetDashboard"),
		web.NSRouter("/statistics/application", &controllers.StatisticsController{}, "get:GetApplicationStats"),
		web.NSRouter("/statistics/logs", &controllers.StatisticsController{}, "get:GetOperationLogs"),
		web.NSRouter("/statistics/activity", &controllers.StatisticsController{}, "get:GetUserActivity"),
		web.NSRouter("/statistics/trends", &controllers.StatisticsController{}, "get:GetDataTrends"),
		web.NSRouter("/statistics/export", &controllers.StatisticsController{}, "post:ExportData"),
		web.NSRouter("/statistics/system", &controllers.StatisticsController{}, "get:GetSystemInfo"),

		// 权限管理
		web.NSRouter("/permissions/roles", &controllers.PermissionController{}, "get:GetRoles"),
		web.NSRouter("/permissions/roles", &controllers.PermissionController{}, "post:CreateRole"),
		web.NSRouter("/permissions/roles/:id", &controllers.PermissionController{}, "get:GetRole"),
		web.NSRouter("/permissions/roles/:id", &controllers.PermissionController{}, "put:UpdateRole"),
		web.NSRouter("/permissions/roles/:id", &controllers.PermissionController{}, "delete:DeleteRole"),
		web.NSRouter("/permissions/permissions", &controllers.PermissionController{}, "get:GetPermissions"),
		web.NSRouter("/permissions/permissions", &controllers.PermissionController{}, "post:CreatePermission"),
		web.NSRouter("/permissions/permissions/:id", &controllers.PermissionController{}, "put:UpdatePermission"),
		web.NSRouter("/permissions/permissions/:id", &controllers.PermissionController{}, "delete:DeletePermission"),
		web.NSRouter("/permissions/tree", &controllers.PermissionController{}, "get:GetPermissionTree"),

		// 系统管理
		web.NSRouter("/system/config", &controllers.SystemController{}, "get:GetSystemConfig"),
		web.NSRouter("/system/config", &controllers.SystemController{}, "put:UpdateSystemConfig"),
		web.NSRouter("/system/status", &controllers.SystemController{}, "get:GetSystemStatus"),
		web.NSRouter("/system/cache", &controllers.SystemController{}, "delete:ClearCache"),
		web.NSRouter("/system/cache/stats", &controllers.SystemController{}, "get:GetCacheStats"),
		web.NSRouter("/system/logs/clean", &controllers.SystemController{}, "post:CleanLogs"),
		web.NSRouter("/system/backup", &controllers.SystemController{}, "post:BackupData"),
		web.NSRouter("/system/backup", &controllers.SystemController{}, "get:GetBackupList"),
		web.NSRouter("/system/backup/restore", &controllers.SystemController{}, "post:RestoreBackup"),
		web.NSRouter("/system/backup/:id", &controllers.SystemController{}, "delete:DeleteBackup"),
		web.NSRouter("/system/server", &controllers.SystemController{}, "get:GetServerInfo"),
		web.NSRouter("/system/database", &controllers.SystemController{}, "get:GetDatabaseInfo"),
		web.NSRouter("/system/database/optimize", &controllers.SystemController{}, "post:OptimizeDatabase"),

		// 文件管理
		web.NSRouter("/files/upload", &controllers.UploadController{}, "post:UploadFile"),
		web.NSRouter("/files", &controllers.UploadController{}, "get:GetFileList"),
		web.NSRouter("/files/:id", &controllers.UploadController{}, "get:GetFileInfo"),
		web.NSRouter("/files/:id", &controllers.UploadController{}, "delete:DeleteFile"),
		web.NSRouter("/files/:id/download", &controllers.UploadController{}, "get:DownloadFile"),
		web.NSRouter("/files/batch/delete", &controllers.UploadController{}, "post:BatchDeleteFiles"),
		web.NSRouter("/files/stats", &controllers.UploadController{}, "get:GetUploadStats"),
		web.NSRouter("/files/cleanup", &controllers.UploadController{}, "post:CleanupFiles"),

		// 通知管理
		web.NSRouter("/notifications", &controllers.NotificationController{}, "get:GetNotifications"),
		web.NSRouter("/notifications", &controllers.NotificationController{}, "post:CreateNotification"),
		web.NSRouter("/notifications/:id", &controllers.NotificationController{}, "get:GetNotification"),
		web.NSRouter("/notifications/:id", &controllers.NotificationController{}, "put:UpdateNotification"),
		web.NSRouter("/notifications/:id", &controllers.NotificationController{}, "delete:DeleteNotification"),
		web.NSRouter("/notifications/:id/send", &controllers.NotificationController{}, "post:SendNotification"),
		web.NSRouter("/notifications/templates", &controllers.NotificationController{}, "get:GetNotificationTemplates"),
		web.NSRouter("/notifications/templates", &controllers.NotificationController{}, "post:CreateNotificationTemplate"),
		web.NSRouter("/notifications/logs", &controllers.NotificationController{}, "get:GetNotificationLogs"),
		web.NSRouter("/notifications/stats", &controllers.NotificationController{}, "get:GetNotificationStats"),
		web.NSRouter("/notifications/mark-read", &controllers.NotificationController{}, "post:MarkAsRead"),
	)

	web.AddNamespace(apiNamespace)

}
