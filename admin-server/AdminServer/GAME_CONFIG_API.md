# 游戏配置功能 API 文档

## 概述

游戏配置功能允许开发者为游戏应用创建远程配置，支持全局配置和版本配置。版本配置优先于全局配置，实现灵活的配置管理。

## 功能特点

- **版本优先**: 支持版本配置，优先于全局配置
- **类型支持**: 支持字符串、数字、布尔值、对象、数组等多种数据类型
- **动态开关**: 支持配置的启用/禁用
- **管理权限**: 需要管理员权限才能操作配置
- **客户端友好**: 提供简洁的客户端获取接口

## 云函数接口

### 1. createGameConfig - 创建游戏配置

**功能**: 创建新的游戏配置

**权限**: 需要 `app_manage` 权限

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 应用ID |
| configKey | string | 是 | 配置键名 |
| configValue | any | 是 | 配置值 |
| version | string | 否 | 游戏版本（为空时为全局配置） |
| description | string | 否 | 配置描述 |
| configType | string | 否 | 配置类型 (string/number/boolean/object/array) |
| isActive | boolean | 否 | 是否激活，默认true |

**请求示例**:
```json
{
    "appId": "test_game_001",
    "configKey": "max_level",
    "configValue": 100,
    "version": "1.0.0",
    "description": "最大关卡数",
    "configType": "number"
}
```

**返回示例**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "id": "config_id_123456",
        "appId": "test_game_001",
        "configKey": "max_level",
        "configValue": 100,
        "version": "1.0.0",
        "description": "最大关卡数",
        "configType": "number",
        "isActive": true,
        "createTime": "2023-10-01 10:00:00",
        "updateTime": "2023-10-01 10:00:00"
    }
}
```

### 2. getGameConfig - 获取游戏配置（客户端使用）

**功能**: 获取游戏配置，优先返回版本配置

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 应用ID |
| version | string | 否 | 游戏版本 |
| configKey | string | 否 | 特定配置键名（为空时返回所有配置） |

**请求示例**:
```json
{
    "appId": "test_game_001",
    "version": "1.0.0",
    "configKey": "max_level"
}
```

**返回示例**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "configs": {
            "max_level": {
                "value": 100,
                "type": "number",
                "source": "version",
                "version": "1.0.0",
                "description": "最大关卡数"
            },
            "game_name": {
                "value": "超级游戏",
                "type": "string",
                "source": "global",
                "version": null,
                "description": "游戏名称"
            }
        }
    }
}
```

### 3. getGameConfigList - 获取游戏配置列表（管理后台使用）

**功能**: 获取游戏配置列表

**权限**: 需要 `app_manage` 权限

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 应用ID |
| version | string | 否 | 过滤特定版本（为空时显示所有版本） |
| configKey | string | 否 | 过滤特定配置键名 |
| page | number | 否 | 页码，默认1 |
| pageSize | number | 否 | 每页数量，默认20 |

**请求示例**:
```json
{
    "appId": "test_game_001",
    "page": 1,
    "pageSize": 20
}
```

**返回示例**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "list": [
            {
                "id": "config_id_123456",
                "appId": "test_game_001",
                "configKey": "max_level",
                "configValue": 100,
                "version": "1.0.0",
                "description": "最大关卡数",
                "configType": "number",
                "isActive": true,
                "createTime": "2023-10-01 10:00:00",
                "updateTime": "2023-10-01 10:00:00"
            }
        ],
        "total": 10,
        "page": 1,
        "pageSize": 20,
        "versions": ["1.0.0", "1.1.0", "2.0.0"]
    }
}
```

### 4. updateGameConfig - 更新游戏配置

**功能**: 更新游戏配置

**权限**: 需要 `app_manage` 权限

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| id | string | 是 | 配置ID |
| configValue | any | 否 | 配置值 |
| description | string | 否 | 配置描述 |
| configType | string | 否 | 配置类型 |
| isActive | boolean | 否 | 是否激活 |

**请求示例**:
```json
{
    "id": "config_id_123456",
    "configValue": 120,
    "description": "更新后的最大关卡数",
    "isActive": true
}
```

**返回示例**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "id": "config_id_123456",
        "updated": true
    }
}
```

### 5. deleteGameConfig - 删除游戏配置

**功能**: 删除游戏配置

**权限**: 需要 `app_manage` 权限

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| id | string | 是 | 配置ID |

**请求示例**:
```json
{
    "id": "config_id_123456"
}
```

**返回示例**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "id": "config_id_123456",
        "deleted": true
    }
}
```

## 数据库结构

### game_config 集合

```javascript
{
    "_id": "config_id_123456",
    "appId": "test_game_001",
    "configKey": "max_level",
    "configValue": 100,
    "version": "1.0.0", // 可选，为空表示全局配置
    "description": "最大关卡数",
    "configType": "number",
    "isActive": true,
    "createTime": "2023-10-01 10:00:00",
    "updateTime": "2023-10-01 10:00:00"
}
```

## 使用场景

### 1. 全局配置
适用于所有版本都需要的配置：
```json
{
    "appId": "my_game",
    "configKey": "server_url",
    "configValue": "https://api.mygame.com",
    "configType": "string",
    "description": "服务器地址"
}
```

### 2. 版本配置
适用于特定版本的配置：
```json
{
    "appId": "my_game",
    "configKey": "max_level",
    "configValue": 50,
    "version": "1.0.0",
    "configType": "number",
    "description": "1.0.0版本最大关卡数"
}
```

### 3. 客户端使用示例

```javascript
// 获取特定版本的配置
const response = await api.post('/gameConfig/get', {
    appId: 'my_game',
    version: '1.0.0'
});

// 配置会按优先级返回：版本配置 > 全局配置
const configs = response.data.configs;
const maxLevel = configs.max_level?.value || 30; // 默认值
const serverUrl = configs.server_url?.value;
```

## 错误码

- **4001**: 参数错误
- **4002**: 配置已存在
- **4003**: 权限不足
- **4004**: 应用/配置不存在
- **5001**: 服务器内部错误

## 注意事项

1. **版本优先级**: 如果同时存在全局配置和版本配置，版本配置优先
2. **配置键名**: 建议使用有意义的键名，如 `max_level`、`server_url` 等
3. **配置类型**: 确保配置值与配置类型匹配
4. **权限控制**: 只有具有 `app_manage` 权限的管理员才能操作配置
5. **客户端缓存**: 建议客户端对配置进行适当缓存，减少请求频率 