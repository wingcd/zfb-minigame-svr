@REM @echo off
REM =============================================================================
REM Minigame Admin Service 部署脚本 (Windows)
REM Go程序环境部署和数据库初始化
REM =============================================================================
chcp 65001

setlocal EnableDelayedExpansion

REM 配置变量
set "PROJECT_NAME=小游戏管理后台服务"
set "REQUIRED_GO_VERSION=1.21"
set "SERVICE_PORT=8080"

REM 数据库配置
set "MYSQL_HOST=127.0.0.1"
set "MYSQL_PORT=3306"
set "MYSQL_DATABASE=minigame_game"
set "MYSQL_USER=minigame"
set "MYSQL_PASSWORD="
set "MYSQL_ROOT_PASSWORD="

REM Redis配置
set "REDIS_HOST=127.0.0.1"
set "REDIS_PORT=6379"
set "REDIS_PASSWORD="

REM =============================================================================
REM 工具函数
REM =============================================================================

:print_title
echo.
echo ==========================================
echo %~1
echo ==========================================
echo.
goto :eof

:print_step
echo [94m🔄 %~1[0m
goto :eof

:print_success
echo [92m✅ %~1[0m
goto :eof

:print_warning
echo [93m⚠️  %~1[0m
goto :eof

:print_error
echo [91m❌ %~1[0m
goto :eof

:command_exists
where %1 >nul 2>&1
goto :eof

REM =============================================================================
REM 环境检查
REM =============================================================================

:check_go
call :print_step "检查Go环境"

call :command_exists go
if errorlevel 1 (
    call :print_error "未找到Go环境，请先安装Go %REQUIRED_GO_VERSION%+"
    echo 下载地址: https://golang.org/dl/
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
call :print_success "Go版本: %GO_VERSION%"
goto :eof

:check_mysql
call :print_step "检查MySQL连接"

call :command_exists mysql
if errorlevel 1 (
    call :print_warning "未找到MySQL客户端"
    goto :eof
)

REM 测试MySQL连接
mysql -h%MYSQL_HOST% -P%MYSQL_PORT% -uroot -p%MYSQL_ROOT_PASSWORD% -e "SELECT 1;" >nul 2>&1
if errorlevel 1 (
    call :print_warning "MySQL连接失败，将在程序启动时尝试自动配置"
) else (
    call :print_success "MySQL连接正常"
)
goto :eof

:check_redis
call :print_step "检查Redis连接"

call :command_exists redis-cli
if errorlevel 1 (
    call :print_warning "未找到Redis客户端"
    goto :eof
)

REM 测试Redis连接
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% ping >nul 2>&1
if errorlevel 1 (
    call :print_warning "Redis连接失败，将使用内存存储"
) else (
    call :print_success "Redis连接正常"
)
goto :eof

REM =============================================================================
REM 安装和配置
REM =============================================================================

:install_dependencies
call :print_step "安装Go依赖"

if not exist "go.mod" (
    call :print_error "未找到go.mod文件"
    exit /b 1
)

go mod tidy
go mod download
call :print_success "依赖安装完成"
goto :eof

:generate_random
REM 生成随机字符串（简化版）
set "chars=ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
set "result="
for /l %%i in (1,1,12) do (
    set /a "rand=!random! %% 62"
    for %%j in (!rand!) do set "result=!result!!chars:~%%j,1!"
)
set "%~1=!result!"
goto :eof

:generate_config
call :print_step "生成配置文件"

REM 生成随机密码
if "%MYSQL_PASSWORD%"=="" call :generate_random MYSQL_PASSWORD
if "%REDIS_PASSWORD%"=="" call :generate_random REDIS_PASSWORD
call :generate_random JWT_SECRET
call :generate_random PASSWORD_SALT
call :generate_random API_SECRET

REM 备份原配置
if exist "conf\app.conf" (
    for /f "tokens=1-3 delims=/ " %%a in ('date /t') do set mydate=%%c%%a%%b
    for /f "tokens=1-2 delims=: " %%a in ('time /t') do set mytime=%%a%%b
    copy "conf\app.conf" "conf\app.conf.backup.!mydate!_!mytime!" >nul
)

REM 创建配置目录
if not exist "conf" mkdir conf

REM 生成配置文件
(
echo # 应用配置
echo appname = admin-service
echo httpport = %SERVICE_PORT%
echo runmode = prod
echo autorender = false
echo copyrequestbody = true
echo EnableDocs = false
echo.
echo # 数据库配置
echo mysql_host = %MYSQL_HOST%
echo mysql_port = %MYSQL_PORT%
echo mysql_user = %MYSQL_USER%
echo mysql_password = %MYSQL_PASSWORD%
echo mysql_database = %MYSQL_DATABASE%
echo mysql_charset = utf8mb4
echo mysql_max_idle = 10
echo mysql_max_open = 100
echo.
echo # Redis配置
echo redis_host = %REDIS_HOST%
echo redis_port = %REDIS_PORT%
echo redis_password = %REDIS_PASSWORD%
echo redis_database = 0
echo redis_pool_size = 10
echo.
echo # JWT配置
echo jwt_secret = %JWT_SECRET%
echo jwt_expire = 86400
echo.
echo # 加密配置
echo password_salt = %PASSWORD_SALT%
echo api_secret = %API_SECRET%
echo.
echo # 日志配置
echo [logs]
echo level = 4
echo separate = ["emergency", "alert", "critical", "error", "warning", "notice", "info"]
echo.
echo # 会话配置
echo sessionon = true
echo sessionprovider = memory
echo sessiongcmaxlifetime = 3600
echo sessioncookielifetime = 3600
echo.
echo # CORS配置
echo cors_allow_origins = *
echo cors_allow_methods = GET,POST,PUT,DELETE,OPTIONS
echo cors_allow_headers = Origin,Content-Type,Accept,Authorization,X-Requested-With
echo.
echo # 限流配置
echo rate_limit_requests = 1000
echo rate_limit_duration = 60
echo.
echo # 文件上传配置
echo max_memory = 67108864
echo max_file_size = 10485760
echo.
echo # 自动部署配置
echo auto_install = true
echo auto_create_database = true
echo auto_create_admin = true
echo default_admin_username = admin
echo default_admin_password = admin123
) > conf\app.conf

call :print_success "配置文件生成完成"

REM 保存安装信息
(
echo # Minigame Admin Service 安装信息
echo # 安装时间: %date% %time%
echo.
echo MYSQL_HOST=%MYSQL_HOST%
echo MYSQL_PORT=%MYSQL_PORT%
echo MYSQL_USER=%MYSQL_USER%
echo MYSQL_PASSWORD=%MYSQL_PASSWORD%
echo MYSQL_DATABASE=%MYSQL_DATABASE%
echo.
echo REDIS_HOST=%REDIS_HOST%
echo REDIS_PORT=%REDIS_PORT%
echo REDIS_PASSWORD=%REDIS_PASSWORD%
echo.
echo SERVICE_PORT=%SERVICE_PORT%
echo.
echo # 默认管理员账号
echo DEFAULT_ADMIN_USERNAME=admin
echo DEFAULT_ADMIN_PASSWORD=admin123
echo.
echo # 注意：请妥善保管此文件，首次运行后建议修改默认密码
) > .install_info

call :print_success "安装信息保存完成"
goto :eof

:build_project
call :print_step "构建项目"

REM 创建输出目录
if not exist "bin" mkdir bin

REM 构建项目
go build -o bin\admin-service.exe .
if errorlevel 1 (
    call :print_error "项目构建失败"
    exit /b 1
)

call :print_success "项目构建成功"
goto :eof

:create_service_script
call :print_step "创建服务脚本"

REM 创建启动脚本
(
echo @echo off
echo REM Minigame Admin Service 启动脚本
echo.
echo cd /d "%%~dp0"
echo.
echo echo 🚀 启动 Minigame Admin Service...
echo.
echo REM 检查配置文件
echo if not exist "conf\app.conf" ^(
echo     echo ❌ 配置文件不存在，请先运行安装脚本
echo     pause
echo     exit /b 1
echo ^)
echo.
echo REM 检查二进制文件
echo if not exist "bin\admin-service.exe" ^(
echo     echo ❌ 程序文件不存在，请先构建项目
echo     pause
echo     exit /b 1
echo ^)
echo.
echo REM 启动服务
echo for /f "tokens=2 delims==" %%%%i in ^('findstr "httpport" conf\app.conf'^) do set PORT=%%%%i
echo set PORT=%%PORT: =%%
echo echo 📝 配置文件: conf\app.conf
echo echo 🌐 服务端口: %%PORT%%
echo echo 📊 管理界面: http://localhost:%%PORT%%
echo echo.
echo.
echo bin\admin-service.exe
echo pause
) > start.bat

REM 创建停止脚本
(
echo @echo off
echo echo 🛑 停止 Minigame Admin Service...
echo.
echo taskkill /f /im admin-service.exe >nul 2>&1
echo if errorlevel 1 ^(
echo     echo ⚠️  服务未运行
echo ^) else ^(
echo     echo ✅ 服务已停止
echo ^)
echo pause
) > stop.bat

call :print_success "服务脚本创建完成"
goto :eof

REM =============================================================================
REM 主安装流程
REM =============================================================================

:show_welcome
cls
call :print_title "%PROJECT_NAME% 部署脚本"

echo     ╔══════════════════════════════════════════════════════════════════╗
echo     ║                🎮 Minigame Admin Service Installer               ║
echo     ║                                                                  ║
echo     ║  本脚本将自动完成以下操作：                                        ║
echo     ║  • 检查Go环境                                                    ║
echo     ║  • 安装项目依赖                                                  ║
echo     ║  • 生成配置文件                                                  ║
echo     ║  • 构建项目                                                      ║
echo     ║  • 创建启动脚本                                                  ║
echo     ║                                                                  ║
echo     ║  数据库初始化将在首次启动时自动完成                               ║
echo     ╚══════════════════════════════════════════════════════════════════╝
echo.
goto :eof

:show_results
call :print_title "部署完成"

echo 🎉 Minigame Admin Service 部署成功！
echo.
echo 📋 部署摘要：
for /f "tokens=3" %%i in ('go version') do echo ├── Go版本: %%i
echo ├── 服务端口: %SERVICE_PORT%
echo ├── 配置文件: conf\app.conf
echo ├── 程序文件: bin\admin-service.exe
echo ├── 安装信息: .install_info
echo.
echo 🚀 启动服务：
echo start.bat
echo.
echo 🛑 停止服务：
echo stop.bat
echo.
echo 🌐 访问地址：
echo http://localhost:%SERVICE_PORT%
echo.
echo 👤 默认管理员账号：
echo 用户名: admin
echo 密码: admin123
echo.
echo ⚠️  重要说明：
echo 1. 首次启动时会自动创建数据库和表结构
echo 2. 如果MySQL未安装，程序会使用SQLite作为后备数据库
echo 3. 请在首次登录后修改默认管理员密码
echo 4. 生产环境请配置真实的数据库连接
echo.
echo 📚 更多操作：
echo ├── 重新构建: go build -o bin\admin-service.exe .
echo ├── 运行测试: go test .\...
echo.

call :print_success "部署完成！现在可以启动服务了！"
goto :eof

:main
call :show_welcome

set /p "response=是否开始部署？[y/N] "
if /i not "%response%"=="y" (
    echo 部署已取消
    pause
    exit /b 0
)

REM 记录开始时间
set START_TIME=%time%

REM 执行部署步骤
call :check_go
if errorlevel 1 (
    pause
    exit /b 1
)

call :check_mysql
call :check_redis
call :install_dependencies
call :generate_config
call :build_project
call :create_service_script

REM 显示结果
call :show_results

echo 部署完成！
pause
exit /b 0

REM 错误处理
:error_handler
call :print_error "部署过程中发生错误"
pause
exit /b 1

REM 运行主程序
call :main %*
