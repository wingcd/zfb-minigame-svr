package config

import (
	"fmt"
	"log"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/config"
	_ "github.com/go-sql-driver/mysql"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Charset  string `json:"charset"`
	MaxIdle  int    `json:"max_idle"`
	MaxOpen  int    `json:"max_open"`
}

// GetDSN 获取数据库连接字符串
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Database, d.Charset)
}

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *DatabaseConfig) error {
	// 注册数据库驱动
	err := orm.RegisterDriver(cfg.Driver, orm.DRMySQL)
	if err != nil {
		return fmt.Errorf("failed to register database driver: %v", err)
	}

	// 注册数据库连接
	err = orm.RegisterDataBase("default", cfg.Driver, cfg.GetDSN())
	if err != nil {
		return fmt.Errorf("failed to register database: %v", err)
	}

	// 设置连接池参数
	orm.SetMaxIdleConns("default", cfg.MaxIdle)
	orm.SetMaxOpenConns("default", cfg.MaxOpen)

	// 设置调试模式
	if runmode, _ := config.String("runmode"); runmode == "dev" {
		orm.Debug = true
	}

	log.Printf("Database connected successfully to %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)
	return nil
}

// CreateTables 创建数据表
func CreateTables(force bool, verbose bool) error {
	return orm.RunSyncdb("default", force, verbose)
}

// GetDefaultDatabaseConfig 获取默认数据库配置
func GetDefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "",
		Database: "game_admin",
		Charset:  "utf8mb4",
		MaxIdle:  10,
		MaxOpen:  100,
	}
}

// LoadDatabaseConfigFromEnv 从环境变量加载数据库配置
func LoadDatabaseConfigFromEnv() *DatabaseConfig {
	cfg := GetDefaultDatabaseConfig()

	if host, _ := config.String("db.host"); host != "" {
		cfg.Host = host
	}

	if port := config.DefaultInt("db.port", 3306); port > 0 {
		cfg.Port = port
	}

	if user, _ := config.String("db.user"); user != "" {
		cfg.User = user
	}

	if password, _ := config.String("db.password"); password != "" {
		cfg.Password = password
	}

	if database, _ := config.String("db.database"); database != "" {
		cfg.Database = database
	}

	if charset, _ := config.String("db.charset"); charset != "" {
		cfg.Charset = charset
	}

	if maxIdle := config.DefaultInt("db.max_idle", 10); maxIdle > 0 {
		cfg.MaxIdle = maxIdle
	}

	if maxOpen := config.DefaultInt("db.max_open", 100); maxOpen > 0 {
		cfg.MaxOpen = maxOpen
	}

	return cfg
}
