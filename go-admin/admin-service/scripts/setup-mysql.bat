@echo off
setlocal enabledelayedexpansion

chcp 65001

:: MySQL配置和测试脚本
title MySQL 配置助手

:: 颜色定义
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

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

:print_warning
echo %YELLOW%⚠️  %~1%NC%
goto :eof

:print_info
echo %BLUE%ℹ️  %~1%NC%
goto :eof

:main
call :print_header "MySQL 配置助手"

:: 检查MySQL客户端
mysql --version >nul 2>&1
if errorlevel 1 (
    call :print_error "MySQL客户端未安装"
    call :print_info "请下载并安装MySQL: https://dev.mysql.com/downloads/mysql/"
    pause
    exit /b 1
)
call :print_success "MySQL客户端已安装"

:: 测试不同的连接配置
call :print_info "测试MySQL连接配置..."

:: 测试配置1: localhost无密码
echo.
call :print_info "测试配置1: mysql -hlocalhost -uroot"
mysql -hlocalhost -uroot -e "SELECT 'Connection successful' as status;" 2>nul
if not errorlevel 1 (
    call :print_success "配置1成功: localhost无密码连接"
    set WORKING_CONFIG=mysql -hlocalhost -uroot
    goto :create_config
)

:: 测试配置2: localhost密码123456
echo.
call :print_info "测试配置2: mysql -hlocalhost -uroot -p123456"
mysql -hlocalhost -uroot -p123456 -e "SELECT 'Connection successful' as status;" 2>nul
if not errorlevel 1 (
    call :print_success "配置2成功: localhost密码为123456"
    set WORKING_CONFIG=mysql -hlocalhost -uroot -p123456
    goto :create_config
)

:: 测试配置3: localhost密码root
echo.
call :print_info "测试配置3: mysql -hlocalhost -uroot -proot"
mysql -hlocalhost -uroot -proot -e "SELECT 'Connection successful' as status;" 2>nul
if not errorlevel 1 (
    call :print_success "配置3成功: localhost密码为root"
    set WORKING_CONFIG=mysql -hlocalhost -uroot -proot
    goto :create_config
)

:: 测试配置4: 127.0.0.1无密码
echo.
call :print_info "测试配置4: mysql -h127.0.0.1 -uroot"
mysql -h127.0.0.1 -uroot -e "SELECT 'Connection successful' as status;" 2>nul
if not errorlevel 1 (
    call :print_success "配置4成功: 127.0.0.1无密码连接"
    set WORKING_CONFIG=mysql -h127.0.0.1 -uroot
    goto :create_config
)

:: 测试配置5: 127.0.0.1密码123456
echo.
call :print_info "测试配置5: mysql -h127.0.0.1 -uroot -p123456"
mysql -h127.0.0.1 -uroot -p123456 -e "SELECT 'Connection successful' as status;" 2>nul
if not errorlevel 1 (
    call :print_success "配置5成功: 127.0.0.1密码为123456"
    set WORKING_CONFIG=mysql -h127.0.0.1 -uroot -p123456
    goto :create_config
)

:: 没有找到工作的配置
call :print_error "未找到可用的MySQL配置"
echo.
call :print_info "请检查以下事项:"
echo   1. MySQL服务是否正在运行
echo   2. root用户是否存在
echo   3. 防火墙是否阻止连接
echo   4. MySQL配置是否正确
echo.
call :print_info "手动测试连接:"
echo   mysql -h127.0.0.1 -uroot -p
echo.
pause
exit /b 1

:create_config
call :print_header "创建测试数据库"

call :print_info "创建admin_service_test数据库..."
%WORKING_CONFIG% -e "CREATE DATABASE IF NOT EXISTS admin_service_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
if errorlevel 1 (
    call :print_error "创建数据库失败"
    pause
    exit /b 1
)

call :print_success "数据库创建成功"

:: 验证数据库
call :print_info "验证数据库..."
%WORKING_CONFIG% -e "SHOW DATABASES LIKE 'admin_service_test';"
call :print_success "数据库验证完成"

call :print_header "配置完成"
call :print_success "MySQL配置成功，可以运行完整测试了"
call :print_info "运行完整测试: scripts\run_tests.bat"

pause
endlocal
exit /b 0
