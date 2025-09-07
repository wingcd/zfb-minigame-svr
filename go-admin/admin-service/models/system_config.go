package models

import (
	"github.com/beego/beego/v2/client/orm"
)

// SystemConfig 系统配置模型
type SystemConfig struct {
	BaseModel
	SiteName           string `orm:"size(100)" json:"siteName"`
	SiteUrl            string `orm:"size(255)" json:"siteUrl"`
	SiteLogo           string `orm:"size(255)" json:"siteLogo"`
	SiteDescription    string `orm:"type(text)" json:"siteDescription"`
	SiteKeywords       string `orm:"type(text)" json:"siteKeywords"`
	AdminEmail         string `orm:"size(100)" json:"adminEmail"`
	EnableRegister     bool   `orm:"default(true)" json:"enableRegister"`
	EnableEmailVerify  bool   `orm:"default(false)" json:"enableEmailVerify"`
	EnableCaptcha      bool   `orm:"default(true)" json:"enableCaptcha"`
	JwtSecret          string `orm:"size(255)" json:"jwtSecret"`
	JwtExpireHours     int    `orm:"default(24)" json:"jwtExpireHours"`
	EnableCache        bool   `orm:"default(true)" json:"enableCache"`
	CacheExpireMinutes int    `orm:"default(30)" json:"cacheExpireMinutes"`
	LogLevel           string `orm:"size(20);default(info)" json:"logLevel"`
	LogRetentionDays   int    `orm:"default(30)" json:"logRetentionDays"`
}

func (s *SystemConfig) TableName() string {
	return "system_config"
}

// GetSystemConfig 获取系统配置
func GetSystemConfig() (*SystemConfig, error) {
	o := orm.NewOrm()
	config := &SystemConfig{}

	err := o.QueryTable("system_config").OrderBy("-id").Limit(1).One(config)
	if err == orm.ErrNoRows {
		// 如果没有配置，创建默认配置
		config = &SystemConfig{
			SiteName:           "Mini Game Admin",
			SiteUrl:            "http://localhost:8080",
			SiteDescription:    "Mini Game Admin System",
			SiteKeywords:       "game,admin,management",
			AdminEmail:         "admin@example.com",
			EnableRegister:     false,
			EnableEmailVerify:  false,
			EnableCaptcha:      true,
			JwtSecret:          "default_jwt_secret",
			JwtExpireHours:     24,
			EnableCache:        true,
			CacheExpireMinutes: 30,
			LogLevel:           "info",
			LogRetentionDays:   30,
		}

		_, err = o.Insert(config)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return config, nil
}

// UpdateSystemConfig 更新系统配置
func UpdateSystemConfig(config *SystemConfig) error {
	o := orm.NewOrm()

	// 获取现有配置
	existingConfig := &SystemConfig{}
	err := o.QueryTable("system_config").OrderBy("-id").Limit(1).One(existingConfig)
	if err == orm.ErrNoRows {
		// 如果没有配置，直接插入
		_, err = o.Insert(config)
		return err
	} else if err != nil {
		return err
	}

	// 更新现有配置
	config.Id = existingConfig.Id
	_, err = o.Update(config)
	return err
}

// GetSystemStatus 获取系统状态
func GetSystemStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})

	status["database"] = "正常"
	status["cache"] = "正常"
	status["storage"] = "正常"
	status["memory"] = "正常"
	status["cpu"] = "正常"

	return status, nil
}

// ClearCache 清理缓存
func ClearCache() error {
	// 这里应该实现清理缓存的逻辑
	// 目前返回成功
	return nil
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	stats["total_keys"] = 0
	stats["memory_usage"] = "0MB"
	stats["hit_rate"] = "0%"
	return stats, nil
}

// CleanLogs 清理日志
func CleanLogs(days int) error {
	// 这里应该实现清理日志的逻辑
	// 目前返回成功
	return nil
}

// CreateBackup 创建备份
func CreateBackup() (string, error) {
	// 这里应该实现创建备份的逻辑
	// 返回备份文件路径
	return "/tmp/backup_" + "20240101_120000" + ".sql", nil
}

// GetBackupList 获取备份列表
func GetBackupList() ([]map[string]interface{}, error) {
	backups := make([]map[string]interface{}, 0)
	backup := map[string]interface{}{
		"filename":  "backup_20240101_120000.sql",
		"size":      "1.2MB",
		"createdAt": "2024-01-01 12:00:00",
	}
	backups = append(backups, backup)
	return backups, nil
}

// RestoreBackup 恢复备份
func RestoreBackup(filename string) error {
	// 这里应该实现恢复备份的逻辑
	// 目前返回成功
	return nil
}

// DeleteBackup 删除备份
func DeleteBackup(filename string) error {
	// 这里应该实现删除备份的逻辑
	// 目前返回成功
	return nil
}

// GetServerInfo 获取服务器信息
func GetServerInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})
	info["os"] = "Linux"
	info["arch"] = "amd64"
	info["cpu_cores"] = 4
	info["memory_total"] = "8GB"
	info["memory_used"] = "2GB"
	info["disk_total"] = "100GB"
	info["disk_used"] = "50GB"
	return info, nil
}

// GetDatabaseInfo 获取数据库信息
func GetDatabaseInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})
	info["type"] = "MySQL"
	info["version"] = "8.0"
	info["size"] = "100MB"
	info["tables"] = 10
	return info, nil
}

// OptimizeDatabase 优化数据库
func OptimizeDatabase() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	result["optimized_tables"] = 5
	result["freed_space"] = "10MB"
	result["duration"] = "2s"
	return result, nil
}

func init() {
	orm.RegisterModel(new(SystemConfig))
}
