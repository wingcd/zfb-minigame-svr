# 后端权限校验安全指南

## 🔒 安全概述

本系统实现了多层次的权限控制机制，确保数据安全和操作规范。前端权限控制只是用户体验层面的优化，**真正的安全保障在后端实现**。

## 🛡️ 权限体系架构

### 权限校验流程
```
客户端请求 → 权限中间件 → 业务逻辑 → 操作日志 → 返回结果
```

### 安全层级
1. **Token验证** - 验证用户身份
2. **权限检查** - 验证操作权限
3. **角色限制** - 特殊操作的角色要求
4. **操作审计** - 记录所有操作日志

## 🔑 权限系统设计

### 内置角色和权限

| 角色 | 角色代码 | 权限列表 | 说明 |
|------|----------|----------|------|
| 超级管理员 | `super_admin` | 所有权限 | 系统最高权限，不受限制 |
| 管理员 | `admin` | `app_manage`, `user_manage`, `leaderboard_manage`, `stats_view` | 大部分管理权限 |
| 运营人员 | `operator` | `user_manage`, `leaderboard_manage`, `stats_view` | 运营相关权限 |
| 查看者 | `viewer` | `stats_view` | 仅查看权限 |

### 权限定义

| 权限代码 | 权限名称 | 说明 |
|----------|----------|------|
| `admin_manage` | 管理员管理 | 创建、编辑、删除管理员账户 |
| `role_manage` | 角色管理 | 创建、编辑、删除角色和权限 |
| `app_manage` | 应用管理 | 管理小游戏应用 |
| `user_manage` | 用户管理 | 管理游戏用户 |
| `leaderboard_manage` | 排行榜管理 | 管理排行榜配置和数据 |
| `stats_view` | 统计查看 | 查看统计数据 |
| `system_config` | 系统配置 | 系统级配置（预留） |

## 🔐 接口权限映射

### 应用管理接口
| 接口路径 | 所需权限 | 额外限制 |
|----------|----------|----------|
| `POST /app/getAllApps` | `app_manage` | - |
| `POST /app/updateApp` | `app_manage` | - |
| `POST /app/deleteApp` | `app_manage` | 仅超级管理员 |

### 用户管理接口
| 接口路径 | 所需权限 | 额外限制 |
|----------|----------|----------|
| `POST /user/getAllUsers` | `user_manage` | - |
| `POST /user/banUser` | `user_manage` | - |
| `POST /user/unbanUser` | `user_manage` | - |
| `POST /user/deleteUser` | `user_manage` | 管理员或超级管理员 |

### 排行榜管理接口
| 接口路径 | 所需权限 | 额外限制 |
|----------|----------|----------|
| `POST /leaderboard/getAllLeaderboards` | `leaderboard_manage` | - |
| `POST /leaderboard/updateLeaderboard` | `leaderboard_manage` | - |
| `POST /leaderboard/deleteLeaderboard` | `leaderboard_manage` | 管理员或超级管理员 |

### 管理员接口
| 接口路径 | 所需权限 | 额外限制 |
|----------|----------|----------|
| `POST /admin/adminLogin` | 无需权限 | 安全处理 + 操作日志 |
| `POST /admin/verifyToken` | 无需权限 | 安全处理 + 操作日志 |
| `POST /admin/initAdmin` | 无需权限 | 系统初始化，特殊安全控制 |
| `POST /admin/getAdminList` | `admin_manage` | - |
| `POST /admin/createAdmin` | `admin_manage` | 不能创建比自己权限更高的管理员 |
| `POST /admin/updateAdmin` | `admin_manage` | 不能修改比自己权限更高的管理员 |
| `POST /admin/deleteAdmin` | `admin_manage` | 仅超级管理员，不能删除自己和最后一个超级管理员 |
| `POST /admin/resetPassword` | `admin_manage` | 不能重置比自己权限更高的管理员密码 |
| `POST /admin/getRoleList` | `role_manage` | - |
| `POST /admin/getAllRoles` | `role_manage` | - |

### 统计接口
| 接口路径 | 所需权限 | 额外限制 |
|----------|----------|----------|
| `POST /stats/getDashboardStats` | `stats_view` | - |

## 🛠️ 实现机制

### 1. 权限中间件 (`common/auth.js`)

#### 核心函数
- `checkPermission(event, requiredPermissions)` - 权限验证
- `requirePermission(handler, requiredPermissions)` - 权限装饰器
- `logOperation(adminInfo, action, resource, details)` - 操作日志

#### 使用方式
```javascript
const { requirePermission } = require("./common/auth");

// 原始业务函数
async function businessHandler(event, context) {
    // 业务逻辑
    // event.adminInfo 包含当前管理员信息
}

// 导出带权限校验的函数
exports.main = requirePermission(businessHandler, 'required_permission');
```

### 2. Token验证流程

1. **获取Token**: 从请求头 `authorization: Bearer <token>` 或参数 `token` 获取
2. **验证Token**: 查询数据库验证token有效性
3. **检查过期**: 验证token是否过期
4. **获取权限**: 查询管理员角色和权限列表
5. **权限判断**: 检查是否具有所需权限

### 3. 操作审计

所有敏感操作都会记录到 `admin_operation_logs` 表：

```javascript
{
    adminId: "管理员ID",
    username: "操作者用户名", 
    action: "操作类型（VIEW/CREATE/UPDATE/DELETE/BAN/UNBAN）",
    resource: "操作资源（APPS/USERS/LEADERBOARDS等）",
    details: "操作详情（JSON对象）",
    severity: "操作风险等级（LOW/MEDIUM/HIGH/CRITICAL）",
    createTime: "操作时间"
}
```

## ⚠️ 安全最佳实践

### 1. 权限分级
- **查看操作**: 基础权限即可
- **修改操作**: 需要对应管理权限
- **删除操作**: 需要高级角色（管理员或超级管理员）

### 2. 敏感操作保护
以下操作有额外的角色限制：
- 删除应用：仅超级管理员
- 删除用户：管理员或超级管理员
- 删除排行榜：管理员或超级管理员

### 3. 操作日志
- 所有操作都有详细的审计日志
- 删除等关键操作标记为 `CRITICAL` 级别
- 封禁用户等操作标记为 `HIGH` 级别

### 4. Token安全
- Token有过期时间（默认7天）
- Token存储在数据库中，支持服务端主动失效
- 支持多端登录管理

## 🚨 安全注意事项

### 1. 部署前检查
- [ ] 所有敏感接口都已添加权限校验
- [ ] 删除操作有适当的角色限制
- [ ] 操作日志记录完整
- [ ] Token验证机制正常工作

### 2. 运营监控
- 定期检查操作日志，特别是 `CRITICAL` 级别的操作
- 监控异常的权限尝试
- 定期轮换超级管理员密码

### 3. 代码审查
- 新增接口必须经过权限评估
- 确保前端权限控制与后端一致
- 避免在业务逻辑中硬编码权限判断

## 📋 权限测试清单

### 基础验证
- [ ] 未登录用户无法访问任何管理接口
- [ ] 过期token会被正确拒绝
- [ ] 无效token会被正确拒绝

### 权限验证
- [ ] 不同角色只能访问对应权限的接口
- [ ] 超级管理员可以访问所有接口
- [ ] 权限不足时返回正确的错误码

### 操作验证
- [ ] 删除操作有正确的角色限制
- [ ] 所有操作都有审计日志
- [ ] 敏感操作标记了正确的风险等级

## 🔧 故障排查

### 常见权限错误
1. **4001 - 缺少认证token**: 检查前端是否正确发送token
2. **4001 - 无效token**: token可能过期或被删除
3. **4003 - 权限不足**: 用户角色权限不够
4. **4003 - 角色限制**: 特殊操作需要更高角色

### 调试建议
1. 检查 `admin_users` 表中的用户状态和角色
2. 检查 `admin_roles` 表中的角色权限配置
3. 查看 `admin_operation_logs` 表的审计记录
4. 验证API请求中的token格式和内容

## 📚 相关文档
- [API接口文档](./API_INTERFACES.md)
- [部署指南](../game-web-admin/DEPLOYMENT.md)
- [前端权限控制](../game-web-admin/src/utils/auth.js) 