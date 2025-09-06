package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	_ "github.com/go-sql-driver/mysql"

	// 导入models包以确保所有模型被注册
	_ "admin-service/models"
)

var testFramework *TestFramework
var testData *TestData

// TestMain 测试入口点
func TestMain(m *testing.M) {
	// 初始化测试环境
	if err := setupTestEnvironment(); err != nil {
		fmt.Printf("❌ 初始化测试环境失败: %v\n", err)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()

	// 清理测试环境
	cleanupTestEnvironment()

	os.Exit(code)
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() error {
	fmt.Println("🚀 初始化测试环境...")

	// 初始化数据库连接（使用测试数据库）
	if err := initTestDatabase(); err != nil {
		return fmt.Errorf("初始化测试数据库失败: %v", err)
	}

	// 创建测试框架
	testFramework = NewTestFramework()
	testData = NewTestData()

	// 创建测试数据
	if err := testData.setupTestData(); err != nil {
		return fmt.Errorf("创建测试数据失败: %v", err)
	}

	fmt.Println("✅ 测试环境初始化完成")
	return nil
}

// cleanupTestEnvironment 清理测试环境
func cleanupTestEnvironment() {
	fmt.Println("🧹 清理测试环境...")

	if testData != nil {
		testData.Cleanup()
	}

	if testFramework != nil {
		testFramework.Close()
	}

	fmt.Println("✅ 测试环境清理完成")
}

// initTestDatabase 初始化测试数据库
func initTestDatabase() error {

	// 设置数据库连接
	dataSource := "root:@tcp(127.0.0.1:3306)/admin_service_test?charset=utf8mb4&parseTime=True&loc=Local"

	// 注册数据库，如果已经注册则跳过
	err := orm.RegisterDataBase("default", "mysql", dataSource)
	if err != nil {
		// 如果数据库已经注册，检查错误信息
		if err.Error() == "DataBase alias name `default` already registered, cannot reuse" {
			// 数据库已经注册，继续执行
			fmt.Println("⚠️  数据库连接已存在，继续使用现有连接")
		} else {
			return fmt.Errorf("注册数据库失败: %v", err)
		}
	}

	// 同步数据库表结构
	if err := orm.RunSyncdb("default", false, true); err != nil {
		fmt.Printf("同步数据库表失败: %v，尝试手动创建表\n", err)
	}

	// 创建表结构
	if err := createTestTables(); err != nil {
		return fmt.Errorf("创建测试表失败: %v", err)
	}

	return nil
}

// createTestTables 创建测试表
func createTestTables() error {
	o := orm.NewOrm()

	// 创建应用表
	createApplicationsTable := `
	CREATE TABLE IF NOT EXISTS applications (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		app_id VARCHAR(100) UNIQUE NOT NULL,
		app_name VARCHAR(200) NOT NULL,
		description TEXT,
		status TINYINT DEFAULT 1,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_app_id (app_id),
		INDEX idx_status (status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	// 创建管理员表
	createAdminsTable := `
	CREATE TABLE IF NOT EXISTS admins (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(100) NOT NULL,
		nickname VARCHAR(100),
		email VARCHAR(100),
		phone VARCHAR(20),
		role VARCHAR(50) DEFAULT 'admin',
		status TINYINT DEFAULT 1,
		last_login_time DATETIME,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_username (username),
		INDEX idx_status (status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	// 创建用户封禁记录表
	createBanRecordsTable := `
	CREATE TABLE IF NOT EXISTS user_ban_records (
		id VARCHAR(64) PRIMARY KEY,
		app_id VARCHAR(100) NOT NULL,
		player_id VARCHAR(100) NOT NULL,
		admin_id BIGINT NOT NULL,
		ban_type VARCHAR(20) NOT NULL,
		ban_reason TEXT,
		ban_start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		ban_end_time DATETIME NULL,
		is_active TINYINT DEFAULT 1,
		unban_admin_id BIGINT NULL,
		unban_time DATETIME NULL,
		unban_reason TEXT,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_app_player (app_id, player_id),
		INDEX idx_active (is_active),
		INDEX idx_admin (admin_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	// 创建游戏配置表
	createGameConfigsTable := `
	CREATE TABLE IF NOT EXISTS game_configs (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		app_id VARCHAR(100) NOT NULL,
		config_key VARCHAR(100) NOT NULL,
		config_value TEXT,
		config_type VARCHAR(20) DEFAULT 'string',
		description TEXT,
		status TINYINT DEFAULT 1,
		is_public TINYINT DEFAULT 1,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uk_app_key (app_id, config_key),
		INDEX idx_app_id (app_id),
		INDEX idx_status (status)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	// 创建操作日志表
	createOperationLogsTable := `
	CREATE TABLE IF NOT EXISTS operation_logs (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		admin_id BIGINT NOT NULL,
		action VARCHAR(100) NOT NULL,
		resource VARCHAR(100),
		resource_id VARCHAR(100),
		details TEXT,
		ip VARCHAR(45),
		user_agent TEXT,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_admin (admin_id),
		INDEX idx_action (action),
		INDEX idx_create_time (create_time)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	// 创建系统配置表
	createSystemConfigsTable := `
	CREATE TABLE IF NOT EXISTS system_configs (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		config_key VARCHAR(100) UNIQUE NOT NULL,
		config_value TEXT,
		config_type VARCHAR(20) DEFAULT 'string',
		description TEXT,
		create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_config_key (config_key)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	// 执行建表语句
	tables := []string{
		createApplicationsTable,
		createAdminsTable,
		createBanRecordsTable,
		createGameConfigsTable,
		createOperationLogsTable,
		createSystemConfigsTable,
	}

	for _, sql := range tables {
		_, err := o.Raw(sql).Exec()
		if err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}

	return nil
}

// setupTestData 设置测试数据
func (td *TestData) setupTestData() error {
	// 创建测试管理员
	if err := td.createTestAdmin(); err != nil {
		return fmt.Errorf("创建测试管理员失败: %v", err)
	}

	// 创建测试应用和相关数据
	testApps := []string{"test_app_001", "test_app_002", "test_app_performance"}
	for _, appId := range testApps {
		if err := CreateFullTestEnvironment(appId); err != nil {
			return fmt.Errorf("创建测试应用环境失败 [%s]: %v", appId, err)
		}
		td.CreatedApps = append(td.CreatedApps, appId)
	}

	// 创建系统配置
	if err := td.createTestSystemConfigs(); err != nil {
		return fmt.Errorf("创建系统配置失败: %v", err)
	}

	return nil
}

// createTestAdmin 创建测试管理员
func (td *TestData) createTestAdmin() error {
	o := orm.NewOrm()

	// 创建普通管理员（简化版密码哈希）
	adminPassword := "test123456_hashed"
	adminSQL := `
	INSERT INTO admins (username, password, nickname, email, role, status, create_time, update_time) 
	VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW()) 
	ON DUPLICATE KEY UPDATE password = VALUES(password), update_time = NOW()
	`

	admins := []struct {
		Username string
		Password string
		Nickname string
		Email    string
		Role     string
		Status   int
	}{
		{"test_admin", adminPassword, "测试管理员", "test_admin@example.com", "admin", 1},
		{"test_super_admin", adminPassword, "测试超级管理员", "test_super_admin@example.com", "super_admin", 1},
		{"test_user", adminPassword, "测试普通用户", "test_user@example.com", "user", 1},
	}

	for _, admin := range admins {
		_, err := o.Raw(adminSQL, admin.Username, admin.Password, admin.Nickname, admin.Email, admin.Role, admin.Status).Exec()
		if err != nil {
			return fmt.Errorf("创建管理员失败 [%s]: %v", admin.Username, err)
		}
	}

	return nil
}

// createTestSystemConfigs 创建测试系统配置
func (td *TestData) createTestSystemConfigs() error {
	o := orm.NewOrm()

	configs := []struct {
		Key   string
		Value string
		Type  string
		Desc  string
	}{
		{"site_name", "测试管理系统", "string", "站点名称"},
		{"site_url", "https://test-admin.example.com", "string", "站点URL"},
		{"site_description", "这是一个用于测试的管理系统", "string", "站点描述"},
		{"enable_register", "false", "boolean", "是否允许注册"},
		{"enable_captcha", "true", "boolean", "是否启用验证码"},
		{"jwt_expire_hours", "24", "number", "JWT过期时间（小时）"},
		{"max_login_attempts", "5", "number", "最大登录尝试次数"},
		{"cache_expire_minutes", "30", "number", "缓存过期时间（分钟）"},
		{"log_retention_days", "30", "number", "日志保留天数"},
	}

	configSQL := `
	INSERT INTO system_configs (config_key, config_value, config_type, description, create_time, update_time) 
	VALUES (?, ?, ?, ?, NOW(), NOW()) 
	ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), update_time = NOW()
	`

	for _, config := range configs {
		_, err := o.Raw(configSQL, config.Key, config.Value, config.Type, config.Desc).Exec()
		if err != nil {
			return fmt.Errorf("创建系统配置失败 [%s]: %v", config.Key, err)
		}
	}

	return nil
}

// TestAllAPIs 运行所有API测试
func TestAllAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	fmt.Println("\n🧪 开始运行API测试...")
	testFramework.RunAllTests(t)
}

// TestUserAPIs 用户管理API测试
func TestUserAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	suite := GetUserTestSuite()
	results := testFramework.ExecuteTestSuite(suite)

	// 统计结果
	var passed, failed int
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
			t.Errorf("Test %s failed: %s", result.TestCase.Name, result.Error)
		}
	}

	t.Logf("用户API测试完成: %d 通过, %d 失败", passed, failed)
}

// TestSystemAPIs 系统管理API测试
func TestSystemAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	suite := GetSystemTestSuite()
	results := testFramework.ExecuteTestSuite(suite)

	// 统计结果
	var passed, failed int
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
			t.Errorf("Test %s failed: %s", result.TestCase.Name, result.Error)
		}
	}

	t.Logf("系统API测试完成: %d 通过, %d 失败", passed, failed)
}

// TestStatisticsAPIs 统计分析API测试
func TestStatisticsAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	suite := GetStatisticsTestSuite()
	results := testFramework.ExecuteTestSuite(suite)

	// 统计结果
	var passed, failed int
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
			t.Errorf("Test %s failed: %s", result.TestCase.Name, result.Error)
		}
	}

	t.Logf("统计API测试完成: %d 通过, %d 失败", passed, failed)
}

// TestApplicationAPIs 应用管理API测试
func TestApplicationAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	suite := GetApplicationTestSuite()
	results := testFramework.ExecuteTestSuite(suite)

	// 统计结果
	var passed, failed int
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
			t.Errorf("Test %s failed: %s", result.TestCase.Name, result.Error)
		}
	}

	t.Logf("应用管理API测试完成: %d 通过, %d 失败", passed, failed)
}

// TestPermissionAPIs 权限管理API测试
func TestPermissionAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	suite := GetPermissionTestSuite()
	results := testFramework.ExecuteTestSuite(suite)

	// 统计结果
	var passed, failed int
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
			t.Errorf("Test %s failed: %s", result.TestCase.Name, result.Error)
		}
	}

	t.Logf("权限管理API测试完成: %d 通过, %d 失败", passed, failed)
}

// TestGameDataAPIs 游戏数据管理API测试
func TestGameDataAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	suite := GetGameDataTestSuite()
	results := testFramework.ExecuteTestSuite(suite)

	// 统计结果
	var passed, failed int
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
			t.Errorf("Test %s failed: %s", result.TestCase.Name, result.Error)
		}
	}

	t.Logf("游戏数据管理API测试完成: %d 通过, %d 失败", passed, failed)
}

// TestFileAPIs 文件管理API测试
func TestFileAPIs(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	suite := GetFileTestSuite()
	results := testFramework.ExecuteTestSuite(suite)

	// 统计结果
	var passed, failed int
	for _, result := range results {
		if result.Success {
			passed++
		} else {
			failed++
			t.Errorf("Test %s failed: %s", result.TestCase.Name, result.Error)
		}
	}

	t.Logf("文件管理API测试完成: %d 通过, %d 失败", passed, failed)
}

// BenchmarkUserAPIs 用户API性能测试
func BenchmarkUserAPIs(b *testing.B) {
	if testFramework == nil {
		b.Fatal("测试框架未初始化")
	}

	// 获取用户列表的性能测试
	testCase := &TestCase{
		Name:        "GetAllUsers_Performance",
		Description: "用户列表接口性能测试",
		Method:      "POST",
		URL:         "/api/user-management/users",
		RequestData: map[string]interface{}{
			"appId":    "test_app_performance",
			"page":     1,
			"pageSize": 20,
		},
		RequiresAuth: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := testFramework.ExecuteTestCase(testCase)
		if !result.Success {
			b.Errorf("Performance test failed: %s", result.Error)
		}
	}
}

// TestErrorHandling 错误处理测试
func TestErrorHandling(t *testing.T) {
	if testFramework == nil {
		t.Fatal("测试框架未初始化")
	}

	// 测试各种错误情况
	errorCases := []*TestCase{
		{
			Name:        "UnauthorizedAccess",
			Description: "未授权访问测试",
			Method:      "POST",
			URL:         "/api/user-management/users",
			RequestData: map[string]interface{}{
				"appId": "test_app_001",
			},
			RequiresAuth: false, // 不提供认证
			ExpectedCode: 4003,  // ValidateJWT返回CodeUnauthorized (4003)
		},
		{
			Name:        "InvalidJSON",
			Description: "无效JSON数据测试",
			Method:      "POST",
			URL:         "/api/user-management/user/data",
			RequestData: map[string]interface{}{
				"appId":    "test_app_001",
				"playerId": "test_player_001",
				"userData": "invalid_json", // 无效的JSON
			},
			RequiresAuth: true,
			ExpectedCode: 4005, // 数据验证错误
		},
		{
			Name:        "NonExistentResource",
			Description: "不存在资源测试",
			Method:      "POST",
			URL:         "/api/user-management/user/detail",
			RequestData: map[string]interface{}{
				"appId":    "non_existent_app",
				"playerId": "non_existent_player",
			},
			RequiresAuth: true,
			ExpectedCode: 4004, // 资源不存在
		},
	}

	for _, testCase := range errorCases {
		result := testFramework.ExecuteTestCase(testCase)
		if !result.Success {
			t.Errorf("Error handling test %s failed: %s", testCase.Name, result.Error)
		} else {
			t.Logf("✅ Error handling test %s passed", testCase.Name)
		}
	}
}
