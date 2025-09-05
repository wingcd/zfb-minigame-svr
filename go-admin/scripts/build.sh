#!/bin/bash

# Minigame Server Build Script for Linux/macOS
# 用于编译管理后台服务和游戏服务

set -e

echo "[INFO] Starting Minigame Server build process..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "[ERROR] Go is not installed or not in PATH"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "[INFO] Go version: $GO_VERSION"

# 创建输出目录
if [ ! -d "bin" ]; then
    mkdir bin
    echo "[INFO] Created output directory: bin"
fi

# 编译admin-service
echo "[INFO] Building admin-service..."
cd admin-service
echo "[INFO] Downloading dependencies for admin-service..."
go mod download
go mod tidy
CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/admin-service .
if [ $? -ne 0 ]; then
    echo "[ERROR] admin-service build failed"
    exit 1
fi
echo "[INFO] admin-service build completed successfully"
cd ..

# 编译game-service
echo "[INFO] Building game-service..."
cd game-service
echo "[INFO] Downloading dependencies for game-service..."
go mod download
go mod tidy
CGO_ENABLED=0 go build -ldflags="-s -w" -o ../bin/game-service .
if [ $? -ne 0 ]; then
    echo "[ERROR] game-service build failed"
    exit 1
fi
echo "[INFO] game-service build completed successfully"
cd ..

# 编译migration-tools（如果存在）
if [ -d "../migration-tools" ]; then
    echo "[INFO] Building migration-tools..."
    cd ../migration-tools
    echo "[INFO] Downloading dependencies for migration-tools..."
    go mod download
    go mod tidy
    CGO_ENABLED=0 go build -ldflags="-s -w" -o ../go-admin/bin/migration-tools .
    if [ $? -ne 0 ]; then
        echo "[ERROR] migration-tools build failed"
        exit 1
    fi
    echo "[INFO] migration-tools build completed successfully"
    cd ../go-admin
else
    echo "[WARN] migration-tools directory not found, skipping..."
fi

# 复制配置文件
echo "[INFO] Copying configuration files..."
if [ -f "admin-service/conf/app.conf" ]; then
    cp "admin-service/conf/app.conf" "bin/admin-app.conf"
    echo "[INFO] Copied admin-service config"
fi

if [ -f "game-service/conf/app.conf" ]; then
    cp "game-service/conf/app.conf" "bin/game-app.conf"
    echo "[INFO] Copied game-service config"
fi

if [ -d "scripts/sql" ]; then
    cp -r "scripts/sql" "bin/"
    echo "[INFO] Copied SQL files"
fi

# 生成启动脚本
echo "[INFO] Generating start scripts..."

cat > bin/start-admin.sh << 'EOF'
#!/bin/bash
echo "Starting Admin Service..."
./admin-service
EOF

cat > bin/start-game.sh << 'EOF'
#!/bin/bash
echo "Starting Game Service..."
./game-service
EOF

chmod +x bin/start-admin.sh
chmod +x bin/start-game.sh
chmod +x bin/admin-service
chmod +x bin/game-service

if [ -f "bin/migration-tools" ]; then
    chmod +x bin/migration-tools
fi

echo "[INFO] Start scripts generated"

# 显示构建信息
echo ""
echo "[INFO] Build completed successfully!"
echo ""
echo "Output files:"
echo "  bin/admin-service      - Admin service (port 8080)"
echo "  bin/game-service       - Game service (port 8081)"
echo "  bin/admin-app.conf     - Admin service config"
echo "  bin/game-app.conf      - Game service config"
echo "  bin/sql/               - Database scripts"
if [ -f "bin/migration-tools" ]; then
    echo "  bin/migration-tools    - Data migration tool"
fi
echo ""
echo "Start services:"
echo "  Linux/macOS: ./start-admin.sh or ./start-game.sh"
echo ""
echo "Database setup:"
echo "  1. Create MySQL database"
echo "  2. Run bin/sql/init.sql"
echo "  3. Run bin/sql/seed.sql (optional)"
echo ""
echo "Migration (if needed):"
echo "  ./migration-tools --config config/migration.yaml --mode full"
echo ""