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
	appconf         config.Configer
)

// CheckInstallStatus 检查安装状态
func CheckInstallStatus() *InstallStatus {
	cfg, err := config.NewConfig("ini", FindConfigFile())
	appconf = cfg
	if err != nil {
		return nil
	}

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

func testCeateMySQLDB() bool {
	mysqlHost := getConfigString(appconf, "mysql_host", "127.0.0.1")
	mysqlPort := getConfigString(appconf, "mysql_port", "3306")
	mysqlUser := getConfigString(appconf, "mysql_user", "root")
	mysqlPassword := getConfigString(appconf, "mysql_password", "")
	mysqlDatabase := getConfigString(appconf, "mysql_database", "minigame_admin")

	// 先连接到MySQL根目录，不指定数据库
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort)

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return false
	}
	defer db.Close()

	// 测试基础连接
	if err := db.Ping(); err != nil {
		return false
	}

	// 尝试创建数据库（如果不存在）
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", mysqlDatabase))
	if err != nil {
		return false
	}

	// 现在测试连接到指定数据库
	dataSourceWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)

	dbWithDB, err := sql.Open("mysql", dataSourceWithDB)
	if err != nil {
		return false
	}
	defer dbWithDB.Close()

	if err := dbWithDB.Ping(); err != nil {
		return false
	}

	return true
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

	// 先连接到MySQL根目录，不指定数据库
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort)

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return false
	}
	defer db.Close()

	// 测试基础连接
	if err := db.Ping(); err != nil {
		return false
	}

	// 检查数据库是否存在，如果不存在则尝试创建
	var dbExists int
	err = db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", mysqlDatabase).Scan(&dbExists)
	if err != nil {
		return false
	}

	if dbExists == 0 {
		// 数据库不存在，尝试创建
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", mysqlDatabase))
		if err != nil {
			return false
		}
	}

	// 现在测试连接到指定数据库
	dataSourceWithDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)

	dbWithDB, err := sql.Open("mysql", dataSourceWithDB)
	if err != nil {
		return false
	}
	defer dbWithDB.Close()

	if err := dbWithDB.Ping(); err != nil {
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

func getMySQLDB() (*sql.DB, error) {
	configPath := FindConfigFile()
	appconf, err := config.NewConfig("ini", configPath)
	if err != nil {
		return nil, err
	}

	mysqlHost := getConfigString(appconf, "mysql_host", "127.0.0.1")
	mysqlPort := getConfigString(appconf, "mysql_port", "3306")
	mysqlUser := getConfigString(appconf, "mysql_user", "root")
	mysqlPassword := getConfigString(appconf, "mysql_password", "")
	mysqlDatabase := getConfigString(appconf, "mysql_database", "minigame_admin")

	// 先连接到MySQL根目录，确保数据库存在
	rootDataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort)

	rootDB, err := sql.Open("mysql", rootDataSource)
	if err != nil {
		return nil, err
	}
	defer rootDB.Close()

	// 测试连接
	if err := rootDB.Ping(); err != nil {
		return nil, fmt.Errorf("MySQL连接失败: %v", err)
	}

	// 确保数据库存在
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", mysqlDatabase))
	if err != nil {
		return nil, fmt.Errorf("创建数据库失败: %v", err)
	}

	// 现在连接到指定数据库
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDatabase)

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// AutoInstallWithParams 使用指定参数自动安装
func AutoInstallWithParams(username, password string) error {
	log.Println("开始系统初始化...")

	var db *sql.DB
	var err error
	if testMySQLConnection() {
		db, err = getMySQLDB()
		if err != nil {
			return fmt.Errorf("获取MySQL连接失败: %v", err)
		}
		defer db.Close()
	} else {
		return fmt.Errorf("MySQL连接失败，当前版本需要MySQL数据库")
	}

	// 1. 创建数据库表
	if err = createTables(db, "mysql"); err != nil {
		return fmt.Errorf("创建数据库表失败: %v", err)
	}

	// 2. 创建默认角色
	if err = createDefaultRoles(db); err != nil {
		return fmt.Errorf("创建默认角色失败: %v", err)
	}

	// 3. 创建默认管理员
	if err = createDefaultAdminWithParams(db, username, password); err != nil {
		return fmt.Errorf("创建默认管理员失败: %v", err)
	}

	log.Println("系统初始化完成")
	return nil
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
	var dbType string = getConfigString(appconf, "database_type", "mysql")
	var err error

	if dbType == "mysql" {
		if testCeateMySQLDB() {
			err = installWithMySQL()
		} else {
			log.Println("MySQL连接失败，使用SQLite数据库")
			err = installWithMySQL()
		}
	} else {
		log.Println("数据库类型错误，使用SQLite数据库")
		err = installWithSQLite()
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

	db, err := getMySQLDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", "minigame_admin"))
	if err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	// 选择数据库
	_, err = db.Exec(fmt.Sprintf("USE %s", "minigame_admin"))
	if err != nil {
		return err
	}

	// 创建表结构
	if err := createTables(db, "mysql"); err != nil {
		return err
	}

	// 创建默认角色
	if err := createDefaultRoles(db); err != nil {
		return fmt.Errorf("创建默认角色失败: %v", err)
	}

	// 创建默认管理员
	if err := createDefaultAdmin(); err != nil {
		return fmt.Errorf("创建默认管理员失败: %v", err)
	}

	log.Println("MySQL数据库初始化完成")
	return nil
}

// installWithSQLite 使用SQLite安装
func installWithSQLite() error {
	log.Println("初始化SQLite数据库...")

	// 连接SQLite数据库（如果文件不存在会自动创建）
	db, err := sql.Open("sqlite3", sqliteDbFile)
	if err != nil {
		return fmt.Errorf("连接SQLite数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("SQLite数据库连接测试失败: %v", err)
	}

	// 创建表结构
	if err := createTables(db, "sqlite"); err != nil {
		return fmt.Errorf("创建SQLite表结构失败: %v", err)
	}

	// 创建默认角色
	if err := createDefaultRoles(db); err != nil {
		return fmt.Errorf("创建默认角色失败: %v", err)
	}

	log.Println("SQLite数据库初始化完成")
	return nil
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
		// 管理员表 - 匹配AdminUser模型
		`CREATE TABLE IF NOT EXISTS admin_users (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			createTime DATETIME NOT NULL,
			updateTime DATETIME NOT NULL,
			username VARCHAR(50) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			email VARCHAR(100) NOT NULL DEFAULT '',
			phone VARCHAR(20) NOT NULL DEFAULT '',
			realName VARCHAR(50) NOT NULL DEFAULT '',
			avatar VARCHAR(255) NOT NULL DEFAULT '',
			status INT NOT NULL DEFAULT 1,
			lastLoginAt DATETIME NULL,
			lastLoginIp VARCHAR(50) NOT NULL DEFAULT '',
			roleId BIGINT NOT NULL DEFAULT 0,
			token VARCHAR(128) NULL,
			tokenExpire DATETIME NULL,
			INDEX idx_username (username),
			INDEX idx_email (email),
			INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 角色表 - 匹配AdminRole模型
		`CREATE TABLE IF NOT EXISTS admin_roles (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			createTime DATETIME NOT NULL,
			updateTime DATETIME NOT NULL,
			roleCode VARCHAR(50) NULL,
			roleName VARCHAR(50) NULL,
			name VARCHAR(50) NOT NULL UNIQUE,
			description VARCHAR(255) DEFAULT '',
			permissions TEXT,
			status INT NOT NULL DEFAULT 1,
			INDEX idx_name (name),
			INDEX idx_role_code (roleCode),
			INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 应用表
		`CREATE TABLE IF NOT EXISTS apps (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			createTime DATETIME NOT NULL,
			updateTime DATETIME NOT NULL,
			appId VARCHAR(50) NOT NULL UNIQUE COMMENT '应用ID（唯一）',
			appName VARCHAR(100) NOT NULL COMMENT '应用名称',
			description TEXT COMMENT '应用描述',
			channelAppId VARCHAR(100) NOT NULL DEFAULT '' COMMENT '渠道应用ID',
			channelAppKey VARCHAR(100) NOT NULL DEFAULT '' COMMENT '渠道应用密钥',
			appSecret VARCHAR(100) NOT NULL DEFAULT '' COMMENT '应用密钥',
			category VARCHAR(50) NOT NULL DEFAULT 'game' COMMENT '应用分类: game/tool/social',
			platform VARCHAR(50) NOT NULL DEFAULT '' COMMENT '平台: alipay/wechat/baidu',
			status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态: active/inactive/pending',
			version VARCHAR(50) NOT NULL DEFAULT '1.0.0' COMMENT '当前版本',
			minVersion VARCHAR(50) NOT NULL DEFAULT '1.0.0' COMMENT '最低支持版本',
			settings TEXT COMMENT '应用设置(JSON格式)',
			userCount BIGINT NOT NULL DEFAULT 0 COMMENT '用户数量',
			scoreCount BIGINT NOT NULL DEFAULT 0 COMMENT '分数记录数',
			dailyActive BIGINT NOT NULL DEFAULT 0 COMMENT '日活跃用户',
			monthlyActive BIGINT NOT NULL DEFAULT 0 COMMENT '月活跃用户',
			createdBy VARCHAR(50) NOT NULL DEFAULT '' COMMENT '创建者',
			INDEX idx_app_id (appId),
			INDEX idx_status (status),
			INDEX idx_category (category),
			INDEX idx_platform (platform),
			INDEX idx_created_by (createdBy)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 用户数据表
		`CREATE TABLE IF NOT EXISTS user_data (
			id INT PRIMARY KEY AUTO_INCREMENT,
			appId VARCHAR(50) NOT NULL,
			playerId VARCHAR(50) NOT NULL,
			data JSON,
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_app_player (appId, playerId),
			INDEX idx_app_id (appId),
			INDEX idx_player_id (playerId)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 系统配置表
		`CREATE TABLE IF NOT EXISTS system_configs (
			id INT PRIMARY KEY AUTO_INCREMENT,
			configKey VARCHAR(100) NOT NULL UNIQUE,
			configValue TEXT,
			description VARCHAR(255),
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_key (configKey)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		// 管理员操作日志表
		`CREATE TABLE IF NOT EXISTS admin_operation_logs (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			createTime DATETIME NOT NULL,
			updateTime DATETIME NOT NULL,
			userId BIGINT NOT NULL,
			username VARCHAR(50) NOT NULL,
			action VARCHAR(100) NOT NULL,
			resource VARCHAR(100) NOT NULL DEFAULT '',
			params TEXT,
			ipAddress VARCHAR(45) NOT NULL DEFAULT '',
			userAgent VARCHAR(500) NOT NULL DEFAULT '',
			INDEX idx_user_id (userId),
			INDEX idx_username (username),
			INDEX idx_action (action),
			INDEX idx_create_time (createTime)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
}

// getSQLiteTables 获取SQLite表结构（与MySQL对齐）
func getSQLiteTables() []string {
	return []string{
		// 管理员表 - 对齐AdminUser模型
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			email TEXT NOT NULL DEFAULT '',
			phone TEXT NOT NULL DEFAULT '',
			realName TEXT NOT NULL DEFAULT '',
			avatar TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1,
			lastLoginAt DATETIME NULL,
			lastLoginIp TEXT NOT NULL DEFAULT '',
			roleId INTEGER NOT NULL DEFAULT 0,
			token TEXT NULL,
			tokenExpire DATETIME NULL
		)`,

		// 角色表 - 对齐AdminRole模型
		`CREATE TABLE IF NOT EXISTS admin_roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			roleCode TEXT NULL,
			roleName TEXT NULL,
			name TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			permissions TEXT,
			status INTEGER NOT NULL DEFAULT 1
		)`,

		// 权限表 - 对齐AdminPermission模型
		`CREATE TABLE IF NOT EXISTS admin_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			permissionCode TEXT NOT NULL UNIQUE,
			permissionName TEXT NOT NULL,
			description TEXT DEFAULT '',
			status INTEGER NOT NULL DEFAULT 1
		)`,

		// 角色权限关联表
		`CREATE TABLE IF NOT EXISTS admin_role_permissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			roleId INTEGER NOT NULL,
			permissionId INTEGER NOT NULL,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(roleId, permissionId)
		)`,

		// 用户角色关联表
		`CREATE TABLE IF NOT EXISTS admin_user_roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			userId INTEGER NOT NULL,
			roleId INTEGER NOT NULL,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(userId, roleId)
		)`,

		// 应用表 - 对齐MySQL版本
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			appId TEXT NOT NULL UNIQUE,
			appName TEXT NOT NULL,
			description TEXT,
			channelAppId TEXT NOT NULL DEFAULT '',
			channelAppKey TEXT NOT NULL DEFAULT '',
			appSecret TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT 'game',
			platform TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'active',
			version TEXT NOT NULL DEFAULT '1.0.0',
			minVersion TEXT NOT NULL DEFAULT '1.0.0',
			settings TEXT,
			userCount INTEGER NOT NULL DEFAULT 0,
			scoreCount INTEGER NOT NULL DEFAULT 0,
			dailyActive INTEGER NOT NULL DEFAULT 0,
			monthlyActive INTEGER NOT NULL DEFAULT 0,
			createdBy TEXT NOT NULL DEFAULT ''
		)`,

		// 用户数据表 - 对齐MySQL版本
		`CREATE TABLE IF NOT EXISTS user_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			appId TEXT NOT NULL,
			playerId TEXT NOT NULL,
			data TEXT,
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(appId, playerId)
		)`,

		// 分数记录表 - 对齐MySQL版本
		`CREATE TABLE IF NOT EXISTS score_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			appId TEXT NOT NULL,
			playerId TEXT NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			level INTEGER NOT NULL DEFAULT 1,
			gameTime INTEGER NOT NULL DEFAULT 0,
			extras TEXT,
			ipAddress TEXT NOT NULL DEFAULT '',
			userAgent TEXT NOT NULL DEFAULT '',
			platform TEXT NOT NULL DEFAULT '',
			version TEXT NOT NULL DEFAULT '1.0.0'
		)`,

		// 系统配置表 - 对齐MySQL版本
		`CREATE TABLE IF NOT EXISTS system_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			configKey TEXT NOT NULL UNIQUE,
			configValue TEXT,
			description TEXT,
			createTime DATETIME DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// 操作日志表 - 对齐MySQL版本
		`CREATE TABLE IF NOT EXISTS operation_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			createTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updateTime DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			userId INTEGER NOT NULL DEFAULT 0,
			username TEXT NOT NULL DEFAULT '',
			action TEXT NOT NULL,
			resource TEXT NOT NULL DEFAULT '',
			resourceId TEXT NOT NULL DEFAULT '',
			details TEXT,
			ipAddress TEXT NOT NULL DEFAULT '',
			userAgent TEXT NOT NULL DEFAULT '',
			result INTEGER NOT NULL DEFAULT 1,
			errorMessage TEXT DEFAULT ''
		)`,

		// 创建索引 - 对齐MySQL版本
		`CREATE INDEX IF NOT EXISTS idx_admin_users_username ON admin_users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_users_status ON admin_users(status)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_roles_name ON admin_roles(name)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_roles_role_code ON admin_roles(roleCode)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_roles_status ON admin_roles(status)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_permissions_permission_code ON admin_permissions(permissionCode)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_permissions_status ON admin_permissions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_role_permissions_role_id ON admin_role_permissions(roleId)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_role_permissions_permission_id ON admin_role_permissions(permissionId)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_user_roles_user_id ON admin_user_roles(userId)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_user_roles_role_id ON admin_user_roles(roleId)`,
		`CREATE INDEX IF NOT EXISTS idx_apps_app_id ON apps(appId)`,
		`CREATE INDEX IF NOT EXISTS idx_apps_status ON apps(status)`,
		`CREATE INDEX IF NOT EXISTS idx_apps_category ON apps(category)`,
		`CREATE INDEX IF NOT EXISTS idx_apps_platform ON apps(platform)`,
		`CREATE INDEX IF NOT EXISTS idx_apps_created_by ON apps(createdBy)`,
		`CREATE INDEX IF NOT EXISTS idx_user_data_app_id ON user_data(appId)`,
		`CREATE INDEX IF NOT EXISTS idx_user_data_player_id ON user_data(playerId)`,
		`CREATE INDEX IF NOT EXISTS idx_score_records_app_id ON score_records(appId)`,
		`CREATE INDEX IF NOT EXISTS idx_score_records_player_id ON score_records(playerId)`,
		`CREATE INDEX IF NOT EXISTS idx_score_records_score ON score_records(score)`,
		`CREATE INDEX IF NOT EXISTS idx_score_records_create_time ON score_records(createTime)`,
		`CREATE INDEX IF NOT EXISTS idx_system_configs_config_key ON system_configs(configKey)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(userId)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_action ON operation_logs(action)`,
		`CREATE INDEX IF NOT EXISTS idx_operation_logs_create_time ON operation_logs(createTime)`,
	}
}

// createDefaultRoles 创建默认角色
func createDefaultRoles(db *sql.DB) error {
	log.Println("创建默认角色...")

	// 定义默认角色
	defaultRoles := []struct {
		ID          int64
		RoleCode    string
		RoleName    string
		Name        string
		Description string
		Permissions string
		Status      int
	}{
		{
			ID:          1,
			RoleCode:    "super_admin",
			RoleName:    "超级管理员",
			Name:        "超级管理员",
			Description: "拥有所有权限的超级管理员",
			Permissions: `["admin_manage","role_manage","app_manage","user_manage","leaderboard_manage","mail_manage","stats_view","system_config","counter_manage"]`,
			Status:      1,
		},
		{
			ID:          2,
			RoleCode:    "admin",
			RoleName:    "管理员",
			Name:        "管理员",
			Description: "普通管理员，拥有大部分管理权限",
			Permissions: `["app_manage","user_manage","leaderboard_manage","mail_manage","stats_view"]`,
			Status:      1,
		},
		{
			ID:          3,
			RoleCode:    "operator",
			RoleName:    "运营人员",
			Name:        "运营人员",
			Description: "运营人员，拥有内容管理权限",
			Permissions: `["user_manage","leaderboard_manage","mail_manage","stats_view"]`,
			Status:      1,
		},
		{
			ID:          4,
			RoleCode:    "viewer",
			RoleName:    "查看者",
			Name:        "查看者",
			Description: "只读权限，可以查看统计数据",
			Permissions: `["stats_view"]`,
			Status:      1,
		},
	}

	now := time.Now()

	for _, role := range defaultRoles {
		// 检查角色是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM admin_roles WHERE id = ?", role.ID).Scan(&count)
		if err != nil {
			return fmt.Errorf("检查角色是否存在失败: %v", err)
		}

		if count > 0 {
			log.Printf("角色 %s 已存在，跳过创建", role.RoleName)
			continue
		}

		// 插入角色
		_, err = db.Exec(`
			INSERT INTO admin_roles (id, createTime, updateTime, roleCode, roleName, name, description, permissions, status) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			role.ID, now, now, role.RoleCode, role.RoleName, role.Name, role.Description, role.Permissions, role.Status)

		if err != nil {
			return fmt.Errorf("创建角色 %s 失败: %v", role.RoleName, err)
		}

		log.Printf("默认角色 %s 创建成功", role.RoleName)
	}

	return nil
}

// createDefaultAdminWithParams 使用指定参数创建默认管理员
func createDefaultAdminWithParams(db *sql.DB, username, password string) error {
	log.Printf("创建默认管理员: %s", username)

	// MD5加密密码 (对齐登录验证逻辑)
	hashedPassword := HashPassword(password)

	// 检查管理员是否已存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Printf("管理员 %s 已存在，跳过创建", username)
		return nil
	}

	// 插入默认管理员
	now := time.Now()
	_, err = db.Exec(`
		INSERT INTO admin_users (createTime, updateTime, username, password, email, phone, realName, avatar, status, roleId) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, username, hashedPassword, "admin@example.com", "", "系统管理员", "", 1, 1)

	if err != nil {
		return err
	}

	log.Printf("默认管理员 %s 创建成功，密码: %s", username, password)
	return nil
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

	// MD5加密密码 (对齐登录验证逻辑)
	hashedPassword := HashPassword(password)

	// 连接数据库
	var db *sql.DB
	if testMySQLConnection() {
		db, err = getMySQLDB()
		if err != nil {
			return err
		}
		defer db.Close()
	} else {
		return fmt.Errorf("MySQL连接失败，当前版本需要MySQL数据库")
	}

	// 检查管理员是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Printf("管理员 %s 已存在，跳过创建", username)
		return nil
	}

	// 插入默认管理员
	now := time.Now()
	_, err = db.Exec(`
		INSERT INTO admin_users (createTime, updateTime, username, password, email, phone, realName, avatar, status, roleId) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, username, hashedPassword, "admin@example.com", "", "系统管理员", "", 1, 1)

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

	now := time.Now()
	_, err = db.Exec(`
		INSERT INTO admin_users (createTime, updateTime, username, password, email, phone, realName, avatar, status, roleId) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, config.AdminUsername, hashedPassword, config.AdminEmail, "", config.AdminUsername, "", 1, 1)

	return err
}

// Uninstall 卸载
func Uninstall() error {
	log.Println("开始卸载...")

	// 清理数据库
	if err := cleanupDatabase(); err != nil {
		log.Printf("清理数据库失败: %v", err)
		// 不返回错误，继续执行其他清理步骤
	}

	// 删除SQLite数据库文件（兼容性保留）
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

// cleanupDatabase 清理数据库
func cleanupDatabase() error {
	log.Println("开始清理数据库...")

	// 检查是否使用MySQL
	if testMySQLConnection() {
		return cleanupMySQLDatabase()
	} else {
		return cleanupSQLiteDatabase()
	}
}

// cleanupMySQLDatabase 清理MySQL数据库
func cleanupMySQLDatabase() error {
	log.Println("清理MySQL数据库...")

	db, err := getMySQLDB()
	if err != nil {
		return fmt.Errorf("获取MySQL连接失败: %v", err)
	}
	defer db.Close()

	// 获取数据库名称
	var dbName string = "minigame_admin"
	// 删除数据库
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("删除数据库失败: %v", err)
	}

	log.Println("MySQL数据库清理完成")
	return nil
}

// cleanupSQLiteDatabase 清理SQLite数据库
func cleanupSQLiteDatabase() error {
	log.Println("清理SQLite数据库...")

	// 检查SQLite数据库文件是否存在
	if _, err := os.Stat(sqliteDbFile); os.IsNotExist(err) {
		log.Println("SQLite数据库文件不存在，跳过清理")
		return nil
	}

	// 连接SQLite数据库
	db, err := sql.Open("sqlite3", sqliteDbFile)
	if err != nil {
		return fmt.Errorf("连接SQLite数据库失败: %v", err)
	}
	defer db.Close()

	dbName := "minigame_admin"
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		return fmt.Errorf("删除数据库失败: %v", err)
	}

	log.Println("SQLite数据库清理完成")
	return nil
}

// ChangeAdminPassword 修改管理员密码
func ChangeAdminPassword(username, oldPassword, newPassword string) error {
	log.Printf("开始修改管理员密码: %s", username)

	// 获取数据库连接
	db, err := getMySQLDB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %v", err)
	}
	defer db.Close()

	// 验证用户是否存在并检查原密码
	var currentPassword string
	err = db.QueryRow("SELECT password FROM admin_users WHERE username = ?", username).Scan(&currentPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("用户 %s 不存在", username)
		}
		return fmt.Errorf("查询用户失败: %v", err)
	}

	// 验证原密码
	oldPasswordHash := HashPassword(oldPassword)
	if currentPassword != oldPasswordHash {
		return fmt.Errorf("原密码不正确")
	}

	// 加密新密码
	newPasswordHash := HashPassword(newPassword)

	// 更新密码
	now := time.Now()
	_, err = db.Exec("UPDATE admin_users SET password = ?, update_time = ? WHERE username = ?",
		newPasswordHash, now, username)
	if err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	log.Printf("管理员 %s 密码修改成功", username)
	return nil
}

// ResetAdminPassword 重置管理员密码（不验证原密码）
func ResetAdminPassword(username, newPassword string) error {
	log.Printf("开始重置管理员密码: %s", username)

	// 获取数据库连接
	db, err := getMySQLDB()
	if err != nil {
		return fmt.Errorf("获取数据库连接失败: %v", err)
	}
	defer db.Close()

	// 验证用户是否存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return fmt.Errorf("查询用户失败: %v", err)
	}

	if count == 0 {
		return fmt.Errorf("用户 %s 不存在", username)
	}

	// 加密新密码
	newPasswordHash := HashPassword(newPassword)

	// 更新密码
	now := time.Now()
	_, err = db.Exec("UPDATE admin_users SET password = ?, update_time = ?, token = NULL, token_expire = NULL WHERE username = ?",
		newPasswordHash, now, username)
	if err != nil {
		return fmt.Errorf("重置密码失败: %v", err)
	}

	log.Printf("管理员 %s 密码重置成功", username)
	return nil
}

// ListAdminUsers 列出所有管理员用户
func ListAdminUsers() ([]map[string]interface{}, error) {
	log.Println("获取管理员用户列表")

	// 获取数据库连接
	db, err := getMySQLDB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %v", err)
	}
	defer db.Close()

	// 查询管理员用户
	rows, err := db.Query(`
		SELECT id, username, email, phone, real_name, status, last_login_at, last_login_ip, role_id, create_time 
		FROM admin_users 
		ORDER BY create_time DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("查询管理员用户失败: %v", err)
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var (
			id          int64
			username    string
			email       string
			phone       string
			realName    string
			status      int
			lastLoginAt sql.NullTime
			lastLoginIP string
			roleID      int64
			createTime  time.Time
		)

		err = rows.Scan(&id, &username, &email, &phone, &realName, &status, &lastLoginAt, &lastLoginIP, &roleID, &createTime)
		if err != nil {
			return nil, fmt.Errorf("扫描用户数据失败: %v", err)
		}

		user := map[string]interface{}{
			"id":            id,
			"username":      username,
			"email":         email,
			"phone":         phone,
			"real_name":     realName,
			"status":        status,
			"last_login_ip": lastLoginIP,
			"role_id":       roleID,
			"create_time":   createTime.Format("2006-01-02 15:04:05"),
		}

		if lastLoginAt.Valid {
			user["last_login_at"] = lastLoginAt.Time.Format("2006-01-02 15:04:05")
		} else {
			user["last_login_at"] = nil
		}

		users = append(users, user)
	}

	return users, nil
}

// getConfigString 获取配置字符串
func getConfigString(conf config.Configer, key, defaultValue string) string {
	if value, _ := conf.String(key); value != "" {
		return value
	}
	return defaultValue
}

// ChangeAdminPasswordCLI 命令行修改管理员密码（不需要验证原密码）
func ChangeAdminPasswordCLI(username, newPassword string) error {
	// 检查系统是否已安装
	status := CheckInstallStatus()
	if !status.IsInstalled {
		return fmt.Errorf("系统未安装，请先运行安装")
	}

	// 验证密码强度
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}

	// 调用重置密码函数（不需要验证原密码）
	return ResetAdminPassword(username, newPassword)
}

// CreateAdminUser 创建新管理员用户
func CreateAdminUser(username, password, email, realName string) error {
	log.Printf("开始创建管理员用户: %s", username)

	// 参数验证
	if username == "" || password == "" {
		return fmt.Errorf("用户名和密码不能为空")
	}

	// 验证密码强度
	if len(password) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}

	// 连接数据库
	var db *sql.DB
	var err error
	var dbType = getConfigString(appconf, "database_type", "mysql")
	if dbType == "mysql" {
		db, err = getMySQLDB()
		if err != nil {
			return fmt.Errorf("获取MySQL连接失败: %v", err)
		}
		defer db.Close()
	} else {
		return fmt.Errorf("MySQL连接失败，当前版本需要MySQL数据库")
	}

	// 检查管理员是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = ?", username).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查用户是否存在失败: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("用户名 '%s' 已存在", username)
	}

	// MD5加密密码 (对齐登录验证逻辑)
	hashedPassword := HashPassword(password)

	// 设置默认值
	if email == "" {
		email = username + "@example.com"
	}
	if realName == "" {
		realName = username
	}

	// 插入新管理员
	now := time.Now()
	_, err = db.Exec(`
		INSERT INTO admin_users (createTime, updateTime, username, password, email, phone, realName, avatar, status, roleId) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, now, username, hashedPassword, email, "", realName, "", 1, 1)

	if err != nil {
		return fmt.Errorf("创建管理员失败: %v", err)
	}

	log.Printf("管理员用户 %s 创建成功", username)
	return nil
}
