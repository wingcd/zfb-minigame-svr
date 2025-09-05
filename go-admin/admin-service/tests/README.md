# Admin Service API 测试环境

基于云函数接口格式，为admin-service的Go接口创建的完整测试环境和测试用例。

## 🎯 项目概述

本测试环境参考了云函数的输入输出格式，为admin-service中的所有接口创建了标准化的测试用例。测试覆盖了用户管理、系统管理、统计分析等核心功能模块。

## 📁 项目结构

```
tests/
├── README.md              # 本文档
├── test_framework.go       # 测试框架核心
├── test_cases.go          # 所有测试用例定义
├── test_data.go           # 测试数据管理
├── main_test.go           # 主测试文件
├── test_config.json       # 测试配置
└── run_tests.sh           # 测试运行脚本
```

## 🚀 快速开始

### 1. 环境准备

确保你的系统已安装：
- Go 1.16+
- MySQL 5.7+
- Git

### 2. 数据库配置

创建测试数据库：
```sql
CREATE DATABASE admin_service_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 运行测试

```bash
# 运行所有测试
./run_tests.sh

# 运行特定测试模块
go test -v ./tests/ -run TestUserAPIs      # 用户管理测试
go test -v ./tests/ -run TestSystemAPIs    # 系统管理测试
go test -v ./tests/ -run TestStatisticsAPIs # 统计分析测试

# 运行性能测试
./run_tests.sh --benchmark

# 运行覆盖率测试
./run_tests.sh --coverage

# 运行集成测试
./run_tests.sh --integration
```

## 📊 接口映射对照

| 云函数名称 | Admin Service接口 | 功能说明 | 测试状态 |
|-----------|------------------|----------|----------|
| `adminLogin` | `POST /api/auth/login` | 管理员登录验证 | ✅ |
| `getAllUsers` | `POST /api/users/list` | 获取用户列表（分页） | ✅ |
| `getUserDetail` | `POST /api/users/detail` | 获取用户详情 | ✅ |
| `setUserDetail` | `POST /api/users/update` | 更新用户数据 | ✅ |
| `banUser` | `POST /api/users/ban` | 封禁用户 | ✅ |
| `unbanUser` | `POST /api/users/unban` | 解封用户 | ✅ |
| `deleteUser` | `DELETE /api/users/delete` | 删除用户 | ✅ |
| `getUserStats` | `POST /api/users/stats` | 获取用户统计 | ✅ |
| `getDashboardStats` | `GET /api/statistics/dashboard` | 获取仪表盘数据 | ✅ |
| `getApplicationStats` | `GET /api/statistics/app` | 获取应用统计 | ✅ |
| `getOperationLogs` | `GET /api/statistics/logs` | 获取操作日志 | ✅ |
| `getSystemConfig` | `GET /api/system/config` | 获取系统配置 | ✅ |
| `updateSystemConfig` | `POST /api/system/config/update` | 更新系统配置 | ✅ |

## 🏗️ 标准响应格式

所有接口都遵循云函数的标准响应格式：

```json
{
  "code": 0,                    // 状态码，0表示成功
  "msg": "success",             // 响应消息
  "timestamp": 1603991234567,   // 时间戳
  "data": {                     // 实际数据
    // 具体的响应内容
  }
}
```

### 常用状态码

| 状态码 | 含义 | 说明 |
|-------|------|------|
| `0` | 成功 | 请求处理成功 |
| `4001` | 参数错误 | 请求参数不正确 |
| `4002` | 认证失败 | Token无效或过期 |
| `4003` | 权限不足 | 没有访问权限 |
| `4004` | 资源不存在 | 请求的资源不存在 |
| `5001` | 服务器错误 | 内部服务器错误 |

## 📝 测试用例结构

每个测试用例包含以下信息：

```go
type TestCase struct {
    Name           string                 // 测试名称
    Description    string                 // 测试描述
    Method         string                 // HTTP方法
    URL            string                 // 请求URL
    Headers        map[string]string      // 请求头
    RequestData    map[string]interface{} // 请求数据
    ExpectedCode   int                    // 期望的状态码
    ExpectedMsg    string                 // 期望的消息
    ValidateData   func(interface{}) bool // 数据验证函数
    SetupFunc      func() error           // 前置条件设置
    CleanupFunc    func() error           // 后置清理
    RequiresAuth   bool                   // 是否需要认证
    RequiresAdmin  bool                   // 是否需要管理员权限
    Tags           []string               // 测试标签
}
```

## 🧪 测试数据

### 测试应用

- `test_app_001`: 主要测试应用，包含完整的测试数据
- `test_app_002`: 次要测试应用，用于多应用场景测试
- `test_app_performance`: 性能测试专用应用

### 测试用户

- `test_player_001`: 普通测试用户
- `test_player_ban`: 用于封禁测试的用户
- `test_player_unban`: 用于解封测试的用户
- `test_player_delete`: 用于删除测试的用户
- `test_player_stats`: 用于统计测试的用户

### 测试管理员

- `test_admin`: 普通管理员账户
- `test_super_admin`: 超级管理员账户
- 密码统一为: `test123456`

## 🔧 配置说明

### 数据库配置

测试使用独立的测试数据库 `admin_service_test`，配置在 `test_config.json` 中：

```json
{
  "database": {
    "driver": "mysql",
    "host": "127.0.0.1",
    "port": 3306,
    "user": "root",
    "password": "",
    "database": "admin_service_test",
    "charset": "utf8mb4"
  }
}
```

### JWT配置

测试环境使用独立的JWT密钥：

```json
{
  "jwt": {
    "secret": "test_jwt_secret_key_2023",
    "expireHours": 24
  }
}
```

## 📈 性能测试

### 基准测试

```bash
# 运行所有性能测试
go test -v ./tests/ -bench=. -benchtime=10s

# 运行特定接口的性能测试
go test -v ./tests/ -bench=BenchmarkUserAPIs
```

### 性能指标

- **响应时间**: < 1000ms
- **吞吐量**: > 100 req/s
- **错误率**: < 1%

## 📊 覆盖率报告

生成覆盖率报告：

```bash
# 生成覆盖率数据
go test -coverprofile=coverage.out ./tests/

# 生成HTML报告
go tool cover -html=coverage.out -o coverage.html

# 查看覆盖率统计
go tool cover -func=coverage.out
```

## 🐛 调试和故障排除

### 常见问题

1. **数据库连接失败**
   ```
   解决方案：检查MySQL服务是否运行，数据库配置是否正确
   ```

2. **测试数据创建失败**
   ```
   解决方案：确保有足够的数据库权限，检查表结构是否正确
   ```

3. **JWT认证失败**
   ```
   解决方案：检查JWT密钥配置，确保token生成逻辑正确
   ```

### 调试模式

启用详细日志：

```bash
# 设置环境变量
export GO_ENV=test
export LOG_LEVEL=debug

# 运行测试
go test -v ./tests/ -timeout=30m
```

## 🔄 CI/CD 集成

### GitHub Actions

```yaml
name: API Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:5.7
        env:
          MYSQL_ALLOW_EMPTY_PASSWORD: yes
          MYSQL_DATABASE: admin_service_test
        ports:
          - 3306:3306
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.18
      - name: Run tests
        run: ./run_tests.sh --coverage
```

## 📚 扩展测试用例

### 添加新的测试用例

1. 在 `test_cases.go` 中添加测试用例定义
2. 在 `test_data.go` 中添加所需的测试数据
3. 更新 `test_config.json` 中的配置（如需要）
4. 运行测试验证

示例：

```go
// 在 GetUserTestSuite() 中添加新测试用例
{
    Name:        "NewFeature_Success",
    Description: "测试新功能的成功场景",
    Method:      "POST",
    URL:         "/api/users/new-feature",
    RequestData: map[string]interface{}{
        "appId":    "test_app_001",
        "playerId": "test_player_001",
        "feature":  "new_feature_data",
    },
    ExpectedCode: 0,
    ExpectedMsg:  "操作成功",
    RequiresAuth: true,
    Tags:         []string{"user", "new-feature", "success"},
},
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/new-test`)
3. 提交更改 (`git commit -am 'Add new test case'`)
4. 推送到分支 (`git push origin feature/new-test`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详细信息请查看 LICENSE 文件。

---

## 🆘 获取帮助

如果你在使用过程中遇到问题：

1. 查看本文档的故障排除部分
2. 检查测试日志输出
3. 在项目中创建 Issue
4. 联系开发团队

---

**Happy Testing! 🎉**
