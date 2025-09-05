# 小游戏服务端 - Go版本

🎮 基于Beego框架开发的企业级小游戏服务端系统，支持多种小游戏平台（微信小游戏、支付宝小游戏、H5游戏等）。

## ✨ 项目特点

- **🏗️ 双服务架构**：管理后台服务(8080) + 游戏SDK服务(8081)
- **🎯 完整功能**：用户系统、排行榜、计数器、邮件、配置管理
- **⚡ 高性能**：MySQL + Redis，支持水平扩展
- **🔄 兼容性强**：保持原有API接口格式，无缝迁移
- **🔐 安全可靠**：JWT认证、API签名验证、参数校验
- **🚀 企业级**：完善的监控、日志、部署方案

## 🛠️ 技术栈

- **后端框架**：Beego v2.2.0
- **数据库**：MySQL 5.7+ / Redis 6.0+
- **编程语言**：Go 1.21+
- **部署方式**：二进制部署 / Docker部署 / Kubernetes

## 📁 项目结构

```
go-admin/
├── admin-service/          # 🎛️ 管理后台服务 (端口8080)
│   ├── controllers/        # 控制器层
│   ├── models/            # 数据模型层
│   ├── services/          # 业务逻辑层
│   ├── middlewares/       # 中间件
│   ├── utils/             # 工具函数
│   ├── routers/           # 路由配置
│   └── conf/              # 配置文件
├── game-service/          # 🎮 游戏SDK服务 (端口8081)
│   ├── controllers/       # 控制器层
│   ├── models/           # 数据模型层
│   ├── services/         # 业务逻辑层
│   ├── middlewares/      # 中间件
│   ├── utils/            # 工具函数
│   ├── routers/          # 路由配置
│   └── conf/             # 配置文件
├── migration-tools/       # 📦 数据迁移工具
├── scripts/              # 🔧 编译和部署脚本
│   ├── build.sh         # Linux编译脚本
│   ├── build.bat        # Windows编译脚本
│   ├── deploy.sh        # 部署脚本
│   └── sql/             # SQL初始化脚本
└── docs/                # 📚 项目文档
```

## 🚀 快速开始

### 1. 环境要求

- **Go**: 1.21+
- **MySQL**: 5.7+ 或 8.0+
- **Redis**: 6.0+
- **系统**: Linux/Windows/macOS

### 2. 数据库初始化

```bash
# 创建数据库
mysql -u root -p < scripts/sql/init.sql

# 插入初始数据
mysql -u root -p < scripts/sql/seed.sql
```

### 3. 编译项目

```bash
# Linux/macOS
chmod +x scripts/build.sh
./scripts/build.sh

# Windows
scripts\build.bat
```

### 4. 修改配置

编辑配置文件，修改数据库连接信息：

```ini
# 管理后台服务配置 (bin/admin-app.conf)
mysql_host = 127.0.0.1
mysql_port = 3306
mysql_user = root
mysql_password = your_password
mysql_database = minigame_server

# 游戏SDK服务配置 (bin/game-app.conf)
mysql_host = 127.0.0.1
mysql_port = 3306
mysql_user = root
mysql_password = your_password
mysql_database = minigame_server
```

### 5. 启动服务

```bash
# 方式一：直接启动
cd bin
./admin-service &    # 管理后台服务 (端口8080)
./game-service &     # 游戏SDK服务 (端口8081)

# 方式二：使用部署脚本 (推荐)
sudo chmod +x scripts/deploy.sh
sudo ./scripts/deploy.sh

# 方式三：Docker部署
docker-compose up -d
```

### 6. 验证安装

```bash
# 检查服务状态
curl http://localhost:8080/health
curl http://localhost:8081/health

# 访问管理后台
# 打开浏览器访问: http://localhost:8080
# 默认账号: admin / admin123
```

## 🎯 核心功能

### 管理后台服务 (8080端口)

- **👤 用户管理**: 管理员账户、角色权限、操作审计
- **📱 应用管理**: 多应用支持、密钥管理、状态控制
- **🎮 游戏数据**: 用户数据查看、排行榜管理、计数器控制
- **📧 邮件系统**: 系统邮件、个人邮件、奖励发放
- **⚙️ 配置管理**: 游戏配置、系统参数、版本控制
- **📊 数据统计**: 用户统计、活跃分析、趋势图表
- **📁 文件管理**: 文件上传、图片管理、资源存储
- **🔔 通知系统**: 系统通知、消息推送、模板管理

### 游戏SDK服务 (8081端口)

- **💾 用户数据**: 保存/获取用户游戏数据，支持JSON格式
- **🏆 排行榜**: 分数提交、排名查询、多榜单支持
- **🔢 计数器**: 全局/用户计数器、自动重置、批量操作
- **📮 邮件**: 获取邮件、阅读状态、奖励领取
- **🎛️ 配置**: 游戏配置获取、版本控制、热更新

## 🔧 API接口

### 管理后台API

```bash
# 管理员登录
POST /auth/login
{
  "username": "admin",
  "password": "admin123"
}

# 创建应用
POST /application/create
{
  "app_id": "my_game",
  "app_name": "我的游戏",
  "description": "游戏描述"
}

# 获取用户列表
POST /game-data/users
{
  "app_id": "my_game",
  "page": 1,
  "page_size": 20
}
```

### 游戏SDK API

```bash
# 保存用户数据
POST /saveData
Headers: App-Id, Timestamp, Sign
{
  "userId": "user123",
  "data": {"level": 10, "score": 1000}
}

# 提交分数到排行榜
POST /commitScore
Headers: App-Id, Timestamp, Sign
{
  "userId": "user123",
  "leaderboard": "global",
  "score": 1000
}

# 获取排行榜
POST /getLeaderboard
Headers: App-Id, Timestamp, Sign
{
  "leaderboard": "global",
  "limit": 10
}
```

## 📦 部署方案

### 传统部署

详见 [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)

### Docker部署

```bash
# 快速启动
docker-compose up -d

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 生产环境部署

- **负载均衡**: Nginx反向代理
- **高可用**: 多实例部署
- **数据库**: MySQL主从复制
- **缓存**: Redis集群
- **监控**: Prometheus + Grafana
- **日志**: ELK Stack

## 🔄 数据迁移

如果您需要从MongoDB迁移到MySQL：

```bash
# 配置迁移工具
vim migration-tools/config/migration.yaml

# 执行全量迁移
./migration-tools --config config/migration.yaml --mode full

# 验证数据
./migration-tools --config config/migration.yaml --mode verify
```

## 📊 监控和维护

### 健康检查

```bash
# 服务健康状态
curl http://localhost:8080/health
curl http://localhost:8081/health
```

### 性能监控

- **CPU使用率**: `top`, `htop`
- **内存使用**: `free -h`
- **磁盘空间**: `df -h`
- **网络连接**: `netstat -tlnp`

### 日志管理

```bash
# 查看应用日志
tail -f /opt/minigame-server/logs/admin-service.log
tail -f /opt/minigame-server/logs/game-service.log

# 查看系统日志
journalctl -u minigame-admin -f
journalctl -u minigame-game -f
```

## 🔐 安全特性

- **API签名验证**: MD5签名防止接口被恶意调用
- **JWT认证**: 管理后台安全认证
- **参数验证**: 严格的输入参数校验
- **SQL注入防护**: ORM框架自动防护
- **XSS防护**: 输出内容自动转义
- **CORS配置**: 跨域请求控制
- **限流保护**: API调用频率限制

## 📈 性能优化

- **数据库连接池**: 复用数据库连接
- **Redis缓存**: 热点数据缓存
- **批量操作**: 减少数据库IO
- **索引优化**: 合理的数据库索引
- **压缩传输**: Gzip响应压缩
- **静态资源**: CDN加速

## 🧪 测试

```bash
# 单元测试
go test ./admin-service/...
go test ./game-service/...

# 集成测试
go test -tags=integration ./tests/...

# 压力测试
ab -n 1000 -c 10 http://localhost:8081/health
```

## 📚 文档

- [部署指南](DEPLOYMENT_GUIDE.md) - 详细的部署说明
- [API文档](API.md) - 完整的API接口文档
- [配置说明](CONFIG.md) - 配置参数详解
- [故障排除](TROUBLESHOOTING.md) - 常见问题解决

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🎉 致谢

感谢以下开源项目：

- [Beego](https://github.com/beego/beego) - Web框架
- [MySQL](https://www.mysql.com/) - 数据库
- [Redis](https://redis.io/) - 缓存数据库
- [Go](https://golang.org/) - 编程语言

---

⭐ 如果这个项目对您有帮助，请给个Star支持一下！ 