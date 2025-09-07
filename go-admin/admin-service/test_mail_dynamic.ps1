# 测试邮件系统动态表名修复
# PowerShell 测试脚本

$baseUrl = "http://localhost:8080"

# 测试数据
$testData = @{
    appId = "test-app"
    force = $true
} | ConvertTo-Json

Write-Host "测试邮件系统初始化 (使用动态表名)..." -ForegroundColor Green
Write-Host "请求数据: $testData" -ForegroundColor Yellow

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/mail/init" -Method POST -Body $testData -ContentType "application/json"
    Write-Host "响应:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
    
    if ($response.code -eq 2000) {
        Write-Host "✅ 邮件系统初始化成功!" -ForegroundColor Green
        Write-Host "动态表名应该为: minigame_test_app_mail" -ForegroundColor Cyan
    } else {
        Write-Host "❌ 初始化失败: $($response.msg)" -ForegroundColor Red
    }
}
catch {
    Write-Host "❌ 请求失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "错误详情: $($_.Exception)" -ForegroundColor Yellow
}

Write-Host "`n按任意键继续..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
