# 云接口文件使用说明

## 📋 概述

本目录包含了所有管理后台系统的云接口文件，每个文件对应一个具体的功能接口。这些文件可以直接部署到云函数平台（如支付宝小程序云、微信小程序云等）。

## 🚀 快速开始

### 统一调用方式

**推荐使用统一入口**：
```javascript
// 部署 /cloud-api/callAPI.js 到云函数
// 调用任意接口：
my.cloud.callFunction({
  name: 'callAPI',
  data: {
    action: 'admin.getList',      // 接口名称
    params: {                     // 接口参数
      page: 1,
      pageSize: 20
    }
  }
})
```

### 直接调用方式

**部署单个接口文件**：
```javascript
// 部署 /cloud-api/admin/getAdminList.js 到云函数
my.cloud.callFunction({
  name: 'getAdminList',
  data: {
    page: 1,
    pageSize: 20
  }
})
```

## 📁 文件结构

```
cloud-api/
├── callAPI.js                 # 🌟 统一调用入口
├── getAPIList.js             # 获取接口列表
├── admin/                    # 管理员相关接口
│   ├── adminLogin.js         # 管理员登录
│   ├── verifyToken.js        # Token验证
│   ├── initAdmin.js          # 初始化系统
│   ├── getAdminList.js       # 获取管理员列表
│   ├── createAdmin.js        # 创建管理员
│   ├── updateAdmin.js        # 更新管理员
│   ├── deleteAdmin.js        # 删除管理员
│   ├── resetPassword.js      # 重置密码
│   ├── getRoleList.js        # 获取角色列表
│   └── getAllRoles.js        # 获取所有角色
├── app/                      # 应用管理接口
│   ├── appInit.js            # 初始化应用
│   ├── queryApp.js           # 查询应用
│   ├── getAllApps.js         # 获取应用列表
│   ├── updateApp.js          # 更新应用
│   └── deleteApp.js          # 删除应用
├── user/                     # 用户管理接口
│   ├── getAllUsers.js        # 获取用户列表
│   ├── banUser.js            # 封禁用户
│   ├── unbanUser.js          # 解封用户
│   └── deleteUser.js         # 删除用户
├── leaderboard/              # 排行榜管理接口
│   ├── getAllLeaderboards.js # 获取排行榜列表
│   ├── updateLeaderboard.js  # 更新排行榜
│   └── deleteLeaderboard.js  # 删除排行榜
└── stats/                    # 统计数据接口
    └── getDashboardStats.js  # 获取仪表板统计
```

## 🎯 接口映射

| 云接口文件 | 对应功能 | API名称 |
|------------|----------|---------|
| `callAPI.js` | 统一调用入口 | - |
| `getAPIList.js` | 获取接口列表 | `api.list` |
| `admin/adminLogin.js` | 管理员登录 | `auth.login` |
| `admin/verifyToken.js` | 验证Token | `auth.verify` |
| `admin/initAdmin.js` | 初始化系统 | `auth.init` |
| `admin/getAdminList.js` | 获取管理员列表 | `admin.getList` |
| `admin/createAdmin.js` | 创建管理员 | `admin.create` |
| `admin/updateAdmin.js` | 更新管理员 | `admin.update` |
| `admin/deleteAdmin.js` | 删除管理员 | `admin.delete` |
| `admin/resetPassword.js` | 重置密码 | `admin.resetPassword` |
| `admin/getRoleList.js` | 获取角色列表 | `role.getList` |
| `admin/getAllRoles.js` | 获取所有角色 | `role.getAll` |
| `app/appInit.js` | 初始化应用 | `app.init` |
| `app/queryApp.js` | 查询应用 | `app.query` |
| `app/getAllApps.js` | 获取应用列表 | `app.getAll` |
| `app/updateApp.js` | 更新应用 | `app.update` |
| `app/deleteApp.js` | 删除应用 | `app.delete` |
| `user/getAllUsers.js` | 获取用户列表 | `user.getAll` |
| `user/banUser.js` | 封禁用户 | `user.ban` |
| `user/unbanUser.js` | 解封用户 | `user.unban` |
| `user/deleteUser.js` | 删除用户 | `user.delete` |
| `leaderboard/getAllLeaderboards.js` | 获取排行榜列表 | `leaderboard.getAll` |
| `leaderboard/updateLeaderboard.js` | 更新排行榜 | `leaderboard.update` |
| `leaderboard/deleteLeaderboard.js` | 删除排行榜 | `leaderboard.delete` |
| `stats/getDashboardStats.js` | 获取仪表板统计 | `stats.dashboard` |

## 🔧 部署说明

### 方案一：统一入口部署（推荐）

1. 部署 `alipay-admin-sever` 整个目录到云函数
2. 部署 `callAPI.js` 作为入口函数
3. 前端统一调用 `callAPI` 函数

### 方案二：分别部署

1. 部署 `alipay-admin-sever` 整个目录到云函数
2. 根据需要部署具体的接口文件
3. 前端分别调用对应的函数

## 📝 示例代码

### 管理员登录
```javascript
my.cloud.callFunction({
  name: 'callAPI',
  data: {
    action: 'auth.login',
    params: {
      username: 'admin',
      password: '123456'
    }
  }
})
```

### 获取管理员列表
```javascript
my.cloud.callFunction({
  name: 'callAPI',
  data: {
    action: 'admin.getList',
    params: {
      page: 1,
      pageSize: 20,
      username: '搜索关键词'
    }
  }
})
```

### 创建应用
```javascript
my.cloud.callFunction({
  name: 'callAPI',
  data: {
    action: 'app.init',
    params: {
      appName: '我的小游戏',
      platform: 'alipay',
      channelAppId: 'your_app_id',
      channelAppKey: 'your_app_key'
    }
  }
})
```

## ⚠️ 注意事项

1. **依赖关系**：所有云接口文件都依赖 `alipay-admin-sever` 目录
2. **权限验证**：除了登录和初始化接口，其他接口都需要Token验证
3. **错误处理**：统一的错误码和响应格式
4. **数据库**：需要配置云数据库连接

## 🛠️ 开发说明

如果需要添加新的接口：

1. 在 `alipay-admin-sever` 中实现业务逻辑
2. 在 `index.js` 的 `AdminAPI` 中添加映射
3. 在 `cloud-api` 中创建对应的云接口文件
4. 更新本文档和接口映射表

## 📞 技术支持

如有问题请参考：
- [完整API文档](../doc/API_USAGE_GUIDE.md)
- [权限说明](../doc/SECURITY_GUIDE.md)
- [接口文档](../doc/API_INTERFACES.md) 