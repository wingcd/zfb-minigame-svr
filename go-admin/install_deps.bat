@echo off
echo ========================================
echo 正在安装Go项目依赖...
echo ========================================

echo.
echo [1/2] 安装admin-service依赖...
cd admin-service
go mod tidy
go mod download
if %ERRORLEVEL% neq 0 (
    echo 错误: admin-service依赖安装失败
    pause
    exit /b 1
)
echo admin-service依赖安装完成!

echo.
echo [2/2] 安装game-service依赖...
cd ..\game-service
go mod tidy
go mod download
if %ERRORLEVEL% neq 0 (
    echo 错误: game-service依赖安装失败
    pause
    exit /b 1
)
echo game-service依赖安装完成!

echo.
echo ========================================
echo 所有依赖安装完成!
echo ========================================

echo.
echo 依赖列表:
echo - Beego v2.0.7 (Web框架)
echo - JWT Go v3.2.0 (JWT认证)
echo - Redis v8.11.5 (缓存数据库)
echo - MySQL Driver v1.7.0 (数据库驱动)
echo - Crypto (加密库)

echo.
echo 下一步: 运行 build.bat 编译项目
pause
