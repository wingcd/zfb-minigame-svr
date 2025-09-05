#!/bin/bash

# Minigame Server Deployment Script
# 用于部署小游戏服务器

set -e

# 配置参数
DEPLOY_DIR="/opt/minigame-server"
SERVICE_USER="minigame"
SYSTEMD_DIR="/etc/systemd/system"
BACKUP_DIR="/opt/minigame-server/backups"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "Please run as root or with sudo"
        exit 1
    fi
}

# 检查系统依赖
check_dependencies() {
    log_info "Checking system dependencies..."
    
    # 检查systemctl
    if ! command -v systemctl &> /dev/null; then
        log_error "systemctl not found. This script requires systemd."
        exit 1
    fi
    
    # 检查MySQL客户端
    if ! command -v mysql &> /dev/null; then
        log_warn "mysql client not found. Database operations may fail."
    fi
    
    log_info "Dependencies check completed"
}

# 创建用户
create_user() {
    if ! id "$SERVICE_USER" &>/dev/null; then
        log_info "Creating user $SERVICE_USER..."
        useradd -r -s /bin/false -d "$DEPLOY_DIR" "$SERVICE_USER"
    else
        log_info "User $SERVICE_USER already exists"
    fi
}

# 创建目录结构
create_directories() {
    log_info "Creating directory structure..."
    
    mkdir -p "$DEPLOY_DIR"/{bin,conf,logs,data}
    mkdir -p "$BACKUP_DIR"
    
    chown -R "$SERVICE_USER:$SERVICE_USER" "$DEPLOY_DIR"
    chmod 755 "$DEPLOY_DIR"
    
    log_info "Directory structure created"
}

# 部署文件
deploy_files() {
    log_info "Deploying application files..."
    
    if [ ! -d "bin" ]; then
        log_error "bin directory not found. Please run build script first."
        exit 1
    fi
    
    # 备份现有文件
    if [ -f "$DEPLOY_DIR/bin/admin-service" ]; then
        log_info "Backing up existing files..."
        BACKUP_NAME="backup-$(date +%Y%m%d-%H%M%S)"
        mkdir -p "$BACKUP_DIR/$BACKUP_NAME"
        cp -r "$DEPLOY_DIR/bin"/* "$BACKUP_DIR/$BACKUP_NAME/" 2>/dev/null || true
        log_info "Backup saved to $BACKUP_DIR/$BACKUP_NAME"
    fi
    
    # 复制新文件
    cp bin/* "$DEPLOY_DIR/bin/"
    
    # 设置权限
    chown -R "$SERVICE_USER:$SERVICE_USER" "$DEPLOY_DIR"
    chmod +x "$DEPLOY_DIR/bin/admin-service"
    chmod +x "$DEPLOY_DIR/bin/game-service"
    
    # 复制配置文件
    if [ -f "bin/admin-app.conf" ]; then
        cp "bin/admin-app.conf" "$DEPLOY_DIR/conf/"
    fi
    if [ -f "bin/game-app.conf" ]; then
        cp "bin/game-app.conf" "$DEPLOY_DIR/conf/"
    fi
    
    log_info "Files deployed successfully"
}

# 创建systemd服务
create_systemd_services() {
    log_info "Creating systemd services..."
    
    # Admin Service
    cat > "$SYSTEMD_DIR/minigame-admin.service" << EOF
[Unit]
Description=Minigame Admin Service
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/bin/admin-service
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=minigame-admin

# Resource limits
LimitNOFILE=65536
LimitNPROC=32768

# Environment
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

    # Game Service
    cat > "$SYSTEMD_DIR/minigame-game.service" << EOF
[Unit]
Description=Minigame Game Service
After=network.target mysql.service
Wants=mysql.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/bin/game-service
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=minigame-game

# Resource limits
LimitNOFILE=65536
LimitNPROC=32768

# Environment
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd
    systemctl daemon-reload
    
    log_info "Systemd services created"
}

# 配置日志轮转
setup_logrotate() {
    log_info "Setting up log rotation..."
    
    cat > "/etc/logrotate.d/minigame-server" << EOF
$DEPLOY_DIR/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $SERVICE_USER $SERVICE_USER
    postrotate
        systemctl reload minigame-admin minigame-game || true
    endscript
}
EOF

    log_info "Log rotation configured"
}

# 启动服务
start_services() {
    log_info "Starting services..."
    
    systemctl enable minigame-admin
    systemctl enable minigame-game
    
    systemctl start minigame-admin
    systemctl start minigame-game
    
    # 等待服务启动
    sleep 3
    
    # 检查服务状态
    if systemctl is-active --quiet minigame-admin; then
        log_info "Admin service started successfully"
    else
        log_error "Admin service failed to start"
        systemctl status minigame-admin
    fi
    
    if systemctl is-active --quiet minigame-game; then
        log_info "Game service started successfully"
    else
        log_error "Game service failed to start"
        systemctl status minigame-game
    fi
}

# 显示状态
show_status() {
    log_info "Deployment completed!"
    echo ""
    echo "Service Status:"
    systemctl status minigame-admin --no-pager -l || true
    echo ""
    systemctl status minigame-game --no-pager -l || true
    echo ""
    
    echo "Service Commands:"
    echo "  Start:   systemctl start minigame-admin minigame-game"
    echo "  Stop:    systemctl stop minigame-admin minigame-game"
    echo "  Restart: systemctl restart minigame-admin minigame-game"
    echo "  Status:  systemctl status minigame-admin minigame-game"
    echo "  Logs:    journalctl -u minigame-admin -f"
    echo "           journalctl -u minigame-game -f"
    echo ""
    
    echo "Application URLs:"
    echo "  Admin Service: http://localhost:8080"
    echo "  Game Service:  http://localhost:8081"
    echo ""
    
    echo "Configuration:"
    echo "  Deploy Directory: $DEPLOY_DIR"
    echo "  Config Files:     $DEPLOY_DIR/conf/"
    echo "  Log Files:        $DEPLOY_DIR/logs/"
    echo "  Backup Files:     $BACKUP_DIR"
}

# 主函数
main() {
    log_info "Starting Minigame Server deployment..."
    
    check_root
    check_dependencies
    create_user
    create_directories
    deploy_files
    create_systemd_services
    setup_logrotate
    start_services
    show_status
}

# 执行主函数
main "$@"
