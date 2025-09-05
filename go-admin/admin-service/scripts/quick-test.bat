@echo off
setlocal enabledelayedexpansion

chcp 65001

:: 确保在正确的目录中运行
set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
cd /d "%PROJECT_DIR%"

:: 快速测试脚本 - 不依赖外部数据库
title Admin Service 快速测试

:: 颜色定义
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

:: 打印函数
:print_header
echo.
echo %BLUE%===================================================%NC%
echo %BLUE%%~1%NC%
echo %BLUE%===================================================%NC%
goto :eof

:print_success
echo %GREEN%✅ %~1%NC%
goto :eof

:print_error
echo %RED%❌ %~1%NC%
goto :eof

:print_info
echo %BLUE%ℹ️  %~1%NC%
goto :eof

:: 主函数
:main
call :print_header "Admin Service 快速测试"

:: 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    call :print_error "Go 未安装或未在PATH中"
    pause
    exit /b 1
)
for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
call :print_success "Go 环境: !GO_VERSION!"

:: 下载依赖
call :print_info "下载依赖..."
go mod tidy
if errorlevel 1 (
    call :print_error "依赖下载失败"
    pause
    exit /b 1
)
call :print_success "依赖下载完成"

:: 构建项目
call :print_info "构建项目..."
if not exist "bin" mkdir bin
go build -o bin\admin-service.exe main.go
if errorlevel 1 (
    call :print_error "项目构建失败"
    pause
    exit /b 1
)
call :print_success "项目构建完成"

:: 运行基础测试（跳过需要数据库的测试）
call :print_info "运行基础测试（跳过数据库测试）..."
go test -v ./... -short -timeout=5m
if errorlevel 1 (
    call :print_error "基础测试失败"
) else (
    call :print_success "基础测试通过"
)

:: 检查代码格式
call :print_info "检查代码格式..."
go fmt ./...
call :print_success "代码格式检查完成"

:: 运行代码检查
call :print_info "运行代码检查..."
go vet ./...
if errorlevel 1 (
    call :print_error "代码检查发现问题"
) else (
    call :print_success "代码检查通过"
)

call :print_header "快速测试完成"
call :print_info "如需完整测试，请运行: scripts\run_tests.bat"

pause
endlocal
exit /b 0
