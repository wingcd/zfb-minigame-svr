# 小游戏服务器部署说明

## 项目概述

这是一个基于Golang + Beego + MySQL + Redis的小游戏服务器，支持动态按应用创建数据表的架构设计。

## 系统架构

### 服务分离
- **管理后台服务 (admin-service)**: 端口8080，提供Web管理界面API
- **游戏SDK服务 (game-service)**: 端口8081，提供游戏SDK接口

### 数据表设计
- **系统固定表**: 管理员、角色、应用等基础管理表
- **动态游戏表**: 为每个应用自动创建独立的游戏数据表
  - `user_data_{app_id}` - 用户数据表
  - `leaderboard_{app_id}` - 排行榜表  
  - `counter_{app_id}` - 计数器表
  - `mail_{app_id}` - 邮件表
  - `game_config_{app_id}` - 游戏配置表

## 环境要求

### 基础环境
- Go 1.17+ (推荐Go 1.19+)
- MySQL 5.7+ 或 8.0+
- Redis 5.0+

### 依赖包
```bash
go mod tidy
```

## 安装部署

### 1. 数据库初始化

```bash
# 连接MySQL并执行初始化脚本
mysql -u root -p < scripts/sql/init.sql
```

初始化脚本会：
- 创建数据库 `minigame_server`
- 创建系统管理表
- 插入默认管理员账户 (用户名: admin, 密码: admin123)

### 2. 配置文件

复制并修改配置文件：

**admin-service/conf/app.conf**
```ini
appname = admin-service
httpport = 8080
runmode = dev

# MySQL配置
mysql_host = localhost
mysql_port = 3306
mysql_user = root
mysql_password = your_password
mysql_database = minigame_server
mysql_charset = utf8mb4

# Redis配置
redis_addr = localhost:6379
redis_password = 
redis_db = 0

# JWT配置
jwt_secret = your_jwt_secret_key
jwt_expire = 24h

# 加密配置
encrypt_key = your_encrypt_key
sign_salt = your_sign_salt
```

**game-service/conf/app.conf**
```ini
appname = game-service
httpport = 8081
runmode = dev

# MySQL配置
mysql_host = localhost
mysql_port = 3306
mysql_user = root
mysql_password = your_password
mysql_database = minigame_server
mysql_charset = utf8mb4

# Redis配置
redis_addr = localhost:6379
redis_password = 
redis_db = 1

# 加密配置
encrypt_key = your_encrypt_key
sign_salt = your_sign_salt
```

### 3. 编译运行

#### Windows
```bash
# 编译
.\scripts\build.bat

# 或直接运行
cd admin-service && go run main.go
cd game-service && go run main.go
```

#### Linux
```bash
# 编译
chmod +x scripts/build.sh
./scripts/build.sh

# 或直接运行
cd admin-service && go run main.go &
cd game-service && go run main.go &
```

### 4. 验证部署

- 管理后台: http://localhost:8080
- 游戏服务: http://localhost:8081/health
- 默认管理员: admin / admin123

## 使用流程

### 1. 创建应用

通过管理后台创建新应用，系统会自动：
- 生成应用ID和密钥
- 创建该应用的专属数据表
- 初始化基础配置

### 2. SDK集成

```javascript
// 设置SDK基础URL
const baseURL = "http://your-server:8081"

// 保存用户数据
await saveData(userId, data, appId, appSecret)

// 获取用户数据
const userData = await getData(userId, appId, appSecret)

// 提交排行榜分数
await submitScore(leaderboardName, score, userId, appId, appSecret)
```

### 3. 数据隔离

每个应用的数据完全隔离：
- 用户数据存储在独立表中
- 排行榜按应用分离
- 配置和邮件系统独立

## API接口

### 游戏服务API (8081端口)

#### 用户数据
- `POST /saveData` - 保存用户数据
- `POST /getData` - 获取用户数据  
- `POST /deleteData` - 删除用户数据

#### 排行榜
- `POST /submitScore` - 提交分数
- `POST /getLeaderboard` - 获取排行榜
- `POST /getUserRank` - 获取用户排名
- `POST /resetLeaderboard` - 重置排行榜

#### 计数器
- `POST /getCounter` - 获取计数器
- `POST /incrementCounter` - 增加计数器
- `POST /decrementCounter` - 减少计数器
- `POST /setCounter` - 设置计数器
- `POST /resetCounter` - 重置计数器

#### 邮件系统
- `POST /getMailList` - 获取邮件列表
- `POST /readMail` - 读取邮件
- `POST /claimRewards` - 领取奖励
- `POST /deleteMail` - 删除邮件

#### 游戏配置
- `POST /getConfig` - 获取配置
- `POST /setConfig` - 设置配置
- `POST /getConfigsByVersion` - 获取版本配置

### 管理后台API (8080端口)

#### 认证
- `POST /api/auth/login` - 管理员登录
- `POST /api/auth/logout` - 登出

#### 应用管理
- `GET /api/applications` - 应用列表
- `POST /api/applications` - 创建应用
- `PUT /api/applications/:id` - 更新应用
- `DELETE /api/applications/:id` - 删除应用

## 签名验证

所有游戏服务API都需要MD5签名验证：

```javascript
// 签名算法
function generateSign(params, appSecret) {
    const sortedKeys = Object.keys(params).sort()
    let signStr = ""
    
    for (const key of sortedKeys) {
        if (key !== "sign" && params[key] !== "") {
            signStr += key + "=" + params[key] + "&"
        }
    }
    
    signStr += "key=" + appSecret
    return md5(signStr).toUpperCase()
}
```

## 监控和维护

### 健康检查
- 游戏服务: `GET /health`
- 管理服务: `GET /api/health`

### 日志管理
- 应用日志: `logs/`目录
- 操作日志: 数据库`admin_operation_logs`表

### 性能优化
- Redis缓存热点数据
- 数据库连接池配置
- 定期清理过期数据

## 常见问题

### 1. 编译错误
```bash
# Go版本兼容性问题
go mod tidy
go clean -modcache
```

### 2. 数据库连接
- 检查MySQL配置和权限
- 确认数据库已正确初始化

### 3. Redis连接
- 确认Redis服务状态
- 检查连接配置

### 4. 应用表创建失败
- 检查数据库权限
- 确认表名规范（app_id不能包含特殊字符）

## 安全建议

1. **修改默认密码**: 首次部署后立即修改admin密码
2. **配置HTTPS**: 生产环境启用HTTPS
3. **防火墙设置**: 限制数据库和Redis的外部访问
4. **定期备份**: 设置数据库自动备份
5. **监控日志**: 监控异常请求和错误日志

## 扩展开发

### 添加新功能模块
1. 在models中定义数据模型
2. 在controllers中实现业务逻辑  
3. 在routers中注册路由
4. 更新数据表创建脚本

### 自定义表结构
修改`admin-service/models/application.go`中的`createAppTables`方法。

---

更多技术细节请参考代码注释和API文档。 