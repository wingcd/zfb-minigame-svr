@echo off
echo === 权限控制测试 ===
echo.

rem 启动服务器（后台运行）
echo 1. 启动服务器...
start /B admin-service.exe
timeout /t 3 > nul

rem 测试1: 登录获取token
echo 2. 测试登录获取token...
curl -s -X POST http://localhost:8080/admin/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\": \"admin\", \"password\": \"admin123\"}" > login_response.txt

type login_response.txt
echo.

rem 提取token（简化处理）
for /f "tokens=2 delims=:" %%i in ('findstr "token" login_response.txt') do (
    set TOKEN_PART=%%i
)

echo 获取到token信息
echo.

rem 测试2: 使用token访问需要权限的接口
echo 3. 测试访问应用管理接口（需要app_manage权限）...
curl -s -X POST http://localhost:8080/app/getAll ^
  -H "Authorization: Bearer %TOKEN_PART%" ^
  -H "Content-Type: application/json" ^
  -d "{}" > app_response.txt

type app_response.txt
echo.

rem 检查是否返回权限错误
findstr "权限不足" app_response.txt > nul
if %errorlevel% == 0 (
    echo ✅ 权限控制正常工作 - 返回了权限不足错误
) else (
    echo ❓ 权限控制状态: 检查响应内容
)

echo.

rem 测试3: 访问统计接口
echo 4. 测试访问统计接口（需要stats_view权限）...
curl -s -X POST http://localhost:8080/stat/dashboard ^
  -H "Authorization: Bearer %TOKEN_PART%" ^
  -H "Content-Type: application/json" ^
  -d "{}" > stats_response.txt

type stats_response.txt
echo.

rem 清理
echo 5. 清理资源...
taskkill /F /IM admin-service.exe > nul 2>&1
del login_response.txt app_response.txt stats_response.txt > nul 2>&1
echo ✅ 测试完成

pause
