# Yalla SDK 集成模块

## 概述

本模块为 Game Service 提供 Yalla SDK 集成功能，支持用户认证、奖励发放、数据同步、事件上报等核心功能。

## 架构设计

```
game-service/yalla/
├── models/          # 数据模型
│   ├── yalla_config.go     # 配置模型
│   ├── yalla_request.go    # 请求结构
│   └── yalla_response.go   # 响应结构
├── services/        # 服务层
│   └── yalla_service.go    # 核心服务逻辑
├── controllers/     # 控制器
│   └── yalla_controller.go # API控制器
├── utils/          # 工具包
│   ├── yalla_crypto.go     # 加密工具
│   └── yalla_http.go       # HTTP客户端
└── README.md       # 说明文档
```

## 核心功能

### 1. 用户认证
- **接口**: `POST /api/yalla/auth`
- **功能**: 验证Yalla用户令牌，获取用户信息
- **参数**: app_id, user_id, auth_token

### 2. 获取用户信息
- **接口**: `GET /api/yalla/user/info`
- **功能**: 获取Yalla用户详细信息
- **参数**: app_id, yalla_user_id

### 3. 发放奖励
- **接口**: `POST /api/yalla/reward/send`
- **功能**: 向Yalla用户发放游戏奖励
- **参数**: app_id, yalla_user_id, reward_type, reward_amount, reward_data, description

### 4. 同步游戏数据
- **接口**: `POST /api/yalla/data/sync`
- **功能**: 同步游戏数据到Yalla平台
- **参数**: app_id, yalla_user_id, data_type, game_data

### 5. 上报事件
- **接口**: `POST /api/yalla/event/report`
- **功能**: 上报游戏事件到Yalla平台
- **参数**: app_id, yalla_user_id, event_type, event_data

### 6. 用户绑定管理
- **接口**: `GET /api/yalla/user/binding`
- **功能**: 获取游戏用户与Yalla用户的绑定关系
- **参数**: app_id, game_user_id

### 7. 配置管理
- **接口**: `GET /api/yalla/config`
- **功能**: 获取Yalla SDK配置信息
- **参数**: app_id

## 数据模型

### YallaConfig - Yalla配置
```go
type YallaConfig struct {
    ID          int       `json:"id"`
    AppID       string    `json:"app_id"`
    APIKey      string    `json:"api_key"`
    SecretKey   string    `json:"secret_key"`
    BaseURL     string    `json:"base_url"`
    Timeout     int       `json:"timeout"`
    RetryCount  int       `json:"retry_count"`
    EnableLog   bool      `json:"enable_log"`
    Status      int       `json:"status"`
    Remark      string    `json:"remark"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### YallaUserBinding - 用户绑定
```go
type YallaUserBinding struct {
    ID          int       `json:"id"`
    AppID       string    `json:"app_id"`
    GameUserID  string    `json:"game_user_id"`
    YallaUserID string    `json:"yalla_user_id"`
    YallaToken  string    `json:"yalla_token"`
    ExpiresAt   time.Time `json:"expires_at"`
    Status      int       `json:"status"`
    BindAt      time.Time `json:"bind_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### YallaCallLog - API调用日志
```go
type YallaCallLog struct {
    ID           int       `json:"id"`
    AppID        string    `json:"app_id"`
    UserID       string    `json:"user_id"`
    Method       string    `json:"method"`
    Endpoint     string    `json:"endpoint"`
    RequestData  string    `json:"request_data"`
    ResponseData string    `json:"response_data"`
    StatusCode   int       `json:"status_code"`
    Duration     int64     `json:"duration"`
    Success      bool      `json:"success"`
    ErrorMsg     string    `json:"error_msg"`
    CreatedAt    time.Time `json:"created_at"`
}
```

## 安全机制

### 1. 签名验证
- 使用MD5算法对请求参数进行签名
- 包含时间戳防止重放攻击
- 签名有效期5分钟

### 2. 令牌管理
- 用户令牌加密存储
- 支持令牌过期检查
- 自动刷新机制

### 3. 请求加密
- 敏感数据传输加密
- HTTPS通信保障
- 参数验证机制

## 配置管理

### 环境配置
```yaml
# conf/yalla_config.yaml
yalla:
  default:
    base_url: "https://api.yalla.com"
    timeout: 30
    retry_count: 3
    enable_log: true
    
  development:
    base_url: "https://dev-api.yalla.com"
    timeout: 10
    retry_count: 1
    enable_log: true
```

### 应用配置
- 每个应用独立配置
- 支持多环境部署
- 动态配置更新

## 错误处理

### 错误码定义
- `400`: 参数错误
- `401`: 认证失败
- `404`: 资源不存在
- `500`: 服务器内部错误
- `501`: 功能未实现

### 重试机制
- 网络错误自动重试
- 可配置重试次数
- 指数退避策略

## 日志记录

### 调用日志
- 记录所有API调用
- 包含请求/响应数据
- 性能监控数据

### 错误日志
- 详细错误信息
- 调用堆栈跟踪
- 便于问题排查

## 测试支持

### 单元测试
- 覆盖核心业务逻辑
- Mock外部依赖
- 自动化测试流程

### 集成测试
- 端到端测试用例
- 真实环境验证
- 性能压力测试

### 测试工具
- Game Services Tester 支持
- 完整的测试面板
- 示例数据生成

## 部署说明

### 数据库初始化
```bash
# 执行SQL初始化脚本
mysql -u username -p database_name < sql/yalla_init.sql
```

### 服务启动
```bash
# 启动Game Service
cd game-service
go run main.go
```

### 测试验证
```bash
# 启动测试工具
cd game-services-tester
go run main.go
```

## 监控与维护

### 性能监控
- API响应时间
- 成功率统计
- 错误率监控

### 数据清理
- 定期清理过期日志
- 清理无效绑定关系
- 数据备份策略

### 版本更新
- 向后兼容保证
- 平滑升级机制
- 回滚方案准备

## 故障排查

### 常见问题
1. **认证失败**: 检查AppID和SecretKey配置
2. **网络超时**: 调整超时时间和重试次数
3. **数据格式错误**: 验证JSON格式和必需字段
4. **令牌过期**: 检查令牌有效期设置

### 日志分析
```bash
# 查看API调用日志
SELECT * FROM yalla_call_logs WHERE success = 0 ORDER BY created_at DESC LIMIT 10;

# 查看错误统计
SELECT endpoint, COUNT(*) as error_count FROM yalla_call_logs 
WHERE success = 0 AND created_at >= DATE_SUB(NOW(), INTERVAL 1 DAY) 
GROUP BY endpoint ORDER BY error_count DESC;
```

## 联系方式

如有问题或建议，请联系开发团队。
