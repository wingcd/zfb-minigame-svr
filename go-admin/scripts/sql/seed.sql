-- 使用管理后台数据库
USE `minigame_admin`;

-- 插入默认管理员角色
INSERT INTO `admin_roles` (`id`, `name`, `description`, `permissions`, `status`) VALUES
(1, '超级管理员', '拥有所有权限的超级管理员', '["*"]', 1),
(2, '管理员', '普通管理员，拥有大部分权限', '["app", "config", "mail", "leaderboard", "counter", "stats"]', 1),
(3, '运营人员', '运营人员，只能管理邮件和配置', '["mail", "config"]', 1);

-- 插入默认超级管理员账户
-- 密码: admin123 (需要在实际使用时修改)
INSERT INTO `admin_users` (`username`, `password`, `email`, `real_name`, `status`, `role_id`) VALUES
('admin', '$2a$10$YourHashedPasswordHere', 'admin@example.com', '系统管理员', 1, 1);

-- 插入示例应用
INSERT INTO `applications` (`app_id`, `app_name`, `app_secret`, `description`, `platform`, `status`, `config`) VALUES
('demo_app_001', '示例小游戏', 'demo_secret_key_12345678901234567890', '这是一个示例小游戏应用', 'wechat', 1, '{"enable_leaderboard": true, "enable_mail": true, "enable_counter": true}');

-- 插入示例游戏配置
INSERT INTO `game_configs` (`app_id`, `config_key`, `config_value`, `config_type`, `description`, `is_public`, `status`) VALUES
('demo_app_001', 'game_name', '示例小游戏', 'string', '游戏名称', 1, 1),
('demo_app_001', 'version', '1.0.0', 'string', '游戏版本', 1, 1),
('demo_app_001', 'max_score', '999999', 'number', '最大分数限制', 1, 1),
('demo_app_001', 'enable_sound', 'true', 'boolean', '是否启用音效', 1, 1),
('demo_app_001', 'server_url', 'https://api.example.com', 'string', '服务器地址', 0, 1),
('demo_app_001', 'api_key', 'secret_api_key_123', 'string', 'API密钥', 0, 1);

-- 使用游戏SDK数据库
USE `minigame_game`;

-- 为示例应用创建用户数据表
CREATE TABLE IF NOT EXISTS `user_data_demo_app_001` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) NOT NULL COMMENT '用户ID',
  `platform` varchar(20) DEFAULT '' COMMENT '平台类型',
  `platform_user_id` varchar(100) DEFAULT '' COMMENT '平台用户ID',
  `nickname` varchar(100) DEFAULT '' COMMENT '用户昵称',
  `avatar` varchar(255) DEFAULT '' COMMENT '用户头像',
  `data` longtext COMMENT '用户数据(JSON格式)',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `idx_platform` (`platform`),
  KEY `idx_last_login_at` (`last_login_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='示例应用用户数据表';

-- 为示例应用创建排行榜表
CREATE TABLE IF NOT EXISTS `leaderboard_demo_app_001_score` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) NOT NULL COMMENT '用户ID',
  `nickname` varchar(100) DEFAULT '' COMMENT '用户昵称',
  `avatar` varchar(255) DEFAULT '' COMMENT '用户头像',
  `score` bigint(20) DEFAULT 0 COMMENT '分数',
  `extra_data` text COMMENT '扩展数据(JSON格式)',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_id` (`user_id`),
  KEY `idx_score` (`score` DESC),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='示例应用分数排行榜';

-- 为示例应用创建计数器表
CREATE TABLE IF NOT EXISTS `counter_demo_app_001` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `counter_key` varchar(100) NOT NULL COMMENT '计数器键',
  `counter_value` bigint(20) DEFAULT 0 COMMENT '计数器值',
  `description` varchar(255) DEFAULT '' COMMENT '计数器描述',
  `reset_type` varchar(20) DEFAULT 'none' COMMENT '重置类型',
  `reset_time` varchar(20) DEFAULT '' COMMENT '重置时间',
  `last_reset_at` datetime DEFAULT NULL COMMENT '最后重置时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `counter_key` (`counter_key`),
  KEY `idx_counter_value` (`counter_value`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='示例应用计数器表';

-- 为示例应用创建邮件表
CREATE TABLE IF NOT EXISTS `mail_demo_app_001` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `mail_id` varchar(50) NOT NULL COMMENT '邮件ID',
  `title` varchar(200) NOT NULL COMMENT '邮件标题',
  `content` text COMMENT '邮件内容',
  `sender` varchar(100) DEFAULT 'system' COMMENT '发送者',
  `recipients` text COMMENT '收件人列表(JSON格式)',
  `rewards` text COMMENT '奖励列表(JSON格式)',
  `type` varchar(20) DEFAULT 'system' COMMENT '邮件类型',
  `expire_at` datetime DEFAULT NULL COMMENT '过期时间',
  `status` tinyint(1) DEFAULT 1 COMMENT '状态 1:正常 0:删除',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `mail_id` (`mail_id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='示例应用邮件表';

-- 为示例应用创建用户邮件关系表
CREATE TABLE IF NOT EXISTS `user_mail_demo_app_001` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(50) NOT NULL COMMENT '用户ID',
  `mail_id` varchar(50) NOT NULL COMMENT '邮件ID',
  `is_read` tinyint(1) DEFAULT 0 COMMENT '是否已读 1:已读 0:未读',
  `is_received` tinyint(1) DEFAULT 0 COMMENT '是否已领取 1:已领取 0:未领取',
  `read_at` datetime DEFAULT NULL COMMENT '阅读时间',
  `received_at` datetime DEFAULT NULL COMMENT '领取时间',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_mail` (`user_id`, `mail_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_is_read` (`is_read`),
  KEY `idx_is_received` (`is_received`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='示例应用用户邮件关系表';

-- 插入示例计数器数据
INSERT INTO `counter_demo_app_001` (`counter_key`, `counter_value`, `description`, `reset_type`) VALUES
('total_games', 0, '总游戏次数', 'none'),
('daily_games', 0, '每日游戏次数', 'daily'),
('weekly_games', 0, '每周游戏次数', 'weekly'),
('online_users', 0, '在线用户数', 'none');

-- 插入示例邮件数据
INSERT INTO `mail_demo_app_001` (`mail_id`, `title`, `content`, `sender`, `recipients`, `rewards`, `type`, `expire_at`) VALUES
('welcome_001', '欢迎来到游戏世界！', '亲爱的玩家，欢迎来到我们的游戏世界！为了感谢您的加入，我们为您准备了新手礼包。', 'system', '["all"]', '[{"type": "coin", "amount": 1000}, {"type": "gem", "amount": 100}]', 'system', DATE_ADD(NOW(), INTERVAL 30 DAY)),
('daily_bonus_001', '每日签到奖励', '恭喜您获得每日签到奖励！记得每天都来签到哦~', 'system', '["all"]', '[{"type": "coin", "amount": 500}]', 'daily', DATE_ADD(NOW(), INTERVAL 1 DAY)),
('update_notice_001', '游戏更新公告', '游戏已更新到1.0.0版本，新增了更多有趣的功能，快来体验吧！', 'system', '["all"]', '[]', 'announcement', DATE_ADD(NOW(), INTERVAL 7 DAY));

-- 注意：在实际部署时，需要：
-- 1. 修改默认管理员密码（使用正确的bcrypt哈希）
-- 2. 修改示例应用的app_secret
-- 3. 根据实际需求调整配置数据 