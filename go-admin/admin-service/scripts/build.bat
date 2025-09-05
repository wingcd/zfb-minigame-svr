@echo off
REM =============================================================================
REM Minigame Admin Service 编译脚本 (Windows)
REM 支持Go后端和前端项目的编译打包
REM =============================================================================
chcp 65001

cd ..

setlocal EnableDelayedExpansion

REM 配置变量
set "PROJECT_NAME=Minigame Admin Service"
set "VERSION=1.0.0"
if defined VERSION_ENV set "VERSION=%VERSION_ENV%"

REM 获取构建时间和Git信息
for /f "tokens=1-3 delims=/ " %%a in ('date /t') do set BUILD_DATE=%%c-%%a-%%b
for /f "tokens=1-2 delims=: " %%a in ('time /t') do set BUILD_TIME_ONLY=%%a:%%b
set "BUILD_TIME=%BUILD_DATE% %BUILD_TIME_ONLY%"

REM 获取Git提交ID
git rev-parse --short HEAD >nul 2>&1
if errorlevel 1 (
    set "GIT_COMMIT=unknown"
) else (
    for /f %%i in ('git rev-parse --short HEAD') do set "GIT_COMMIT=%%i"
)

REM 构建目标
if not defined BUILD_BACKEND set "BUILD_BACKEND=true"
if not defined BUILD_FRONTEND set "BUILD_FRONTEND=true"
if not defined CREATE_RELEASE set "CREATE_RELEASE=false"

REM 路径配置
set "SCRIPT_DIR=%~dp0"
set "SCRIPT_DIR=%SCRIPT_DIR:~0,-1%"
for %%i in ("%SCRIPT_DIR%\..") do set "ROOT_DIR=%%~fi"
set "ADMIN_SERVICE_DIR=%SCRIPT_DIR%"
set "FRONTEND_DIR=%ROOT_DIR%\game-admin"
set "BUILD_OUTPUT_DIR=%SCRIPT_DIR%\dist"
set "RELEASE_DIR=%SCRIPT_DIR%\release"

REM 支持的平台
set PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

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

:check_environment
call :print_step "检查构建环境"

REM 检查Go环境
call :command_exists go
if errorlevel 1 (
    call :print_error "未找到Go环境，请先安装Go 1.21+"
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
call :print_success "Go版本: %GO_VERSION%"

REM 检查Node.js环境（如果需要构建前端）
if "%BUILD_FRONTEND%"=="true" (
    call :command_exists node
    if errorlevel 1 (
        call :print_warning "未找到Node.js环境，跳过前端构建"
        set "BUILD_FRONTEND=false"
    ) else (
        for /f %%i in ('node --version') do set NODE_VERSION=%%i
        call :print_success "Node.js版本: !NODE_VERSION!"
    )
)

REM 检查Git
call :command_exists git
if not errorlevel 1 (
    for /f "tokens=3" %%i in ('git --version') do call :print_success "Git版本: %%i"
)
goto :eof

REM =============================================================================
REM Go后端编译
REM =============================================================================

:build_backend
call :print_title "编译Go后端服务"

cd /d "%ADMIN_SERVICE_DIR%"

REM 清理旧的构建文件
if exist "%BUILD_OUTPUT_DIR%\backend" rmdir /s /q "%BUILD_OUTPUT_DIR%\backend"
mkdir "%BUILD_OUTPUT_DIR%\backend" 2>nul

REM 安装依赖
call :print_step "安装Go依赖"
go mod tidy
go mod download

REM 编译不同平台的二进制文件
for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        set "GOOS=%%a"
        set "GOARCH=%%b"
        
        call :print_step "编译 !GOOS!/!GOARCH!"
        
        set "output_name=admin-service"
        if "!GOOS!"=="windows" set "output_name=admin-service.exe"
        
        set "output_dir=%BUILD_OUTPUT_DIR%\backend\!GOOS!-!GOARCH!"
        mkdir "!output_dir!" 2>nul
        
        REM 设置构建标签
        set "ldflags=-X main.version=%VERSION% -X 'main.buildTime=%BUILD_TIME%' -X main.gitCommit=%GIT_COMMIT%"
        
        REM 编译
        set GOOS=!GOOS!
        set GOARCH=!GOARCH!
        set CGO_ENABLED=0
        go build -ldflags "!ldflags!" -o "!output_dir!\!output_name!" .
        
        if errorlevel 1 (
            call :print_error "!GOOS!/!GOARCH! 编译失败"
            exit /b 1
        )
        
        REM 复制配置文件和资源
        if exist "conf" xcopy /e /i /q "conf" "!output_dir!\conf\" >nul
        if exist "views" xcopy /e /i /q "views" "!output_dir!\views\" >nul
        if exist "static" xcopy /e /i /q "static" "!output_dir!\static\" >nul
        
        REM 创建启动脚本
        if "!GOOS!"=="windows" (
            (
                echo @echo off
                echo echo 🚀 启动 Minigame Admin Service...
                echo admin-service.exe
                echo pause
            ) > "!output_dir!\start.bat"
        ) else (
            (
                echo #!/bin/bash
                echo echo "🚀 启动 Minigame Admin Service..."
                echo ./admin-service
            ) > "!output_dir!\start.sh"
        )
        
        call :print_success "!GOOS!/!GOARCH! 编译完成"
    )
)

call :print_success "Go后端编译完成"
goto :eof

REM =============================================================================
REM 前端编译
REM =============================================================================

:build_frontend
call :print_title "编译Vue前端项目"

if not exist "%FRONTEND_DIR%" (
    call :print_warning "前端目录不存在，跳过前端构建"
    goto :eof
)

cd /d "%FRONTEND_DIR%"

REM 检查package.json
if not exist "package.json" (
    call :print_warning "未找到package.json，跳过前端构建"
    goto :eof
)

REM 安装依赖
call :print_step "安装前端依赖"
call :command_exists yarn
if not errorlevel 1 (
    yarn install
) else (
    npm install
)

REM 构建生产版本
call :print_step "构建生产版本"
call :command_exists yarn
if not errorlevel 1 (
    yarn build
) else (
    npm run build
)

REM 复制构建结果
if exist "dist" (
    mkdir "%BUILD_OUTPUT_DIR%\frontend" 2>nul
    xcopy /e /i /q "dist\*" "%BUILD_OUTPUT_DIR%\frontend\" >nul
    call :print_success "前端构建完成"
) else (
    call :print_error "前端构建失败，未找到dist目录"
)
goto :eof

REM =============================================================================
REM 创建发布包
REM =============================================================================

:create_release
call :print_title "创建发布包"

if exist "%RELEASE_DIR%" rmdir /s /q "%RELEASE_DIR%"
mkdir "%RELEASE_DIR%" 2>nul

REM 为每个平台创建发布包
for %%p in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%p") do (
        set "GOOS=%%a"
        set "GOARCH=%%b"
        
        call :print_step "创建 !GOOS!-!GOARCH! 发布包"
        
        set "platform_dir=!GOOS!-!GOARCH!"
        set "release_name=minigame-admin-service-%VERSION%-!platform_dir!"
        set "release_path=%RELEASE_DIR%\!release_name!"
        
        mkdir "!release_path!" 2>nul
        
        REM 复制后端文件
        if exist "%BUILD_OUTPUT_DIR%\backend\!platform_dir!" (
            xcopy /e /i /q "%BUILD_OUTPUT_DIR%\backend\!platform_dir!\*" "!release_path!\" >nul
        )
        
        REM 复制前端文件
        if exist "%BUILD_OUTPUT_DIR%\frontend" (
            mkdir "!release_path!\static\admin" 2>nul
            xcopy /e /i /q "%BUILD_OUTPUT_DIR%\frontend\*" "!release_path!\static\admin\" >nul
        )
        
        REM 复制文档和脚本
        copy "%SCRIPT_DIR%\install.sh" "!release_path!\" >nul 2>&1
        copy "%SCRIPT_DIR%\install.bat" "!release_path!\" >nul 2>&1
        
        REM 创建README
        (
            echo # Minigame Admin Service %VERSION%
            echo.
            echo ## 快速开始
            echo.
            echo ### 自动安装
            echo ```bash
            echo # Linux/macOS
            echo chmod +x install.sh
            echo ./install.sh
            echo.
            echo # Windows
            echo install.bat
            echo ```
            echo.
            echo ### 手动启动
            echo ```bash
            echo # Linux/macOS
            echo ./start.sh
            echo.
            echo # Windows
            echo start.bat
            echo ```
            echo.
            echo ## 访问地址
            echo - 管理界面: http://localhost:8080
            echo - 默认账号: admin/admin123
            echo.
            echo ## 版本信息
            echo - 版本: %VERSION%
            echo - 构建时间: %BUILD_TIME%
            echo - Git提交: %GIT_COMMIT%
            echo - 平台: !GOOS!/!GOARCH!
            echo.
            echo ## 更多信息
            echo 请访问项目文档了解更多配置和使用方法。
        ) > "!release_path!\README.md"
        
        REM 创建压缩包
        cd /d "%RELEASE_DIR%"
        
        REM 使用PowerShell创建ZIP文件
        powershell -command "Compress-Archive -Path '!release_name!' -DestinationPath '!release_name!.zip' -Force" >nul 2>&1
        if not errorlevel 1 (
            call :print_success "!release_name!.zip 创建完成"
        )
        
        cd /d "%SCRIPT_DIR%"
    )
)

call :print_success "发布包创建完成"
goto :eof

REM =============================================================================
REM 清理函数
REM =============================================================================

:clean_build
call :print_step "清理构建文件"
if exist "%BUILD_OUTPUT_DIR%" rmdir /s /q "%BUILD_OUTPUT_DIR%"
if exist "%RELEASE_DIR%" rmdir /s /q "%RELEASE_DIR%"
call :print_success "清理完成"
goto :eof

REM =============================================================================
REM 显示构建信息
REM =============================================================================

:show_build_info
call :print_title "构建信息"

echo 📋 构建配置：
echo ├── 项目名称: %PROJECT_NAME%
echo ├── 版本: %VERSION%
echo ├── 构建时间: %BUILD_TIME%
echo ├── Git提交: %GIT_COMMIT%
echo ├── 后端构建: %BUILD_BACKEND%
echo ├── 前端构建: %BUILD_FRONTEND%
echo └── 创建发布包: %CREATE_RELEASE%
echo.

echo 🎯 支持平台：
for %%p in (%PLATFORMS%) do echo ├── %%p
echo.
goto :eof

:show_build_results
call :print_title "构建完成"

echo 🎉 Minigame Admin Service 构建成功！
echo.

if exist "%BUILD_OUTPUT_DIR%" (
    echo 📁 构建输出目录: %BUILD_OUTPUT_DIR%
    echo ├── 后端文件: %BUILD_OUTPUT_DIR%\backend\
    echo └── 前端文件: %BUILD_OUTPUT_DIR%\frontend\
    echo.
)

if exist "%RELEASE_DIR%" (
    echo 📦 发布包目录: %RELEASE_DIR%
    for %%f in ("%RELEASE_DIR%\*.zip") do echo ├── %%~nxf
    echo.
)

echo 🚀 快速启动：
echo cd %BUILD_OUTPUT_DIR%\backend\windows-amd64
echo start.bat
echo.
goto :eof

REM =============================================================================
REM 主函数
REM =============================================================================

:show_help
echo 用法: %~nx0 [选项]
echo.
echo 选项:
echo   /h, /help               显示帮助信息
echo   /v VERSION              设置版本号 (默认: 1.0.0)
echo   /backend-only           只构建后端
echo   /frontend-only          只构建前端
echo   /release                创建发布包
echo   /clean                  清理构建文件
echo.
echo 环境变量:
echo   VERSION_ENV             版本号
echo   BUILD_BACKEND           是否构建后端 (true/false)
echo   BUILD_FRONTEND          是否构建前端 (true/false)
echo   CREATE_RELEASE          是否创建发布包 (true/false)
echo.
echo 示例:
echo   %~nx0                   # 完整构建
echo   %~nx0 /backend-only     # 只构建后端
echo   %~nx0 /release          # 构建并创建发布包
echo   %~nx0 /clean            # 清理构建文件
goto :eof

:main
REM 解析命令行参数
:parse_args
if "%~1"=="" goto start_build
if /i "%~1"=="/h" goto show_help
if /i "%~1"=="/help" goto show_help
if /i "%~1"=="/v" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_args
)
if /i "%~1"=="/backend-only" (
    set "BUILD_BACKEND=true"
    set "BUILD_FRONTEND=false"
    shift
    goto parse_args
)
if /i "%~1"=="/frontend-only" (
    set "BUILD_BACKEND=false"
    set "BUILD_FRONTEND=true"
    shift
    goto parse_args
)
if /i "%~1"=="/release" (
    set "CREATE_RELEASE=true"
    shift
    goto parse_args
)
if /i "%~1"=="/clean" (
    call :clean_build
    exit /b 0
)
echo 未知选项: %~1
call :show_help
exit /b 1

:start_build
REM 显示构建信息
call :show_build_info

REM 检查环境
call :check_environment
if errorlevel 1 exit /b 1

REM 记录开始时间
set START_TIME=%time%

REM 执行构建
if "%BUILD_BACKEND%"=="true" (
    call :build_backend
    if errorlevel 1 exit /b 1
)

if "%BUILD_FRONTEND%"=="true" (
    call :build_frontend
    if errorlevel 1 exit /b 1
)

if "%CREATE_RELEASE%"=="true" (
    call :create_release
    if errorlevel 1 exit /b 1
)

REM 显示结果
call :show_build_results

call :print_success "构建完成！"
goto :eof

REM 运行主函数
call :main %*
