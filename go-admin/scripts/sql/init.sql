-- 小游戏服务器数据库初始化脚本
-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `minigame_server` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `minigame_server`;

-- ================================
-- 系统管理表（固定表，不按应用创建）
-- ================================

-- 应用管理表
CREATE TABLE IF NOT EXISTS `applications` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `app_id` varchar(50) NOT NULL COMMENT '应用ID',
  `app_name` varchar(100) NOT NULL COMMENT '应用名称',
  `appSecret` varchar(100) NOT NULL COMMENT '应用密钥',
  `platform` varchar(50) DEFAULT NULL COMMENT '平台类型',
  `channel_app_id` varchar(100) DEFAULT NULL COMMENT '渠道应用ID',
  `channel_app_key` varchar(100) DEFAULT NULL COMMENT '渠道应用密钥',
  `description` text COMMENT '应用描述',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 1:启用 0:禁用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_id` (`app_id`),
  KEY `idx_status` (`status`),
  KEY `idx_platform` (`platform`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用管理表';

-- 管理员用户表
CREATE TABLE IF NOT EXISTS `admin_users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password` varchar(100) NOT NULL COMMENT '密码hash',
  `nickname` varchar(50) DEFAULT NULL COMMENT '昵称',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 1:启用 0:禁用',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `last_login_ip` varchar(45) DEFAULT NULL COMMENT '最后登录IP',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员用户表';

-- 管理员角色表
CREATE TABLE IF NOT EXISTS `admin_roles` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '角色名称',
  `description` varchar(255) DEFAULT NULL COMMENT '角色描述',
  `permissions` text COMMENT '权限列表（JSON格式）',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态 1:启用 0:禁用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员角色表';

-- 管理员角色关联表
CREATE TABLE IF NOT EXISTS `admin_user_roles` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `role_id` bigint(20) NOT NULL COMMENT '角色ID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员角色关联表';

-- 管理员操作日志表
CREATE TABLE IF NOT EXISTS `admin_operation_logs` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL COMMENT '操作用户ID',
  `username` varchar(50) NOT NULL COMMENT '操作用户名',
  `action` varchar(100) NOT NULL COMMENT '操作动作',
  `resource` varchar(100) DEFAULT NULL COMMENT '操作资源',
  `params` text COMMENT '请求参数（JSON格式）',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(500) DEFAULT NULL COMMENT '用户代理',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员操作日志表';

-- ================================
-- 初始化数据
-- ================================

-- 插入默认管理员账户（用户名: admin, 密码: admin123）
INSERT IGNORE INTO `admin_users` (`username`, `password`, `nickname`, `status`) VALUES
('admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iKyqOyZpx8ROB8h6BdqhN7nkGYla', '超级管理员', 1);

-- 插入默认角色
INSERT IGNORE INTO `admin_roles` (`name`, `description`, `permissions`, `status`) VALUES
('超级管理员', '拥有所有权限', '["*"]', 1),
('应用管理员', '管理应用和游戏数据', '["app:*", "game:*"]', 1);

-- 给默认管理员分配超级管理员角色
INSERT IGNORE INTO `admin_user_roles` (`user_id`, `role_id`) VALUES
(1, 1);

-- ================================
-- 动态表模板（用于程序创建应用时参考）
-- ================================

/*
-- 用户数据表模板 (user_data_{app_id})
CREATE TABLE `user_data_{app_id}` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(100) NOT NULL COMMENT '用户ID',
  `data` longtext COMMENT '用户数据（JSON格式）',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户数据表';

-- 排行榜表模板 (leaderboard_{app_id})
CREATE TABLE `leaderboard_{app_id}` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `leaderboard_name` varchar(100) NOT NULL COMMENT '排行榜名称',
  `user_id` varchar(100) NOT NULL COMMENT '用户ID',
  `score` bigint(20) NOT NULL DEFAULT '0' COMMENT '分数',
  `extra_data` text COMMENT '额外数据（JSON格式）',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_leaderboard_user` (`leaderboard_name`, `user_id`),
  KEY `idx_leaderboard_score` (`leaderboard_name`, `score` DESC),
  KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排行榜数据表';

-- 计数器表模板 (counter_{app_id})
CREATE TABLE `counter_{app_id}` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `counter_name` varchar(100) NOT NULL COMMENT '计数器名称',
  `user_id` varchar(100) DEFAULT NULL COMMENT '用户ID（全局计数器为空）',
  `count` bigint(20) NOT NULL DEFAULT '0' COMMENT '计数值',
  `reset_time` datetime DEFAULT NULL COMMENT '重置时间',
  `reset_interval` int(11) DEFAULT NULL COMMENT '重置间隔（秒）',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_counter_user` (`counter_name`, `user_id`),
  KEY `idx_counter_name` (`counter_name`),
  KEY `idx_reset_time` (`reset_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计数器数据表';

-- 邮件表模板 (mail_{app_id})
CREATE TABLE `mail_{app_id}` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(100) NOT NULL COMMENT '收件人用户ID',
  `title` varchar(200) NOT NULL COMMENT '邮件标题',
  `content` text COMMENT '邮件内容',
  `rewards` text COMMENT '奖励物品（JSON格式）',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态 0:未读 1:已读 2:已领取',
  `expire_at` datetime DEFAULT NULL COMMENT '过期时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮件数据表';

-- 游戏配置表模板 (game_config_{app_id})
CREATE TABLE `game_config_{app_id}` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `config_key` varchar(100) NOT NULL COMMENT '配置键',
  `config_value` longtext COMMENT '配置值（JSON格式）',
  `version` varchar(50) DEFAULT NULL COMMENT '版本号',
  `description` varchar(255) DEFAULT NULL COMMENT '配置描述',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`),
  KEY `idx_version` (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏配置表';
*/ 