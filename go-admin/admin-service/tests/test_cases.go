package tests

// GetAllTestSuites 获取所有测试套件
func GetAllTestSuites() []*TestSuite {
	return []*TestSuite{
		GetUserTestSuite(),
		GetSystemTestSuite(),
		GetStatisticsTestSuite(),
		GetApplicationTestSuite(),
		GetPermissionTestSuite(),
		GetGameDataTestSuite(),
		GetFileTestSuite(),
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
				Method:      "POST",
				URL:         "/api/user-management/users",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 20,
					"keyword":  "",
					"status":   "",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				ValidateData: ValidateListResponse,
				RequiresAuth: true,
				Tags:         []string{"user", "list", "success"},
			},
			// 获取用户列表 - 参数错误
			{
				Name:        "GetAllUsers_InvalidParams",
				Description: "测试无效参数的处理",
				Method:      "POST",
				URL:         "/api/user-management/users",
				RequestData: map[string]interface{}{
					"appId": "", // 空的appId
					"page":  1,
				},
				ExpectedCode: 4001,
				ExpectedMsg:  "参数错误",
				RequiresAuth: true,
				Tags:         []string{"user", "list", "error"},
			},
			// 获取用户列表 - 分页测试
			{
				Name:        "GetAllUsers_Pagination",
				Description: "测试分页功能",
				Method:      "POST",
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
				Method:      "POST",
				URL:         "/api/user-management/user/detail",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_001",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "success",
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
				Method:      "POST",
				URL:         "/api/user-management/user/detail",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "non_existent_player",
				},
				ExpectedCode: 4004,
				ExpectedMsg:  "资源不存在",
				RequiresAuth: true,
				Tags:         []string{"user", "detail", "error"},
			},
			// 设置用户详情 - 成功案例
			{
				Name:        "SetUserDetail_Success",
				Description: "成功设置用户详情",
				Method:      "POST",
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
				ExpectedMsg:  "success",
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
				Method:      "POST",
				URL:         "/api/user-management/user/data",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_001",
					"userData": "invalid_json_data", // 无效的数据格式
				},
				ExpectedCode: 4005,
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
				ExpectedMsg:   "success",
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
				ExpectedCode:  4001,
				ExpectedMsg:   "参数错误",
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
				ExpectedMsg:   "success",
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
				Method:      "POST",
				URL:         "/api/user-management/user/delete",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_delete",
				},
				ExpectedCode:  0,
				ExpectedMsg:   "success",
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
				Method:      "POST",
				URL:         "/api/user-management/user/stats",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"playerId": "test_player_stats",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "success",
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
				ExpectedMsg:   "success",
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
				ExpectedMsg:   "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:   "success",
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
				ExpectedMsg:   "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:   "success",
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
				ExpectedMsg:   "success",
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
				ExpectedMsg:   "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:  "success",
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
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"statistics", "system", "success"},
			},
		},
	}
}

// GetApplicationTestSuite 应用管理测试套件
func GetApplicationTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "ApplicationManagement",
		Description: "应用管理相关接口测试，包括应用创建、更新、删除等功能",
		TestCases: []*TestCase{
			{
				Name:         "GetApplications_Success",
				Description:  "成功获取应用列表",
				Method:       "GET",
				URL:          "/api/applications",
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"application", "list", "success"},
			},
			{
				Name:        "CreateApplication_Success",
				Description: "成功创建应用",
				Method:      "POST",
				URL:         "/api/applications",
				RequestData: map[string]interface{}{
					"appName":     "Test Application",
					"description": "Test Description",
					"appType":     "game",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"application", "create", "success"},
			},
		},
	}
}

// GetPermissionTestSuite 权限管理测试套件
func GetPermissionTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "PermissionManagement",
		Description: "权限管理相关接口测试，包括角色和权限管理",
		TestCases: []*TestCase{
			{
				Name:         "GetRoles_Success",
				Description:  "成功获取角色列表",
				Method:       "GET",
				URL:          "/api/permissions/roles",
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"permission", "roles", "success"},
			},
			{
				Name:         "GetPermissions_Success",
				Description:  "成功获取权限列表",
				Method:       "GET",
				URL:          "/api/permissions/permissions",
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"permission", "permissions", "success"},
			},
		},
	}
}

// GetGameDataTestSuite 游戏数据管理测试套件
func GetGameDataTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "GameDataManagement",
		Description: "游戏数据管理相关接口测试，包括排行榜、邮件、配置等",
		TestCases: []*TestCase{
			{
				Name:        "GetLeaderboard_Success",
				Description: "成功获取排行榜数据",
				Method:      "POST",
				URL:         "/api/game-data/leaderboard",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 20,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"gamedata", "leaderboard", "success"},
			},
			{
				Name:        "GetMailList_Success",
				Description: "成功获取邮件列表",
				Method:      "POST",
				URL:         "/api/game-data/mail",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 20,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"gamedata", "mail", "success"},
			},
			{
				Name:        "SendMail_Success",
				Description: "成功发送邮件",
				Method:      "POST",
				URL:         "/api/game-data/mail/send",
				RequestData: map[string]interface{}{
					"appId":       "test_app_001",
					"playerId":    "test_player_001",
					"title":       "Test Mail",
					"content":     "Test Content",
					"attachments": "{}",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"gamedata", "mail", "send", "success"},
			},
		},
	}
}

// GetFileTestSuite 文件管理测试套件
func GetFileTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "FileManagement",
		Description: "文件管理相关接口测试，包括文件上传、下载、删除等功能",
		TestCases: []*TestCase{
			{
				Name:         "GetFileList_Success",
				Description:  "成功获取文件列表",
				Method:       "GET",
				URL:          "/api/files",
				ExpectedCode: 0,
				ExpectedMsg:  "success",
				RequiresAuth: true,
				Tags:         []string{"file", "list", "success"},
			},
			// 注意：文件上传测试需要特殊处理，这里先跳过
		},
	}
}

// GetCounterTestSuite 计数器管理测试套件
func GetCounterTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "CounterManagement",
		Description: "计数器管理相关接口测试，包括计数器创建、更新、删除等功能",
		TestCases: []*TestCase{
			{
				Name:        "GetCounterList_Success",
				Description: "成功获取计数器列表",
				Method:      "POST",
				URL:         "/counter/getList",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 10,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"counter", "list", "success"},
			},
			{
				Name:        "GetCounterStats_Success",
				Description: "成功获取计数器统计",
				Method:      "POST",
				URL:         "/counter/getAllStats",
				RequestData: map[string]interface{}{
					"appId": "test_app_001",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"counter", "stats", "success"},
			},
		},
	}
}

// GetLeaderboardTestSuite 排行榜管理测试套件
func GetLeaderboardTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "LeaderboardManagement",
		Description: "排行榜管理相关接口测试，包括排行榜创建、更新、删除等功能",
		TestCases: []*TestCase{
			{
				Name:        "GetLeaderboardList_Success",
				Description: "成功获取排行榜列表",
				Method:      "POST",
				URL:         "/leaderboard/getAll",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 10,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"leaderboard", "list", "success"},
			},
			{
				Name:        "GetLeaderboardData_Success",
				Description: "成功获取排行榜数据",
				Method:      "POST",
				URL:         "/leaderboard/getData",
				RequestData: map[string]interface{}{
					"appId":           "test_app_001",
					"leaderboardName": "daily_score",
					"page":            1,
					"pageSize":        10,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"leaderboard", "data", "success"},
			},
		},
	}
}

// GetMailSystemTestSuite 邮件系统测试套件
func GetMailSystemTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "MailSystem",
		Description: "邮件系统相关接口测试，包括邮件发送、获取、统计等功能",
		TestCases: []*TestCase{
			{
				Name:        "GetMailList_Success",
				Description: "成功获取邮件列表",
				Method:      "POST",
				URL:         "/mail/getAll",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 10,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"mail", "list", "success"},
			},
			{
				Name:        "GetMailStats_Success",
				Description: "成功获取邮件统计",
				Method:      "POST",
				URL:         "/mail/getStats",
				RequestData: map[string]interface{}{
					"appId": "test_app_001",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"mail", "stats", "success"},
			},
			{
				Name:        "SendMail_Success",
				Description: "成功发送邮件",
				Method:      "POST",
				URL:         "/mail/send",
				RequestData: map[string]interface{}{
					"appId":       "test_app_001",
					"userId":      "test_player_001",
					"title":       "Test Mail",
					"content":     "Test Content",
					"attachments": "{}",
				},
				ExpectedCode: 0,
				ExpectedMsg:  "发送成功",
				RequiresAuth: true,
				Tags:         []string{"mail", "send", "success"},
			},
		},
	}
}

// GetGameConfigTestSuite 游戏配置测试套件
func GetGameConfigTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "GameConfig",
		Description: "游戏配置相关接口测试，包括配置获取、更新、删除等功能",
		TestCases: []*TestCase{
			{
				Name:        "GetConfigList_Success",
				Description: "成功获取配置列表",
				Method:      "POST",
				URL:         "/gameConfig/getList",
				RequestData: map[string]interface{}{
					"appId":    "test_app_001",
					"page":     1,
					"pageSize": 10,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"config", "list", "success"},
			},
		},
	}
}

// GetAdminTestSuite 管理员管理测试套件
func GetAdminTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "AdminManagement",
		Description: "管理员管理相关接口测试，包括管理员创建、更新、删除等功能",
		TestCases: []*TestCase{
			{
				Name:        "GetAdminList_Success",
				Description: "成功获取管理员列表",
				Method:      "POST",
				URL:         "/admin/getList",
				RequestData: map[string]interface{}{
					"page":     1,
					"pageSize": 10,
				},
				ExpectedCode: 0,
				ExpectedMsg:  "获取成功",
				RequiresAuth: true,
				Tags:         []string{"admin", "list", "success"},
			},
		},
	}
}

// GetHealthTestSuite 健康检查测试套件
func GetHealthTestSuite() *TestSuite {
	return &TestSuite{
		Name:        "HealthCheck",
		Description: "健康检查相关接口测试",
		TestCases: []*TestCase{
			{
				Name:         "Health_Success",
				Description:  "健康检查成功",
				Method:       "GET",
				URL:          "/health",
				ExpectedCode: 0,
				ExpectedMsg:  "healthy",
				RequiresAuth: false,
				Tags:         []string{"health", "success"},
			},
		},
	}
}
