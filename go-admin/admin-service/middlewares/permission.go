package middlewares

import (
	"admin-service/models"
	"encoding/json"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(ctx *context.Context) {
	// 获取请求路径
	requestPath := ctx.Request.URL.Path

	// 对于不需要认证的接口，直接跳过权限检查
	skipAuthPaths := []string{
		"/admin/login",
		"/admin/verifyToken",
		"/admin/init",
		"/api/auth/login",
		"/install",
		"/health",
	}

	for _, path := range skipAuthPaths {
		if requestPath == path {
			logs.Debug("Skipping permission check for auth path: %s", requestPath)
			return
		}
	}

	// 获取用户信息（必须在AuthMiddleware之后执行）
	userID := ctx.Input.GetData("user_id")
	roleID := ctx.Input.GetData("role_id")

	// 如果没有用户信息，说明是不需要认证的接口，直接跳过权限检查
	if userID == nil || roleID == nil {
		logs.Debug("No user info found, skipping permission check for %s %s", ctx.Input.Method(), requestPath)
		return
	}

	userIDVal := userID.(int64)
	roleIDVal := roleID.(int64)

	// 获取请求方法
	requestMethod := ctx.Input.Method()

	// 检查是否为超级管理员（超级管理员拥有所有权限）
	if isSuperAdmin(roleIDVal) {
		logs.Debug("Super admin access granted for user %d to %s %s", userIDVal, requestMethod, requestPath)
		return
	}

	// 获取用户角色权限
	_, permissions, err := models.GetAdminRolePermissions(roleIDVal)
	if err != nil {
		logs.Error("Failed to get permissions for role %d: %v", roleIDVal, err)
		responsePermissionError(ctx, 5001, "权限检查失败")
		ctx.Abort(403, "")
		return
	}

	// 检查权限
	requiredPermission := getRequiredPermission(requestPath, requestMethod)
	if requiredPermission == "" {
		// 如果没有定义所需权限，允许访问（兼容性考虑）
		logs.Debug("No permission required for %s %s", requestMethod, requestPath)
		return
	}

	// 检查用户是否具有所需权限
	if !hasPermission(permissions, requiredPermission) {
		logs.Warning("Permission denied for user %d (role %d) to access %s %s. Required: %s, Has: %v",
			userIDVal, roleIDVal, requestMethod, requestPath, requiredPermission, permissions)
		responsePermissionError(ctx, 4003, "权限不足")
		ctx.Abort(403, "")
		return
	}

	logs.Debug("Permission granted for user %d to access %s %s with permission %s",
		userIDVal, requestMethod, requestPath, requiredPermission)
}

// isSuperAdmin 检查是否为超级管理员
func isSuperAdmin(roleID int64) bool {
	// 获取角色信息
	role := &models.AdminRole{}
	err := role.GetById(roleID)
	if err != nil {
		return false
	}

	// 检查是否为超级管理员角色
	return role.RoleCode == "super_admin"
}

// getRequiredPermission 根据请求路径和方法获取所需权限
func getRequiredPermission(path, method string) string {
	// 定义路径权限映射
	permissionMap := map[string]string{
		// 管理员管理
		"/admin/create":   "admin_manage",
		"/admin/getList":  "admin_manage",
		"/admin/update":   "admin_manage",
		"/admin/delete":   "admin_manage",
		"/admin/resetPwd": "admin_manage",

		// 应用管理
		"/app/getAll":    "app_manage",
		"/app/create":    "app_manage",
		"/app/update":    "app_manage",
		"/app/delete":    "app_manage",
		"/app/init":      "app_manage",
		"/app/query":     "app_manage",
		"/app/getDetail": "app_manage",

		// 用户管理
		"/user/getAll":    "user_manage",
		"/user/ban":       "user_manage",
		"/user/unban":     "user_manage",
		"/user/delete":    "user_manage",
		"/user/getDetail": "user_manage",
		"/user/setDetail": "user_manage",
		"/user/getStats":  "user_manage",

		// 排行榜管理
		"/leaderboard/getAll":      "leaderboard_manage",
		"/leaderboard/create":      "leaderboard_manage",
		"/leaderboard/update":      "leaderboard_manage",
		"/leaderboard/delete":      "leaderboard_manage",
		"/leaderboard/getData":     "leaderboard_manage",
		"/leaderboard/updateScore": "leaderboard_manage",
		"/leaderboard/deleteScore": "leaderboard_manage",

		// 计数器管理
		"/counter/getList":     "leaderboard_manage",
		"/counter/create":      "leaderboard_manage",
		"/counter/update":      "leaderboard_manage",
		"/counter/delete":      "leaderboard_manage",
		"/counter/getAllStats": "leaderboard_manage",

		// 统计查看
		"/stat/dashboard":           "stats_view",
		"/stat/getTopApps":          "stats_view",
		"/stat/getRecentActivity":   "stats_view",
		"/stat/getUserGrowth":       "stats_view",
		"/stat/getAppStats":         "stats_view",
		"/stat/getLeaderboardStats": "stats_view",

		// 邮件管理
		"/mail/getAll":       "mail_manage",
		"/mail/create":       "mail_manage",
		"/mail/update":       "mail_manage",
		"/mail/delete":       "mail_manage",
		"/mail/send":         "mail_manage",
		"/mail/getStats":     "mail_manage",
		"/mail/getUserMails": "mail_manage",
		"/mail/initSystem":   "mail_manage",

		// 游戏配置管理
		"/gameConfig/getList": "game_config_manage",
		"/gameConfig/create":  "game_config_manage",
		"/gameConfig/update":  "game_config_manage",
		"/gameConfig/delete":  "game_config_manage",
		"/gameConfig/get":     "game_config_manage",

		// 权限管理（旧路由）
		"/permission/getRoles":       "role_manage",
		"/permission/createRole":     "role_manage",
		"/permission/updateRole":     "role_manage",
		"/permission/deleteRole":     "role_manage",
		"/permission/getPermissions": "role_manage",
	}

	// API命名空间权限映射
	apiPermissionMap := map[string]string{
		// 认证相关（所有已认证用户都可以访问，不需要特殊权限）
		"/api/auth/logout":   "",
		"/api/auth/profile":  "",
		"/api/auth/password": "",

		// 应用管理
		"/api/applications": "app_manage",

		// 管理员管理
		"/api/admins": "admin_manage",

		// 游戏数据管理
		"/api/game-data/user-data":   "user_manage",
		"/api/game-data/leaderboard": "leaderboard_manage",
		"/api/game-data/counter":     "counter_manage",
		"/api/game-data/mail":        "mail_manage",
		"/api/game-data/config":      "game_config_manage",

		// 用户管理
		"/api/user-management/users":              "user_manage",
		"/api/user-management/user/detail":        "user_manage",
		"/api/user-management/user/data":          "user_manage",
		"/api/user-management/user/ban":           "user_manage",
		"/api/user-management/user/unban":         "user_manage",
		"/api/user-management/user/delete":        "user_manage",
		"/api/user-management/user/stats":         "user_manage",
		"/api/user-management/stats/registration": "stats_view",

		// 统计分析
		"/api/statistics/dashboard":   "stats_view",
		"/api/statistics/application": "stats_view",
		"/api/statistics/logs":        "stats_view",
		"/api/statistics/activity":    "stats_view",
		"/api/statistics/trends":      "stats_view",
		"/api/statistics/export":      "stats_view",
		"/api/statistics/system":      "stats_view",

		// 权限管理
		"/api/permissions/roles":       "role_manage",
		"/api/permissions/permissions": "role_manage",
		"/api/permissions/tree":        "role_manage",

		// 系统管理
		"/api/system/config":   "system_config",
		"/api/system/status":   "system_config",
		"/api/system/cache":    "system_config",
		"/api/system/logs":     "system_config",
		"/api/system/backup":   "system_config",
		"/api/system/server":   "system_config",
		"/api/system/database": "system_config",

		// 文件管理
		"/api/files/upload": "file_manage",
		"/api/files":        "file_manage",

		// 通知管理
		"/api/notifications": "notification_manage",
	}

	// 首先检查完整路径
	if permission, exists := permissionMap[path]; exists {
		return permission
	}

	// 检查API路径
	if permission, exists := apiPermissionMap[path]; exists {
		return permission
	}

	// 检查API路径前缀匹配
	for apiPath, permission := range apiPermissionMap {
		if strings.HasPrefix(path, apiPath) {
			return permission
		}
	}

	// 特殊处理带参数的API路径
	if strings.HasPrefix(path, "/api/applications/") {
		return "app_manage"
	}
	if strings.HasPrefix(path, "/api/admins/") {
		return "admin_manage"
	}
	if strings.HasPrefix(path, "/api/permissions/") {
		return "role_manage"
	}
	if strings.HasPrefix(path, "/api/files/") {
		return "file_manage"
	}
	if strings.HasPrefix(path, "/api/notifications/") {
		return "notification_manage"
	}

	// 默认返回空字符串（表示不需要特定权限）
	return ""
}

// hasPermission 检查用户是否具有指定权限
func hasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission {
			return true
		}
	}
	return false
}

// responsePermissionError 返回权限错误响应
func responsePermissionError(ctx *context.Context, code int, message string) {
	response := map[string]interface{}{
		"code": code,
		"msg":  message,
		"data": nil,
	}

	ctx.Output.Header("Content-Type", "application/json")
	ctx.Output.SetStatus(403)

	jsonData, _ := json.Marshal(response)
	ctx.Output.Body(jsonData)
}
