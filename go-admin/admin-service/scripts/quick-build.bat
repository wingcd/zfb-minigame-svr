@echo off
REM =============================================================================
REM 快速构建脚本 - 只构建当前平台用于开发测试 (Windows)
REM =============================================================================
chcp 65001

setlocal EnableDelayedExpansion

REM 配置
set "VERSION=dev"
if defined VERSION_ENV set "VERSION=%VERSION_ENV%"

REM 获取构建时间和Git信息
for /f "tokens=1-3 delims=/ " %%a in ('date /t') do set BUILD_DATE=%%c-%%a-%%b
for /f "tokens=1-2 delims=: " %%a in ('time /t') do set BUILD_TIME_ONLY=%%a:%%b
set "BUILD_TIME=%BUILD_DATE% %BUILD_TIME_ONLY%"

git rev-parse --short HEAD >nul 2>&1
if errorlevel 1 (
    set "GIT_COMMIT=unknown"
) else (
    for /f %%i in ('git rev-parse --short HEAD') do set "GIT_COMMIT=%%i"
)

REM 颜色输出函数
:print_step
echo 🔄 %~1
goto :eof

:print_success
echo ✅ %~1
goto :eof

:print_error
echo ❌ %~1
goto :eof

REM 检查Go环境
where go >nul 2>&1
if errorlevel 1 (
    call :print_error "未找到Go环境"
    pause
    exit /b 1
)

call :print_step "快速构建 Minigame Admin Service"
echo 版本: %VERSION%
echo 时间: %BUILD_TIME%
echo 提交: %GIT_COMMIT%
echo.

REM 安装依赖
call :print_step "安装依赖"
go mod tidy

REM 构建
call :print_step "构建程序"
if not exist "bin" mkdir bin

set "ldflags=-X main.version=%VERSION% -X 'main.buildTime=%BUILD_TIME%' -X main.gitCommit=%GIT_COMMIT%"
go build -ldflags "%ldflags%" -o bin\admin-service.exe .

if errorlevel 1 (
    call :print_error "构建失败"
    pause
    exit /b 1
)

call :print_success "构建完成: bin\admin-service.exe"
echo.
echo 🚀 运行服务:
echo   bin\admin-service.exe
echo.
echo 🌐 访问地址:
echo   http://localhost:8080
echo.
pause
