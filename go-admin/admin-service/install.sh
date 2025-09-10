#!/bin/bash

# =============================================================================
# Minigame Admin Service 部署脚本
# Go程序环境部署和数据库初始化
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
PROJECT_NAME="Minigame Admin Service"
REQUIRED_GO_VERSION="1.21"
SERVICE_PORT="8080"

# 数据库配置
MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3306"
MYSQL_DATABASE="minigame_game"
MYSQL_USER="minigame"
MYSQL_PASSWORD=""
MYSQL_ROOT_PASSWORD=""

# Redis配置
REDIS_HOST="127.0.0.1"
REDIS_PORT="6379"
REDIS_PASSWORD=""

# =============================================================================
# 工具函数
# =============================================================================

print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_title() {
    echo
    print_message $BLUE "=========================================="
    print_message $BLUE "$1"
    print_message $BLUE "=========================================="
    echo
}

print_step() {
    print_message $BLUE "🔄 $1"
}

print_success() {
    print_message $GREEN "✅ $1"
}

print_warning() {
    print_message $YELLOW "⚠️  $1"
}

print_error() {
    print_message $RED "❌ $1"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

get_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "macos";;
        CYGWIN*|MINGW*|MSYS*) echo "windows";;
        *)          echo "unknown";;
    esac
}

# =============================================================================
# 环境检查
# =============================================================================

check_go() {
    print_step "检查Go环境"
    
    if ! command_exists go; then
        print_error "未找到Go环境，请先安装Go $REQUIRED_GO_VERSION+"
        print_message $YELLOW "下载地址: https://golang.org/dl/"
        return 1
    fi
    
    local version=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | head -1 | sed 's/go//')
    print_success "Go版本: $version"
    
    # 简单版本检查
    if [[ $(printf '%s\n' "$REQUIRED_GO_VERSION" "$version" | sort -V | head -n1) != "$REQUIRED_GO_VERSION" ]]; then
        print_warning "Go版本可能过低，推荐使用 $REQUIRED_GO_VERSION+"
    fi
    
    return 0
}

check_mysql() {
    print_step "检查MySQL连接"
    
    if ! command_exists mysql; then
        print_warning "未找到MySQL客户端"
        return 1
    fi
    
    # 测试MySQL连接
    if mysql -h"$MYSQL_HOST" -P"$MYSQL_PORT" -u root -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1;" >/dev/null 2>&1; then
        print_success "MySQL连接正常"
        return 0
    else
        print_warning "MySQL连接失败，将在程序启动时尝试自动配置"
        return 1
    fi
}

check_redis() {
    print_step "检查Redis连接"
    
    if ! command_exists redis-cli; then
        print_warning "未找到Redis客户端"
        return 1
    fi
    
    # 测试Redis连接
    if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping >/dev/null 2>&1; then
        print_success "Redis连接正常"
        return 0
    else
        print_warning "Redis连接失败，将使用内存存储"
        return 1
    fi
}

# =============================================================================
# 安装和配置
# =============================================================================

install_dependencies() {
    print_step "安装Go依赖"
    
    if [[ -f "go.mod" ]]; then
        go mod tidy
        go mod download
        print_success "依赖安装完成"
    else
        print_error "未找到go.mod文件"
        return 1
    fi
}

generate_config() {
    print_step "生成配置文件"
    
    # 生成随机密码和密钥
    if [[ -z "$MYSQL_PASSWORD" ]]; then
        MYSQL_PASSWORD=$(openssl rand -base64 12 2>/dev/null || date +%s | sha256sum | base64 | head -c 12)
    fi
    
    if [[ -z "$REDIS_PASSWORD" ]]; then
        REDIS_PASSWORD=$(openssl rand -base64 12 2>/dev/null || date +%s | sha256sum | base64 | head -c 12)
    fi
    
    local jwt_secret=$(openssl rand -base64 32 2>/dev/null || date +%s | sha256sum | base64 | head -c 32)
    local password_salt=$(openssl rand -base64 32 2>/dev/null || date +%s | sha256sum | base64 | head -c 32)
    local api_secret=$(openssl rand -base64 32 2>/dev/null || date +%s | sha256sum | base64 | head -c 32)
    
    # 备份原配置
    if [[ -f "conf/app.conf" ]]; then
        cp "conf/app.conf" "conf/app.conf.backup.$(date +%Y%m%d_%H%M%S)"
    fi
    
    # 创建配置目录
    mkdir -p conf
    
    # 生成配置文件
    cat > conf/app.conf << EOF
# 应用配置
appname = admin-service
httpport = $SERVICE_PORT
runmode = prod
autorender = false
copyrequestbody = true
EnableDocs = false

# 数据库配置
mysql_host = $MYSQL_HOST
mysql_port = $MYSQL_PORT
mysql_user = $MYSQL_USER
mysql_password = $MYSQL_PASSWORD
mysql_database = $MYSQL_DATABASE
mysql_charset = utf8mb4
mysql_max_idle = 10
mysql_max_open = 100

# Redis配置
redis_host = $REDIS_HOST
redis_port = $REDIS_PORT
redis_password = $REDIS_PASSWORD
redis_database = 0
redis_pool_size = 10

# JWT配置
jwt_secret = $jwt_secret
jwt_expire = 86400

# 加密配置
password_salt = $password_salt
api_secret = $api_secret

# 日志配置
[logs]
level = 4
separate = ["emergency", "alert", "critical", "error", "warning", "notice", "info"]

# 会话配置
sessionon = true
sessionprovider = memory
sessiongcmaxlifetime = 3600
sessioncookielifetime = 3600

# CORS配置
cors_allow_origins = *
cors_allow_methods = GET,POST,PUT,DELETE,OPTIONS
cors_allow_headers = Origin,Content-Type,Accept,Authorization,X-Requested-With

# 限流配置
rate_limit_requests = 1000
rate_limit_duration = 60

# 文件上传配置
max_memory = 67108864
max_file_size = 10485760

# 自动部署配置
auto_install = true
auto_create_database = true
auto_create_admin = true
default_admin_username = admin
default_admin_password = admin123
EOF
    
    print_success "配置文件生成完成"
    
    # 保存安装信息
    cat > .install_info << EOF
# Minigame Admin Service 安装信息
# 安装时间: $(date)

MYSQL_HOST=$MYSQL_HOST
MYSQL_PORT=$MYSQL_PORT
MYSQL_USER=$MYSQL_USER
MYSQL_PASSWORD=$MYSQL_PASSWORD
MYSQL_DATABASE=$MYSQL_DATABASE

REDIS_HOST=$REDIS_HOST
REDIS_PORT=$REDIS_PORT
REDIS_PASSWORD=$REDIS_PASSWORD

SERVICE_PORT=$SERVICE_PORT

# 默认管理员账号
DEFAULT_ADMIN_USERNAME=admin
DEFAULT_ADMIN_PASSWORD=admin123

# 注意：请妥善保管此文件，首次运行后建议修改默认密码
EOF
    
    chmod 600 .install_info
    print_success "安装信息保存完成"
}

build_project() {
    print_step "构建项目"
    
    # 创建输出目录
    mkdir -p bin
    
    # 构建项目
    if go build -o bin/admin-service .; then
        chmod +x bin/admin-service
        print_success "项目构建成功"
    else
        print_error "项目构建失败"
        return 1
    fi
}

create_service_script() {
    print_step "创建服务脚本"
    
    # 创建启动脚本
    cat > start.sh << 'EOF'
#!/bin/bash

# Minigame Admin Service 启动脚本

cd "$(dirname "$0")"

echo "🚀 启动 Minigame Admin Service..."

# 检查配置文件
if [[ ! -f "conf/app.conf" ]]; then
    echo "❌ 配置文件不存在，请先运行安装脚本"
    exit 1
fi

# 检查二进制文件
if [[ ! -f "bin/admin-service" ]]; then
    echo "❌ 程序文件不存在，请先构建项目"
    exit 1
fi

# 启动服务
echo "📝 配置文件: conf/app.conf"
echo "🌐 服务端口: $(grep httpport conf/app.conf | cut -d'=' -f2 | tr -d ' ')"
echo "📊 管理界面: http://localhost:$(grep httpport conf/app.conf | cut -d'=' -f2 | tr -d ' ')"
echo

./bin/admin-service
EOF
    
    chmod +x start.sh
    
    # 创建停止脚本
    cat > stop.sh << 'EOF'
#!/bin/bash

echo "🛑 停止 Minigame Admin Service..."

# 查找进程
PID=$(pgrep -f "admin-service" || true)

if [[ -n "$PID" ]]; then
    kill $PID
    echo "✅ 服务已停止 (PID: $PID)"
else
    echo "⚠️  服务未运行"
fi
EOF
    
    chmod +x stop.sh
    
    print_success "服务脚本创建完成"
}

# =============================================================================
# 主安装流程
# =============================================================================

show_welcome() {
    clear
    print_title "$PROJECT_NAME 部署脚本"
    
    cat << 'EOF'
    ╔══════════════════════════════════════════════════════════════════╗
    ║                🎮 Minigame Admin Service Installer               ║
    ║                                                                  ║
    ║  本脚本将自动完成以下操作：                                        ║
    ║  • 检查Go环境                                                    ║
    ║  • 安装项目依赖                                                  ║
    ║  • 生成配置文件                                                  ║
    ║  • 构建项目                                                      ║
    ║  • 创建启动脚本                                                  ║
    ║                                                                  ║
    ║  数据库初始化将在首次启动时自动完成                               ║
    ╚══════════════════════════════════════════════════════════════════╝
EOF
    echo
}

show_results() {
    print_title "部署完成"
    
    cat << EOF
🎉 Minigame Admin Service 部署成功！

📋 部署摘要：
├── Go版本: $(go version | cut -d' ' -f3)
├── 服务端口: $SERVICE_PORT
├── 配置文件: conf/app.conf
├── 程序文件: bin/admin-service
├── 安装信息: .install_info

🚀 启动服务：
./start.sh

🛑 停止服务：
./stop.sh

🌐 访问地址：
http://localhost:$SERVICE_PORT

👤 默认管理员账号：
用户名: admin
密码: admin123

⚠️  重要说明：
1. 首次启动时会自动创建数据库和表结构
2. 如果MySQL未安装，程序会使用SQLite作为后备数据库
3. 请在首次登录后修改默认管理员密码
4. 生产环境请配置真实的数据库连接

📚 更多操作：
├── 查看日志: tail -f logs/admin-service.log
├── 重新构建: go build -o bin/admin-service .
├── 运行测试: go test ./...

EOF
    
    print_success "部署完成！现在可以启动服务了！"
}

main() {
    show_welcome
    
    # 确认部署
    echo -n "是否开始部署？[y/N] "
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_message $YELLOW "部署已取消"
        exit 0
    fi
    
    # 记录开始时间
    START_TIME=$(date +%s)
    
    # 执行部署步骤
    check_go || exit 1
    check_mysql
    check_redis
    install_dependencies
    generate_config
    build_project
    create_service_script
    
    # 计算部署时间
    END_TIME=$(date +%s)
    DEPLOY_TIME=$((END_TIME - START_TIME))
    
    # 显示结果
    show_results
    
    print_message $BLUE "部署耗时: ${DEPLOY_TIME}秒"
    
    return 0
}

# 错误处理
trap 'print_error "部署过程中发生错误"; exit 1' ERR

# 运行主程序
main "$@"
