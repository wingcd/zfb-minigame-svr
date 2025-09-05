# Minigame Admin Service

一个功能完整的小游戏管理后台服务，支持自动化部署和数据库初始化。

## ✨ 特性

- 🚀 **一键部署**: 支持自动安装和配置
- 🗄️ **数据库管理**: 自动初始化MySQL数据库和表结构
- 👤 **用户管理**: 管理员账号和权限管理
- 🎮 **游戏数据**: 用户数据、排行榜、计数器管理
- 📊 **统计分析**: 数据统计和可视化
- 🔧 **系统配置**: 灵活的配置管理
- 🌐 **Web界面**: 现代化的安装和管理界面

## 🚀 快速开始

### 1. 下载和构建

```bash
# 克隆项目
git clone <repository-url>
cd admin-service

# 安装依赖
go mod tidy

# 构建项目
go build -o bin/admin-service .
```

### 2. 安装部署

#### 选项1: 使用安装脚本（推荐）

**Linux/macOS:**
```bash
chmod +x install.sh
./install.sh
```

**Windows:**
```cmd
install.bat
```

#### 选项2: 命令行安装
```bash
# 自动安装
./bin/admin-service -install

# 启动服务
./bin/admin-service
```

#### 选项3: Web界面安装
```bash
# 启动服务
./bin/admin-service

# 浏览器访问
http://localhost:8080/install
```

### 3. 访问服务

- **管理界面**: http://localhost:8080
- **默认账号**: admin / admin123
- **健康检查**: http://localhost:8080/health

## 📋 系统要求

- **Go**: 1.21+
- **MySQL**: 8.0+ (推荐)
- **Redis**: 6.0+ (可选)
- **操作系统**: Linux, macOS, Windows

## 🛠️ 配置说明

### 数据库配置

```ini
# MySQL配置
mysql_host = 127.0.0.1
mysql_port = 3306
mysql_user = root
mysql_password = your_password
mysql_database = minigame_admin
```

### 服务配置

```ini
# 基本配置
httpport = 8080
runmode = prod

# 自动安装配置
auto_install = true
auto_create_database = true
auto_create_admin = true
default_admin_username = admin
default_admin_password = admin123
```

## 📁 项目结构

```
admin-service/
├── bin/                    # 可执行文件
├── conf/                   # 配置文件
│   └── app.conf
├── controllers/            # 控制器
│   ├── auth.go
│   ├── install.go
│   └── ...
├── models/                 # 数据模型
├── utils/                  # 工具函数
│   ├── installer.go        # 安装工具
│   └── crypto.go          # 加密工具
├── views/                  # 模板文件
│   └── install/
├── routers/                # 路由配置
├── data/                   # 数据文件
├── logs/                   # 日志文件
└── uploads/                # 上传文件
```

## 🔧 命令行工具

```bash
# 显示版本信息
./bin/admin-service -version

# 显示帮助
./bin/admin-service -help

# 检查安装状态
./bin/admin-service -status

# 自动安装
./bin/admin-service -install

# 卸载系统
./bin/admin-service -uninstall
```

## 🌐 API接口

### 安装相关
- `GET /install` - 安装页面
- `GET /install/status` - 检查安装状态
- `POST /install/auto` - 自动安装
- `POST /install/manual` - 手动安装
- `POST /install/test` - 测试数据库连接

### 认证相关
- `POST /admin/login` - 管理员登录
- `POST /admin/verifyToken` - 验证Token

### 应用管理
- `POST /app/getAll` - 获取所有应用
- `POST /app/create` - 创建应用
- `POST /app/update` - 更新应用
- `POST /app/delete` - 删除应用

### 用户管理
- `POST /user/getAll` - 获取所有用户
- `POST /user/ban` - 封禁用户
- `POST /user/unban` - 解封用户

## 📊 数据库表结构

| 表名 | 说明 |
|------|------|
| `admins` | 管理员账号 |
| `apps` | 应用信息 |
| `user_data` | 用户数据 |
| `leaderboards` | 排行榜 |
| `system_configs` | 系统配置 |

## 🔒 安全特性

- JWT令牌认证
- 密码加密存储
- CORS跨域保护
- 请求频率限制
- SQL注入防护

## 🚨 故障排除

### 常见问题

1. **MySQL连接失败**
   - 检查MySQL服务状态
   - 确认用户名密码
   - 检查网络连接

2. **端口被占用**
   ```bash
   # 查找占用进程
   lsof -i :8080
   # 修改配置端口
   vim conf/app.conf
   ```

3. **权限问题**
   ```bash
   chmod +x bin/admin-service
   chmod +x *.sh
   ```

### 日志查看
```bash
# 服务日志
tail -f logs/admin-service.log

# 安装日志
cat logs/install.log
```

## 🔄 更新升级

```bash
# 1. 停止服务
./stop.sh

# 2. 备份数据
cp -r data/ data_backup/

# 3. 更新程序
# 替换新版本文件

# 4. 启动服务
./start.sh
```

## 🏗️ 开发指南

### 环境准备
```bash
# 安装Go依赖
go mod tidy

# 运行测试
go test ./...

# 开发模式运行
go run main.go
```

### 添加新功能
1. 在 `controllers/` 添加控制器
2. 在 `models/` 添加数据模型
3. 在 `routers/router.go` 添加路由
4. 更新数据库表结构（如需要）

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 技术支持

如遇问题，请提供：
- 操作系统版本
- Go版本信息
- 错误日志内容
- 安装状态信息

---

Made with ❤️ for Minigame Developers
