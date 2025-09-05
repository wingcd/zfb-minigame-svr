# 小游戏服务器部署指南

## 概述

本指南将帮助您完整部署小游戏服务器系统，包括管理后台服务（端口8080）和游戏SDK服务（端口8081）。

## 系统要求

### 硬件要求
- **CPU**: 2核心以上
- **内存**: 4GB以上（推荐8GB）
- **存储**: 20GB以上可用空间
- **网络**: 稳定的互联网连接

### 软件要求
- **操作系统**: Linux (Ubuntu 18.04+, CentOS 7+) 或 Windows Server 2016+
- **Go语言**: 1.21+
- **MySQL**: 5.7+ 或 8.0+
- **Redis**: 6.0+
- **Docker**: 20.10+ (可选，用于容器化部署)

## 部署方式

### 方式一：传统部署

#### 1. 环境准备

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y mysql-server redis-server git

# CentOS/RHEL
sudo yum install -y mysql-server redis git
sudo systemctl start mysqld redis
sudo systemctl enable mysqld redis
```

#### 2. 数据库初始化

```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE minigame_server DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'minigame'@'%' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON minigame_server.* TO 'minigame'@'%';
FLUSH PRIVILEGES;
EXIT;

# 初始化表结构
mysql -u root -p minigame_server < scripts/sql/init.sql
mysql -u root -p minigame_server < scripts/sql/seed.sql
```

#### 3. 编译项目

```bash
# Linux/macOS
chmod +x scripts/build.sh
./scripts/build.sh

# Windows
scripts\build.bat
```

#### 4. 配置文件

编辑配置文件 `bin/admin-app.conf` 和 `bin/game-app.conf`：

```ini
# 数据库配置
mysql_host = 127.0.0.1
mysql_port = 3306
mysql_user = minigame
mysql_password = your_password
mysql_database = minigame_server

# Redis配置
redis_host = 127.0.0.1
redis_port = 6379
redis_password = your_redis_password
```

#### 5. 部署服务

```bash
# 使用部署脚本（推荐）
sudo chmod +x scripts/deploy.sh
sudo ./scripts/deploy.sh

# 或手动启动
cd bin
./admin-service &
./game-service &
```

### 方式二：Docker部署

#### 1. 准备Docker环境

```bash
# 安装Docker和Docker Compose
curl -fsSL https://get.docker.com | sh
sudo curl -L "https://github.com/docker-compose/docker-compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 2. 配置Docker Compose

编辑 `docker-compose.yml` 中的环境变量：

```yaml
environment:
  MYSQL_ROOT_PASSWORD: your_root_password
  MYSQL_PASSWORD: your_mysql_password
```

#### 3. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

## 配置说明

### 管理后台服务配置 (admin-service)

```ini
# 基本配置
appname = admin-service
httpport = 8080
runmode = prod

# 数据库配置
mysql_host = 127.0.0.1
mysql_port = 3306
mysql_user = minigame
mysql_password = your_password
mysql_database = minigame_server

# JWT配置
jwt_secret = your_jwt_secret_key
jwt_expire = 86400

# 安全配置
password_salt = your_password_salt
api_secret = your_api_secret_key
```

### 游戏SDK服务配置 (game-service)

```ini
# 基本配置
appname = game-service
httpport = 8081
runmode = prod

# 数据库配置
mysql_host = 127.0.0.1
mysql_port = 3306
mysql_user = minigame
mysql_password = your_password
mysql_database = minigame_server

# API签名配置
api_secret = your_api_secret_key
enable_sign_check = true
sign_timeout = 300
```

## 数据迁移

如果您需要从现有的MongoDB系统迁移数据：

### 1. 配置迁移工具

编辑 `migration-tools/config/migration.yaml`：

```yaml
source:
  type: mongodb
  connection: "mongodb://username:password@localhost:27017/gamedb"
  
target:
  type: mysql
  connection: "minigame:password@tcp(localhost:3306)/minigame_server?charset=utf8mb4&parseTime=True&loc=Local"
```

### 2. 执行迁移

```bash
# 全量迁移
./migration-tools --config config/migration.yaml --mode full

# 增量迁移
./migration-tools --config config/migration.yaml --mode incremental --since "2023-10-01 00:00:00"

# 验证数据
./migration-tools --config config/migration.yaml --mode verify

# 生成报告
./migration-tools --config config/migration.yaml --mode report
```

## 服务管理

### Systemd服务管理

```bash
# 启动服务
sudo systemctl start minigame-admin
sudo systemctl start minigame-game

# 停止服务
sudo systemctl stop minigame-admin
sudo systemctl stop minigame-game

# 重启服务
sudo systemctl restart minigame-admin
sudo systemctl restart minigame-game

# 查看状态
sudo systemctl status minigame-admin
sudo systemctl status minigame-game

# 查看日志
sudo journalctl -u minigame-admin -f
sudo journalctl -u minigame-game -f

# 开机自启
sudo systemctl enable minigame-admin
sudo systemctl enable minigame-game
```

### Docker服务管理

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart admin-service
docker-compose restart game-service

# 查看日志
docker-compose logs -f admin-service
docker-compose logs -f game-service

# 更新服务
docker-compose pull
docker-compose up -d
```

## 监控和维护

### 健康检查

```bash
# 检查服务状态
curl http://localhost:8080/health
curl http://localhost:8081/health
```

### 日志管理

日志文件位置：
- 传统部署: `/opt/minigame-server/logs/`
- Docker部署: `./logs/`

日志轮转已自动配置，保留30天的日志文件。

### 数据库维护

```bash
# 数据库备份
mysqldump -u root -p minigame_server > backup_$(date +%Y%m%d).sql

# 数据库优化
mysql -u root -p minigame_server -e "OPTIMIZE TABLE admin_users, applications;"
```

### 性能优化

1. **MySQL优化**：
   ```ini
   [mysqld]
   innodb_buffer_pool_size = 1G
   max_connections = 1000
   query_cache_size = 256M
   ```

2. **Redis优化**：
   ```ini
   maxmemory 512mb
   maxmemory-policy allkeys-lru
   ```

## 故障排除

### 常见问题

1. **服务启动失败**
   ```bash
   # 检查端口占用
   netstat -tlnp | grep :8080
   netstat -tlnp | grep :8081
   
   # 检查配置文件
   ./admin-service -t
   ./game-service -t
   ```

2. **数据库连接失败**
   ```bash
   # 测试数据库连接
   mysql -h localhost -u minigame -p minigame_server
   
   # 检查防火墙
   sudo ufw status
   sudo firewall-cmd --list-all
   ```

3. **Redis连接失败**
   ```bash
   # 测试Redis连接
   redis-cli -h localhost -p 6379 ping
   
   # 检查Redis状态
   sudo systemctl status redis
   ```

### 日志分析

关键日志位置：
- 应用日志: `/opt/minigame-server/logs/`
- 系统日志: `/var/log/syslog` 或 `journalctl`
- 数据库日志: `/var/log/mysql/`
- Redis日志: `/var/log/redis/`

## 安全建议

1. **防火墙配置**
   ```bash
   # 只开放必要端口
   sudo ufw allow 8080
   sudo ufw allow 8081
   sudo ufw enable
   ```

2. **SSL证书**
   - 使用Let's Encrypt获取免费SSL证书
   - 配置Nginx反向代理处理HTTPS

3. **定期更新**
   - 定期更新系统补丁
   - 更新应用依赖包
   - 监控安全漏洞

## 扩展部署

### 负载均衡

使用Nginx配置负载均衡：

```nginx
upstream admin_backend {
    server 127.0.0.1:8080;
    server 127.0.0.1:8082;  # 第二个实例
}

upstream game_backend {
    server 127.0.0.1:8081;
    server 127.0.0.1:8083;  # 第二个实例
}

server {
    listen 80;
    
    location /admin {
        proxy_pass http://admin_backend;
    }
    
    location /api {
        proxy_pass http://game_backend;
    }
}
```

### 数据库集群

配置MySQL主从复制或使用MySQL Cluster提高可用性。

## 支持

如有问题，请参考：
- 项目文档: `README.md`
- API文档: `API.md`
- 故障排除: `TROUBLESHOOTING.md`

或联系技术支持团队。
