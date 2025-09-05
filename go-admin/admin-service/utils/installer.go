package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/config"
	_ "github.com/go-sql-driver/mysql"
)

// InstallStatus 安装状态
type InstallStatus struct {
	IsInstalled    bool   `json:"is_installed"`
	DatabaseType   string `json:"database_type"`
	DatabaseStatus string `json:"database_status"`
	AdminExists    bool   `json:"admin_exists"`
	InstallTime    string `json:"install_time"`
	Version        string `json:"version"`
}

// InstallConfig 安装配置
type InstallConfig struct {
	DatabaseType     string `json:"database_type"` // mysql, sqlite
	MySQLHost        string `json:"mysql_host"`
	MySQLPort        string `json:"mysql_port"`
	MySQLUser        string `json:"mysql_user"`
	MySQLPassword    string `json:"mysql_password"`
	MySQLDatabase    string `json:"mysql_database"`
	AdminUsername    string `json:"admin_username"`
	AdminPassword    string `json:"admin_password"`
	AdminEmail       string `json:"admin_email"`
	CreateSampleData bool   `json:"create_sample_data"`
}

var (
	installLockFile = ".installed"
	sqliteDbFile    = "data/minigame.db"
)

// CheckInstallStatus 检查安装状态
func CheckInstallStatus() *InstallStatus {
	status := &InstallStatus{
		IsInstalled:    false,
		DatabaseType:   "unknown",
		DatabaseStatus: "not_configured",
		AdminExists:    false,
		Version:        "1.0.0",
	}

	// 检查是否已安装
	if _, err := os.Stat(installLockFile); err == nil {
		status.IsInstalled = true

		// 读取安装时间
		if data, err := os.ReadFile(installLockFile); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "install_time=") {
					status.InstallTime = strings.TrimPrefix(line, "install_time=")
				}
				if strings.HasPrefix(line, "database_type=") {
					status.DatabaseType = strings.TrimPrefix(line, "database_type=")
				}
			}
		}
	}

	// 检查数据库状态
	status.DatabaseStatus = checkDatabaseStatus()

	// 检查管理员是否存在
	status.AdminExists = checkAdminExists()

	return status
}

// checkDatabaseStatus 检查数据库状态
func checkDatabaseStatus() string {
	// 尝试连接MySQL
	if testMySQLConnection() {
		return "mysql_connected"
	}

	// 当前版本不支持SQLite
	// if _, err := os.Stat(sqliteDbFile); err == nil {
	//     return "sqlite_available"
	// }

	return "not_configured"
}

// testMySQLConnection 测试MySQL连接
func testMySQLConnection() bool {
	configPath := FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)
	if err != nil {
		return false
	}

	mysqlHost := getConfigString(appconf, "mysql_host", "127.0.0.1")
	mysqlPort := getConfigString(appconf, "mysql_port", "3306")
	mysqlUser := getConfigString(appconf, "mysql_user", "root")
	mysqlPassword := getConfigString(appconf, "mysql_password", "")
	mysqlDatabase := getConfigString(appconf, "mysql_database", "minigame_admin")

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return false
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return false
	}

	return true
}

// checkAdminExists 检查管理员是否存在
func checkAdminExists() bool {
	// 这里需要根据实际的数据库表结构来检查
	// 暂时返回false，实际实现时需要查询admin表
	return false
}

// AutoInstall 自动安装
func AutoInstall() error {
	log.Println("开始自动安装...")

	// 检查是否已安装
	if status := CheckInstallStatus(); status.IsInstalled {
		log.Println("系统已安装，跳过安装过程")
		return nil
	}

	// 创建数据目录
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 尝试MySQL安装，失败则使用SQLite
	var dbType string
	var err error

	if testMySQLConnection() {
		log.Println("检测到MySQL连接，使用MySQL数据库")
		err = installWithMySQL()
		dbType = "mysql"
	} else {
		log.Println("MySQL连接失败，使用SQLite数据库")
		err = installWithSQLite()
		dbType = "sqlite"
	}

	if err != nil {
		return fmt.Errorf("数据库安装失败: %v", err)
	}

	// 创建默认管理员
	if err := createDefaultAdmin(); err != nil {
		log.Printf("创建默认管理员失败: %v", err)
		// 不返回错误，允许后续手动创建
	}

	// 创建安装锁文件
	if err := createInstallLock(dbType); err != nil {
		return fmt.Errorf("创建安装锁文件失败: %v", err)
	}

	log.Println("自动安装完成")
	return nil
}

// installWithMySQL 使用MySQL安装
func installWithMySQL() error {
	log.Println("初始化MySQL数据库...")

	configPath := FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)
	if err != nil {
		return err
	}

	mysqlHost := getConfigString(appconf, "mysql_host", "127.0.0.1")
	mysqlPort := getConfigString(appconf, "mysql_port", "3306")
	mysqlUser := getConfigString(appconf, "mysql_user", "root")
	mysqlPassword := getConfigString(appconf, "mysql_password", "")
	mysqlDatabase := getConfigString(appconf, "mysql_database", "minigame_admin")

	// 连接MySQL（不指定数据库）
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort)

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return err
	}
	defer db.Close()

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", mysqlDatabase))
	if err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	// 选择数据库
	_, err = db.Exec(fmt.Sprintf("USE %s", mysqlDatabase))
	if err != nil {
		return err
	}

	// 创建表结构
	if err := createTables(db, "mysql"); err != nil {
		return err
	}

	log.Println("MySQL数据库初始化完成")
	return nil
}

// installWithSQLite 使用SQLite安装
func installWithSQLite() error {
	log.Println("SQLite支持需要CGO，当前版本使用MySQL作为主要数据库")
	return fmt.Errorf("SQLite支持需要CGO编译，请配置MySQL数据库")
}

// createTables 创建数据表
func createTables(db *sql.DB, dbType string) error {
	log.Println("创建数据表...")

	var tables []string

	if dbType == "mysql" {
		tables = getMySQLTables()
	} else {
		tables = getSQLiteTables()
	}

	for _, tableSQL := range tables {
		if _, err := db.Exec(tableSQL); err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}

	log.Println("数据表创建完成")
	return nil
}

// getMySQLTables 获取MySQL表结构
func getMySQLTables() []string {
	return []string{
		// 管理员表
		`CREATE TABLE IF NOT EXISTS admins (
			id INT PRIMARY KEY AUTO_INCREMENT,
			username VARCHAR(50) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			email VARCHAR(100),
			role VARCHAR(20) DEFAULT 'admin',
			status TINYINT DEFAULT 1,
			last_login DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_username (username),
			INDEX idx_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 应用表
		`CREATE TABLE IF NOT EXISTS apps (
			id INT PRIMARY KEY AUTO_INCREMENT,
			app_id VARCHAR(50) NOT NULL UNIQUE,
			app_name VARCHAR(100) NOT NULL,
			description TEXT,
			status TINYINT DEFAULT 1,
			user_count INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_app_id (app_id),
			INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 用户数据表
		`CREATE TABLE IF NOT EXISTS user_data (
			id INT PRIMARY KEY AUTO_INCREMENT,
			app_id VARCHAR(50) NOT NULL,
			player_id VARCHAR(50) NOT NULL,
			data JSON,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_app_player (app_id, player_id),
			INDEX idx_app_id (app_id),
			INDEX idx_player_id (player_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 排行榜表
		`CREATE TABLE IF NOT EXISTS leaderboards (
			id INT PRIMARY KEY AUTO_INCREMENT,
			app_id VARCHAR(50) NOT NULL,
			leaderboard_id VARCHAR(50) NOT NULL,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			sort_order ENUM('asc', 'desc') DEFAULT 'desc',
			max_entries INT DEFAULT 100,
			status TINYINT DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_app_leaderboard (app_id, leaderboard_id),
			INDEX idx_app_id (app_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 系统配置表
		`CREATE TABLE IF NOT EXISTS system_configs (
			id INT PRIMARY KEY AUTO_INCREMENT,
			config_key VARCHAR(100) NOT NULL UNIQUE,
			config_value TEXT,
			description VARCHAR(255),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_key (config_key)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
}

// getSQLiteTables 获取SQLite表结构
func getSQLiteTables() []string {
	return []string{
		// 管理员表
		`CREATE TABLE IF NOT EXISTS admins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			email TEXT,
			role TEXT DEFAULT 'admin',
			status INTEGER DEFAULT 1,
			last_login DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 应用表
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id TEXT NOT NULL UNIQUE,
			app_name TEXT NOT NULL,
			description TEXT,
			status INTEGER DEFAULT 1,
			user_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 用户数据表
		`CREATE TABLE IF NOT EXISTS user_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id TEXT NOT NULL,
			player_id TEXT NOT NULL,
			data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(app_id, player_id)
		)`,

		// 排行榜表
		`CREATE TABLE IF NOT EXISTS leaderboards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id TEXT NOT NULL,
			leaderboard_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			sort_order TEXT DEFAULT 'desc',
			max_entries INTEGER DEFAULT 100,
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(app_id, leaderboard_id)
		)`,

		// 系统配置表
		`CREATE TABLE IF NOT EXISTS system_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			config_key TEXT NOT NULL UNIQUE,
			config_value TEXT,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 创建索引
		`CREATE INDEX IF NOT EXISTS idx_admins_username ON admins(username)`,
		`CREATE INDEX IF NOT EXISTS idx_apps_app_id ON apps(app_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_data_app_id ON user_data(app_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_data_player_id ON user_data(player_id)`,
		`CREATE INDEX IF NOT EXISTS idx_leaderboards_app_id ON leaderboards(app_id)`,
	}
}

// createDefaultAdmin 创建默认管理员
func createDefaultAdmin() error {
	log.Println("创建默认管理员...")

	// 读取配置
	configPath := FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)
	if err != nil {
		return err
	}

	username := getConfigString(appconf, "default_admin_username", "admin")
	password := getConfigString(appconf, "default_admin_password", "admin123")

	// 加密密码 (简化实现，实际应该使用bcrypt)
	hashedPassword := password

	// 连接数据库
	var db *sql.DB
	if testMySQLConnection() {
		// 使用MySQL
		mysqlHost := getConfigString(appconf, "mysql_host", "127.0.0.1")
		mysqlPort := getConfigString(appconf, "mysql_port", "3306")
		mysqlUser := getConfigString(appconf, "mysql_user", "root")
		mysqlPassword := getConfigString(appconf, "mysql_password", "")
		mysqlDatabase := getConfigString(appconf, "mysql_database", "minigame_admin")

		dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)

		db, err = sql.Open("mysql", dataSource)
	} else {
		return fmt.Errorf("MySQL连接失败，当前版本需要MySQL数据库")
	}

	if err != nil {
		return err
	}
	defer db.Close()

	// 检查管理员是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admins WHERE username = ?", username).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Printf("管理员 %s 已存在，跳过创建", username)
		return nil
	}

	// 插入默认管理员
	_, err = db.Exec(`
		INSERT INTO admins (username, password, email, role, status, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		username, hashedPassword, "admin@example.com", "super_admin", 1, time.Now())

	if err != nil {
		return err
	}

	log.Printf("默认管理员创建成功: %s", username)
	return nil
}

// createInstallLock 创建安装锁文件
func createInstallLock(dbType string) error {
	content := fmt.Sprintf(`install_time=%s
database_type=%s
version=1.0.0
`, time.Now().Format("2006-01-02 15:04:05"), dbType)

	return os.WriteFile(installLockFile, []byte(content), 0644)
}

// ManualInstall 手动安装
func ManualInstall(config *InstallConfig) error {
	log.Println("开始手动安装...")

	// 检查是否已安装
	if status := CheckInstallStatus(); status.IsInstalled {
		return fmt.Errorf("系统已安装，请先卸载")
	}

	// 创建目录
	for _, dir := range []string{"data", "logs", "uploads"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", dir, err)
		}
	}

	// 根据配置安装数据库
	var err error
	if config.DatabaseType == "mysql" {
		err = installMySQLWithConfig(config)
	} else {
		err = installWithSQLite()
	}

	if err != nil {
		return fmt.Errorf("数据库安装失败: %v", err)
	}

	// 创建管理员
	if err := createAdminWithConfig(config); err != nil {
		return fmt.Errorf("创建管理员失败: %v", err)
	}

	// 创建安装锁
	if err := createInstallLock(config.DatabaseType); err != nil {
		return fmt.Errorf("创建安装锁失败: %v", err)
	}

	log.Println("手动安装完成")
	return nil
}

// installMySQLWithConfig 使用配置安装MySQL
func installMySQLWithConfig(config *InstallConfig) error {
	// 连接MySQL
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		config.MySQLUser, config.MySQLPassword, config.MySQLHost, config.MySQLPort)

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return err
	}
	defer db.Close()

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", config.MySQLDatabase))
	if err != nil {
		return err
	}

	// 选择数据库
	_, err = db.Exec(fmt.Sprintf("USE %s", config.MySQLDatabase))
	if err != nil {
		return err
	}

	// 创建表
	return createTables(db, "mysql")
}

// createAdminWithConfig 使用配置创建管理员
func createAdminWithConfig(config *InstallConfig) error {
	// 简化实现，实际应该使用bcrypt
	hashedPassword := config.AdminPassword

	var db *sql.DB
	var err error
	if config.DatabaseType == "mysql" {
		dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			config.MySQLUser, config.MySQLPassword, config.MySQLHost, config.MySQLPort, config.MySQLDatabase)
		db, err = sql.Open("mysql", dataSource)
	} else {
		return fmt.Errorf("当前版本仅支持MySQL数据库")
	}

	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO admins (username, password, email, role, status, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		config.AdminUsername, hashedPassword, config.AdminEmail, "super_admin", 1, time.Now())

	return err
}

// Uninstall 卸载
func Uninstall() error {
	log.Println("开始卸载...")

	// 删除SQLite数据库文件
	if err := os.RemoveAll("data"); err != nil {
		log.Printf("删除数据目录失败: %v", err)
	}

	// 删除日志文件
	if err := os.RemoveAll("logs"); err != nil {
		log.Printf("删除日志目录失败: %v", err)
	}

	// 删除上传文件
	if err := os.RemoveAll("uploads"); err != nil {
		log.Printf("删除上传目录失败: %v", err)
	}

	// 删除安装锁文件
	if err := os.Remove(installLockFile); err != nil {
		log.Printf("删除安装锁文件失败: %v", err)
	}

	log.Println("卸载完成")
	return nil
}

// getConfigString 获取配置字符串
func getConfigString(conf config.Configer, key, defaultValue string) string {
	if value, _ := conf.String(key); value != "" {
		return value
	}
	return defaultValue
}
