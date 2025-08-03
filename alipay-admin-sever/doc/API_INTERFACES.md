# 云函数接口文档

## 概述

本文档列出了小游戏管理后台系统所有的云函数接口。这些接口为前端管理后台提供完整的数据管理功能。

## 应用管理接口

### 1. 应用初始化 - `app/appInit.js`
- **功能**: 创建新应用
- **路径**: `POST /app/appInit`
- **参数**: appName, platform, appId, appKey
- **状态**: ✅ 已存在

### 2. 查询应用 - `app/queryApp.js`
- **功能**: 查询单个应用信息
- **路径**: `POST /app/queryApp`
- **参数**: appName, appId, channelAppId
- **状态**: ✅ 已存在

### 3. 获取应用列表 - `app/getAllApps.js`
- **功能**: 获取所有应用列表，支持分页和搜索
- **路径**: `POST /app/getAllApps`
- **参数**: page, pageSize, appName, appId, platform
- **状态**: ✅ 新增

### 4. 更新应用 - `app/updateApp.js`
- **功能**: 更新应用信息
- **路径**: `POST /app/updateApp`
- **参数**: appId, appName, description, status, channelAppKey
- **状态**: ✅ 新增

### 5. 删除应用 - `app/deleteApp.js`
- **功能**: 删除应用及所有相关数据
- **路径**: `POST /app/deleteApp`
- **参数**: appId, force
- **状态**: ✅ 新增

## 用户管理接口

### 1. 用户登录 - `user/login.js`
- **功能**: 用户登录（支付宝小程序）
- **路径**: `POST /user/login`
- **参数**: appId, code
- **状态**: ✅ 已存在

### 2. 微信用户登录 - `user/login.wx.js`
- **功能**: 用户登录（微信小程序）
- **路径**: `POST /user/login.wx`
- **参数**: appId, code
- **状态**: ✅ 已存在

### 3. 获取用户数据 - `user/getData.js`
- **功能**: 获取用户游戏数据
- **路径**: `POST /user/getData`
- **参数**: appId, playerId
- **状态**: ✅ 已存在

### 4. 保存用户数据 - `user/saveData.js`
- **功能**: 保存用户游戏数据
- **路径**: `POST /user/saveData`
- **参数**: appId, playerId, data
- **状态**: ✅ 已存在

### 5. 获取用户列表 - `user/getAllUsers.js`
- **功能**: 获取用户列表，支持分页和搜索
- **路径**: `POST /user/getAllUsers`
- **参数**: appId, page, pageSize, playerId, openId
- **状态**: ✅ 新增

### 6. 封禁用户 - `user/banUser.js`
- **功能**: 封禁用户
- **路径**: `POST /user/banUser`
- **参数**: appId, playerId, reason, duration
- **状态**: ✅ 新增

### 7. 解封用户 - `user/unbanUser.js`
- **功能**: 解封用户
- **路径**: `POST /user/unbanUser`
- **参数**: appId, playerId, reason
- **状态**: ✅ 新增

### 8. 删除用户 - `user/deleteUser.js`
- **功能**: 删除用户及相关数据
- **路径**: `POST /user/deleteUser`
- **参数**: appId, playerId, force
- **状态**: ✅ 新增

## 排行榜管理接口

### 1. 排行榜初始化 - `leaderboard/leaderboardInit.js`
- **功能**: 创建排行榜配置
- **路径**: `POST /leaderboard/leaderboardInit`
- **参数**: leaderboardName, appId, leaderboardType, updateStrategy, sort
- **状态**: ✅ 已存在

### 2. 提交分数 - `leaderboard/commitScore.js`
- **功能**: 提交玩家分数
- **路径**: `POST /leaderboard/commitScore`
- **参数**: appId, playerId, type, score, playerInfo
- **状态**: ✅ 已存在

### 3. 查询分数 - `leaderboard/queryScore.js`
- **功能**: 查询玩家分数
- **路径**: `POST /leaderboard/queryScore`
- **参数**: appId, playerId, leaderboardType
- **状态**: ✅ 已存在

### 4. 获取排行榜 - `leaderboard/getLeaderboardTopRank.js`
- **功能**: 获取排行榜前N名
- **路径**: `POST /leaderboard/getLeaderboardTopRank`
- **参数**: appId, type, startRank, count, sort
- **状态**: ✅ 已存在

### 5. 删除分数 - `leaderboard/deleteScore.js`
- **功能**: 删除玩家分数记录
- **路径**: `POST /leaderboard/deleteScore`
- **参数**: appId, playerId, leaderboardType
- **状态**: ✅ 已存在

### 6. 获取排行榜列表 - `leaderboard/getAllLeaderboards.js`
- **功能**: 获取排行榜配置列表
- **路径**: `POST /leaderboard/getAllLeaderboards`
- **参数**: appId, page, pageSize, leaderboardType
- **状态**: ✅ 新增

### 7. 更新排行榜配置 - `leaderboard/updateLeaderboard.js`
- **功能**: 更新排行榜配置
- **路径**: `POST /leaderboard/updateLeaderboard`
- **参数**: appId, leaderboardType, name, updateStrategy, sort, enabled
- **状态**: ✅ 新增

### 8. 删除排行榜 - `leaderboard/deleteLeaderboard.js`
- **功能**: 删除排行榜配置及数据
- **路径**: `POST /leaderboard/deleteLeaderboard`
- **参数**: appId, leaderboardType, force
- **状态**: ✅ 新增

## 统计数据接口

### 1. 获取仪表板统计 - `stats/getDashboardStats.js`
- **功能**: 获取仪表板统计数据
- **路径**: `POST /stats/getDashboardStats`
- **参数**: timeRange
- **状态**: ✅ 新增

## 数据结构说明

### 通用返回格式
```json
{
  "code": 0,
  "msg": "success",
  "timestamp": 1603991234567,
  "data": {}
}
```

### 错误代码说明
- `0`: 成功
- `4001`: 参数错误
- `4003`: 业务逻辑错误（如重复操作）
- `4004`: 资源不存在
- `5001`: 服务器内部错误

### 数据库表结构

#### app_config (应用配置表)
- appId: 应用ID
- appName: 应用名称
- platform: 平台（wechat, alipay, douyin）
- channelAppId: 渠道应用ID
- channelAppKey: 渠道应用密钥
- status: 状态（active, inactive）
- createTime: 创建时间
- updateTime: 更新时间

#### user_{appId} (用户表，每个应用一个表)
- playerId: 玩家ID
- openId: 用户OpenID
- token: 登录令牌
- data: 游戏数据（JSON字符串）
- banned: 是否封禁
- banReason: 封禁原因
- banTime: 封禁时间
- banUntil: 封禁到期时间
- gmtCreate: 创建时间
- gmtModify: 修改时间

#### leaderboard_config (排行榜配置表)
- appId: 应用ID
- leaderboardType: 排行榜类型
- name: 排行榜名称
- updateStrategy: 更新策略（0:最高值, 1:最近, 2:总和）
- sort: 排序方式（0:升序, 1:降序）
- enabled: 是否启用
- createTime: 创建时间
- updateTime: 更新时间

#### leaderboard_score (排行榜分数表)
- appId: 应用ID
- playerId: 玩家ID
- leaderboardType: 排行榜类型
- score: 分数
- playerInfo: 玩家信息（JSON）
- gmtCreate: 创建时间
- gmtModify: 修改时间

## 部署说明

1. 将所有云函数文件部署到对应的云服务平台
2. 配置API网关路由，确保路径映射正确
3. 设置数据库权限，确保云函数可以访问数据库
4. 配置CORS，允许前端域名访问

## 注意事项

1. 所有接口都需要进行参数校验
2. 删除操作建议使用soft delete或force参数确认
3. 统计接口可能比较耗时，建议增加缓存机制
4. 用户数据涉及隐私，注意数据安全
5. 排行榜数据可能会很大，注意分页处理 