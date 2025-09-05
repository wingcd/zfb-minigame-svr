#!/bin/bash
# =============================================================================
# 快速构建脚本 - 只构建当前平台用于开发测试
# =============================================================================
cd ..

set -e

# 配置
VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 颜色输出
print_step() {
    echo -e "\033[94m🔄 $1\033[0m"
}

print_success() {
    echo -e "\033[92m✅ $1\033[0m"
}

print_error() {
    echo -e "\033[91m❌ $1\033[0m"
}

# 检查Go环境
if ! command -v go >/dev/null 2>&1; then
    print_error "未找到Go环境"
    exit 1
fi

print_step "快速构建 Minigame Admin Service"
echo "版本: $VERSION"
echo "时间: $BUILD_TIME"
echo "提交: $GIT_COMMIT"
echo ""

# 安装依赖
print_step "安装依赖"
go mod tidy

# 构建
print_step "构建程序"
mkdir -p bin
ldflags="-X main.version=$VERSION -X 'main.buildTime=$BUILD_TIME' -X main.gitCommit=$GIT_COMMIT"
go build -ldflags "$ldflags" -o bin/admin-service .

if [ $? -eq 0 ]; then
    print_success "构建完成: bin/admin-service"
    echo ""
    echo "🚀 运行服务:"
    echo "  ./bin/admin-service"
    echo ""
    echo "🌐 访问地址:"
    echo "  http://localhost:8080"
else
    print_error "构建失败"
    exit 1
fi
