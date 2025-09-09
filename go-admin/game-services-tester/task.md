这是game-service的服务器测试

需要有简单的web服务器，跑一个web页面，
需要测试所有的游戏接口（除了平台相关的接口，比如微信）。需要模拟登录等
用ts接口，接口在：..\..\zy-sdk里面

✅ 已完成功能：
- ✅ 简单的Web服务器 (Go实现)
- ✅ 完整的测试页面 (HTML + CSS + JavaScript)
- ✅ 所有游戏接口测试（对齐zy-sdk）
- ✅ 模拟登录功能
- ✅ API签名生成
- ✅ 批量测试功能
- ✅ 实时结果显示
- ✅ 美观的用户界面

启动方法：
1. 确保game-service运行在 http://localhost:8081
2. 运行: go run main.go 或使用 start.bat/start.sh
3. 访问: http://localhost:8082