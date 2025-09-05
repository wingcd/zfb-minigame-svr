# API迁移清单

## 管理后台服务 API (端口8080)

### 管理员认证模块
- `POST /adminLogin` - 管理员登录
- `POST /verifyToken` - Token验证
- `POST /createAdmin` - 创建管理员
- `POST /getAllRoles` - 获取所有角色
- `POST /resetPassword` - 重置密码
- `POST /initAdmin` - 初始化管理员系统

### 应用管理模块
- `POST /createApp` - 创建应用
- `POST /getAllApps` - 获取所有应用
- `POST /updateApp` - 更新应用信息
- `POST /deleteApp` - 删除应用
- `POST /getAppStats` - 获取应用统计

### 游戏配置模块
- `POST /createGameConfig` - 创建游戏配置
- `POST /getGameConfigList` - 获取配置列表
- `POST /updateGameConfig` - 更新游戏配置
- `POST /deleteGameConfig` - 删除游戏配置

### 邮件管理模块
- `POST /initMailSystem` - 初始化邮件系统
- `POST /createMail` - 创建邮件
- `POST /sendMail` - 发送邮件
- `POST /getMailList` - 获取邮件列表

### 排行榜管理模块
- `POST /createLeaderboard` - 创建排行榜
- `POST /leaderboardInit` - 初始化排行榜
- `POST /getLeaderboardStats` - 获取排行榜统计

### 计数器管理模块
- `POST /createCounter` - 创建计数器
- `POST /getCounterList` - 获取计数器列表
- `POST /updateCounter` - 更新计数器配置
- `POST /deleteCounter` - 删除计数器
- `POST /resetCounter` - 重置计数器值
- `POST /getCounterStats` - 获取计数器统计

### 排行榜高级管理模块
- `POST /updateLeaderboard` - 更新排行榜配置
- `POST /deleteLeaderboard` - 删除排行榜
- `POST /resetLeaderboard` - 重置排行榜数据
- `POST /getLeaderboardList` - 获取排行榜列表
- `POST /setLeaderboardResetSchedule` - 设置重置计划

### 玩家管理模块
- `POST /getAllUsers` - 获取用户列表（分页、搜索）
- `POST /getUserDetail` - 获取用户详细信息
- `POST /updateUserData` - 更新用户游戏数据
- `POST /setUserDetail` - 设置用户详细信息
- `POST /banUser` - 封禁用户
- `POST /unbanUser` - 解封用户
- `POST /deleteUser` - 删除用户（危险操作）
- `POST /getUserStats` - 获取用户统计信息

### 数据统计模块
- `POST /getRecentActivity` - 获取最近活动
- `POST /getDataStats` - 获取数据统计
- `POST /getUserRegistrationStats` - 获取用户注册统计
- `POST /getGameActivityStats` - 获取游戏活动统计

## 游戏SDK服务 API (端口8081)

### 用户管理模块
- `POST /login` - 用户登录
- `POST /wx_login` - 微信登录
- `POST /saveData` - 保存用户数据
- `POST /getData` - 获取用户数据

### 排行榜模块
- `POST /commitScore` - 提交分数
- `POST /queryScore` - 查询分数
- `POST /getLeaderboardTopRank` - 获取排行榜前几名
- `POST /deleteScore` - 删除分数
- `POST /getLeaderboardData` - 获取排行榜数据

### 计数器模块
- `POST /incrementCounter` - 增加计数器
- `POST /getCounter` - 获取计数器值

### 游戏配置模块
- `POST /getGameConfig` - 获取游戏配置（客户端用）

### 邮件模块
- `POST /getMails` - 获取用户邮件
- `POST /readMail` - 阅读邮件
- `POST /receiveMail` - 领取邮件奖励
- `POST /deleteMail` - 删除邮件
- `POST /getUnreadCount` - 获取未读数量

### 应用初始化模块
- `POST /appInit` - 初始化应用

## 数据库表结构映射

### 系统管理表
```sql
-- 管理员表
CREATE TABLE admin_users (
    id VARCHAR(64) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(32) NOT NULL,
    nickname VARCHAR(100),
    role VARCHAR(50),
    email VARCHAR(100),
    phone VARCHAR(20),
    status ENUM('active', 'inactive') DEFAULT 'active',
    token VARCHAR(128),
    token_expire DATETIME,
    last_login_time DATETIME,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 角色表
CREATE TABLE admin_roles (
    id VARCHAR(64) PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    role_description TEXT,
    permissions JSON,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 操作日志表
CREATE TABLE admin_operation_logs (
    id VARCHAR(64) PRIMARY KEY,
    admin_id VARCHAR(64),
    username VARCHAR(50),
    action VARCHAR(100),
    resource VARCHAR(100),
    details JSON,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_admin_id (admin_id),
    INDEX idx_create_time (create_time)
);
```

### 应用管理表
```sql
-- 应用表
CREATE TABLE applications (
    id VARCHAR(64) PRIMARY KEY,
    app_id VARCHAR(100) UNIQUE NOT NULL,
    app_name VARCHAR(200) NOT NULL,
    app_description TEXT,
    platform VARCHAR(50),
    app_key VARCHAR(128),
    status ENUM('active', 'inactive') DEFAULT 'active',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_id (app_id)
);
```

### 用户数据表（动态创建）
```sql
-- 用户表模板（每个应用一个表：user_{app_id}）
CREATE TABLE user_{app_id} (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    player_id VARCHAR(100) NOT NULL,
    data LONGTEXT,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_player_id (player_id)
);
```

### 排行榜表（动态创建）
```sql
-- 排行榜表模板（每个应用一个表：leaderboard_{app_id}）
CREATE TABLE leaderboard_{app_id} (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    leaderboard_id VARCHAR(100) NOT NULL,
    player_id VARCHAR(100) NOT NULL,
    player_info JSON,
    score BIGINT NOT NULL,
    additional_data JSON,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_leaderboard_score (leaderboard_id, score DESC),
    INDEX idx_player_id (player_id)
);
```

### 计数器表（动态创建）
```sql
-- 计数器表模板（每个应用一个表：counter_{app_id}）
CREATE TABLE counter_{app_id} (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    counter_key VARCHAR(100) NOT NULL,
    location VARCHAR(100) DEFAULT 'default',
    value BIGINT DEFAULT 0,
    reset_type ENUM('daily', 'weekly', 'monthly', 'custom', 'permanent') DEFAULT 'permanent',
    reset_value INT NULL,
    reset_time DATETIME NULL,
    description TEXT,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_key_location (counter_key, location)
);
```

### 排行榜配置表
```sql
-- 排行榜配置表
CREATE TABLE leaderboard_config (
    id VARCHAR(64) PRIMARY KEY,
    app_id VARCHAR(100) NOT NULL,
    leaderboard_id VARCHAR(100) NOT NULL,
    leaderboard_name VARCHAR(200) NOT NULL,
    update_strategy TINYINT DEFAULT 0 COMMENT '0=最高分, 1=最新分, 2=累计分',
    sort_order TINYINT DEFAULT 1 COMMENT '1=降序, 0=升序',
    reset_type ENUM('daily', 'weekly', 'monthly', 'custom', 'permanent') DEFAULT 'permanent',
    reset_value INT NULL COMMENT '自定义重置时间(小时)',
    next_reset_time DATETIME NULL,
    is_active BOOLEAN DEFAULT TRUE,
    description TEXT,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_id (app_id),
    UNIQUE KEY uk_app_leaderboard (app_id, leaderboard_id)
);
```

### 用户封禁记录表
```sql
-- 用户封禁记录表
CREATE TABLE user_ban_records (
    id VARCHAR(64) PRIMARY KEY,
    app_id VARCHAR(100) NOT NULL,
    player_id VARCHAR(100) NOT NULL,
    admin_id VARCHAR(64) NOT NULL,
    ban_type ENUM('temporary', 'permanent') DEFAULT 'temporary',
    ban_reason TEXT,
    ban_start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    ban_end_time DATETIME NULL,
    is_active BOOLEAN DEFAULT TRUE,
    unban_admin_id VARCHAR(64) NULL,
    unban_time DATETIME NULL,
    unban_reason TEXT,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_player (app_id, player_id),
    INDEX idx_admin_id (admin_id)
);
```

### 邮件表（动态创建）
```sql
-- 邮件表模板（每个应用一个表：mail_{app_id}）
CREATE TABLE mail_{app_id} (
    id VARCHAR(64) PRIMARY KEY,
    player_id VARCHAR(100) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    rewards JSON,
    mail_type ENUM('system', 'personal') DEFAULT 'system',
    is_read BOOLEAN DEFAULT FALSE,
    is_received BOOLEAN DEFAULT FALSE,
    expire_time DATETIME,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_player_id (player_id),
    INDEX idx_create_time (create_time)
);
```

### 游戏配置表
```sql
-- 游戏配置表
CREATE TABLE game_config (
    id VARCHAR(64) PRIMARY KEY,
    app_id VARCHAR(100) NOT NULL,
    config_key VARCHAR(100) NOT NULL,
    config_value LONGTEXT,
    version VARCHAR(50) NULL,
    description TEXT,
    config_type ENUM('string', 'number', 'boolean', 'object', 'array') DEFAULT 'string',
    is_active BOOLEAN DEFAULT TRUE,
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_app_id (app_id),
    INDEX idx_config_key (config_key),
    UNIQUE KEY uk_app_key_version (app_id, config_key, version)
);
```

## 新增模块详细说明

### 玩家管理模块详情

#### 核心功能
- **用户列表查询**: 支持分页、模糊搜索（playerId、openId）
- **用户详情管理**: 查看和编辑用户基本信息和游戏数据
- **用户状态管理**: 封禁/解封用户，记录操作历史
- **数据安全**: 敏感信息过滤，权限验证

#### 关键实现要点
```javascript
// 用户搜索支持模糊匹配
whereCondition.playerId = new RegExp(playerId, 'i');

// 用户数据JSON解析和验证
if (user.data && typeof user.data === 'string') {
    try {
        user.playerInfo = JSON.parse(user.data);
    } catch (e) {
        user.playerInfo = null;
    }
}

// 权限检查
if (!['super_admin', 'admin'].includes(adminInfo.role)) {
    return { code: 4003, msg: "权限不足" };
}
```

### 排行榜高级管理详情

#### 核心功能
- **配置管理**: 创建、更新、删除排行榜配置
- **重置策略**: 支持5种重置类型（daily/weekly/monthly/custom/permanent）
- **自动重置**: 系统自动检查并执行重置操作
- **统计分析**: 排行榜数据统计和分析

#### 重置类型详解
```javascript
// 每日重置：每天0点
resetType: 'daily' -> resetTime: '00:00:00'

// 每周重置：每周一0点  
resetType: 'weekly' -> resetTime: 'Monday 00:00:00'

// 每月重置：每月1号0点
resetType: 'monthly' -> resetTime: '1st day 00:00:00'

// 自定义重置：指定小时数后重置
resetType: 'custom', resetValue: 48 -> 48小时后重置

// 永久排行榜：永不重置
resetType: 'permanent' -> 数据永久保存
```

### 计数器高级管理详情

#### 核心功能
- **点位支持**: 同一计数器支持多个location（如不同服务器、地区）
- **原子操作**: 并发安全的增减操作
- **重置策略**: 与排行榜相同的5种重置类型
- **排行统计**: 支持location排行榜功能

#### 使用场景示例
```javascript
// 服务器在线人数统计
await incrementCounter('online_players', 1, 'server_01');
await incrementCounter('online_players', 1, 'server_02');

// 全服活动参与次数（每日重置）
await incrementCounter('daily_event_count', 1, 'default', 'daily');

// 地区活跃度统计
await incrementCounter('activity_score', 10, 'beijing');
await incrementCounter('activity_score', 15, 'shanghai');
```

### 数据统计模块详情

#### 新增统计功能
- **用户注册统计**: 按时间维度统计新用户注册数
- **游戏活动统计**: 统计各种游戏内活动的参与情况
- **实时活动监控**: 获取最近的用户活动记录

## 兼容性检查清单

### 接口兼容性
- [ ] 请求参数格式保持一致
- [ ] 响应数据结构保持一致
- [ ] 错误码和错误信息保持一致
- [ ] 分页参数和返回格式保持一致

### 认证兼容性
- [ ] MD5密码加密方式保持一致
- [ ] Token生成和验证机制保持一致
- [ ] 权限验证逻辑保持一致

### 数据兼容性
- [ ] 用户数据JSON格式保持一致
- [ ] 排行榜数据结构保持一致
- [ ] 邮件奖励格式保持一致
- [ ] 配置数据类型保持一致

### 功能兼容性
- [ ] 动态表创建机制保持一致
- [ ] 计数器重置逻辑保持一致
- [ ] 排行榜排序规则保持一致
- [ ] 排行榜自动重置机制保持一致
- [ ] 邮件系统行为保持一致
- [ ] 用户封禁状态检查机制保持一致
- [ ] 分页查询逻辑保持一致

## 测试验证方案

### 单元测试
- 各模块核心逻辑测试
- 数据库操作测试
- 工具函数测试

### 集成测试
- API接口功能测试
- 数据库事务测试
- 缓存一致性测试

### 兼容性测试
- 现有SDK接入测试
- 管理后台功能测试
- 数据迁移验证测试

### 新增模块测试
#### 玩家管理模块测试
- 用户列表分页和搜索功能测试
- 用户封禁/解封功能测试
- 用户数据编辑权限验证测试
- 敏感信息过滤测试

#### 排行榜管理模块测试
- 排行榜配置CRUD操作测试
- 自动重置功能测试（各种重置类型）
- 排行榜数据一致性测试
- 并发提交分数测试

#### 计数器管理模块测试
- 多点位计数器功能测试
- 原子操作并发安全测试
- 计数器重置功能测试
- Location排行榜功能测试

### 性能测试
- 并发访问测试
- 数据库查询性能测试
- 缓存效果测试 