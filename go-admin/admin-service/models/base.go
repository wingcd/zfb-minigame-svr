package models

import (
	"admin-service/utils"
	"fmt"
	"log"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/config"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

var (
	RedisClient *redis.Client
)

// BaseModel 基础模型
type BaseModel struct {
	Id        int64     `orm:"auto" json:"id"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime);column(create_time)" json:"create_time"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime);column(update_time)" json:"update_time"`
}

// Response 通用响应结构 - 兼容云函数格式
type Response struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// InitDB 初始化数据库
func InitDB() error {
	// 获取配置
	configPath := utils.FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)
	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}

	// 使用默认值处理配置获取
	mysqlHost := getConfigString(appconf, "mysql_host", "localhost")
	mysqlPort := getConfigString(appconf, "mysql_port", "3306")
	mysqlUser := getConfigString(appconf, "mysql_user", "root")
	mysqlPassword := getConfigString(appconf, "mysql_password", "")
	mysqlDatabase := getConfigString(appconf, "mysql_database", "minigame_server")
	mysqlCharset := getConfigString(appconf, "mysql_charset", "utf8mb4")

	// 构建数据库连接字符串
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase, mysqlCharset)

	// 注册数据库
	if err := orm.RegisterDataBase("default", "mysql", dataSource); err != nil {
		return fmt.Errorf("register database failed: %v", err)
	}

	// 开发模式下显示SQL
	orm.Debug = true

	return nil
}

// InitRedis 初始化Redis
func InitRedis() error {
	// 获取配置
	configPath := utils.FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)
	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}

	redisAddr := getConfigString(appconf, "redis_addr", "localhost:6379")
	redisPassword := getConfigString(appconf, "redis_password", "")
	redisDB := getConfigInt(appconf, "redis_db", 0)

	// 创建Redis客户端
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return nil
}

// getConfigString 获取字符串配置，支持默认值
func getConfigString(conf config.Configer, key, defaultValue string) string {
	if value, _ := conf.String(key); value != "" {
		return value
	}
	return defaultValue
}

// getConfigInt 获取整数配置，支持默认值
func getConfigInt(conf config.Configer, key string, defaultValue int) int {
	if value, err := conf.Int(key); err == nil {
		return value
	}
	return defaultValue
}

// CreateTables 创建数据表
func CreateTables() error {
	// 自动建表
	return orm.RunSyncdb("default", false, true)
}

func init() {
	// 注册模型
	orm.RegisterModel(new(BaseModel))

	// 初始化数据库
	if err := InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 初始化Redis
	if err := InitRedis(); err != nil {
		log.Fatal("Failed to initialize Redis:", err)
	}
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) Response {
	return Response{
		Code:      0,
		Msg:       "success",
		Timestamp: time.Now().UnixNano() / 1e6,
		Data:      data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) Response {
	return Response{
		Code:      code,
		Msg:       message,
		Timestamp: time.Now().UnixNano() / 1e6,
		Data:      nil,
	}
}

// PageResponse 分页响应
func PageResponse(list interface{}, total int64, page, pageSize int) Response {
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	return Response{
		Code:      0,
		Msg:       "success",
		Timestamp: time.Now().UnixNano() / 1e6,
		Data: map[string]interface{}{
			"list":       list,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		},
	}
}
