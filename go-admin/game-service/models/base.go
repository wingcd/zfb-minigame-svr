package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/config"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

var (
	DB          *sql.DB
	RedisClient *redis.Client
)

// BaseModel 基础模型
type BaseModel struct {
	Id        int64     `orm:"auto" json:"id"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"create_time"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"update_time"`
}

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// 初始化数据库连接
func init() {
	// 注册数据库驱动
	orm.RegisterDriver("mysql", orm.DRMySQL)

	// 获取配置
	appconf, _ := config.NewConfig("ini", "conf/app.conf")

	mysqlHost := appconf.DefaultString("mysql_host", "localhost")
	mysqlPort := appconf.DefaultString("mysql_port", "3306")
	mysqlUser := appconf.DefaultString("mysql_user", "root")
	mysqlPassword := appconf.DefaultString("mysql_password", "")
	mysqlDatabase := appconf.DefaultString("mysql_database", "minigame_game")
	mysqlCharset := appconf.DefaultString("mysql_charset", "utf8mb4")

	// 构建数据库连接字符串
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase, mysqlCharset)

	// 注册模型
	orm.RegisterModel(new(Application))

	// 注册数据库
	orm.RegisterDataBase("default", "mysql", dataSource)

	// 初始化Redis连接
	initRedis()
}

// 初始化Redis连接
func initRedis() {
	appconf, _ := config.NewConfig("ini", "conf/app.conf")

	redisHost := appconf.DefaultString("redis_host", "localhost")
	redisPort := appconf.DefaultString("redis_port", "6379")
	redisPassword := appconf.DefaultString("redis_password", "")
	redisDatabase, _ := appconf.Int("redis_database")

	// 创建Redis客户端选项
	options := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
		DB:   redisDatabase,
	}

	// 只有在密码不为空时才设置密码
	if redisPassword != "" {
		options.Password = redisPassword
	}

	RedisClient = redis.NewClient(options)
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// PageResponse 分页响应
func PageResponse(list interface{}, total int64, page, pageSize int) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	}
}

// GetDynamicTableName 获取动态表名
func GetDynamicTableName(appId, tableName string) string {
	return fmt.Sprintf("%s_%s", tableName, appId)
}
