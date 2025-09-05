package tests

// GetAllTestSuites 获取所有测试套件
func GetAllTestSuites() []*TestSuite {
	return []*TestSuite{
		GetUserTestSuite(),
		GetSystemTestSuite(),
		GetStatisticsTestSuite(),
	}
}

// GetUserTestSuite 用户管理测试套件
func GetUserTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "UserManagement",
		Description: "用户管理相关接口测试，包括用户列表、封禁、解封、删除等功能",
		TestCases: []*TestCase{
			// 获取用户列表 - 成功案例
			{
				Name:        "GetAllUsers_Success",
				Description: "成功获取用户列表，验证分页和数据格式",
				Method:      "GET",
				URL:         "/api/user-management/users",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 20,
					"keyword":  "",
					"status":   "",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: ValidateListResponse,
				RequiresAuth: true,
				Tags:         []string{"user", "list", "success"},
			},
			// 获取用户列表 - 参数错误
			{
				Name:        "GetAllUsers_InvalidParams",
				Description: "测试无效参数的处理",
				Method:      "GET",
				URL:         "/api/user-management/users",
				RequestData: map[string]interface{}{
					"appId": "", // 空的appId
					"page":  1,
				},
				ExpectedCode: 1002,
				ExpectedMsg:  "应用ID不能为空",
				RequiresAuth: true,
				Tags:         []string{"user", "list", "error"},
			},
			// 获取用户列表 - 分页测试
			{
				Name:        "GetAllUsers_Pagination",
				Description: "测试分页功能",
				Method:      "GET",
				URL:         "/api/user-management/users",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     2,
					"pageSize": 5,
				},
				ExpectedCode: 0,
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					page, exists := dataMap["page"]
					if !exists || page != float64(2) {
						return false
					}

					pageSize, exists := dataMap["pageSize"]
					if !exists || pageSize != float64(5) {
						return false
					}

					return true
				},
				RequiresAuth: true,
				Tags:         []string{"user", "list", "pagination"},
			},
			// 获取用户详情 - 成功案例
			{
				Name:        "GetUserDetail_Success",
				Description: "成功获取用户详情",
				Method:      "GET",
				URL:         "/api/user-management/user/detail",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_001",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: ValidateUserData,
				RequiresAuth: true,
				SetupFunc: func() error {
					// 确保测试用户存在
					return CreateTestUser("test_app_001", "test_player_001")
				},
				Tags: []string{"user", "detail", "success"},
			},
			// 获取用户详情 - 用户不存在
			{
				Name:        "GetUserDetail_NotFound",
				Description: "获取不存在用户的详情",
				Method:      "GET",
				URL:         "/api/user-management/user/detail",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "non_existent_player",
				},
				ExpectedCode: 1003,
				ExpectedMsg:  "获取用户详情失败",
				RequiresAuth: true,
				Tags:         []string{"user", "detail", "error"},
			},
			// 设置用户详情 - 成功案例
			{
				Name:        "SetUserDetail_Success",
				Description: "成功设置用户详情",
				Method:      "PUT",
				URL:         "/api/user-management/user/data",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_001",
					"userData": map[string]interface{}{
						"level":     5,
						"score":     1000,
						"coins":     200,
						"nickname":  "测试玩家",
						"lastLogin": "2023-10-01 10:00:00",
					},
				},
				ExpectedCode: 0,
				ExpectedMsg:  "设置成功",
				RequiresAuth: true,
				SetupFunc: func() error {
					return CreateTestUser("test_app_001", "test_player_001")
				},
				Tags: []string{"user", "update", "success"},
			},
			// 设置用户详情 - 参数错误
			{
				Name:        "SetUserDetail_InvalidData",
				Description: "测试无效用户数据的处理",
				Method:      "PUT",
				URL:         "/api/user-management/user/data",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_001",
					"userData": "invalid_json_data", // 无效的数据格式
				},
				ExpectedCode: 1001,
				RequiresAuth: true,
				Tags:         []string{"user", "update", "error"},
			},
			// 封禁用户 - 成功案例
			{
				Name:        "BanUser_Success",
				Description: "成功封禁用户",
				Method:      "POST",
				URL:         "/api/user-management/user/ban",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_ban",
					"reason":   "违反游戏规则",
					"duration": 24,
				},
				ExpectedCode:  0,
				ExpectedMsg:   "封禁成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				SetupFunc: func() error {
					return CreateTestUser("test_app_001", "test_player_ban")
				},
				CleanupFunc: func() error {
					// 清理封禁记录
					return CleanupBanRecord("test_app_001", "test_player_ban")
				},
				Tags: []string{"user", "ban", "success"},
			},
			// 封禁用户 - 缺少参数
			{
				Name:        "BanUser_MissingParams",
				Description: "测试缺少必要参数的情况",
				Method:      "POST",
				URL:         "/api/user-management/user/ban",
				RequestData: map[string]interface{}{
					"appId":  "test_app_001",
					"reason": "违反规则",
					// 缺少playerId
				},
				ExpectedCode:  1002,
				ExpectedMsg:   "应用ID、玩家ID不能为空",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"user", "ban", "error"},
			},
			// 解封用户 - 成功案例
			{
				Name:        "UnbanUser_Success",
				Description: "成功解封用户",
				Method:      "POST",
				URL:         "/api/user-management/user/unban",
				RequestData: map[string]interface{}{
					"appId":       "test_app_001",
					"playerId":    "test_player_unban",
					"unbanReason": "申诉成功",
				},
				ExpectedCode:  0,
				ExpectedMsg:   "解封成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				SetupFunc: func() error {
					// 先创建用户再封禁
					if err := CreateTestUser("test_app_001", "test_player_unban"); err != nil {
						return err
					}
					return BanTestUser("test_app_001", "test_player_unban")
				},
				CleanupFunc: func() error {
					return CleanupBanRecord("test_app_001", "test_player_unban")
				},
				Tags: []string{"user", "unban", "success"},
			},
			// 删除用户 - 成功案例
			{
				Name:        "DeleteUser_Success",
				Description: "成功删除用户",
				Method:      "DELETE",
				URL:         "/api/user-management/user/delete",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_delete",
				},
				ExpectedCode:  0,
				ExpectedMsg:   "删除成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				SetupFunc: func() error {
					return CreateTestUser("test_app_001", "test_player_delete")
				},
				Tags: []string{"user", "delete", "success"},
			},
			// 获取用户统计 - 成功案例
			{
				Name:        "GetUserStats_Success",
				Description: "成功获取用户统计信息",
				Method:      "GET",
				URL:         "/api/user-management/user/stats",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_stats",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查统计数据字段
					requiredFields := []string{"playerId", "registrationTime"}
					for _, field := range requiredFields {
						if _, exists := dataMap[field]; !exists {
							return false
						}
					}
					return true
				},
				RequiresAuth: true,
				SetupFunc: func() error {
					return CreateTestUser("test_app_001", "test_player_stats")
				},
				Tags: []string{"user", "stats", "success"},
			},
		},
	}
}

// GetSystemTestSuite 系统管理测试套件
func GetSystemTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "SystemManagement",
		Description: "系统管理相关接口测试，包括配置管理、缓存操作、备份恢复等功能",
		TestCases: []*TestCase{
			// 获取系统配置 - 成功案例
			{
				Name:          "GetSystemConfig_Success",
				Description:   "成功获取系统配置",
				Method:        "GET",
				URL:           "/api/system/config",
				ExpectedCode:  0,
				ExpectedMsg:   "获取成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "config", "success"},
			},
			// 更新系统配置 - 成功案例
			{
				Name:        "UpdateSystemConfig_Success",
				Description: "成功更新系统配置",
				Method:      "PUT",
				URL:         "/api/system/config",
				RequestData: map[string]interface{}{
					"siteName":        "测试站点",
					"siteUrl":         "https://test.example.com",
					"siteDescription": "这是一个测试站点",
					"enableRegister":  true,
					"enableCaptcha":   false,
					"jwtExpireHours":  24,
				},
				ExpectedCode:  0,
				ExpectedMsg:   "更新成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "config", "update", "success"},
			},
			// 获取系统状态 - 成功案例
			{
				Name:         "GetSystemStatus_Success",
				Description:  "成功获取系统状态信息",
				Method:       "GET",
				URL:          "/api/system/status",
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: func(data interface{}) bool {
					// 验证系统状态数据结构
					return data != nil
				},
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "status", "success"},
			},
			// 清理缓存 - 成功案例
			{
				Name:        "ClearCache_Success",
				Description: "成功清理系统缓存",
				Method:      "DELETE",
				URL:         "/api/system/cache",
				RequestData: map[string]interface{}{
					"cacheType": "all",
				},
				ExpectedCode:  0,
				ExpectedMsg:   "清理成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "cache", "success"},
			},
			// 获取缓存统计 - 成功案例
			{
				Name:          "GetCacheStats_Success",
				Description:   "成功获取缓存统计信息",
				Method:        "GET",
				URL:           "/api/system/cache/stats",
				ExpectedCode:  0,
				ExpectedMsg:   "获取成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "cache", "stats", "success"},
			},
			// 创建备份 - 成功案例
			{
				Name:        "BackupData_Success",
				Description: "成功创建数据备份",
				Method:      "POST",
				URL:         "/api/system/backup",
				RequestData: map[string]interface{}{
					"backupType": "full",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "备份成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查备份文件字段
					if _, exists := dataMap["backupFile"]; !exists {
						return false
					}
					if _, exists := dataMap["createTime"]; !exists {
						return false
					}
					return true
				},
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "backup", "success"},
			},
			// 获取备份列表 - 成功案例
			{
				Name:         "GetBackupList_Success",
				Description:  "成功获取备份列表",
				Method:       "GET",
				URL:          "/api/system/backup",
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查备份列表字段
					requiredFields := []string{"backups", "total", "page", "pageSize"}
					for _, field := range requiredFields {
						if _, exists := dataMap[field]; !exists {
							return false
						}
					}
					return true
				},
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "backup", "list", "success"},
			},
			// 获取服务器信息 - 成功案例
			{
				Name:          "GetServerInfo_Success",
				Description:   "成功获取服务器信息",
				Method:        "GET",
				URL:           "/api/system/server",
				ExpectedCode:  0,
				ExpectedMsg:   "获取成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "server", "info", "success"},
			},
			// 获取数据库信息 - 成功案例
			{
				Name:          "GetDatabaseInfo_Success",
				Description:   "成功获取数据库信息",
				Method:        "GET",
				URL:           "/api/system/database",
				ExpectedCode:  0,
				ExpectedMsg:   "获取成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "database", "info", "success"},
			},
			// 优化数据库 - 成功案例
			{
				Name:          "OptimizeDatabase_Success",
				Description:   "成功优化数据库",
				Method:        "POST",
				URL:           "/api/system/database/optimize",
				ExpectedCode:  0,
				ExpectedMsg:   "优化成功",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"system", "database", "optimize", "success"},
			},
		},
	}
}

// GetStatisticsTestSuite 统计分析测试套件
func GetStatisticsTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "Statistics",
		Description: "统计分析相关接口测试，包括仪表盘数据、应用统计、操作日志等功能",
		TestCases: []*TestCase{
			// 获取仪表盘数据 - 成功案例
			{
				Name:         "GetDashboard_Success",
				Description:  "成功获取仪表盘统计数据",
				Method:       "GET",
				URL:          "/api/statistics/dashboard",
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查仪表盘数据字段
					requiredFields := []string{"totalApps", "totalAdmins", "activeApps", "todayOperations"}
					for _, field := range requiredFields {
						if _, exists := dataMap[field]; !exists {
							return false
						}
					}
					return true
				},
				RequiresAuth: true,
				Tags:         []string{"statistics", "dashboard", "success"},
			},
			// 获取应用统计 - 成功案例
			{
				Name:        "GetApplicationStats_Success",
				Description: "成功获取应用统计数据",
				Method:      "GET",
				URL:         "/api/statistics/application",
				Headers: map[string]string{
					"appId": "test_app_001",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查应用统计字段
					requiredFields := []string{"userDataCount", "leaderboardCount", "counterCount", "mailCount", "configCount"}
					for _, field := range requiredFields {
						if _, exists := dataMap[field]; !exists {
							return false
						}
					}
					return true
				},
				RequiresAuth: true,
				SetupFunc: func() error {
					return CreateTestApp("test_app_001")
				},
				Tags: []string{"statistics", "app", "success"},
			},
			// 获取应用统计 - 参数错误
			{
				Name:         "GetApplicationStats_MissingAppId",
				Description:  "测试缺少应用ID参数的情况",
				Method:       "GET",
				URL:          "/api/statistics/application",
				ExpectedCode: 1002,
				ExpectedMsg:  "应用ID不能为空",
				RequiresAuth: true,
				Tags:         []string{"statistics", "app", "error"},
			},
			// 获取操作日志 - 成功案例
			{
				Name:        "GetOperationLogs_Success",
				Description: "成功获取操作日志",
				Method:      "GET",
				URL:         "/api/statistics/logs",
				Headers: map[string]string{
					"page":     "1",
					"pageSize": "20",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查日志列表字段
					requiredFields := []string{"logs", "total", "page", "pageSize"}
					for _, field := range requiredFields {
						if _, exists := dataMap[field]; !exists {
							return false
						}
					}
					return true
				},
				RequiresAuth: true,
				Tags:         []string{"statistics", "logs", "success"},
			},
			// 获取用户活跃度 - 成功案例
			{
				Name:        "GetUserActivity_Success",
				Description: "成功获取用户活跃度统计",
				Method:      "GET",
				URL:         "/api/statistics/activity",
				Headers: map[string]string{
					"appId": "test_app_001",
					"days":  "7",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				SetupFunc: func() error {
					return CreateTestApp("test_app_001")
				},
				Tags: []string{"statistics", "activity", "success"},
			},
			// 获取数据趋势 - 成功案例
			{
				Name:        "GetDataTrends_Success",
				Description: "成功获取数据趋势统计",
				Method:      "GET",
				URL:         "/api/statistics/trends",
				Headers: map[string]string{
					"appId":    "test_app_001",
					"dataType": "userData",
					"days":     "7",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查趋势数据字段
					requiredFields := []string{"dataType", "days", "trends"}
					for _, field := range requiredFields {
						if _, exists := dataMap[field]; !exists {
							return false
						}
					}
					return true
				},
				RequiresAuth: true,
				SetupFunc: func() error {
					return CreateTestApp("test_app_001")
				},
				Tags: []string{"statistics", "trends", "success"},
			},
			// 导出数据 - 成功案例
			{
				Name:        "ExportData_Success",
				Description: "成功导出数据",
				Method:      "POST",
				URL:         "/api/statistics/export",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"dataType": "userData",
					"format":   "csv",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "导出成功",
				ValidateData: func(data interface{}) bool {
					dataMap, ok := data.(map[string]interface{})
					if !ok {
						return false
					}

					// 检查导出结果字段
					requiredFields := []string{"filePath", "dataType", "format"}
					for _, field := range requiredFields {
						if _, exists := dataMap[field]; !exists {
							return false
						}
					}
					return true
				},
				RequiresAuth:  true,
				RequiresAdmin: true,
				SetupFunc: func() error {
					return CreateTestApp("test_app_001")
				},
				Tags: []string{"statistics", "export", "success"},
			},
			// 导出数据 - 参数错误
			{
				Name:        "ExportData_MissingParams",
				Description: "测试缺少必要参数的情况",
				Method:      "POST",
				URL:         "/api/statistics/export",
				RequestData: map[string]interface{}{
					"format": "csv",
					// 缺少appId和dataType
				},
				ExpectedCode:  1002,
				ExpectedMsg:   "应用ID和数据类型不能为空",
				RequiresAuth:  true,
				RequiresAdmin: true,
				Tags:          []string{"statistics", "export", "error"},
			},
			// 获取系统信息 - 成功案例
			{
				Name:         "GetSystemInfo_Success",
				Description:  "成功获取系统信息统计",
				Method:       "GET",
				URL:          "/api/statistics/system",
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"statistics", "system", "success"},
			},
		},
	}
}
