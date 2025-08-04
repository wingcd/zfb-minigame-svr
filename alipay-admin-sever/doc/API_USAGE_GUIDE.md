# 管理后台API使用指南

## 🚀 动态API调用入口

现在您可以通过**单一统一接口**调用所有管理后台功能！

### 📍 统一调用入口

**接口地址**: `POST /admin/callAPI`

**请求格式**:
```json
{
  "action": "接口名称",
  "params": {
    // 接口参数
  }
}
```

## 📋 可用接口列表

### 🔐 管理员管理

| 接口名称 | 说明 | 示例参数 |
|----------|------|----------|
| `admin.getList` | 获取管理员列表 | `{"page": 1, "pageSize": 20}` |
| `admin.create` | 创建管理员 | `{"username": "newadmin", "password": "123456", "role": "admin"}` |
| `admin.update` | 更新管理员信息 | `{"id": "admin_id", "nickname": "新昵称"}` |
| `admin.delete` | 删除管理员 | `{"id": "admin_id"}` |
| `admin.resetPassword` | 重置管理员密码 | `{"id": "admin_id", "newPassword": "newpass123"}` |

### 👥 角色管理

| 接口名称 | 说明 | 示例参数 |
|----------|------|----------|
| `role.getList` | 获取角色列表 | `{"page": 1, "pageSize": 20}` |
| `role.getAll` | 获取所有角色 | `{}` |

### 🔑 认证管理

| 接口名称 | 说明 | 示例参数 |
|----------|------|----------|
| `auth.login` | 管理员登录 | `{"username": "admin", "password": "123456"}` |
| `auth.verify` | 验证Token | `{"token": "your_token"}` |
| `auth.init` | 初始化系统 | `{"force": false}` |

### 📱 应用管理

| 接口名称 | 说明 | 示例参数 |
|----------|------|----------|
| `app.init` | 初始化应用 | `{"appName": "小游戏", "platform": "wechat", "channelAppId": "wx123", "channelAppKey": "key123"}` |
| `app.query` | 查询应用详情 | `{"appId": "app123"}` 或 `{"appName": "小游戏"}` |
| `app.getAll` | 获取应用列表 | `{"page": 1, "pageSize": 20}` |
| `app.update` | 更新应用信息 | `{"appId": "app123", "appName": "新名称"}` |
| `app.delete` | 删除应用 | `{"appId": "app123", "force": true}` |

### 👤 用户管理

| 接口名称 | 说明 | 示例参数 |
|----------|------|----------|
| `user.getAll` | 获取用户列表 | `{"appId": "app123", "page": 1}` |
| `user.ban` | 封禁用户 | `{"appId": "app123", "playerId": "player001", "reason": "违规"}` |
| `user.unban` | 解封用户 | `{"appId": "app123", "playerId": "player001"}` |
| `user.delete` | 删除用户 | `{"appId": "app123", "playerId": "player001"}` |

### 🏆 排行榜管理

| 接口名称 | 说明 | 示例参数 |
|----------|------|----------|
| `leaderboard.getAll` | 获取排行榜列表 | `{"appId": "app123", "page": 1}` |
| `leaderboard.update` | 更新排行榜配置 | `{"appId": "app123", "leaderboardType": "score", "name": "新名称"}` |
| `leaderboard.delete` | 删除排行榜 | `{"appId": "app123", "leaderboardType": "score"}` |

### 📊 统计数据

| 接口名称 | 说明 | 示例参数 |
|----------|------|----------|
| `stats.dashboard` | 获取仪表板统计 | `{"timeRange": "week"}` |

## 💡 使用示例

### 1. 获取管理员列表

```bash
curl -X POST https://your-domain/admin/callAPI \
  -H "Content-Type: application/json" \
  -H "authorization: Bearer your_token" \
  -d '{
    "action": "admin.getList",
    "params": {
      "page": 1,
      "pageSize": 10,
      "username": "admin"
    }
  }'
```

### 2. 创建新管理员

```bash
curl -X POST https://your-domain/admin/callAPI \
  -H "Content-Type: application/json" \
  -H "authorization: Bearer your_token" \
  -d '{
    "action": "admin.create",
    "params": {
      "username": "newadmin",
      "password": "123456",
      "nickname": "新管理员",
      "role": "admin",
      "email": "admin@example.com"
    }
  }'
```

### 3. 封禁用户

```bash
curl -X POST https://your-domain/admin/callAPI \
  -H "Content-Type: application/json" \
  -H "authorization: Bearer your_token" \
  -d '{
    "action": "user.ban",
    "params": {
      "appId": "your_app_id",
      "playerId": "player001",
      "reason": "违规行为",
      "duration": 24
    }
  }'
```

### 4. 获取可用接口列表

```bash
curl -X POST https://your-domain/admin/callAPI \
  -H "Content-Type: application/json" \
  -d '{
    "action": "api.list"
  }'
```

### 5. 初始化新应用

```bash
curl -X POST https://your-domain/admin/callAPI \
  -H "Content-Type: application/json" \
  -H "authorization: Bearer your_token" \
  -d '{
    "action": "app.init",
    "params": {
      "appName": "我的小游戏",
      "platform": "wechat",
      "channelAppId": "wx1234567890",
      "channelAppKey": "your_app_secret",
      "force": false
    }
  }'
```

### 6. 查询应用详情

```bash
curl -X POST https://your-domain/admin/callAPI \
  -H "Content-Type: application/json" \
  -H "authorization: Bearer your_token" \
  -d '{
    "action": "app.query",
    "params": {
      "appId": "your_app_id"
    }
  }'
```

### 7. 封禁用户

```bash
curl -X POST https://your-domain/admin/callAPI \
  -H "Content-Type: application/json" \
  -H "authorization: Bearer your_token" \
  -d '{
    "action": "user.ban",
    "params": {
      "appId": "your_app_id",
      "playerId": "player001",
      "reason": "违规行为",
      "duration": 24
    }
  }'
```

## 🔒 权限说明

### 权限要求
- 所有接口（除 `auth.login`, `auth.verify`, `auth.init`）都需要有效的管理员Token
- 不同接口需要不同的权限级别
- 详见 [安全指南](./SECURITY_GUIDE.md)

### Token使用
```javascript
// 前端示例
const token = localStorage.getItem('admin_token');

fetch('/admin/callAPI', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    action: 'admin.getList',
    params: { page: 1, pageSize: 20 }
  })
});
```

## 📝 响应格式

### 成功响应
```json
{
  "code": 0,
  "msg": "success",
  "timestamp": 1698765432000,
  "calledAPI": "admin.getList",
  "callTime": 1698765432100,
  "data": {
    // 接口返回的具体数据
  }
}
```

### 错误响应
```json
{
  "code": 4004,
  "msg": "API接口 'invalid.api' 不存在",
  "timestamp": 1698765432000,
  "data": {
    "availableAPIs": ["admin.getList", "admin.create", ...],
    "categories": {
      "管理员管理": ["admin.getList", "admin.create", ...],
      // ... 其他分类
    }
  }
}
```

## 🎯 优势特点

### ✅ 统一入口
- 一个接口调用所有管理功能
- 统一的参数格式和响应结构
- 便于维护和扩展

### ✅ 类型安全
- 明确的接口名称和参数结构
- 完整的错误提示和可用接口列表
- 自动参数验证

### ✅ 权限保护
- 每个接口都有对应的权限验证
- 操作审计和日志记录
- 安全的Token管理

### ✅ 开发友好
- 清晰的文档和示例
- 一致的错误处理
- 便于测试和调试

## 🔧 开发提示

### JavaScript/TypeScript 封装示例

```javascript
class AdminAPI {
  constructor(baseURL, token) {
    this.baseURL = baseURL;
    this.token = token;
  }

  async call(action, params = {}) {
    const response = await fetch(`${this.baseURL}/admin/callAPI`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'authorization': `Bearer ${this.token}`
      },
      body: JSON.stringify({ action, params })
    });
    
    return await response.json();
  }

  // 管理员相关
  async getAdminList(params) { return this.call('admin.getList', params); }
  async createAdmin(params) { return this.call('admin.create', params); }
  async updateAdmin(params) { return this.call('admin.update', params); }
  async deleteAdmin(params) { return this.call('admin.delete', params); }

  // 用户相关
  async getUserList(params) { return this.call('user.getAll', params); }
  async banUser(params) { return this.call('user.ban', params); }
  async unbanUser(params) { return this.call('user.unban', params); }

  // 获取可用接口
  async getAPIList() { return this.call('api.list'); }
}

// 使用示例
const api = new AdminAPI('https://your-domain', 'your_token');
const adminList = await api.getAdminList({ page: 1, pageSize: 20 });
```

## 📚 相关文档

- [API接口文档](./API_INTERFACES.md) - 详细的接口参数说明
- [安全指南](./SECURITY_GUIDE.md) - 权限和安全配置
- [部署指南](../game-web-admin/DEPLOYMENT.md) - 部署和配置说明 