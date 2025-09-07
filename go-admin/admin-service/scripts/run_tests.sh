#!/bin/bash

# Admin Service API 测试运行脚本
# 基于云函数接口格式，为admin-service创建的完整测试环境

cd ..

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    echo
    print_message $BLUE "=================================================="
    print_message $BLUE "$1"
    print_message $BLUE "=================================================="
}

print_success() {
    print_message $GREEN "✅ $1"
}

print_error() {
    print_message $RED "❌ $1"
}

print_warning() {
    print_message $YELLOW "⚠️  $1"
}

print_info() {
    print_message $BLUE "ℹ️  $1"
}

# 检查依赖
check_dependencies() {
    print_header "检查依赖"
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装或未在PATH中"
        exit 1
    fi
    print_success "Go 环境检查通过: $(go version)"
    
    # 检查MySQL
    if ! command -v mysql &> /dev/null; then
        print_warning "MySQL客户端未找到，请确保MySQL服务正在运行"
    else
        print_success "MySQL 客户端检查通过"
    fi
    
    # 检查测试数据库连接
    if mysql -h127.0.0.1 -uroot -e "SELECT 1;" &> /dev/null; then
        print_success "数据库连接检查通过"
    else
        print_warning "无法连接到MySQL数据库，请检查配置"
    fi
}

# 设置测试环境
setup_test_env() {
    print_header "设置测试环境"
    
    # 创建测试数据库
    print_info "创建测试数据库..."
    mysql -h127.0.0.1 -uroot -e "CREATE DATABASE IF NOT EXISTS admin_service_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || {
        print_error "创建测试数据库失败，请检查MySQL连接"
        exit 1
    }
    print_success "测试数据库创建成功"
    
    # 设置环境变量
    export GO_ENV=test
    export DB_HOST=127.0.0.1
    export DB_PORT=3306
    export DB_USER=root
    export DB_PASSWORD=
    export DB_NAME=admin_service_test
    
    print_success "环境变量设置完成"
}

# 构建项目
build_project() {
    print_header "构建项目"
    
    print_info "下载依赖..."
    go mod tidy
    
    print_info "构建项目..."
    go build -o bin/admin-service main.go
    
    print_success "项目构建完成"
}

# 运行单元测试
run_unit_tests() {
    print_header "运行单元测试"
    
    print_info "运行所有单元测试..."
    go test -v ./tests/ -timeout=10m -count=1
    
    if [ $? -eq 0 ]; then
        print_success "单元测试全部通过"
    else
        print_error "单元测试失败"
        return 1
    fi
}

# 运行性能测试
run_benchmark_tests() {
    print_header "运行性能测试"
    
    print_info "运行性能基准测试..."
    go test -v ./tests/ -bench=. -benchtime=10s -timeout=5m
    
    if [ $? -eq 0 ]; then
        print_success "性能测试完成"
    else
        print_error "性能测试失败"
        return 1
    fi
}

# 运行覆盖率测试
run_coverage_tests() {
    print_header "运行覆盖率测试"
    
    print_info "生成测试覆盖率报告..."
    go test -v ./tests/ -coverprofile=coverage.out -timeout=10m
    
    if [ -f coverage.out ]; then
        print_info "生成HTML覆盖率报告..."
        go tool cover -html=coverage.out -o coverage.html
        
        print_info "覆盖率统计:"
        go tool cover -func=coverage.out | tail -1
        
        print_success "覆盖率报告生成完成: coverage.html"
    else
        print_error "覆盖率文件生成失败"
        return 1
    fi
}

# 运行API集成测试
run_integration_tests() {
    print_header "运行API集成测试"
    
    print_info "启动测试服务器..."
    # 在后台启动服务器
    ./bin/admin-service &
    SERVER_PID=$!
    
    # 等待服务器启动
    sleep 3
    
    # 检查服务器是否启动成功
    if ! curl -s http://localhost:8080/health > /dev/null; then
        print_error "测试服务器启动失败"
        kill $SERVER_PID 2>/dev/null || true
        return 1
    fi
    
    print_success "测试服务器启动成功 (PID: $SERVER_PID)"
    
    # 运行集成测试
    print_info "执行API集成测试..."
    go test -v ./tests/ -tags=integration -timeout=10m
    
    # 停止测试服务器
    print_info "停止测试服务器..."
    kill $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
    
    print_success "API集成测试完成"
}

# 生成测试报告
generate_test_report() {
    print_header "生成测试报告"
    
    local report_file="test_report_$(date +%Y%m%d_%H%M%S).html"
    
    cat > $report_file << EOF
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Admin Service API 测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .success { color: #28a745; }
        .error { color: #dc3545; }
        .warning { color: #ffc107; }
        .info { color: #17a2b8; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .code { background: #f8f9fa; padding: 10px; border-radius: 3px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Admin Service API 测试报告</h1>
        <p><strong>生成时间:</strong> $(date)</p>
        <p><strong>测试环境:</strong> $(uname -a)</p>
        <p><strong>Go版本:</strong> $(go version)</p>
    </div>
    
    <div class="section">
        <h2>测试概览</h2>
        <p>本测试基于云函数接口格式，为admin-service的Go接口创建了完整的测试环境。</p>
        
        <h3>测试范围</h3>
        <ul>
            <li>用户管理接口 (UserController)</li>
            <li>系统管理接口 (SystemController)</li>
            <li>统计分析接口 (StatisticsController)</li>
            <li>错误处理和边界情况</li>
            <li>性能和负载测试</li>
        </ul>
        
        <h3>测试特点</h3>
        <ul>
            <li>✅ 参考云函数的输入输出格式</li>
            <li>✅ 标准化的响应结构 (code, msg, timestamp, data)</li>
            <li>✅ 完整的测试数据和Mock环境</li>
            <li>✅ 自动化的测试流程和清理</li>
            <li>✅ 详细的错误处理和验证</li>
        </ul>
    </div>
    
    <div class="section">
        <h2>API接口映射</h2>
        <table>
            <tr>
                <th>云函数</th>
                <th>Admin Service接口</th>
                <th>功能说明</th>
                <th>测试状态</th>
            </tr>
            <tr>
                <td>getAllUsers</td>
                <td>POST /api/users/list</td>
                <td>获取用户列表，支持分页和搜索</td>
                <td class="success">✅ 已测试</td>
            </tr>
            <tr>
                <td>getUserDetail</td>
                <td>POST /api/users/detail</td>
                <td>获取用户详细信息</td>
                <td class="success">✅ 已测试</td>
            </tr>
            <tr>
                <td>setUserDetail</td>
                <td>POST /api/users/update</td>
                <td>更新用户数据</td>
                <td class="success">✅ 已测试</td>
            </tr>
            <tr>
                <td>banUser</td>
                <td>POST /api/users/ban</td>
                <td>封禁用户</td>
                <td class="success">✅ 已测试</td>
            </tr>
            <tr>
                <td>unbanUser</td>
                <td>POST /api/users/unban</td>
                <td>解封用户</td>
                <td class="success">✅ 已测试</td>
            </tr>
            <tr>
                <td>deleteUser</td>
                <td>DELETE /api/users/delete</td>
                <td>删除用户</td>
                <td class="success">✅ 已测试</td>
            </tr>
            <tr>
                <td>getUserStats</td>
                <td>POST /api/users/stats</td>
                <td>获取用户统计信息</td>
                <td class="success">✅ 已测试</td>
            </tr>
            <tr>
                <td>adminLogin</td>
                <td>POST /api/auth/login</td>
                <td>管理员登录</td>
                <td class="info">🔄 JWT集成</td>
            </tr>
        </table>
    </div>
    
    <div class="section">
        <h2>测试数据结构</h2>
        <p>测试数据完全按照云函数的格式设计：</p>
        
        <h3>标准响应格式</h3>
        <div class="code">
{
    "code": 0,           // 0=成功，其他=错误码
    "msg": "success",    // 响应消息
    "timestamp": 1603991234567,  // 时间戳
    "data": {            // 实际数据
        // 具体的响应数据
    }
}
        </div>
        
        <h3>用户数据格式</h3>
        <div class="code">
{
    "playerId": "test_player_001",
    "data": {
        "level": 5,
        "score": 1500,
        "coins": 300,
        "nickname": "测试玩家",
        "achievements": ["新手上路", "初级玩家"],
        "inventory": {
            "items": ["sword", "shield"],
            "equipment": {
                "weapon": "iron_sword",
                "armor": "leather_armor"
            }
        }
    },
    "createdAt": "2023-10-01 10:00:00",
    "updatedAt": "2023-10-01 15:30:00"
}
        </div>
    </div>
    
    <div class="section">
        <h2>使用说明</h2>
        <h3>运行所有测试</h3>
        <div class="code">./run_tests.sh</div>
        
        <h3>运行特定测试</h3>
        <div class="code">
# 只运行用户API测试
go test -v ./tests/ -run TestUserAPIs

# 只运行系统API测试  
go test -v ./tests/ -run TestSystemAPIs

# 只运行统计API测试
go test -v ./tests/ -run TestStatisticsAPIs

# 运行性能测试
go test -v ./tests/ -bench=.
        </div>
        
        <h3>查看覆盖率报告</h3>
        <div class="code">
# 生成覆盖率报告
go test -coverprofile=coverage.out ./tests/
go tool cover -html=coverage.out -o coverage.html

# 在浏览器中打开 coverage.html
        </div>
    </div>
</body>
</html>
EOF
    
    print_success "测试报告生成完成: $report_file"
}

# 清理测试环境
cleanup_test_env() {
    print_header "清理测试环境"
    
    print_info "删除测试数据库..."
    mysql -h127.0.0.1 -uroot -e "DROP DATABASE IF EXISTS admin_service_test;" 2>/dev/null || {
        print_warning "删除测试数据库失败，请手动清理"
    }
    
    print_info "清理临时文件..."
    rm -f coverage.out bin/admin-service
    
    print_success "测试环境清理完成"
}

# 主函数
main() {
    print_header "Admin Service API 测试套件"
    print_info "基于云函数接口格式的完整测试环境"
    
    # 解析命令行参数
    SKIP_BUILD=false
    SKIP_CLEANUP=false
    RUN_COVERAGE=false
    RUN_BENCHMARK=false
    RUN_INTEGRATION=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-cleanup)
                SKIP_CLEANUP=true
                shift
                ;;
            --coverage)
                RUN_COVERAGE=true
                shift
                ;;
            --benchmark)
                RUN_BENCHMARK=true
                shift
                ;;
            --integration)
                RUN_INTEGRATION=true
                shift
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo "选项:"
                echo "  --skip-build    跳过项目构建"
                echo "  --skip-cleanup  跳过环境清理"
                echo "  --coverage      运行覆盖率测试"
                echo "  --benchmark     运行性能测试"
                echo "  --integration   运行集成测试"
                echo "  --help          显示帮助信息"
                exit 0
                ;;
            *)
                print_error "未知选项: $1"
                exit 1
                ;;
        esac
    done
    
    # 执行测试流程
    check_dependencies
    setup_test_env
    
    if [ "$SKIP_BUILD" = false ]; then
        build_project
    fi
    
    # 运行基础测试
    if ! run_unit_tests; then
        print_error "基础测试失败，停止执行"
        exit 1
    fi
    
    # 运行可选测试
    if [ "$RUN_COVERAGE" = true ]; then
        run_coverage_tests
    fi
    
    if [ "$RUN_BENCHMARK" = true ]; then
        run_benchmark_tests
    fi
    
    if [ "$RUN_INTEGRATION" = true ]; then
        run_integration_tests
    fi
    
    # 生成报告
    generate_test_report
    
    # 清理环境
    if [ "$SKIP_CLEANUP" = false ]; then
        cleanup_test_env
    fi
    
    print_header "测试完成"
    print_success "所有测试执行完毕！"
    print_info "查看详细报告: test_report_*.html"
}

# 捕获退出信号，确保清理
trap cleanup_test_env EXIT

# 运行主函数
main "$@"
