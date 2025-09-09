# Redis排行榜系统使用指南

## 概述

Redis排行榜系统利用Redis的Sorted Set数据结构提供高性能的排行榜功能，支持实时更新和快速查询。该系统使用Redis作为主要存储，MySQL用于配置管理和数据持久化。

## 主要特性

### 1. 高性能排行榜操作
- 使用Redis Sorted Set实现O(log n)的插入和查询性能
- 支持范围查询、排名查询等高效操作
- 支持大量并发读写操作

### 2. 数据持久化保证
- Redis作为主要存储
- 异步同步到MySQL保证数据持久化
- 支持从MySQL恢复数据

### 3. 灵活的更新策略
- **最高分策略**: 只保留最高分数
- **最新分策略**: 总是使用最新分数
- **累计分策略**: 累加分数

### 4. 完整的排行榜管理
- 支持多个排行榜类型
- 支持排行榜大小限制
- 支持排行榜清空和重建

## API接口

### 1. 更新玩家分数
```bash
POST /leaderboard/updateScore
```

**请求参数:**
```json
{
  "appId": "your_app_id",
  "leaderboardType": "weekly_score",
  "playerId": "player_123",
  "score": 1000,
  "extraData": {
    "nickname": "玩家昵称",
    "avatar": "头像URL",
    "level": 10
  }
}
```

**响应:**
```json
{
  "code": 0,
  "msg": "success",
  "timestamp": 1640995200000,
  "data": null
}
```

### 2. 获取排行榜数据
```bash
POST /leaderboard/getData
```

**请求参数:**
```json
{
  "appId": "your_app_id",
  "leaderboardType": "weekly_score",
  "page": 1,
  "pageSize": 20
}
```

**响应:**
```json
{
  "code": 0,
  "msg": "success",
  "timestamp": 1640995200000,
  "data": {
    "list": [
      {
        "playerId": "player_123",
        "score": 1000,
        "rank": 1,
        "extraData": {
          "nickname": "玩家昵称",
          "avatar": "头像URL",
          "level": 10
        },
        "updatedAt": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "pageSize": 20,
    "totalPages": 5
  }
}
```

### 3. 查询玩家分数和排名
```bash
POST /leaderboard/queryScore
```

**请求参数:**
```json
{
  "appId": "your_app_id",
  "leaderboardType": "weekly_score",
  "playerId": "player_123"
}
```

**响应:**
```json
{
  "code": 0,
  "msg": "success",
  "timestamp": 1640995200000,
  "data": {
    "playerId": "player_123",
    "score": 1000,
    "rank": 1
  }
}
```

### 4. 删除玩家分数
```bash
POST /leaderboard/deleteScore
```

**请求参数:**
```json
{
  "appId": "your_app_id",
  "leaderboardType": "weekly_score",
  "playerId": "player_123"
}
```

### 5. 提交分数（别名接口）
```bash
POST /leaderboard/commitScore
```

**请求参数:**
```json
{
  "appId": "your_app_id",
  "leaderboardType": "weekly_score",
  "playerId": "player_123",
  "score": 1000,
  "extraData": {
    "nickname": "玩家昵称",
    "level": 10
  }
}
```

## Redis数据结构

### 1. 排行榜主数据
**键名格式:** `leaderboard:{appId}:{leaderboardType}`
**数据类型:** Sorted Set
**成员:** playerId
**分数:** 玩家分数

### 2. 玩家详细数据
**键名格式:** `leaderboard_data:{appId}:{leaderboardType}:{playerId}`
**数据类型:** String (JSON)
**内容:** 
```json
{
  "playerId": "player_123",
  "score": 1000,
  "extraData": {...},
  "updatedAt": "2024-01-01T12:00:00Z"
}
```

### 3. 排行榜配置缓存
**键名格式:** `leaderboard_config:{appId}:{leaderboardType}`
**数据类型:** String (JSON)
**过期时间:** 10分钟

## 配置说明

### Redis配置 (conf/app.conf)
```ini
# Redis配置
redis_host = 127.0.0.1
redis_port = 6379
redis_password = your_password
redis_database = 0
redis_pool_size = 10
```

### 排行榜配置表 (leaderboard_config)
```sql
CREATE TABLE leaderboard_config (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    app_id VARCHAR(100) NOT NULL,
    leaderboard_type VARCHAR(100) NOT NULL,
    enabled TINYINT(1) DEFAULT 1,
    max_rank INT DEFAULT 0,
    update_strategy TINYINT DEFAULT 0,  -- 0:最高分 1:最新分 2:累计分
    score_type VARCHAR(20) DEFAULT 'higher_better',  -- higher_better, lower_better
    sort TINYINT DEFAULT 1,  -- 0:升序 1:降序
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_app_leaderboard (app_id, leaderboard_type)
);
```

## 性能优化建议

### 1. Redis内存优化
- 设置合理的排行榜大小限制 (max_rank)
- 定期清理过期的排行榜数据
- 使用适当的Redis内存回收策略

### 2. 数据同步策略
- 异步同步避免阻塞主流程
- 批量同步提高效率
- 失败重试机制保证数据一致性

### 3. 配置缓存策略
- 排行榜配置缓存10分钟
- 玩家详细数据在Redis中长期保存
- 支持手动数据重建

## 监控和运维

### 1. 关键指标
- Redis连接数和内存使用
- 排行榜操作延迟
- 同步失败率

### 2. 故障恢复
- Redis故障时需要重启服务
- 支持从MySQL恢复数据
- 完整的操作日志记录

### 3. 容量规划
- 估算每个排行榜的内存需求
- 规划Redis集群扩展
- 监控排行榜增长趋势

## 最佳实践

1. **合理设置排行榜大小**: 根据业务需求设置max_rank，避免无限增长
2. **选择合适的更新策略**: 根据游戏特性选择最高分、最新分或累计分策略
3. **定期数据同步**: 建议每小时进行一次MySQL同步，确保数据安全
4. **监控Redis性能**: 关注内存使用和操作延迟，及时优化
5. **备份策略**: 定期备份MySQL数据，Redis作为缓存层可重建

## 故障排查

### 常见问题

1. **Redis连接失败**
   - 检查Redis服务状态
   - 验证连接配置
   - 查看网络连接

2. **数据不一致**
   - 手动执行同步操作
   - 检查异步同步日志
   - 验证配置是否正确

3. **性能问题**
   - 检查Redis内存使用
   - 优化查询范围
   - 考虑分片策略

## 版本更新说明

### v1.0.0
- 实现基础Redis排行榜功能
- 支持异步MySQL同步
- 提供完整的管理API
- 支持多种更新策略
