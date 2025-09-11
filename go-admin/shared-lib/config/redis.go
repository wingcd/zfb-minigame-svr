package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/beego/beego/v2/core/config"
	"github.com/go-redis/redis/v8"
)

// RedisConfig Redis配置结构
type RedisConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Password     string        `json:"password"`
	Database     int           `json:"database"`
	PoolSize     int           `json:"pool_size"`
	MinIdleConns int           `json:"min_idle_conns"`
	DialTimeout  time.Duration `json:"dial_timeout"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	PoolTimeout  time.Duration `json:"pool_timeout"`
}

// GetAddress 获取Redis地址
func (r *RedisConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// RedisClient Redis客户端管理器
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

var (
	defaultRedisClient *RedisClient
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *RedisConfig) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.GetAddress(),
		Password:     cfg.Password,
		DB:           cfg.Database,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolTimeout:  cfg.PoolTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	redisClient := &RedisClient{
		client: rdb,
		config: cfg,
	}

	defaultRedisClient = redisClient
	log.Printf("Redis connected successfully to %s", cfg.GetAddress())
	return redisClient, nil
}

// GetClient 获取Redis客户端
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Close 关闭Redis连接
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetDefaultRedisClient 获取默认Redis客户端
func GetDefaultRedisClient() *RedisClient {
	return defaultRedisClient
}

// GetDefaultRedisConfig 获取默认Redis配置
func GetDefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		Database:     0,
		PoolSize:     10,
		MinIdleConns: 3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	}
}

// LoadRedisConfigFromEnv 从环境变量加载Redis配置
func LoadRedisConfigFromEnv() *RedisConfig {
	cfg := GetDefaultRedisConfig()

	if host, _ := config.String("redis.host"); host != "" {
		cfg.Host = host
	}

	if port := config.DefaultInt("redis.port", 6379); port > 0 {
		cfg.Port = port
	}

	if password, _ := config.String("redis.password"); password != "" {
		cfg.Password = password
	}

	if database := config.DefaultInt("redis.database", 0); database >= 0 {
		cfg.Database = database
	}

	if poolSize := config.DefaultInt("redis.pool_size", 10); poolSize > 0 {
		cfg.PoolSize = poolSize
	}

	if minIdleConns := config.DefaultInt("redis.min_idle_conns", 3); minIdleConns > 0 {
		cfg.MinIdleConns = minIdleConns
	}

	// 超时配置（秒）
	if dialTimeout := config.DefaultInt("redis.dial_timeout", 5); dialTimeout > 0 {
		cfg.DialTimeout = time.Duration(dialTimeout) * time.Second
	}

	if readTimeout := config.DefaultInt("redis.read_timeout", 3); readTimeout > 0 {
		cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	}

	if writeTimeout := config.DefaultInt("redis.write_timeout", 3); writeTimeout > 0 {
		cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second
	}

	if poolTimeout := config.DefaultInt("redis.pool_timeout", 4); poolTimeout > 0 {
		cfg.PoolTimeout = time.Duration(poolTimeout) * time.Second
	}

	return cfg
}

// CacheKey 缓存键管理
type CacheKey struct {
	prefix string
}

// NewCacheKey 创建缓存键管理器
func NewCacheKey(prefix string) *CacheKey {
	return &CacheKey{prefix: prefix}
}

// Leaderboard 排行榜缓存键
func (c *CacheKey) Leaderboard(appId string) string {
	return fmt.Sprintf("%s:leaderboard:%s", c.prefix, appId)
}

// UserRank 用户排名缓存键
func (c *CacheKey) UserRank(appId, userId string) string {
	return fmt.Sprintf("%s:user_rank:%s:%s", c.prefix, appId, userId)
}

// Statistics 统计缓存键
func (c *CacheKey) Statistics(appId, statType, date string) string {
	return fmt.Sprintf("%s:stats:%s:%s:%s", c.prefix, appId, statType, date)
}

// Session 会话缓存键
func (c *CacheKey) Session(sessionId string) string {
	return fmt.Sprintf("%s:session:%s", c.prefix, sessionId)
}

// Lock 分布式锁键
func (c *CacheKey) Lock(resource string) string {
	return fmt.Sprintf("%s:lock:%s", c.prefix, resource)
}
