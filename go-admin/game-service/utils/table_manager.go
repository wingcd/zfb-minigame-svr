package utils

import (
	"fmt"
	"strings"
)

// CleanAppId 清理应用ID，替换特殊字符为下划线
func CleanAppId(appId string) string {
	cleanAppId := strings.ReplaceAll(appId, "-", "_")
	cleanAppId = strings.ReplaceAll(cleanAppId, ".", "_")
	return cleanAppId
}

// GetMailTableName 获取邮件表名
func GetMailTableName(appId string) string {
	return fmt.Sprintf("mail_%s", CleanAppId(appId))
}

// GetMailRelationTableName 获取邮件关联表名
func GetMailRelationTableName(appId string) string {
	return fmt.Sprintf("mail_player_relation_%s", CleanAppId(appId))
}

// GetLeaderboardTableName 获取排行榜表名
func GetLeaderboardTableName(appId string) string {
	return fmt.Sprintf("leaderboard_%s", CleanAppId(appId))
}

// GetUserTableName 获取用户表名（新的正确命名）
func GetUserTableName(appId string) string {
	return fmt.Sprintf("user_%s", CleanAppId(appId))
}

// GetUserDataTableName 获取用户数据表名（兼容旧版本）
// Deprecated: 请使用 GetUserTableName
func GetUserDataTableName(appId string) string {
	return GetUserTableName(appId)
}

// GetCounterTableName 获取计数器表名
func GetCounterTableName(appId string) string {
	return fmt.Sprintf("counter_%s", CleanAppId(appId))
}
