@echo off
setlocal enabledelayedexpansion

chcp 65001

:: 确保在正确的目录中运行
set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
cd /d "%PROJECT_DIR%"

:: Admin Service API 测试运行脚本 (Windows版本)
:: 基于云函数接口格式，为admin-service创建的完整测试环境

title Admin Service API 测试套件

:: 颜色定义 (使用echo命令的颜色代码)
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

:: 打印带颜色的消息
goto :main

:print_message
echo %~2
goto :eof

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

:: 检测MySQL配置
:detect_mysql_config
set MYSQL_CMD=mysql -hlocalhost -uroot
set MYSQL_CONFIG_FOUND=false

:: 尝试localhost无密码连接
mysql -hlocalhost -uroot -e "SELECT 1;" >nul 2>&1
if not errorlevel 1 (
    set MYSQL_CONFIG_FOUND=true
    set MYSQL_CMD=mysql -hlocalhost -uroot
    goto :eof
)

:: 尝试localhost常见密码
for %%p in ("123456" "root" "password" "") do (
    mysql -hlocalhost -uroot -p%%p -e "SELECT 1;" >nul 2>&1
    if not errorlevel 1 (
        set MYSQL_CONFIG_FOUND=true
        set MYSQL_CMD=mysql -hlocalhost -uroot -p%%p
        goto :eof
    )
)

:: 尝试127.0.0.1无密码连接
mysql -h127.0.0.1 -uroot -e "SELECT 1;" >nul 2>&1
if not errorlevel 1 (
    set MYSQL_CONFIG_FOUND=true
    set MYSQL_CMD=mysql -h127.0.0.1 -uroot
    goto :eof
)

:: 尝试127.0.0.1常见密码
for %%p in ("123456" "root" "password" "") do (
    mysql -h127.0.0.1 -uroot -p%%p -e "SELECT 1;" >nul 2>&1
    if not errorlevel 1 (
        set MYSQL_CONFIG_FOUND=true
        set MYSQL_CMD=mysql -h127.0.0.1 -uroot -p%%p
        goto :eof
    )
)
goto :eof

:: 检查依赖
:check_dependencies
call :print_header "检查依赖"

:: 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    call :print_error "Go 未安装或未在PATH中"
    exit /b 1
)
for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
call :print_success "Go 环境检查通过: !GO_VERSION!"

:: 检查MySQL
mysql --version >nul 2>&1
if errorlevel 1 (
    call :print_warning "MySQL客户端未找到，请确保MySQL服务正在运行"
) else (
    call :print_success "MySQL 客户端检查通过"
)

:: 检查测试数据库连接
call :detect_mysql_config
if "%MYSQL_CONFIG_FOUND%"=="true" (
    call :print_success "数据库连接检查通过"
) else (
    call :print_warning "无法连接到MySQL数据库，请检查配置"
    call :print_info "请确保MySQL服务正在运行，并检查用户名密码"
)
goto :eof

:: 设置测试环境
:setup_test_env
call :print_header "设置测试环境"

:: 创建测试数据库
call :print_info "创建测试数据库..."
if "%MYSQL_CONFIG_FOUND%"=="false" (
    call :print_error "MySQL配置未找到，无法创建测试数据库"
    exit /b 1
)

%MYSQL_CMD% -e "CREATE DATABASE IF NOT EXISTS admin_service_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" >nul 2>&1
if errorlevel 1 (
    call :print_error "创建测试数据库失败，请检查MySQL连接"
    exit /b 1
)
call :print_success "测试数据库创建成功"

:: 设置环境变量
set GO_ENV=test
set DB_HOST=127.0.0.1
set DB_PORT=3306
set DB_USER=root
set DB_PASSWORD=
set DB_NAME=admin_service_test

call :print_success "环境变量设置完成"
goto :eof

:: 构建项目
:build_project
call :print_header "构建项目"

call :print_info "下载依赖..."
go mod tidy
if errorlevel 1 (
    call :print_error "依赖下载失败"
    exit /b 1
)

call :print_info "构建项目..."
if not exist "bin" mkdir bin
go build -o bin\admin-service.exe main.go
if errorlevel 1 (
    call :print_error "项目构建失败"
    exit /b 1
)

call :print_success "项目构建完成"
goto :eof

:: 运行单元测试
:run_unit_tests
call :print_header "运行单元测试"

call :print_info "运行所有单元测试..."
go test -v ./tests/ -timeout=10m -count=1
if errorlevel 1 (
    call :print_error "单元测试失败"
    exit /b 1
) else (
    call :print_success "单元测试全部通过"
)
goto :eof

:: 运行性能测试
:run_benchmark_tests
call :print_header "运行性能测试"

call :print_info "运行性能基准测试..."
go test -v ./tests/ -bench=. -benchtime=10s -timeout=5m
if errorlevel 1 (
    call :print_error "性能测试失败"
    exit /b 1
) else (
    call :print_success "性能测试完成"
)
goto :eof

:: 运行覆盖率测试
:run_coverage_tests
call :print_header "运行覆盖率测试"

call :print_info "生成测试覆盖率报告..."
go test -v ./tests/ -coverprofile=coverage.out -timeout=10m
if errorlevel 1 (
    call :print_error "覆盖率测试失败"
    exit /b 1
)

if exist coverage.out (
    call :print_info "生成HTML覆盖率报告..."
    go tool cover -html=coverage.out -o coverage.html
    
    call :print_info "覆盖率统计:"
    go tool cover -func=coverage.out | findstr "total"
    
    call :print_success "覆盖率报告生成完成: coverage.html"
) else (
    call :print_error "覆盖率文件生成失败"
    exit /b 1
)
goto :eof

:: 运行API集成测试
:run_integration_tests
call :print_header "运行API集成测试"

call :print_info "启动测试服务器..."
:: 在后台启动服务器
start /b "" bin\admin-service.exe
set SERVER_STARTED=1

:: 等待服务器启动
timeout /t 3 /nobreak >nul

:: 检查服务器是否启动成功
curl -s http://localhost:8080/health >nul 2>&1
if errorlevel 1 (
    call :print_error "测试服务器启动失败"
    taskkill /im admin-service.exe /f >nul 2>&1
    exit /b 1
)

call :print_success "测试服务器启动成功"

:: 运行集成测试
call :print_info "执行API集成测试..."
go test -v ./tests/ -tags=integration -timeout=10m

:: 停止测试服务器
call :print_info "停止测试服务器..."
taskkill /im admin-service.exe /f >nul 2>&1

call :print_success "API集成测试完成"
goto :eof

:: 生成测试报告
:generate_test_report
call :print_header "生成测试报告"

:: 获取当前时间戳
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "YY=%dt:~2,2%" & set "YYYY=%dt:~0,4%" & set "MM=%dt:~4,2%" & set "DD=%dt:~6,2%"
set "HH=%dt:~8,2%" & set "Min=%dt:~10,2%" & set "Sec=%dt:~12,2%"
set "timestamp=%YYYY%%MM%%DD%_%HH%%Min%%Sec%"

set "report_file=test_report_%timestamp%.html"

:: 创建HTML报告
(
echo ^<!DOCTYPE html^>
echo ^<html^>
echo ^<head^>
echo     ^<meta charset="utf-8"^>
echo     ^<title^>Admin Service API 测试报告^</title^>
echo     ^<style^>
echo         body { font-family: 'Microsoft YaHei', Arial, sans-serif; margin: 20px; }
echo         .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
echo         .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
echo         .success { color: #28a745; }
echo         .error { color: #dc3545; }
echo         .warning { color: #ffc107; }
echo         .info { color: #17a2b8; }
echo         table { width: 100%%; border-collapse: collapse; margin: 10px 0; }
echo         th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
echo         th { background-color: #f2f2f2; }
echo         .code { background: #f8f9fa; padding: 10px; border-radius: 3px; font-family: 'Consolas', monospace; }
echo     ^</style^>
echo ^</head^>
echo ^<body^>
echo     ^<div class="header"^>
echo         ^<h1^>Admin Service API 测试报告 ^(Windows版本^)^</h1^>
echo         ^<p^>^<strong^>生成时间:^</strong^> %date% %time%^</p^>
echo         ^<p^>^<strong^>测试环境:^</strong^> Windows %OS%^</p^>
for /f "tokens=*" %%i in ('go version') do echo         ^<p^>^<strong^>Go版本:^</strong^> %%i^</p^>
echo     ^</div^>
echo     
echo     ^<div class="section"^>
echo         ^<h2^>测试概览^</h2^>
echo         ^<p^>本测试基于云函数接口格式，为admin-service的Go接口创建了完整的测试环境。^</p^>
echo         
echo         ^<h3^>测试范围^</h3^>
echo         ^<ul^>
echo             ^<li^>用户管理接口 ^(UserController^)^</li^>
echo             ^<li^>系统管理接口 ^(SystemController^)^</li^>
echo             ^<li^>统计分析接口 ^(StatisticsController^)^</li^>
echo             ^<li^>错误处理和边界情况^</li^>
echo             ^<li^>性能和负载测试^</li^>
echo         ^</ul^>
echo         
echo         ^<h3^>测试特点^</h3^>
echo         ^<ul^>
echo             ^<li^>✅ 参考云函数的输入输出格式^</li^>
echo             ^<li^>✅ 标准化的响应结构 ^(code, msg, timestamp, data^)^</li^>
echo             ^<li^>✅ 完整的测试数据和Mock环境^</li^>
echo             ^<li^>✅ 自动化的测试流程和清理^</li^>
echo             ^<li^>✅ 详细的错误处理和验证^</li^>
echo         ^</ul^>
echo     ^</div^>
echo     
echo     ^<div class="section"^>
echo         ^<h2^>Windows环境特殊说明^</h2^>
echo         ^<ul^>
echo             ^<li^>使用 .bat 批处理脚本代替 shell 脚本^</li^>
echo             ^<li^>使用 taskkill 命令管理进程^</li^>
echo             ^<li^>使用 timeout 命令代替 sleep^</li^>
echo             ^<li^>支持中文字符显示^</li^>
echo             ^<li^>兼容 Windows 10/11 系统^</li^>
echo         ^</ul^>
echo     ^</div^>
echo     
echo     ^<div class="section"^>
echo         ^<h2^>使用说明^</h2^>
echo         ^<h3^>运行所有测试^</h3^>
echo         ^<div class="code"^>run_tests.bat^</div^>
echo         
echo         ^<h3^>运行特定测试^</h3^>
echo         ^<div class="code"^>
echo REM 只运行用户API测试
echo go test -v ./tests/ -run TestUserAPIs
echo.
echo REM 只运行系统API测试  
echo go test -v ./tests/ -run TestSystemAPIs
echo.
echo REM 只运行统计API测试
echo go test -v ./tests/ -run TestStatisticsAPIs
echo.
echo REM 运行性能测试
echo go test -v ./tests/ -bench=.
echo         ^</div^>
echo         
echo         ^<h3^>命令行选项^</h3^>
echo         ^<div class="code"^>
echo run_tests.bat --help          显示帮助信息
echo run_tests.bat --skip-build    跳过项目构建
echo run_tests.bat --skip-cleanup  跳过环境清理
echo run_tests.bat --coverage      运行覆盖率测试
echo run_tests.bat --benchmark     运行性能测试
echo run_tests.bat --integration   运行集成测试
echo         ^</div^>
echo     ^</div^>
echo     
echo     ^<div class="section"^>
echo         ^<h2^>故障排除^</h2^>
echo         ^<h3^>常见问题^</h3^>
echo         ^<ul^>
echo             ^<li^>^<strong^>MySQL连接失败:^</strong^> 请确保MySQL服务正在运行，用户名密码正确^</li^>
echo             ^<li^>^<strong^>Go命令未找到:^</strong^> 请检查Go是否正确安装并添加到PATH^</li^>
echo             ^<li^>^<strong^>端口占用:^</strong^> 请确保8080端口未被其他程序占用^</li^>
echo             ^<li^>^<strong^>权限问题:^</strong^> 请以管理员身份运行批处理文件^</li^>
echo         ^</ul^>
echo         
echo         ^<h3^>日志查看^</h3^>
echo         ^<div class="code"^>
echo REM 查看详细测试日志
echo go test -v ./tests/ -timeout=10m
echo.
echo REM 查看特定测试的输出
echo go test -v ./tests/ -run TestUserAPIs -timeout=5m
echo         ^</div^>
echo     ^</div^>
echo ^</body^>
echo ^</html^>
) > "%report_file%"

call :print_success "测试报告生成完成: %report_file%"
goto :eof

:: 清理测试环境
:cleanup_test_env
call :print_header "清理测试环境"

call :print_info "删除测试数据库..."
if "%MYSQL_CONFIG_FOUND%"=="true" (
    %MYSQL_CMD% -e "DROP DATABASE IF EXISTS admin_service_test;" >nul 2>&1
    if errorlevel 1 (
        call :print_warning "删除测试数据库失败，请手动清理"
    )
) else (
    call :print_warning "MySQL配置未找到，请手动删除测试数据库"
)

call :print_info "清理临时文件..."
if exist coverage.out del /f coverage.out
if exist bin\admin-service.exe del /f bin\admin-service.exe

:: 停止可能残留的服务器进程
taskkill /im admin-service.exe /f >nul 2>&1

call :print_success "测试环境清理完成"
goto :eof

:: 显示帮助信息
:show_help
echo 用法: %~nx0 [选项]
echo.
echo 选项:
echo   --skip-build    跳过项目构建
echo   --skip-cleanup  跳过环境清理
echo   --coverage      运行覆盖率测试
echo   --benchmark     运行性能测试
echo   --integration   运行集成测试
echo   --help          显示帮助信息
echo.
echo 示例:
echo   %~nx0                    运行基础测试
echo   %~nx0 --coverage         运行覆盖率测试
echo   %~nx0 --benchmark        运行性能测试
echo   %~nx0 --integration      运行集成测试
echo   %~nx0 --skip-cleanup     运行测试但不清理环境
goto :eof

:: 主函数
:main
call :print_header "Admin Service API 测试套件 (Windows版本)"
call :print_info "基于云函数接口格式的完整测试环境"

:: 解析命令行参数
set SKIP_BUILD=false
set SKIP_CLEANUP=false
set RUN_COVERAGE=false
set RUN_BENCHMARK=false
set RUN_INTEGRATION=false

:parse_args
if "%~1"=="" goto :start_tests
if "%~1"=="--skip-build" (
    set SKIP_BUILD=true
    shift
    goto :parse_args
)
if "%~1"=="--skip-cleanup" (
    set SKIP_CLEANUP=true
    shift
    goto :parse_args
)
if "%~1"=="--coverage" (
    set RUN_COVERAGE=true
    shift
    goto :parse_args
)
if "%~1"=="--benchmark" (
    set RUN_BENCHMARK=true
    shift
    goto :parse_args
)
if "%~1"=="--integration" (
    set RUN_INTEGRATION=true
    shift
    goto :parse_args
)
if "%~1"=="--help" (
    call :show_help
    exit /b 0
)
call :print_error "未知选项: %~1"
call :show_help
exit /b 1

:start_tests
:: 执行测试流程
call :check_dependencies
if errorlevel 1 exit /b 1

call :setup_test_env
if errorlevel 1 exit /b 1

if "%SKIP_BUILD%"=="false" (
    call :build_project
    if errorlevel 1 exit /b 1
)

:: 运行基础测试
call :run_unit_tests
if errorlevel 1 (
    call :print_error "基础测试失败，停止执行"
    goto :cleanup_and_exit
)

:: 运行可选测试
if "%RUN_COVERAGE%"=="true" (
    call :run_coverage_tests
)

if "%RUN_BENCHMARK%"=="true" (
    call :run_benchmark_tests
)

if "%RUN_INTEGRATION%"=="true" (
    call :run_integration_tests
)

:: 生成报告
call :generate_test_report

:cleanup_and_exit
:: 清理环境
if "%SKIP_CLEANUP%"=="false" (
    call :cleanup_test_env
)

call :print_header "测试完成"
call :print_success "所有测试执行完毕！"
call :print_info "查看详细报告: test_report_*.html"

:: 如果是双击运行，暂停以便查看结果
if "%~1"=="" pause

endlocal
exit /b 0

