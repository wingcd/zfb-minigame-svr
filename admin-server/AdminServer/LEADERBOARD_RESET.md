# 排行榜重置功能文档

## 概述

排行榜重置功能允许管理员配置排行榜的自动重置策略，支持多种重置周期，适用于不同类型的游戏场景。

## 功能特点

- **自动重置**: 根据配置的重置类型自动清空排行榜数据
- **灵活配置**: 支持5种不同的重置周期
- **实时检查**: 每次获取排行榜或提交分数时都会检查重置条件
- **数据一致性**: 重置后重新计算下次重置时间

## 重置类型

### 1. daily（每日重置）
- **描述**: 每天0点自动重置排行榜
- **适用场景**: 每日排行榜、日常活动等
- **重置时间**: 每日 00:00:00

### 2. weekly（每周重置）
- **描述**: 每周一0点自动重置排行榜
- **适用场景**: 周排行榜、周赛等
- **重置时间**: 每周一 00:00:00

### 3. monthly（每月重置）
- **描述**: 每月1号0点自动重置排行榜
- **适用场景**: 月排行榜、月度竞赛等
- **重置时间**: 每月1号 00:00:00

### 4. custom（自定义重置）
- **描述**: 根据resetValue参数指定的小时数后重置
- **适用场景**: 限时活动、自定义周期竞赛等
- **配置参数**: 需要配合resetValue参数使用
- **重置时间**: 从创建时间开始，每隔resetValue小时重置一次

### 5. permanent（永久排行榜）
- **描述**: 排行榜永远不会自动重置
- **适用场景**: 总排行榜、历史最高分等
- **特点**: 数据会一直累积，除非手动清除

## API 接口

### leaderboardInit - 初始化排行榜

**参数说明:**

| 参数名 | 类型 | 必选 | 说明 |
| --- | --- | --- | --- |
| leaderboardName | string | 是 | 排行榜名字 |
| appId | string | 是 | 小程序id |
| leaderboardType | string | 是 | 排行榜类型 |
| updateStrategy | number | 否 | 更新策略 (0=最高分, 1=最新分, 2=累计分) |
| sort | number | 否 | 排序方式 (1=降序, 0=升序) |
| resetType | string | 否 | 重置类型，默认为"permanent" |
| resetValue | number | 否 | 自定义重置时间(小时)，仅在resetType为custom时有效 |

**请求示例:**

```json
{
    "leaderboardName": "每日排行榜",
    "appId": "your-app-id",
    "leaderboardType": "daily_rank",
    "updateStrategy": 0,
    "sort": 1,
    "resetType": "daily"
}
```

### commitScore - 提交分数

提交分数时会自动检查是否需要重置排行榜。如果到达重置时间，会先清空数据再提交新分数。

### getLeaderboardTopRank - 获取排行榜

获取排行榜时会自动检查是否需要重置。如果需要重置，会先清空数据再返回空的排行榜。

## 使用场景示例

### 1. 每日排行榜

```json
{
    "leaderboardName": "每日排行榜",
    "appId": "your-app-id",
    "leaderboardType": "daily_rank",
    "resetType": "daily",
    "updateStrategy": 0,
    "sort": 1
}
```

**特点**: 每天0点自动重置，适合日常竞赛活动。

### 2. 限时活动排行榜（48小时后重置）

```json
{
    "leaderboardName": "限时活动排行榜",
    "appId": "your-app-id",
    "leaderboardType": "event_rank",
    "resetType": "custom",
    "resetValue": 48,
    "updateStrategy": 0,
    "sort": 1
}
```

**特点**: 创建后48小时自动重置，适合短期活动。

### 3. 周排行榜

```json
{
    "leaderboardName": "周排行榜",
    "appId": "your-app-id",
    "leaderboardType": "weekly_rank",
    "resetType": "weekly",
    "updateStrategy": 0,
    "sort": 1
}
```

**特点**: 每周一0点重置，适合周赛活动。

### 4. 总排行榜（永久）

```json
{
    "leaderboardName": "总排行榜",
    "appId": "your-app-id",
    "leaderboardType": "total_rank",
    "resetType": "permanent",
    "updateStrategy": 0,
    "sort": 1
}
```

**特点**: 永不重置，数据持续累积。

## 技术实现

### 数据库字段

在 `leaderboard_config` 集合中新增以下字段：

- `resetType`: 重置类型
- `resetValue`: 自定义重置间隔（小时）
- `resetTime`: 下次重置时间

### 重置逻辑

1. **检查时机**: 在 `commitScore` 和 `getLeaderboardTopRank` 函数中检查
2. **检查条件**: 当前时间 > resetTime 且 resetType ≠ 'permanent'
3. **重置操作**: 清空 `leaderboard_score` 集合中对应的数据
4. **更新时间**: 重新计算并更新下次重置时间

### 时间计算

```javascript
// 每日重置 - 明天0点
resetTime = moment().startOf('day').add(1, 'day').format("YYYY-MM-DD HH:mm:ss");

// 每周重置 - 下周一0点
resetTime = moment().startOf('week').add(1, 'week').format("YYYY-MM-DD HH:mm:ss");

// 每月重置 - 下月1号0点
resetTime = moment().startOf('month').add(1, 'month').format("YYYY-MM-DD HH:mm:ss");

// 自定义重置 - 当前时间 + resetValue小时
resetTime = moment().add(resetValue, 'hours').format("YYYY-MM-DD HH:mm:ss");
```

## 注意事项

1. **时区问题**: 所有时间计算基于服务器时区
2. **数据备份**: 重置前建议备份重要数据
3. **并发处理**: 多个请求同时检查重置时，可能出现重复重置，但不会影响数据一致性
4. **性能考虑**: 重置检查在每次请求时进行，对性能影响很小

## 错误处理

- **5001**: 数据库操作失败
- **重置排行榜失败**: 清空数据时出现错误
- **更新重置时间失败**: 更新配置时出现错误

## 版本历史

- **v1.0**: 初始版本，支持5种重置类型
- 支持自动重置检查和时间计算
- 兼容现有排行榜功能 