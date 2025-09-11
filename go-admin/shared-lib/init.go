package sharedlib

import (
	"shared-lib/config"
	"shared-lib/models"
)

// SharedLibConfig 共享库配置
type SharedLibConfig struct {
	Database *config.DatabaseConfig
	Redis    *config.RedisConfig
}

// Init 初始化共享库
func Init(cfg *SharedLibConfig) error {
	// 初始化数据库
	if cfg.Database != nil {
		if err := config.InitDatabase(cfg.Database); err != nil {
			return err
		}
	}

	// 初始化Redis
	if cfg.Redis != nil {
		if _, err := config.InitRedis(cfg.Redis); err != nil {
			return err
		}
	}

	// 注册所有模型
	return RegisterAllModels()
}

// RegisterAllModels 注册所有共享模型
func RegisterAllModels() error {
	return models.RegisterModels(
		new(models.BaseModel),
		new(models.Application),
		new(models.LeaderboardConfig),
		new(models.Leaderboard),
		new(models.Statistics),
	)
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *SharedLibConfig {
	return &SharedLibConfig{
		Database: config.GetDefaultDatabaseConfig(),
		Redis:    config.GetDefaultRedisConfig(),
	}
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *SharedLibConfig {
	return &SharedLibConfig{
		Database: config.LoadDatabaseConfigFromEnv(),
		Redis:    config.LoadRedisConfigFromEnv(),
	}
}
