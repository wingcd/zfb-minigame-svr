# 用户管理模块

现在 go-admin/admin-service 已经包含完整的用户管理功能！

## 新增的API接口

### 1. 获取用户列表
```http
GET /api/user-management/users?appId=your_app_id&page=1&pageSize=10&keyword=player123&status=active
Authorization: Bearer <token>
```

### 2. 获取用户详情
```http
GET /api/user-management/user/detail?appId=your_app_id&playerId=player123
Authorization: Bearer <token>
```

### 3. 更新用户数据
```http
PUT /api/user-management/user/data
Authorization: Bearer <token>
Content-Type: application/json

{
  "appId": "your_app_id",
  "playerId": "player123",
  "data": {
    "level": 10,
    "score": 1000,
    "coins": 500
  }
}
```

### 4. 封禁用户
```http
POST /api/user-management/user/ban
Authorization: Bearer <token>
Content-Type: application/json

{
  "appId": "your_app_id",
  "playerId": "player123",
  "banType": "temporary",
  "banReason": "违规行为",
  "banHours": 24
}
```

### 5. 解封用户
```http
POST /api/user-management/user/unban
Authorization: Bearer <token>
Content-Type: application/json

{
  "appId": "your_app_id",
  "playerId": "player123",
  "unbanReason": "申诉通过"
}
```

### 6. 删除用户（超级管理员专用）
```http
DELETE /api/user-management/user/delete?appId=your_app_id&playerId=player123
Authorization: Bearer <token>
```

### 7. 获取用户统计
```http
GET /api/user-management/user/stats?appId=your_app_id&playerId=player123
Authorization: Bearer <token>
```

### 8. 获取注册统计
```http
GET /api/user-management/stats/registration?appId=your_app_id&days=7
Authorization: Bearer <token>
```

## 功能特性

### ✅ 完整的用户管理功能
- [x] 用户列表查询（支持分页、搜索、状态过滤）
- [x] 用户详细信息查看
- [x] 用户游戏数据编辑
- [x] 用户封禁/解封管理
- [x] 用户数据删除（权限控制）
- [x] 用户统计信息查看
- [x] 用户注册统计分析

### ✅ 安全功能
- [x] JWT权限验证
- [x] 管理员角色权限控制
- [x] 封禁状态检查
- [x] 操作日志记录
- [x] 数据完整性保护

### ✅ 高级功能
- [x] 自动解封机制（临时封禁过期）
- [x] 批量数据操作（事务保护）
- [x] 多维度数据统计
- [x] JSON数据解析和验证

## 数据库支持

新增数据表：
- `user_ban_records` - 用户封禁记录表
- 支持动态用户表 `user_{app_id}`
- 关联排行榜、邮件、计数器数据

## 权限说明

| 操作 | 普通管理员 | 超级管理员 |
|------|-----------|-----------|
| 查看用户列表 | ✅ | ✅ |
| 查看用户详情 | ✅ | ✅ |
| 编辑用户数据 | ✅ | ✅ |
| 封禁/解封用户 | ✅ | ✅ |
| 删除用户 | ❌ | ✅ |
| 查看统计数据 | ✅ | ✅ |

## 使用说明

1. **数据库初始化**：系统会自动创建 `user_ban_records` 表
2. **权限验证**：所有接口都需要 JWT Token 验证
3. **应用隔离**：不同 appId 的用户数据完全隔离
4. **安全删除**：删除用户会同时清理相关数据（排行榜、邮件等）

这样就完成了从 admin-server（Node.js）到 go-admin/admin-service（Go）的用户管理模块迁移！ 