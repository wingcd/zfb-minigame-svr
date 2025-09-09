@echo off
echo ========================================
echo   Game Service API Tester 启动脚本
echo ========================================
echo.
echo 正在启动测试工具...
echo 请确保 game-service 运行在 http://localhost:8081
echo.
echo 启动后请访问: http://localhost:8082
echo.
echo 按任意键启动，或Ctrl+C取消...
pause >nul

go run main.go
