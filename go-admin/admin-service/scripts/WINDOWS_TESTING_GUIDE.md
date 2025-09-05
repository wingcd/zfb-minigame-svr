# Windows 测试环境使用指南

## 🚀 快速开始

本项目提供了三种Windows测试方式，选择最适合你的：

### 1. 批处理脚本 (.bat) - 兼容性最好
```cmd
# 双击运行或命令行执行
run_tests.bat

# 带参数运行
run_tests.bat --coverage
run_tests.bat --benchmark
run_tests.bat --help
```

### 2. PowerShell脚本 (.ps1) - 功能最强大 ⭐ 推荐
```powershell
# 基础运行
.\run_tests.ps1

# 高级用法
.\run_tests.ps1 -Coverage -Benchmark
.\run_tests.ps1 -SkipBuild -SkipCleanup
.\run_tests.ps1 -Help
```

### 3. 直接Go命令 - 最简单
```cmd
go test -v ./tests/ -timeout=10m
```

## 📋 环境要求

### 必需组件
- ✅ **Go 1.18+** - [下载地址](https://golang.org/dl/)
- ✅ **MySQL 5.7+** - [下载地址](https://dev.mysql.com/downloads/mysql/)
- ✅ **Git** - [下载地址](https://git-scm.com/download/win)

### 可选组件
- 🔧 **PowerShell 5.1+** (Windows 10自带)
- 🔧 **PowerShell Core 7+** (更好的体验)
- 🔧 **Visual Studio Code** (代码编辑)

## 🛠️ 环境配置

### 1. MySQL 配置
```sql
-- 创建测试用户（可选，默认使用root）
CREATE USER 'test_user'@'localhost' IDENTIFIED BY 'test_password';
GRANT ALL PRIVILEGES ON admin_service_test.* TO 'test_user'@'localhost';
FLUSH PRIVILEGES;
```

### 2. PowerShell 执行策略
如果遇到执行策略限制：
```powershell
# 方法1：设置当前用户策略
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 方法2：临时绕过（单次使用）
PowerShell -ExecutionPolicy Bypass -File .\run_tests.ps1

# 方法3：管理员设置（全局）
Set-ExecutionPolicy RemoteSigned
```

### 3. 环境变量（可选）
```cmd
set GO_ENV=test
set DB_HOST=127.0.0.1
set DB_PORT=3306
set DB_USER=root
set DB_PASSWORD=
set DB_NAME=admin_service_test
```

## 📊 测试选项对比

| 功能 | .bat | .ps1 | go test |
|------|------|------|---------|
| 基础测试 | ✅ | ✅ | ✅ |
| 覆盖率测试 | ✅ | ✅ | ✅ |
| 性能测试 | ✅ | ✅ | ✅ |
| 集成测试 | ✅ | ✅ | ❌ |
| 彩色输出 | ❌ | ✅ | ❌ |
| 参数化调用 | 基础 | 高级 | 手动 |
| 错误处理 | 基础 | 强大 | 基础 |
| 报告生成 | ✅ | ✅ | ❌ |
| 自动清理 | ✅ | ✅ | ❌ |

## 🎯 使用场景

### 开发调试
```powershell
# 快速测试（跳过构建和清理）
.\run_tests.ps1 -SkipBuild -SkipCleanup

# 只测试特定模块
go test -v ./tests/ -run TestUserAPIs
```

### CI/CD 集成
```cmd
# 批处理脚本适合CI环境
run_tests.bat --coverage --skip-cleanup

# 或使用PowerShell
powershell -ExecutionPolicy Bypass -File run_tests.ps1 -Coverage
```

### 性能分析
```powershell
# 完整的性能测试
.\run_tests.ps1 -Benchmark -Coverage

# 只运行性能测试
go test -v ./tests/ -bench=. -benchtime=30s
```

### 集成测试
```powershell
# 完整的集成测试
.\run_tests.ps1 -Integration

# 手动启动服务器进行测试
start bin\admin-service.exe
go test -v ./tests/ -tags=integration
```

## 🔧 故障排除

### 常见问题

#### 1. PowerShell 执行策略限制
```
错误: 无法加载文件 run_tests.ps1，因为在此系统上禁止运行脚本
```
**解决方案:**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### 2. MySQL 连接失败
```
错误: 无法连接到MySQL数据库
```
**解决方案:**
- 检查MySQL服务是否启动：`net start mysql`
- 检查端口是否正确：`netstat -an | findstr :3306`
- 验证用户名密码：`mysql -uroot -p`

#### 3. 端口占用
```
错误: 测试服务器启动失败
```
**解决方案:**
```cmd
# 查看端口占用
netstat -ano | findstr :8080

# 结束占用进程
taskkill /PID <进程ID> /F
```

#### 4. Go 模块问题
```
错误: go mod tidy 失败
```
**解决方案:**
```cmd
# 清理模块缓存
go clean -modcache

# 重新初始化模块
go mod init admin-service
go mod tidy
```

### 调试技巧

#### 1. 详细日志
```cmd
# 启用详细输出
set GO_ENV=test
set LOG_LEVEL=debug
go test -v ./tests/ -timeout=30m
```

#### 2. 单独测试
```cmd
# 测试特定函数
go test -v ./tests/ -run TestGetUserList

# 测试特定文件
go test -v ./tests/test_framework_test.go
```

#### 3. 内存和性能分析
```cmd
# 内存分析
go test -v ./tests/ -memprofile=mem.prof

# CPU分析
go test -v ./tests/ -cpuprofile=cpu.prof

# 查看分析结果
go tool pprof mem.prof
```

## 📈 性能优化

### 系统优化
1. **使用SSD硬盘** - 可提升50%+的测试速度
2. **增加内存** - 建议8GB+，16GB最佳
3. **关闭实时杀毒** - 临时关闭可避免文件扫描延迟
4. **使用本地MySQL** - 避免网络延迟

### 测试优化
```cmd
# 并行测试
go test -v ./tests/ -parallel 4

# 缓存测试结果
go test -v ./tests/ -count=1

# 跳过慢速测试
go test -v ./tests/ -short
```

## 📝 配置文件

### 测试配置 (test_config.json)
```json
{
  "database": {
    "host": "127.0.0.1",
    "port": 3306,
    "user": "root",
    "password": "",
    "database": "admin_service_test"
  },
  "server": {
    "port": 8080,
    "timeout": 30
  },
  "test": {
    "parallel": 4,
    "timeout": "10m",
    "verbose": true
  }
}
```

### 环境配置 (.env)
```env
GO_ENV=test
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=admin_service_test
SERVER_PORT=8080
LOG_LEVEL=info
```

## 🎉 最佳实践

### 1. 开发工作流
```cmd
# 1. 修改代码后快速测试
.\run_tests.ps1 -SkipBuild

# 2. 提交前完整测试
.\run_tests.ps1 -Coverage -Benchmark

# 3. 发布前集成测试
.\run_tests.ps1 -Integration -Coverage
```

### 2. 团队协作
- 使用相同的Go版本和依赖
- 统一的MySQL配置
- 共享测试配置文件
- 定期更新测试用例

### 3. 持续集成
```yaml
# GitHub Actions 示例
- name: Run Windows Tests
  run: |
    powershell -ExecutionPolicy Bypass -File run_tests.ps1 -Coverage
  shell: cmd
```

## 📚 相关文档

- [测试框架详细说明](tests/README.md)
- [API测试用例文档](API_TEST_SUMMARY.md)
- [云函数接口对照表](API_TEST_SUMMARY.md#接口映射对照表)

## 🆘 获取帮助

### 查看帮助
```cmd
# 批处理版本
run_tests.bat --help

# PowerShell版本
.\run_tests.ps1 -Help

# Go测试帮助
go test -h
```

### 社区支持
- GitHub Issues: 报告问题和建议
- 技术文档: 查看详细的技术说明
- 代码示例: 参考existing测试用例

---

**🎯 推荐使用PowerShell版本 (`run_tests.ps1`)，功能最完整，体验最佳！**
