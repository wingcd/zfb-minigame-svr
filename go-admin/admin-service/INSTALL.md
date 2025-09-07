# Minigame Admin Service 安装指南

## 概述

Minigame Admin Service 是一个小游戏管理后台服务，支持自动化部署和数据库初始化。

## 系统要求

- **Go**: 1.21+ 
- **数据库**: MySQL 8.0+ 或 SQLite (自动后备)
- **缓存**: Redis 6.0+ (可选，未安装时使用内存缓存)
- **操作系统**: Linux, macOS, Windows

## 快速安装

### 方法一：使用安装脚本（推荐）

#### Linux/macOS
```bash
# 运行安装脚本
chmod +x install.sh
./install.sh
```

#### Windows
```cmd
# 运行安装脚本
install.bat
```

### 方法二：命令行安装

```bash
# 1. 构建项目
go build -o bin/admin-service .

# 2. 自动安装
./bin/admin-service -install

# 3. 启动服务
./bin/admin-service
```

### 方法三：Web界面安装

```bash
# 1. 构建并启动服务
go build -o bin/admin-service .
./bin/admin-service

# 2. 访问安装页面
# 浏览器打开: http://localhost:8080/install
```

## 安装选项

### 自动安装
- 自动检测MySQL连接
- MySQL不可用时自动使用SQLite
- 创建默认管理员账号 (admin/admin123)
- 生成随机密钥和密码

### 手动安装
- 自定义数据库配置
- 自定义管理员账号
- 选择是否创建示例数据

## 配置说明

### 数据库配置

#### MySQL配置
```ini
mysql_host = 127.0.0.1
mysql_port = 3306
mysql_user = minigame
mysql_password = your_password
mysql_database = minigame_admin
```

#### SQLite配置
SQLite无需额外配置，数据文件位于 `data/minigame.db`

### 服务配置
```ini
httpport = 8080
runmode = prod
auto_install = true
auto_create_database = true
auto_create_admin = true
```

## 启动和管理

### 启动服务
```bash
# 使用生成的脚本
./start.sh          # Linux/macOS
start.bat           # Windows

# 或直接运行
./bin/admin-service
```

### 停止服务
```bash
# 使用生成的脚本
./stop.sh           # Linux/macOS
stop.bat            # Windows

# 或手动停止
pkill admin-service  # Linux/macOS
taskkill /f /im admin-service.exe  # Windows
```

### 命令行选项
```bash
# 显示版本信息
./bin/admin-service -version

# 显示帮助信息
./bin/admin-service -help

# 检查安装状态
./bin/admin-service -status

# 自动安装
./bin/admin-service -install

# 卸载系统
./bin/admin-service -uninstall
```

## 访问服务

### 管理界面
- URL: http://localhost:8080
- 默认账号: admin
- 默认密码: admin123

### API接口
- 健康检查: http://localhost:8080/health
- API文档: http://localhost:8080/swagger/ (开发模式)

## 目录结构

```
admin-service/
├── bin/                    # 可执行文件
│   └── admin-service
├── conf/                   # 配置文件
│   └── app.conf
├── data/                   # 数据文件 (SQLite)
│   └── minigame.db
├── logs/                   # 日志文件
├── uploads/                # 上传文件
├── views/                  # 模板文件
│   └── install/
│       └── index.html
├── .install_info          # 安装信息
├── .installed             # 安装锁文件
├── start.sh               # 启动脚本 (Linux/macOS)
├── stop.sh                # 停止脚本 (Linux/macOS)
├── start.bat              # 启动脚本 (Windows)
├── stop.bat               # 停止脚本 (Windows)
├── install.sh             # 安装脚本 (Linux/macOS)
└── install.bat            # 安装脚本 (Windows)
```

## 数据库表结构

安装完成后会自动创建以下数据表：

- `admins`: 管理员账号
- `apps`: 应用信息
- `system_configs`: 系统配置

## 故障排除

### 常见问题

#### 1. Go环境未找到
```bash
# 安装Go (以Ubuntu为例)
wget https://golang.org/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 2. MySQL连接失败
- 检查MySQL服务是否启动
- 确认用户名密码正确
- 检查防火墙设置
- 系统会自动使用SQLite作为后备

#### 3. 端口被占用
```bash
# 查找占用端口的进程
netstat -tulpn | grep :8080
lsof -i :8080

# 修改配置文件中的端口
vim conf/app.conf
# httpport = 8081
```

#### 4. 权限问题
```bash
# 给予执行权限
chmod +x bin/admin-service
chmod +x *.sh

# 或使用sudo运行
sudo ./bin/admin-service
```

### 日志查看
```bash
# 查看服务日志
tail -f logs/admin-service.log

# 查看安装日志
cat logs/install.log
```

### 重新安装
```bash
# 卸载现有安装
./bin/admin-service -uninstall

# 或手动删除
rm -rf data/ logs/ .installed .install_info

# 重新安装
./install.sh
```

## 生产环境部署

### 安全建议
1. 修改默认管理员密码
2. 使用HTTPS (配置反向代理)
3. 配置防火墙规则
4. 定期备份数据库
5. 更新JWT密钥和API密钥

### 性能优化
1. 使用MySQL而非SQLite
2. 配置Redis缓存
3. 调整数据库连接池大小
4. 启用日志轮转

### 监控和维护
```bash
# 检查服务状态
./bin/admin-service -status

# 查看系统信息
curl http://localhost:8080/health

# 数据库优化
# 访问管理界面 -> 系统管理 -> 数据库优化
```

## 更新升级

```bash
# 1. 停止服务
./stop.sh

# 2. 备份数据
cp -r data/ data_backup/
cp conf/app.conf conf/app.conf.backup

# 3. 更新程序文件
# 下载新版本并替换

# 4. 启动服务
./start.sh
```

## 技术支持

如遇到问题，请提供以下信息：
- 操作系统版本
- Go版本 (`go version`)
- 错误日志 (`logs/admin-service.log`)
- 安装状态 (`./bin/admin-service -status`)

## 许可证

本项目采用 MIT 许可证。
