# Game Service API Tester

一个用于测试 game-service 所有API接口的Web测试工具。

## 🚀 功能特性

- **完整的API覆盖**: 测试所有game-service接口（除微信等平台相关接口）
- **模拟登录**: 支持模拟用户登录和身份验证
- **签名生成**: 自动生成API签名，支持zy-sdk风格和标准签名
- **批量测试**: 一键运行所有测试用例
- **实时结果**: 实时显示测试结果和统计信息
- **美观界面**: 现代化的Web界面，支持响应式设计

## 📋 支持的接口

### 🏥 健康检查
- `GET /health` - 服务健康检查
- `GET /heartbeat` - 心跳检测

### 👤 用户数据接口 (zy-sdk对齐)
- `POST /user/login` - 模拟用户登录
- `POST /user/saveData` - 保存用户游戏数据
- `POST /user/getData` - 获取用户游戏数据
- `POST /user/saveUserInfo` - 保存用户基本信息

### 🏆 排行榜接口 (zy-sdk对齐)
- `POST /leaderboard/commit` - 提交分数到排行榜
- `POST /leaderboard/queryTopRank` - 查询排行榜数据

### 🔢 计数器接口 (zy-sdk对齐)
- `POST /counter/increment` - 增加计数器值
- `GET /counter/get` - 获取计数器值

### 📧 邮件接口 (zy-sdk对齐)
- `GET /mail/getUserMails` - 获取用户邮件列表
- `POST /mail/updateStatus` - 更新邮件状态（阅读/领取/删除）

### ⚙️ 配置接口
- `POST /getConfig` - 获取指定配置
- `POST /getAllConfigs` - 获取所有配置

## 🛠️ 使用方法

### 1. 启动 Game Service
确保 game-service 服务正在运行：
```bash
cd ../game-service
go run main.go
```
默认端口：`8081`

### 2. 启动测试工具
```bash
# 进入测试工具目录
cd game-services-tester

# 运行测试服务器
go run main.go
```
默认端口：`8082`

### 3. 打开测试页面
在浏览器中访问：`http://localhost:8082`

## ⚙️ 配置说明

在测试页面顶部的配置区域中设置：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| 服务器地址 | game-service的运行地址 | `http://localhost:8081` |
| App ID | 应用标识符 | `test_app` |
| App Secret | 应用密钥 | `test_secret` |
| Player ID | 测试玩家ID | `test_player_001` |

## 🔐 签名机制

测试工具支持两种签名方式：

### 1. zy-sdk风格签名（用户接口）
用于 `/user/`、`/leaderboard/`、`/mail/` 等接口：
```javascript
// 构建参数: appId + userId + appSecret
// 按key排序后拼接: appId=xxx&appSecret=xxx&userId=xxx
// MD5加密生成签名
```

### 2. 标准API签名（其他接口）
用于配置、计数器等标准接口：
```javascript
// 将请求参数按key排序
// 添加timestamp和key参数
// MD5加密生成签名
```

## 📊 测试功能

### 单独测试
- 每个接口都有独立的测试按钮
- 可以自定义测试参数
- 实时显示测试结果

### 批量测试
- 点击"运行所有测试"按钮
- 自动执行所有测试用例
- 显示成功/失败统计

### 测试结果
- ✅ 绿色表示测试成功
- ❌ 红色表示测试失败
- 显示完整的JSON响应数据
- 实时更新测试统计

## 🔧 开发说明

### 项目结构
```
game-services-tester/
├── main.go                 # Go测试服务器
├── static/
│   ├── index.html          # 主测试页面
│   ├── api-tester.js       # API测试逻辑
│   ├── crypto.js           # 签名生成工具
│   └── style.css           # 页面样式
├── task.md                 # 任务说明
└── README.md               # 使用文档
```

### 自定义配置
你可以修改 `static/api-tester.js` 中的默认配置：
```javascript
let config = {
    baseUrl: 'http://localhost:8081',
    appId: 'test_app',
    appSecret: 'test_secret',
    playerId: 'test_player_001'
};
```

### 添加新测试
1. 在 `index.html` 中添加测试界面
2. 在 `api-tester.js` 中添加测试函数
3. 在 `runAllTests()` 中加入新测试

## 🚨 注意事项

1. **确保服务运行**: 测试前确保 game-service 正在 8081 端口运行
2. **应用配置**: 需要在 admin-service 中创建对应的测试应用
3. **数据库连接**: 确保 MySQL 和 Redis 服务正常运行
4. **跨域问题**: game-service 已配置CORS，支持跨域访问

## 🔍 故障排除

### 常见问题

1. **连接失败**
   - 检查 game-service 是否运行在 8081 端口
   - 确认服务器地址配置正确

2. **签名验证失败**
   - 检查 App ID 和 App Secret 是否正确
   - 确认应用在 admin-service 中已创建并启用

3. **数据库错误**
   - 确认 MySQL 服务运行正常
   - 检查数据库连接配置

4. **权限错误**
   - 确认测试应用有相应接口权限
   - 检查Player ID是否有效

### 调试技巧

1. 打开浏览器开发者工具查看网络请求
2. 检查控制台错误信息
3. 查看 game-service 的运行日志
4. 使用"清空结果"按钮重置测试状态

## 📝 更新日志

- **v1.0.0**: 初始版本，支持所有主要API接口测试
- 完整的zy-sdk接口对齐
- 美观的Web界面
- 批量测试功能
- 实时结果统计

## 🤝 贡献

欢迎提交Issue和Pull Request来改进这个测试工具！

## 📄 许可证

本项目采用 MIT 许可证。
