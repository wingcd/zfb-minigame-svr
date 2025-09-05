# Admin Service API 测试运行脚本 (PowerShell版本)
# 基于云函数接口格式，为admin-service创建的完整测试环境

param(
    [switch]$SkipBuild,
    [switch]$SkipCleanup,
    [switch]$Coverage,
    [switch]$Benchmark,
    [switch]$Integration,
    [switch]$Help
)

# 设置控制台编码为UTF-8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# 颜色定义
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    Cyan = "Cyan"
    White = "White"
}

# 打印函数
function Write-ColoredMessage {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Header {
    param([string]$Title)
    Write-Host ""
    Write-ColoredMessage "==================================================" "Blue"
    Write-ColoredMessage $Title "Blue"
    Write-ColoredMessage "==================================================" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColoredMessage "✅ $Message" "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColoredMessage "❌ $Message" "Red"
}

function Write-Warning {
    param([string]$Message)
    Write-ColoredMessage "⚠️  $Message" "Yellow"
}

function Write-Info {
    param([string]$Message)
    Write-ColoredMessage "ℹ️  $Message" "Cyan"
}

# 检查依赖
function Test-Dependencies {
    Write-Header "检查依赖"
    
    # 检查Go环境
    try {
        $goVersion = go version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Go 环境检查通过: $goVersion"
        } else {
            throw "Go command failed"
        }
    } catch {
        Write-Error "Go 未安装或未在PATH中"
        return $false
    }
    
    # 检查MySQL
    try {
        $mysqlVersion = mysql --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "MySQL 客户端检查通过"
        } else {
            Write-Warning "MySQL客户端未找到，请确保MySQL服务正在运行"
        }
    } catch {
        Write-Warning "MySQL客户端未找到，请确保MySQL服务正在运行"
    }
    
    # 检查测试数据库连接
    try {
        mysql -h127.0.0.1 -uroot -e "SELECT 1;" 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "数据库连接检查通过"
        } else {
            Write-Warning "无法连接到MySQL数据库，请检查配置"
        }
    } catch {
        Write-Warning "无法连接到MySQL数据库，请检查配置"
    }
    
    return $true
}

# 设置测试环境
function Initialize-TestEnvironment {
    Write-Header "设置测试环境"
    
    # 创建测试数据库
    Write-Info "创建测试数据库..."
    try {
        mysql -h127.0.0.1 -uroot -e "CREATE DATABASE IF NOT EXISTS admin_service_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>$null | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Database creation failed"
        }
        Write-Success "测试数据库创建成功"
    } catch {
        Write-Error "创建测试数据库失败，请检查MySQL连接"
        return $false
    }
    
    # 设置环境变量
    $env:GO_ENV = "test"
    $env:DB_HOST = "127.0.0.1"
    $env:DB_PORT = "3306"
    $env:DB_USER = "root"
    $env:DB_PASSWORD = ""
    $env:DB_NAME = "admin_service_test"
    
    Write-Success "环境变量设置完成"
    return $true
}

# 构建项目
function Build-Project {
    Write-Header "构建项目"
    
    Write-Info "下载依赖..."
    try {
        go mod tidy
        if ($LASTEXITCODE -ne 0) {
            throw "go mod tidy failed"
        }
    } catch {
        Write-Error "依赖下载失败"
        return $false
    }
    
    Write-Info "构建项目..."
    try {
        if (!(Test-Path "bin")) {
            New-Item -ItemType Directory -Path "bin" | Out-Null
        }
        go build -o "bin/admin-service.exe" main.go
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed"
        }
        Write-Success "项目构建完成"
    } catch {
        Write-Error "项目构建失败"
        return $false
    }
    
    return $true
}

# 运行单元测试
function Invoke-UnitTests {
    Write-Header "运行单元测试"
    
    Write-Info "运行所有单元测试..."
    try {
        go test -v ./tests/ -timeout=10m -count=1
        if ($LASTEXITCODE -ne 0) {
            throw "Unit tests failed"
        }
        Write-Success "单元测试全部通过"
    } catch {
        Write-Error "单元测试失败"
        return $false
    }
    
    return $true
}

# 运行性能测试
function Invoke-BenchmarkTests {
    Write-Header "运行性能测试"
    
    Write-Info "运行性能基准测试..."
    try {
        go test -v ./tests/ -bench=. -benchtime=10s -timeout=5m
        if ($LASTEXITCODE -ne 0) {
            throw "Benchmark tests failed"
        }
        Write-Success "性能测试完成"
    } catch {
        Write-Error "性能测试失败"
        return $false
    }
    
    return $true
}

# 运行覆盖率测试
function Invoke-CoverageTests {
    Write-Header "运行覆盖率测试"
    
    Write-Info "生成测试覆盖率报告..."
    try {
        go test -v ./tests/ -coverprofile=coverage.out -timeout=10m
        if ($LASTEXITCODE -ne 0) {
            throw "Coverage tests failed"
        }
        
        if (Test-Path "coverage.out") {
            Write-Info "生成HTML覆盖率报告..."
            go tool cover -html=coverage.out -o coverage.html
            
            Write-Info "覆盖率统计:"
            $coverageStats = go tool cover -func=coverage.out | Select-String "total"
            Write-Host $coverageStats
            
            Write-Success "覆盖率报告生成完成: coverage.html"
        } else {
            throw "Coverage file not generated"
        }
    } catch {
        Write-Error "覆盖率测试失败: $_"
        return $false
    }
    
    return $true
}

# 运行API集成测试
function Invoke-IntegrationTests {
    Write-Header "运行API集成测试"
    
    $serverProcess = $null
    try {
        Write-Info "启动测试服务器..."
        # 启动服务器进程
        $serverProcess = Start-Process -FilePath "bin/admin-service.exe" -PassThru -WindowStyle Hidden
        
        # 等待服务器启动
        Start-Sleep -Seconds 3
        
        # 检查服务器是否启动成功
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -TimeoutSec 5 -UseBasicParsing
            Write-Success "测试服务器启动成功 (PID: $($serverProcess.Id))"
        } catch {
            throw "Server health check failed"
        }
        
        # 运行集成测试
        Write-Info "执行API集成测试..."
        go test -v ./tests/ -tags=integration -timeout=10m
        if ($LASTEXITCODE -ne 0) {
            throw "Integration tests failed"
        }
        
        Write-Success "API集成测试完成"
    } catch {
        Write-Error "集成测试失败: $_"
        return $false
    } finally {
        # 停止测试服务器
        if ($serverProcess -and !$serverProcess.HasExited) {
            Write-Info "停止测试服务器..."
            Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
        }
    }
    
    return $true
}

# 生成测试报告
function New-TestReport {
    Write-Header "生成测试报告"
    
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    $reportFile = "test_report_$timestamp.html"
    
    $goVersion = go version
    $osInfo = [System.Environment]::OSVersion.VersionString
    $currentDate = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    
    $htmlContent = @"
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Admin Service API 测试报告</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .success { color: #28a745; }
        .error { color: #dc3545; }
        .warning { color: #ffc107; }
        .info { color: #17a2b8; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .code { background: #f8f9fa; padding: 10px; border-radius: 3px; font-family: 'Consolas', monospace; }
        .powershell { background: #012456; color: #ffffff; padding: 15px; border-radius: 5px; font-family: 'Consolas', monospace; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Admin Service API 测试报告 (PowerShell版本)</h1>
        <p><strong>生成时间:</strong> $currentDate</p>
        <p><strong>测试环境:</strong> Windows $osInfo</p>
        <p><strong>Go版本:</strong> $goVersion</p>
        <p><strong>PowerShell版本:</strong> $($PSVersionTable.PSVersion)</p>
    </div>
    
    <div class="section">
        <h2>PowerShell版本特性</h2>
        <ul>
            <li>✅ 现代化的PowerShell脚本，支持参数化调用</li>
            <li>✅ 彩色输出和更好的错误处理</li>
            <li>✅ 强类型参数和更安全的进程管理</li>
            <li>✅ 更好的异常处理和资源清理</li>
            <li>✅ UTF-8编码支持，完美显示中文</li>
            <li>✅ 兼容PowerShell 5.1+ 和 PowerShell Core 7+</li>
        </ul>
    </div>
    
    <div class="section">
        <h2>使用说明</h2>
        <h3>基础用法</h3>
        <div class="powershell">
# 运行所有测试
.\run_tests.ps1

# 运行特定测试类型
.\run_tests.ps1 -Coverage
.\run_tests.ps1 -Benchmark
.\run_tests.ps1 -Integration

# 跳过构建和清理
.\run_tests.ps1 -SkipBuild -SkipCleanup

# 组合使用
.\run_tests.ps1 -Coverage -Benchmark -SkipCleanup
        </div>
        
        <h3>参数说明</h3>
        <table>
            <tr>
                <th>参数</th>
                <th>说明</th>
                <th>示例</th>
            </tr>
            <tr>
                <td>-SkipBuild</td>
                <td>跳过项目构建步骤</td>
                <td>.\run_tests.ps1 -SkipBuild</td>
            </tr>
            <tr>
                <td>-SkipCleanup</td>
                <td>跳过环境清理步骤</td>
                <td>.\run_tests.ps1 -SkipCleanup</td>
            </tr>
            <tr>
                <td>-Coverage</td>
                <td>运行代码覆盖率测试</td>
                <td>.\run_tests.ps1 -Coverage</td>
            </tr>
            <tr>
                <td>-Benchmark</td>
                <td>运行性能基准测试</td>
                <td>.\run_tests.ps1 -Benchmark</td>
            </tr>
            <tr>
                <td>-Integration</td>
                <td>运行API集成测试</td>
                <td>.\run_tests.ps1 -Integration</td>
            </tr>
            <tr>
                <td>-Help</td>
                <td>显示帮助信息</td>
                <td>.\run_tests.ps1 -Help</td>
            </tr>
        </table>
        
        <h3>执行策略设置</h3>
        <div class="powershell">
# 如果遇到执行策略限制，请运行：
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 或者临时绕过：
PowerShell -ExecutionPolicy Bypass -File .\run_tests.ps1
        </div>
    </div>
    
    <div class="section">
        <h2>故障排除</h2>
        <h3>常见问题及解决方案</h3>
        <table>
            <tr>
                <th>问题</th>
                <th>可能原因</th>
                <th>解决方案</th>
            </tr>
            <tr>
                <td>执行策略限制</td>
                <td>PowerShell安全设置</td>
                <td>Set-ExecutionPolicy RemoteSigned</td>
            </tr>
            <tr>
                <td>MySQL连接失败</td>
                <td>服务未启动或配置错误</td>
                <td>检查MySQL服务状态和连接参数</td>
            </tr>
            <tr>
                <td>端口占用</td>
                <td>8080端口被占用</td>
                <td>netstat -ano | findstr :8080</td>
            </tr>
            <tr>
                <td>Go命令未找到</td>
                <td>Go未安装或PATH未设置</td>
                <td>重新安装Go并设置PATH</td>
            </tr>
        </table>
        
        <h3>调试命令</h3>
        <div class="powershell">
# 查看详细测试输出
go test -v ./tests/ -timeout=10m

# 查看特定测试
go test -v ./tests/ -run TestUserAPIs

# 查看进程占用
Get-Process | Where-Object {$_.ProcessName -like "*admin*"}

# 查看端口占用
netstat -ano | Select-String ":8080"
        </div>
    </div>
    
    <div class="section">
        <h2>性能优化建议</h2>
        <ul>
            <li>使用SSD硬盘可显著提升测试速度</li>
            <li>确保足够的内存(建议8GB+)</li>
            <li>关闭不必要的杀毒软件实时扫描</li>
            <li>使用本地MySQL实例而非远程连接</li>
            <li>定期清理临时文件和日志</li>
        </ul>
    </div>
</body>
</html>
"@
    
    try {
        $htmlContent | Out-File -FilePath $reportFile -Encoding UTF8
        Write-Success "测试报告生成完成: $reportFile"
        
        # 尝试打开报告
        if (Get-Command "Start-Process" -ErrorAction SilentlyContinue) {
            Write-Info "尝试在默认浏览器中打开报告..."
            Start-Process $reportFile -ErrorAction SilentlyContinue
        }
    } catch {
        Write-Error "生成测试报告失败: $_"
        return $false
    }
    
    return $true
}

# 清理测试环境
function Clear-TestEnvironment {
    Write-Header "清理测试环境"
    
    Write-Info "删除测试数据库..."
    try {
        mysql -h127.0.0.1 -uroot -e "DROP DATABASE IF EXISTS admin_service_test;" 2>$null | Out-Null
    } catch {
        Write-Warning "删除测试数据库失败，请手动清理"
    }
    
    Write-Info "清理临时文件..."
    try {
        if (Test-Path "coverage.out") { Remove-Item "coverage.out" -Force }
        if (Test-Path "bin/admin-service.exe") { Remove-Item "bin/admin-service.exe" -Force }
        
        # 停止可能残留的服务器进程
        Get-Process | Where-Object {$_.ProcessName -eq "admin-service"} | Stop-Process -Force -ErrorAction SilentlyContinue
    } catch {
        Write-Warning "清理临时文件时出现错误: $_"
    }
    
    Write-Success "测试环境清理完成"
}

# 显示帮助信息
function Show-Help {
    Write-Host @"

Admin Service API 测试套件 (PowerShell版本)
基于云函数接口格式的完整测试环境

用法:
    .\run_tests.ps1 [参数...]

参数:
    -SkipBuild      跳过项目构建
    -SkipCleanup    跳过环境清理
    -Coverage       运行覆盖率测试
    -Benchmark      运行性能测试
    -Integration    运行集成测试
    -Help           显示此帮助信息

示例:
    .\run_tests.ps1                          # 运行基础测试
    .\run_tests.ps1 -Coverage                # 运行覆盖率测试
    .\run_tests.ps1 -Benchmark -Integration  # 运行性能和集成测试
    .\run_tests.ps1 -SkipBuild -SkipCleanup  # 快速测试(跳过构建和清理)

注意:
    如果遇到执行策略限制，请运行:
    Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

"@ -ForegroundColor Cyan
}

# 主函数
function Main {
    # 显示帮助
    if ($Help) {
        Show-Help
        return
    }
    
    Write-Header "Admin Service API 测试套件 (PowerShell版本)"
    Write-Info "基于云函数接口格式的完整测试环境"
    
    $success = $true
    
    try {
        # 检查依赖
        if (!(Test-Dependencies)) {
            throw "依赖检查失败"
        }
        
        # 设置测试环境
        if (!(Initialize-TestEnvironment)) {
            throw "测试环境初始化失败"
        }
        
        # 构建项目
        if (!$SkipBuild) {
            if (!(Build-Project)) {
                throw "项目构建失败"
            }
        }
        
        # 运行基础测试
        if (!(Invoke-UnitTests)) {
            throw "基础测试失败"
        }
        
        # 运行可选测试
        if ($Coverage) {
            if (!(Invoke-CoverageTests)) {
                Write-Warning "覆盖率测试失败，但继续执行其他测试"
            }
        }
        
        if ($Benchmark) {
            if (!(Invoke-BenchmarkTests)) {
                Write-Warning "性能测试失败，但继续执行其他测试"
            }
        }
        
        if ($Integration) {
            if (!(Invoke-IntegrationTests)) {
                Write-Warning "集成测试失败，但继续执行其他测试"
            }
        }
        
        # 生成报告
        if (!(New-TestReport)) {
            Write-Warning "测试报告生成失败"
        }
        
    } catch {
        Write-Error "测试执行过程中出现错误: $_"
        $success = $false
    } finally {
        # 清理环境
        if (!$SkipCleanup) {
            Clear-TestEnvironment
        }
    }
    
    # 显示结果
    Write-Header "测试完成"
    if ($success) {
        Write-Success "所有测试执行完毕！"
        Write-Info "查看详细报告: test_report_*.html"
    } else {
        Write-Error "部分测试执行失败，请查看上述错误信息"
        exit 1
    }
}

# 运行主函数
Main
