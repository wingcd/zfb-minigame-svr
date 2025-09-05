@echo off
REM Minigame Server Build Script for Windows
REM 用于编译管理后台服务和游戏服务

setlocal enabledelayedexpansion

echo [INFO] Starting Minigame Server build process...

REM 检查Go环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo [INFO] Go version: %GO_VERSION%

REM 创建输出目录
if not exist "bin" (
    mkdir bin
    echo [INFO] Created output directory: bin
)

REM 编译admin-service
echo [INFO] Building admin-service...
cd admin-service
echo [INFO] Downloading dependencies for admin-service...
go mod download
go mod tidy
set CGO_ENABLED=0
go build -ldflags="-s -w" -o ..\bin\admin-service.exe .
if %errorlevel% neq 0 (
    echo [ERROR] admin-service build failed
    cd ..
    exit /b 1
)
echo [INFO] admin-service build completed successfully
cd ..

REM 编译game-service
echo [INFO] Building game-service...
cd game-service
echo [INFO] Downloading dependencies for game-service...
go mod download
go mod tidy
set CGO_ENABLED=0
go build -ldflags="-s -w" -o ..\bin\game-service.exe .
if %errorlevel% neq 0 (
    echo [ERROR] game-service build failed
    cd ..
    exit /b 1
)
echo [INFO] game-service build completed successfully
cd ..

REM 复制配置文件
echo [INFO] Copying configuration files...
if exist "admin-service\conf\app.conf" (
    copy "admin-service\conf\app.conf" "bin\admin-app.conf" >nul
    echo [INFO] Copied admin-service config
)

if exist "game-service\conf\app.conf" (
    copy "game-service\conf\app.conf" "bin\game-app.conf" >nul
    echo [INFO] Copied game-service config
)

if exist "scripts\sql" (
    xcopy "scripts\sql" "bin\sql\" /E /I /Q >nul
    echo [INFO] Copied SQL files
)

REM 生成启动脚本
echo [INFO] Generating start scripts...

echo @echo off > bin\start-admin.bat
echo echo Starting Admin Service... >> bin\start-admin.bat
echo admin-service.exe >> bin\start-admin.bat
echo pause >> bin\start-admin.bat

echo @echo off > bin\start-game.bat
echo echo Starting Game Service... >> bin\start-game.bat
echo game-service.exe >> bin\start-game.bat
echo pause >> bin\start-game.bat

echo [INFO] Start scripts generated

REM 显示构建信息
echo.
echo [INFO] Build completed successfully!
echo.
echo Output files:
echo   bin\admin-service.exe  - Admin service (port 8080)
echo   bin\game-service.exe   - Game service (port 8081)
echo   bin\admin-app.conf     - Admin service config
echo   bin\game-app.conf      - Game service config
echo   bin\sql\               - Database scripts
echo.
echo Start services:
echo   Windows: run start-admin.bat or start-game.bat
echo.
echo Database setup:
echo   1. Create MySQL database
echo   2. Run bin\sql\init.sql
echo   3. Run bin\sql\seed.sql (optional)
echo.

pause 