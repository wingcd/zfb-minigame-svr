# 计数器功能 API 文档

## 概述

计数器功能允许前端为对应的app增加一个key，所有玩家共同使用该计数器，可以调用接口增加传入次数，然后通过另外一个接口获取对应的key值，并支持设置自动清空时间（比如全服每日活动参与次数，全服挑战次数等）。

**注意：计数器是绑定到游戏(appId)的，所有玩家共享同一个计数器，不是每个玩家独立的计数器。**

**新增功能：现在支持点位参数(location)，一个计数器可以记录不同点位的值，可用于地区排序、服务器排行等功能。**

## 功能特点

- **灵活计数**: 支持任意key的计数器
- **点位支持**: 支持为同一计数器设置不同的点位，实现地区排序等功能
- **自动重置**: 支持5种重置类型
- **原子操作**: 增加操作是原子性的，确保并发安全
- **批量查询**: 支持获取单个点位或所有点位的计数器值
- **排行榜功能**: 提供便捷的地区排行榜API

## 云函数接口

### 1. incrementCounter - 增加计数器

**功能**: 增加指定key和点位的计数器值

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 小程序id |
| key | string | 是 | 计数器key |
| increment | number | 否 | 增加的数量，默认1 |
| location | string | 否 | 点位参数，用于地区排序等，默认为"default" |
| resetType | string | 否 | 重置类型：daily(每日)、weekly(每周)、monthly(每月)、custom(自定义)、permanent(永久) |
| resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |

**请求示例**:
```json
{
    "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
    "key": "daily_challenge",
    "increment": 1,
    "location": "beijing",
    "resetType": "daily"
}
```

**返回示例**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "key": "daily_challenge",
        "location": "beijing",
        "currentValue": 5,
        "resetTime": "2023-10-29 00:00:00",
        "wasReset": false
    }
}
```

### 2. getCounter - 获取计数器值

**功能**: 获取指定key的计数器所有点位的值

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 小程序id |
| key | string | 是 | 计数器key |

**请求示例**:
```json
{
    "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
    "key": "daily_challenge"
}
```

**返回示例**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "key": "daily_challenge",
        "locations": {
            "beijing": {
                "value": 5
            },
            "shanghai": {
                "value": 10
            },
            "default": {
                "value": 15
            }
        },
        "resetType": "daily",
        "resetValue": null,
        "resetTime": "2023-10-29 00:00:00",
        "timeToReset": 36000000,
        "description": "每日挑战计数器"
    }
}
```

### 3. createCounter - 创建计数器（管理员权限）

**功能**: 创建新的计数器

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 应用ID |
| key | string | 是 | 计数器key |
| resetType | string | 是 | 重置类型：daily、weekly、monthly、custom、permanent |
| resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |
| description | string | 否 | 计数器描述 |
| location | string | 否 | 点位参数，默认为"default" |

**请求示例**:
```json
{
    "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
    "key": "server_events",
    "resetType": "daily",
    "description": "全服每日活动次数",
    "location": "default"
}
```

## SDK 接口使用

### 基础使用

```typescript
// 增加计数器（默认点位）
await ZYSDK.counter.incrementCounter('daily_task', 1);

// 增加计数器（指定点位）
await ZYSDK.counter.incrementCounter('daily_task', 1, 'beijing');

// 获取指定点位的计数器值
const counter = await ZYSDK.counter.getCounter('daily_task', 'beijing');

// 获取所有点位的计数器值
const allCounters = await ZYSDK.counter.getCounter('daily_task');

// 获取地区排行榜
const ranking = await ZYSDK.counter.getLocationRanking('daily_task');
```

### 地区排行榜功能

SDK提供了便捷的地区排行榜API：

```typescript
// 获取地区排行榜
const ranking = await ZYSDK.counter.getLocationRanking('server_activity');

console.log(ranking.data);
// 输出:
// [
//   { key: 'server_activity', location: 'shanghai', value: 150, rank: 1 },
//   { key: 'server_activity', location: 'beijing', value: 120, rank: 2 },
//   { key: 'server_activity', location: 'guangzhou', value: 100, rank: 3 }
// ]
```

## 使用场景示例

### 1. 地区活动竞赛

```typescript
// 各地区参与活动
await ZYSDK.counter.incrementCounter('region_event', 5, 'beijing');
await ZYSDK.counter.incrementCounter('region_event', 3, 'shanghai');
await ZYSDK.counter.incrementCounter('region_event', 8, 'guangzhou');

// 获取地区排行榜
const ranking = await ZYSDK.counter.getLocationRanking('region_event');
console.log('地区活动排行榜:', ranking.data);
```

### 2. 服务器统计

```typescript
// 不同服务器的在线玩家数量统计
await ZYSDK.counter.incrementCounter('online_players', 1, 'server_01');
await ZYSDK.counter.incrementCounter('online_players', 1, 'server_02');

// 获取所有服务器统计
const serverStats = await ZYSDK.counter.getCounter('online_players');
console.log('服务器统计:', serverStats.data);
```

### 3. 传统全服计数器（向后兼容）

```typescript
// 不传location参数，使用默认点位
await ZYSDK.counter.incrementCounter('global_events', 1);

// 获取全服统计
const globalCounter = await ZYSDK.counter.getCounter('global_events', 'default');
```

## 重置类型说明

### daily（每日重置）
- **描述**: 每天 00:00:00 重置计数器
- **适用场景**: 每日任务、每日活动次数等
- **重置时间**: 每天 00:00:00

### weekly（每周重置）
- **描述**: 每周一 00:00:00 重置计数器
- **适用场景**: 周常任务、周赛等
- **重置时间**: 每周一 00:00:00

### monthly（每月重置）
- **描述**: 每月1号 00:00:00 重置计数器
- **适用场景**: 月度活动、月度统计等
- **重置时间**: 每月1号 00:00:00

### custom（自定义重置）
- **描述**: 根据resetValue参数指定的小时数后重置
- **适用场景**: 限时活动、自定义周期任务等
- **配置参数**: 需要配合resetValue参数使用

### permanent（永久计数器）
- **描述**: 计数器永远不会自动重置
- **适用场景**: 总积分、历史统计等
- **特点**: 数据会一直累积

## 数据存储结构

计数器数据存储在MongoDB集合中，集合名格式为：`counter_{appId}`

数据结构：
```json
{
    "_id": "记录ID",
    "appId": "应用ID",
    "key": "计数器key",
    "location": "点位标识",
    "value": 100,
    "resetType": "daily",
    "resetValue": null,
    "resetTime": "2023-10-29 00:00:00",
    "description": "计数器描述",
    "gmtCreate": "2023-10-28 10:30:00",
    "gmtModify": "2023-10-28 15:45:00"
}
```

## 注意事项

1. **点位参数**: location参数用于区分同一计数器的不同点位，如地区、服务器等
2. **向后兼容**: 不传location参数时默认使用"default"点位，保持向后兼容
3. **自动创建**: 当访问不存在的location时，系统会自动基于默认配置创建新的location记录
4. **原子操作**: 所有计数器操作都是原子性的，确保并发安全
5. **重置机制**: 每个location的重置时间是独立计算的
6. **排行榜**: 可以使用getLocationRanking方法快速获取地区排行榜 