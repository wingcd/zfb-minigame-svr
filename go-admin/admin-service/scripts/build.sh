#!/bin/bash
# =============================================================================
# Minigame Admin Service 编译脚本 (Linux/macOS)
# 支持Go后端、云函数和前端项目的编译打包
# =============================================================================

cd ..

set -e

# 配置变量
PROJECT_NAME="Minigame Admin Service"
VERSION=${VERSION:-"1.0.0"}
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建目标
BUILD_BACKEND=${BUILD_BACKEND:-true}
BUILD_FRONTEND=${BUILD_FRONTEND:-true}
CREATE_RELEASE=${CREATE_RELEASE:-false}

# 路径配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
ADMIN_SERVICE_DIR="$SCRIPT_DIR"
FRONTEND_DIR="$ROOT_DIR/game-admin"
BUILD_OUTPUT_DIR="$SCRIPT_DIR/dist"
RELEASE_DIR="$SCRIPT_DIR/release"

# 支持的平台
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# =============================================================================
# 工具函数
# =============================================================================

print_title() {
    echo ""
    echo "=========================================="
    echo "$1"
    echo "=========================================="
    echo ""
}

print_step() {
    echo -e "\033[94m🔄 $1\033[0m"
}

print_success() {
    echo -e "\033[92m✅ $1\033[0m"
}

print_warning() {
    echo -e "\033[93m⚠️  $1\033[0m"
}

print_error() {
    echo -e "\033[91m❌ $1\033[0m"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# =============================================================================
# 环境检查
# =============================================================================

check_environment() {
    print_step "检查构建环境"
    
    # 检查Go环境
    if ! command_exists go; then
        print_error "未找到Go环境，请先安装Go 1.21+"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go版本: $GO_VERSION"
    
    # 检查Node.js环境（如果需要构建前端）
    if [ "$BUILD_FRONTEND" = "true" ]; then
        if ! command_exists node; then
            print_warning "未找到Node.js环境，跳过前端构建"
            BUILD_FRONTEND=false
        else
            NODE_VERSION=$(node --version)
            print_success "Node.js版本: $NODE_VERSION"
        fi
    fi
    
    # 检查Git
    if command_exists git; then
        print_success "Git版本: $(git --version | awk '{print $3}')"
    fi
}

# =============================================================================
# Go后端编译
# =============================================================================

build_backend() {
    print_title "编译Go后端服务"
    
    cd "$ADMIN_SERVICE_DIR"
    
    # 清理旧的构建文件
    rm -rf "$BUILD_OUTPUT_DIR/backend"
    mkdir -p "$BUILD_OUTPUT_DIR/backend"
    
    # 安装依赖
    print_step "安装Go依赖"
    go mod tidy
    go mod download
    
    # 编译不同平台的二进制文件
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "$platform"
        
        print_step "编译 $GOOS/$GOARCH"
        
        output_name="admin-service"
        if [ "$GOOS" = "windows" ]; then
            output_name="admin-service.exe"
        fi
        
        output_dir="$BUILD_OUTPUT_DIR/backend/$GOOS-$GOARCH"
        mkdir -p "$output_dir"
        
        # 设置构建标签
        ldflags="-X main.version=$VERSION -X 'main.buildTime=$BUILD_TIME' -X main.gitCommit=$GIT_COMMIT"
        
        # 编译
        env GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build \
            -ldflags "$ldflags" \
            -o "$output_dir/$output_name" \
            .
        
        # 复制配置文件和资源
        if [ -d "conf" ]; then
            cp -r conf "$output_dir/"
        fi
        
        if [ -d "views" ]; then
            cp -r views "$output_dir/"
        fi
        
        if [ -d "static" ]; then
            cp -r static "$output_dir/"
        fi
        
        # 创建启动脚本
        if [ "$GOOS" = "windows" ]; then
            cat > "$output_dir/start.bat" << 'EOF'
@echo off
echo 🚀 启动 Minigame Admin Service...
admin-service.exe
pause
EOF
        else
            cat > "$output_dir/start.sh" << 'EOF'
#!/bin/bash
echo "🚀 启动 Minigame Admin Service..."
./admin-service
EOF
            chmod +x "$output_dir/start.sh"
        fi
        
        print_success "$GOOS/$GOARCH 编译完成"
    done
    
    print_success "Go后端编译完成"
}

# =============================================================================
# 前端编译
# =============================================================================

build_frontend() {
    print_title "编译Vue前端项目"
    
    if [ ! -d "$FRONTEND_DIR" ]; then
        print_warning "前端目录不存在，跳过前端构建"
        return
    fi
    
    cd "$FRONTEND_DIR"
    
    # 检查package.json
    if [ ! -f "package.json" ]; then
        print_warning "未找到package.json，跳过前端构建"
        return
    fi
    
    # 安装依赖
    print_step "安装前端依赖"
    if command_exists yarn; then
        yarn install
    else
        npm install
    fi
    
    # 构建生产版本
    print_step "构建生产版本"
    if command_exists yarn; then
        yarn build
    else
        npm run build
    fi
    
    # 复制构建结果
    if [ -d "dist" ]; then
        mkdir -p "$BUILD_OUTPUT_DIR/frontend"
        cp -r dist/* "$BUILD_OUTPUT_DIR/frontend/"
        print_success "前端构建完成"
    else
        print_error "前端构建失败，未找到dist目录"
    fi
}


# =============================================================================
# 创建发布包
# =============================================================================

create_release() {
    print_title "创建发布包"
    
    rm -rf "$RELEASE_DIR"
    mkdir -p "$RELEASE_DIR"
    
    # 为每个平台创建发布包
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r GOOS GOARCH <<< "$platform"
        
        print_step "创建 $GOOS-$GOARCH 发布包"
        
        platform_dir="$GOOS-$GOARCH"
        release_name="minigame-admin-service-$VERSION-$platform_dir"
        release_path="$RELEASE_DIR/$release_name"
        
        mkdir -p "$release_path"
        
        # 复制后端文件
        if [ -d "$BUILD_OUTPUT_DIR/backend/$platform_dir" ]; then
            cp -r "$BUILD_OUTPUT_DIR/backend/$platform_dir"/* "$release_path/"
        fi
        
        # 复制前端文件
        if [ -d "$BUILD_OUTPUT_DIR/frontend" ]; then
            mkdir -p "$release_path/static/admin"
            cp -r "$BUILD_OUTPUT_DIR/frontend"/* "$release_path/static/admin/"
        fi
        
        
        # 复制文档和脚本
        cp "$SCRIPT_DIR/install.sh" "$release_path/" 2>/dev/null || true
        cp "$SCRIPT_DIR/install.bat" "$release_path/" 2>/dev/null || true
        
        # 创建README
        cat > "$release_path/README.md" << EOF
# Minigame Admin Service $VERSION

## 快速开始

### 自动安装
\`\`\`bash
# Linux/macOS
chmod +x install.sh
./install.sh

# Windows
install.bat
\`\`\`

### 手动启动
\`\`\`bash
# Linux/macOS
./start.sh

# Windows
start.bat
\`\`\`

## 访问地址
- 管理界面: http://localhost:8080
- 默认账号: admin/admin123

## 版本信息
- 版本: $VERSION
- 构建时间: $BUILD_TIME
- Git提交: $GIT_COMMIT
- 平台: $GOOS/$GOARCH

## 更多信息
请访问项目文档了解更多配置和使用方法。
EOF
        
        # 创建压缩包
        cd "$RELEASE_DIR"
        if command_exists zip; then
            zip -r "$release_name.zip" "$release_name/"
            print_success "$release_name.zip 创建完成"
        fi
        
        if command_exists tar; then
            tar -czf "$release_name.tar.gz" "$release_name/"
            print_success "$release_name.tar.gz 创建完成"
        fi
    done
    
    print_success "发布包创建完成"
}

# =============================================================================
# 清理函数
# =============================================================================

clean_build() {
    print_step "清理构建文件"
    rm -rf "$BUILD_OUTPUT_DIR"
    rm -rf "$RELEASE_DIR"
    print_success "清理完成"
}

# =============================================================================
# 显示构建信息
# =============================================================================

show_build_info() {
    print_title "构建信息"
    
    echo "📋 构建配置："
    echo "├── 项目名称: $PROJECT_NAME"
    echo "├── 版本: $VERSION"
    echo "├── 构建时间: $BUILD_TIME"
    echo "├── Git提交: $GIT_COMMIT"
    echo "├── 后端构建: $BUILD_BACKEND"
    echo "├── 前端构建: $BUILD_FRONTEND"
    echo "└── 创建发布包: $CREATE_RELEASE"
    echo ""
    
    echo "🎯 支持平台："
    for platform in "${PLATFORMS[@]}"; do
        echo "├── $platform"
    done
    echo ""
}

show_build_results() {
    print_title "构建完成"
    
    echo "🎉 Minigame Admin Service 构建成功！"
    echo ""
    
    if [ -d "$BUILD_OUTPUT_DIR" ]; then
        echo "📁 构建输出目录: $BUILD_OUTPUT_DIR"
        echo "├── 后端文件: $BUILD_OUTPUT_DIR/backend/"
        echo "└── 前端文件: $BUILD_OUTPUT_DIR/frontend/"
        echo ""
    fi
    
    if [ -d "$RELEASE_DIR" ]; then
        echo "📦 发布包目录: $RELEASE_DIR"
        ls -la "$RELEASE_DIR" | grep -E '\.(zip|tar\.gz)$' | while read -r line; do
            echo "├── $(echo "$line" | awk '{print $9}')"
        done
        echo ""
    fi
    
    echo "🚀 快速启动："
    echo "cd $BUILD_OUTPUT_DIR/backend/linux-amd64"
    echo "./start.sh"
    echo ""
}

# =============================================================================
# 主函数
# =============================================================================

show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  -v, --version VERSION   设置版本号 (默认: 1.0.0)"
    echo "  --backend-only          只构建后端"
    echo "  --frontend-only         只构建前端"
    echo ""
    echo "  --release               创建发布包"
    echo "  --clean                 清理构建文件"
    echo "  --platforms PLATFORMS   指定构建平台 (用逗号分隔)"
    echo ""
    echo "环境变量:"
    echo "  VERSION                 版本号"
    echo "  BUILD_BACKEND           是否构建后端 (true/false)"
    echo "  BUILD_FRONTEND          是否构建前端 (true/false)"
    echo ""
    echo "  CREATE_RELEASE          是否创建发布包 (true/false)"
    echo ""
    echo "示例:"
    echo "  $0                      # 完整构建"
    echo "  $0 --backend-only       # 只构建后端"
    echo "  $0 --release            # 构建并创建发布包"
    echo "  $0 --clean              # 清理构建文件"
}

main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            --backend-only)
                BUILD_BACKEND=true
                BUILD_FRONTEND=false
                BUILD_CLOUD_FUNCTIONS=false
                shift
                ;;
            --frontend-only)
                BUILD_BACKEND=false
                BUILD_FRONTEND=true
                BUILD_CLOUD_FUNCTIONS=false
                shift
                ;;
            --release)
                CREATE_RELEASE=true
                shift
                ;;
            --clean)
                clean_build
                exit 0
                ;;
            --platforms)
                IFS=',' read -ra PLATFORMS <<< "$2"
                shift 2
                ;;
            *)
                echo "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 显示构建信息
    show_build_info
    
    # 检查环境
    check_environment
    
    # 记录开始时间
    START_TIME=$(date +%s)
    
    # 执行构建
    if [ "$BUILD_BACKEND" = "true" ]; then
        build_backend
    fi
    
    if [ "$BUILD_FRONTEND" = "true" ]; then
        build_frontend
    fi
    
    
    if [ "$CREATE_RELEASE" = "true" ]; then
        create_release
    fi
    
    # 计算构建时间
    END_TIME=$(date +%s)
    BUILD_DURATION=$((END_TIME - START_TIME))
    
    # 显示结果
    show_build_results
    
    print_success "构建完成！耗时: ${BUILD_DURATION}秒"
}

# 运行主函数
main "$@"
