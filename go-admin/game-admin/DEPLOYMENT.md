# 部署指南

## 快速启动

### 1. 安装依赖
```bash
cd game-web-admin
npm install
```

### 2. 启动开发服务器
```bash
npm run dev
```

访问 `http://localhost:5173` 即可查看管理后台。

## 云函数配置

### 1. 配置API地址
编辑 `src/config/index.js`，修改 `baseURL` 为您的云函数部署地址：

```javascript
export default {
  api: {
    baseURL: 'https://your-cloud-function-domain.com',
    timeout: 10000
  }
}
```

### 2. 云函数路由映射
确保您的云函数网关配置了以下路由映射：

```
POST /user/login → alipay-server/user/login.js
POST /user/login.wx → alipay-server/user/login.wx.js
POST /user/getData → alipay-server/user/getData.js
POST /user/saveData → alipay-server/user/saveData.js

POST /app/appInit → alipay-server/app/appInit.js
POST /app/queryApp → alipay-server/app/queryApp.js

POST /leaderboard/commitScore → alipay-server/leaderboard/commitScore.js
POST /leaderboard/queryScore → alipay-server/leaderboard/queryScore.js
POST /leaderboard/getLeaderboardTopRank → alipay-server/leaderboard/getLeaderboardTopRank.js
POST /leaderboard/deleteScore → alipay-server/leaderboard/deleteScore.js
```

### 3. CORS 配置
确保云函数支持跨域请求，在响应头中添加：
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, authorization
```

## 生产部署

### 1. 构建项目
```bash
npm run build
```

### 2. 部署静态文件
将 `dist` 文件夹中的内容部署到您的静态文件服务器或CDN。

### 3. 配置反向代理（可选）
如果需要避免跨域问题，可以配置 Nginx 反向代理：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        root /path/to/dist;
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass https://your-cloud-function-domain.com;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 测试数据

为了帮助您快速测试系统功能，以下是一些示例API调用：

### 创建应用
```bash
curl -X POST "https://your-domain.com/app/appInit" \
  -H "Content-Type: application/json" \
  -d '{
    "appName": "测试小游戏",
    "platform": "wechat",
    "appId": "wx1234567890",
    "appKey": "your-app-key"
  }'
```

### 用户登录
```bash
curl -X POST "https://your-domain.com/user/login" \
  -H "Content-Type: application/json" \
  -d '{
    "appId": "your-app-id",
    "code": "test_player_001"
  }'
```

### 提交分数
```bash
curl -X POST "https://your-domain.com/leaderboard/commitScore" \
  -H "Content-Type: application/json" \
  -d '{
    "appId": "your-app-id",
    "playerId": "test_player_001",
    "type": "easy",
    "score": 1000,
    "playerInfo": {
      "name": "测试玩家",
      "avatar": "https://example.com/avatar.jpg"
    }
  }'
```

## 常见问题

### Q: 页面显示空白或加载失败
A: 检查以下几点：
1. 确认云函数正常运行
2. 检查 API 地址配置是否正确
3. 查看浏览器控制台是否有错误信息
4. 确认 CORS 配置正确

### Q: 数据不显示或接口调用失败
A: 
1. 打开浏览器开发者工具，查看 Network 面板
2. 检查 API 请求的状态码和响应内容
3. 确认云函数返回的数据格式符合前端预期

### Q: 图表不显示
A: 
1. 确认 ECharts 依赖已正确安装
2. 检查容器元素是否有有效的宽高
3. 查看控制台是否有 ECharts 相关错误

## 支持

如遇到问题，请：
1. 查看浏览器控制台错误信息
2. 检查网络请求状态
3. 确认云函数日志
4. 参考项目 README.md 文档 