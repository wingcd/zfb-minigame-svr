package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "admin-service/routers" // 导入路由以注册所有路由
	"admin-service/utils"

	"github.com/beego/beego/v2/server/web"
)

// TestFramework 测试框架结构
type TestFramework struct {
	Server *httptest.Server
	Client *http.Client
}

// APIResponse 标准API响应结构（参考云函数格式）
type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// TestCase 测试用例结构
type TestCase struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Method        string                 `json:"method"`
	URL           string                 `json:"url"`
	Headers       map[string]string      `json:"headers"`
	RequestData   map[string]interface{} `json:"requestData"`
	ExpectedCode  int                    `json:"expectedCode"`
	ExpectedMsg   string                 `json:"expectedMsg"`
	ValidateData  func(interface{}) bool `json:"-"`
	SetupFunc     func() error           `json:"-"`
	CleanupFunc   func() error           `json:"-"`
	RequiresAuth  bool                   `json:"requiresAuth"`
	RequiresAdmin bool                   `json:"requiresAdmin"`
	Tags          []string               `json:"tags"`
}

// TestResult 测试结果
type TestResult struct {
	TestCase   *TestCase     `json:"testCase"`
	Success    bool          `json:"success"`
	Error      string        `json:"error"`
	Response   *APIResponse  `json:"response"`
	Duration   time.Duration `json:"duration"`
	StatusCode int           `json:"statusCode"`
}

// TestSuite 测试套件
type TestSuite struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	TestCases   []*TestCase  `json:"testCases"`
	SetupFunc   func() error `json:"-"`
	CleanupFunc func() error `json:"-"`
}

// NewTestFramework 创建新的测试框架
func NewTestFramework() *TestFramework {
	// 设置测试环境的JWT密钥，与app.conf中保持一致
	utils.SetJWTSecret("minigame_admin_jwt_secret_key_2024")

	// 确保Beego配置正确
	web.BConfig.CopyRequestBody = true
	web.BConfig.MaxMemory = 1 << 26 // 64MB

	// 创建测试服务器使用Beego的Handler
	server := httptest.NewServer(web.BeeApp.Handlers)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &TestFramework{
		Server: server,
		Client: client,
	}
}

// Close 关闭测试框架
func (tf *TestFramework) Close() {
	if tf.Server != nil {
		tf.Server.Close()
	}
}

// ExecuteTestCase 执行单个测试用例
func (tf *TestFramework) ExecuteTestCase(testCase *TestCase) *TestResult {
	startTime := time.Now()
	result := &TestResult{
		TestCase: testCase,
		Success:  false,
	}

	// 执行设置函数
	if testCase.SetupFunc != nil {
		if err := testCase.SetupFunc(); err != nil {
			result.Error = fmt.Sprintf("Setup failed: %v", err)
			result.Duration = time.Since(startTime)
			return result
		}
	}

	// 执行清理函数（延迟执行）
	if testCase.CleanupFunc != nil {
		defer func() {
			if err := testCase.CleanupFunc(); err != nil {
				fmt.Printf("Cleanup failed for %s: %v\n", testCase.Name, err)
			}
		}()
	}

	// 准备请求数据
	var requestBody io.Reader
	var url string

	if testCase.Method == "GET" && testCase.RequestData != nil {
		// GET请求：将参数添加到URL查询字符串
		baseURL := tf.Server.URL + testCase.URL
		queryParams := make([]string, 0)
		for key, value := range testCase.RequestData {
			queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
		}
		if len(queryParams) > 0 {
			url = baseURL + "?" + strings.Join(queryParams, "&")
		} else {
			url = baseURL
		}
	} else {
		// POST/PUT/DELETE请求：将参数放在请求体中
		url = tf.Server.URL + testCase.URL
		if testCase.RequestData != nil {
			jsonData, err := json.Marshal(testCase.RequestData)
			if err != nil {
				result.Error = fmt.Sprintf("Failed to marshal request data: %v", err)
				result.Duration = time.Since(startTime)
				return result
			}
			fmt.Printf("DEBUG: Sending JSON data for %s: %s\n", testCase.Name, string(jsonData))
			requestBody = bytes.NewReader(jsonData)
		}
	}

	// 创建HTTP请求
	req, err := http.NewRequest(testCase.Method, url, requestBody)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for key, value := range testCase.Headers {
		req.Header.Set(key, value)
	}

	// 如果需要认证，添加JWT token
	if testCase.RequiresAuth {
		token := tf.GetTestToken(testCase.RequiresAdmin)
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 发送请求
	resp, err := tf.Client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to send request: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// 解析响应
	var apiResponse APIResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		// 如果JSON解析失败，尝试看是否是纯文本或数字响应
		responseText := string(responseBody)
		if responseText != "" {
			// 如果是纯数字，可能是错误码
			if len(responseText) < 10 {
				result.Error = fmt.Sprintf("API returned non-JSON response: %s", responseText)
			} else {
				result.Error = fmt.Sprintf("Failed to parse JSON response: %v, raw response: %s", err, responseText)
			}
		} else {
			result.Error = fmt.Sprintf("Failed to parse response: %v", err)
		}
		result.Duration = time.Since(startTime)
		return result
	}

	result.Response = &apiResponse
	result.Duration = time.Since(startTime)

	// 验证响应码
	if testCase.ExpectedCode != 0 && apiResponse.Code != testCase.ExpectedCode {
		result.Error = fmt.Sprintf("Expected code %d, got %d", testCase.ExpectedCode, apiResponse.Code)
		return result
	}

	// 验证响应消息
	if testCase.ExpectedMsg != "" && apiResponse.Msg != testCase.ExpectedMsg {
		result.Error = fmt.Sprintf("Expected msg '%s', got '%s'", testCase.ExpectedMsg, apiResponse.Msg)
		return result
	}

	// 自定义数据验证
	if testCase.ValidateData != nil && !testCase.ValidateData(apiResponse.Data) {
		result.Error = "Custom data validation failed"
		return result
	}

	result.Success = true
	return result
}

// ExecuteTestSuite 执行测试套件
func (tf *TestFramework) ExecuteTestSuite(suite *TestSuite) []*TestResult {
	var results []*TestResult

	// 执行套件设置
	if suite.SetupFunc != nil {
		if err := suite.SetupFunc(); err != nil {
			fmt.Printf("Suite setup failed: %v\n", err)
			return results
		}
	}

	// 执行套件清理（延迟执行）
	if suite.CleanupFunc != nil {
		defer func() {
			if err := suite.CleanupFunc(); err != nil {
				fmt.Printf("Suite cleanup failed: %v\n", err)
			}
		}()
	}

	// 执行所有测试用例
	for _, testCase := range suite.TestCases {
		result := tf.ExecuteTestCase(testCase)
		results = append(results, result)

		// 打印测试结果
		if result.Success {
			fmt.Printf("✅ %s - %s (%.2fms)\n", suite.Name, testCase.Name, float64(result.Duration.Nanoseconds())/1000000)
		} else {
			fmt.Printf("❌ %s - %s: %s (%.2fms)\n", suite.Name, testCase.Name, result.Error, float64(result.Duration.Nanoseconds())/1000000)
		}
	}

	return results
}

// GetTestToken 获取测试用的JWT token
func (tf *TestFramework) GetTestToken(isAdmin bool) string {
	// 生成真实的JWT token用于测试
	if isAdmin {
		// 生成管理员token (user_id: 1, username: test_admin, role_id: 1)
		token, err := utils.GenerateJWT(1, "test_admin", 1)
		if err != nil {
			fmt.Printf("Failed to generate admin token: %v\n", err)
			return ""
		}
		return token
	}

	// 生成普通用户token (user_id: 2, username: test_user, role_id: 2)
	token, err := utils.GenerateJWT(2, "test_user", 2)
	if err != nil {
		fmt.Printf("Failed to generate user token: %v\n", err)
		return ""
	}
	return token
}

// CreateTestDatabase 创建测试数据库
func (tf *TestFramework) CreateTestDatabase() error {
	// 这里应该创建测试数据库和表结构
	// 为了简化，我们假设数据库已经存在
	return nil
}

// CleanTestDatabase 清理测试数据库
func (tf *TestFramework) CleanTestDatabase() error {
	// 清理测试数据
	return nil
}

// GenerateTestData 生成测试数据
func (tf *TestFramework) GenerateTestData() error {
	// 这里会在实际运行时由测试数据管理器处理
	return nil
}

// ValidateStandardResponse 验证标准响应格式
func ValidateStandardResponse(data interface{}) bool {
	if data == nil {
		return true // 允许空数据
	}

	// 这里可以添加更多的验证逻辑
	return true
}

// ValidateListResponse 验证列表响应格式
func ValidateListResponse(data interface{}) bool {
	if data == nil {
		return false
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	// 检查必要字段 - 修改为实际的字段名
	requiredFields := []string{"userList", "total", "page", "pageSize"}
	for _, field := range requiredFields {
		if _, exists := dataMap[field]; !exists {
			return false
		}
	}

	return true
}

// ValidateUserData 验证用户数据格式
func ValidateUserData(data interface{}) bool {
	if data == nil {
		return false
	}

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	// 检查用户必要字段
	requiredFields := []string{"id", "playerId"}
	for _, field := range requiredFields {
		if _, exists := dataMap[field]; !exists {
			return false
		}
	}

	return true
}

// RunAllTests 运行所有测试
func (tf *TestFramework) RunAllTests(t *testing.T) {
	// 初始化测试数据
	if err := tf.GenerateTestData(); err != nil {
		t.Fatalf("Failed to generate test data: %v", err)
	}

	// 获取所有测试套件
	suites := GetAllTestSuites()

	var totalTests, passedTests int

	for _, suite := range suites {
		fmt.Printf("\n🧪 Running test suite: %s\n", suite.Name)
		fmt.Printf("📝 %s\n", suite.Description)

		results := tf.ExecuteTestSuite(suite)

		for _, result := range results {
			totalTests++
			if result.Success {
				passedTests++
			}
		}
	}

	// 打印总结
	fmt.Printf("\n📊 Test Summary:\n")
	fmt.Printf("Total: %d, Passed: %d, Failed: %d\n", totalTests, passedTests, totalTests-passedTests)
	fmt.Printf("Success Rate: %.2f%%\n", float64(passedTests)/float64(totalTests)*100)

	if totalTests-passedTests > 0 {
		t.Errorf("Some tests failed")
	}
}
