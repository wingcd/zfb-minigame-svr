# 计数器状态切换功能测试说明

## 功能概述
为计数器添加了基于 `isActive` 字段的状态切换功能，可以启用或禁用计数器。

## API 接口

### 切换计数器状态
- **路由**: `POST /counter/toggleStatus`
- **功能**: 切换计数器的启用/禁用状态
- **请求参数**:
```json
{
    "appId": "your-app-id",
    "key": "your-counter-key"
}
```

- **响应示例**:
```json
{
    "code": 0,
    "msg": "计数器已启用", // 或 "计数器已禁用"
    "timestamp": 1694764800000,
    "data": {
        "appId": "your-app-id",
        "key": "your-counter-key",
        "isActive": true // 或 false
    }
}
```

## 测试步骤

1. **创建计数器**（如果还没有）:
```bash
curl -X POST http://localhost:8080/counter/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "appId": "test-app",
    "key": "test-counter",
    "resetType": "permanent",
    "description": "测试计数器"
  }'
```

2. **切换计数器状态**:
```bash
curl -X POST http://localhost:8080/counter/toggleStatus \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "appId": "test-app",
    "key": "test-counter"
  }'
```

3. **验证状态变化**:
```bash
curl -X POST http://localhost:8080/counter/getList \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "appId": "test-app",
    "page": 1,
    "pageSize": 20,
    "groupByKey": true
  }'
```

## 实现细节

### 控制器方法
- 添加了 `ToggleCounterStatus()` 方法
- 包含完整的参数验证和错误处理
- 支持获取包括已禁用配置在内的所有计数器配置
- 返回友好的状态变更消息

### 数据库操作
- 直接使用 ORM 查询来获取计数器配置（包括已禁用的）
- 使用 `UpdateCounterConfig` 方法更新 `isActive` 字段
- 保持数据一致性

### 路由配置
- 添加了 `/counter/toggleStatus` 路由
- 继承现有的认证和权限中间件

## 错误处理
- 参数验证失败：返回 4001 错误码
- 计数器不存在：返回 4004 错误码  
- 数据库操作失败：返回 5001 错误码

## 注意事项
1. 该功能会切换计数器的 `isActive` 状态
2. 已禁用的计数器在 `GetCounterConfig` 查询中会被过滤掉
3. 但在状态切换功能中可以查询到所有状态的计数器
4. 切换操作是原子性的，要么成功要么失败
