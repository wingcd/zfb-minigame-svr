# 计数器功能 API 文档

## 概述

计数器功能允许前端为对应的app增加一个key，所有玩家共同使用该计数器，可以调用接口增加传入次数，然后通过另外一个接口获取对应的key值，并支持设置自动清空时间（比如全服每日活动参与次数，全服挑战次数等）。

**注意：计数器是绑定到游戏(appId)的，所有玩家共享同一个计数器，不是每个玩家独立的计数器。**

## 功能特点

- **灵活计数**: 支持任意key的计数器
- **自动重置**: 支持5种重置类型
- **原子操作**: 增加操作是原子性的，确保并发安全
- **批量查询**: 支持获取单个或所有计数器

## 云函数接口

### 1. incrementCounter - 增加计数器

**功能**: 增加指定key的计数器值

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 小程序id |
| key | string | 是 | 计数器key |
| increment | number | 否 | 增加的数量，默认1 |
| resetType | string | 否 | 重置类型：daily(每日)、weekly(每周)、monthly(每月)、custom(自定义)、permanent(永久) |
| resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |

**请求示例**:
```json
{
    "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
    "key": "daily_challenge",
    "increment": 1,
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
        "currentValue": 5,
        "resetTime": "2023-10-29 00:00:00",
        "wasReset": false
    }
}
```

### 2. getCounter - 获取计数器值

**功能**: 获取指定key的计数器值，或获取游戏所有计数器

**参数**:

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| appId | string | 是 | 小程序id |
| key | string | 否 | 计数器key，不传则获取该游戏所有计数器 |

**请求示例**:
```json
{
    "appId": "6a5f86e9-d59b-4a2a-a63b-c06c772bcee9",
    "key": "daily_challenge"
}
```

**返回示例（单个计数器）**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": {
        "key": "daily_challenge",
        "value": 5,
        "resetType": "daily",
        "resetTime": "2023-10-29 00:00:00",
        "timeToReset": 36000000,
        "lastModified": "2023-10-28 14:30:00"
    }
}
```

**返回示例（所有计数器）**:
```json
{
    "code": 0,
    "msg": "success",
    "timestamp": 1603991234567,
    "data": [
        {
            "key": "daily_challenge",
            "value": 5,
            "resetType": "daily",
            "resetTime": "2023-10-29 00:00:00",
            "timeToReset": 36000000,
            "lastModified": "2023-10-28 14:30:00"
        },
        {
            "key": "weekly_battle",
            "value": 10,
            "resetType": "weekly",
            "resetTime": "2023-10-30 00:00:00",
            "timeToReset": 122400000,
            "lastModified": "2023-10-28 10:15:00"
        }
    ]
}
```

## ZY-SDK 接口

### 基础接口

#### incrementCounter(key, increment?, resetType?, resetValue?)

增加计数器值

**参数**:
- `key: string` - 计数器key
- `increment: number = 1` - 增加的数量，默认1
- `resetType?: 'daily' | 'weekly' | 'monthly' | 'custom' | 'permanent'` - 重置类型
- `resetValue?: number` - 自定义重置时间(小时)

**示例**:
```typescript
// 增加每日挑战次数
const result = await ZYSDK.counter.incrementCounter('daily_challenge', 1, 'daily');

// 增加自定义计数器（48小时后重置）
const result = await ZYSDK.counter.incrementCounter('event_counter', 1, 'custom', 48);
```

#### getCounter(key?)

获取计数器值

**参数**:
- `key?: string` - 计数器key，不传则获取所有计数器

**示例**:
```typescript
// 获取单个计数器
const result = await ZYSDK.counter.getCounter('daily_challenge');

// 获取所有计数器
const allCounters = await ZYSDK.counter.getCounter();
```

### 便捷接口

#### incrementDailyChallenge(key?, increment?)

增加每日挑战次数

**示例**:
```typescript
// 使用默认key 'daily_challenge'
await ZYSDK.counter.incrementDailyChallenge();

// 使用自定义key
await ZYSDK.counter.incrementDailyChallenge('custom_daily', 2);
```

#### incrementWeeklyBattle(key?, increment?)

增加每周战斗次数

**示例**:
```typescript
// 使用默认key 'weekly_battle'
await ZYSDK.counter.incrementWeeklyBattle();
```

#### incrementScore(key?, increment?)

增加积分（永久累积）

**示例**:
```typescript
// 使用默认key 'total_score'
await ZYSDK.counter.incrementScore('total_score', 100);
```

#### getDailyChallenge(key?)

获取每日挑战次数

**示例**:
```typescript
const dailyCount = await ZYSDK.counter.getDailyChallenge();
```

#### getWeeklyBattle(key?)

获取每周战斗次数

#### getTotalScore(key?)

获取总积分

#### getAllCounters()

获取游戏所有计数器

**示例**:
```typescript
const allCounters = await ZYSDK.counter.getAllCounters();
```

## 重置类型说明

### daily（每日重置）
- **描述**: 每天0点自动重置计数器
- **适用场景**: 每日任务、日常挑战等
- **重置时间**: 每日 00:00:00

### weekly（每周重置）
- **描述**: 每周一0点自动重置计数器
- **适用场景**: 周任务、周赛等
- **重置时间**: 每周一 00:00:00

### monthly（每月重置）
- **描述**: 每月1号0点自动重置计数器
- **适用场景**: 月任务、月度活动等
- **重置时间**: 每月1号 00:00:00

### custom（自定义重置）
- **描述**: 根据resetValue参数指定的小时数后重置
- **适用场景**: 限时活动、自定义周期任务等
- **配置参数**: 需要配合resetValue参数使用

### permanent（永久计数器）
- **描述**: 计数器永远不会自动重置
- **适用场景**: 总积分、历史统计等
- **特点**: 数据会一直累积

## 使用场景示例

### 1. 每日任务系统

```typescript
// 初始化SDK
ZYSDK.init({
    appId: 'your-app-id'
});

// 完成每日任务时增加次数
await ZYSDK.counter.incrementDailyChallenge('daily_task', 1);

// 获取今日完成次数
const taskCount = await ZYSDK.counter.getDailyChallenge('daily_task');
console.log(`今日已完成任务: ${taskCount.data.value} 次`);
```

### 2. 游戏积分系统

```typescript
// 获得积分时增加
await ZYSDK.counter.incrementScore('game_score', 50);

// 查看总积分
const totalScore = await ZYSDK.counter.getTotalScore('game_score');
console.log(`总积分: ${totalScore.data.value}`);
```

### 3. 限时活动计数

```typescript
// 参与活动时增加计数（24小时后重置）
await ZYSDK.counter.incrementCounter('event_2023', 1, 'custom', 24);

// 查看活动参与次数
const eventCount = await ZYSDK.counter.getCounter('event_2023');
console.log(`活动参与次数: ${eventCount.data.value}`);
console.log(`剩余时间: ${eventCount.data.timeToReset}ms`);
```

### 4. 多计数器管理

```typescript
// 获取玩家所有计数器
const allCounters = await ZYSDK.counter.getAllCounters();
allCounters.data.forEach(counter => {
    console.log(`${counter.key}: ${counter.value} (${counter.resetType})`);
});
```

## 数据存储结构

计数器数据存储在MongoDB集合中，集合名格式为：`counter_{appId}`

数据结构：
```json
{
    "_id": "记录ID",
    "appId": "小程序ID",
    "key": "计数器key",
    "value": "当前计数值",
    "resetType": "重置类型",
    "resetValue": "自定义重置时间（小时）",
    "resetTime": "下次重置时间",
    "gmtCreate": "创建时间",
    "gmtModify": "修改时间"
}
```

## 注意事项

1. **自动重置**：当玩家调用任何接口时，系统会自动检查是否到达重置时间，如果到达则自动重置计数器
2. **时区处理**：所有时间处理基于服务器时区
3. **原子操作**：增加操作是原子性的，确保并发安全
4. **性能考虑**：建议为计数器集合创建复合索引：`{appId: 1, key: 1}`
5. **错误处理**：接口包含完整的参数校验和错误处理机制

## 错误码说明

- **4001**: 参数错误
- **5001**: 数据库操作失败

## 版本历史

- **v1.0**: 初始版本，支持基础计数和5种重置类型
- 支持云函数和SDK接口
- 兼容现有系统架构 