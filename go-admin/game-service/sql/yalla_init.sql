-- Yalla SDK 数据库初始化脚本

-- 创建 Yalla 配置表
CREATE TABLE IF NOT EXISTS `yalla_config` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `app_id` varchar(50) NOT NULL COMMENT '应用ID',
  `app_game_id` varchar(100) NOT NULL COMMENT '游戏ID',
  `api_key` varchar(200) NOT NULL COMMENT 'API密钥',
  `secret_key` varchar(200) NOT NULL COMMENT '秘钥',
  `base_url` varchar(200) NOT NULL COMMENT 'API基础URL',
  `push_url` varchar(200) NOT NULL DEFAULT '' COMMENT '推送域名URL',
  `timeout` int(11) NOT NULL DEFAULT '30' COMMENT '请求超时时间(秒)',
  `retry_count` int(11) NOT NULL DEFAULT '3' COMMENT '重试次数',
  `enable_log` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否开启日志',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '状态 1启用 0禁用',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Yalla SDK配置表';

-- 创建 Yalla API调用日志表
CREATE TABLE IF NOT EXISTS `yalla_call_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `app_id` varchar(50) NOT NULL COMMENT '应用ID',
  `user_id` varchar(100) DEFAULT NULL COMMENT '用户ID',
  `method` varchar(20) NOT NULL COMMENT '请求方法',
  `endpoint` varchar(200) NOT NULL COMMENT '接口端点',
  `request_data` text COMMENT '请求数据',
  `response_data` text COMMENT '响应数据',
  `status_code` int(11) DEFAULT NULL COMMENT 'HTTP状态码',
  `duration` bigint(20) DEFAULT NULL COMMENT '请求耗时(毫秒)',
  `success` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否成功',
  `error_msg` varchar(500) DEFAULT NULL COMMENT '错误信息',
  `created_at` datetime NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Yalla API调用日志表';

-- 创建 Yalla 用户绑定表
CREATE TABLE IF NOT EXISTS `yalla_user_bindings` (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `app_id` varchar(50) NOT NULL COMMENT '应用ID',
  `game_user_id` varchar(100) NOT NULL COMMENT '游戏用户ID',
  `yalla_user_id` varchar(100) NOT NULL COMMENT 'Yalla用户ID',
  `yalla_token` varchar(500) DEFAULT NULL COMMENT 'Yalla用户令牌',
  `expires_at` datetime DEFAULT NULL COMMENT '令牌过期时间',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT '状态 1有效 0无效',
  `bind_at` datetime NOT NULL COMMENT '绑定时间',
  `updated_at` datetime NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_game_user` (`app_id`, `game_user_id`),
  KEY `idx_yalla_user_id` (`yalla_user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Yalla用户绑定表';

-- 插入测试配置数据
INSERT INTO `yalla_config` (`app_id`, `app_game_id`, `api_key`, `secret_key`, `base_url`, `push_url`, `timeout`, `retry_count`, `enable_log`, `status`, `remark`, `created_at`, `updated_at`) VALUES
('test_app', 'test_api_key_12345', 'test_secret_key_67890', 'https://sdkapitest.yallagame.com', 'https://sdklogapitest.yallagame.com', 30, 3, 1, 1, '测试应用配置', NOW(), NOW()),
('test_app_1', 'prod_api_key_12345', 'prod_secret_key_67890', 'https://sdkapi.yallagame.com', 'https://sdklogapi.yallagame.com', 30, 3, 0, 1, '生产应用配置1', NOW(), NOW()),
('test_app_2', 'dev_api_key_12345', 'dev_secret_key_67890', 'https://sdkapitest.yallagame.com', 'https://sdklogapitest.yallagame.com', 10, 1, 1, 1, '开发应用配置2', NOW(), NOW());

-- 插入测试用户绑定数据
INSERT INTO `yalla_user_bindings` (`app_id`,`game_user_id`, `yalla_user_id`, `yalla_token`, `expires_at`, `status`, `bind_at`, `updated_at`) VALUES
('test_app', 'test_code', 'test_yalla_user_001', 'test_auth_token_123', DATE_ADD(NOW(), INTERVAL 24 HOUR), 1, NOW(), NOW()),
('test_app', 'test_player_001', 'test_yalla_user_002', 'test_auth_token_456', DATE_ADD(NOW(), INTERVAL 24 HOUR), 1, NOW(), NOW()),
('test_app_1', 'player_123', 'yalla_user_123', 'auth_token_123', DATE_ADD(NOW(), INTERVAL 24 HOUR), 1, NOW(), NOW());

-- 创建索引优化查询性能
ALTER TABLE `yalla_call_logs` ADD INDEX `idx_app_success` (`app_id`, `success`);
ALTER TABLE `yalla_call_logs` ADD INDEX `idx_endpoint` (`endpoint`);
ALTER TABLE `yalla_user_bindings` ADD INDEX `idx_expires_at` (`expires_at`);

-- 显示创建结果
SELECT 'Yalla SDK 数据库初始化完成' as message;
SELECT COUNT(*) as config_count FROM yalla_config;
SELECT COUNT(*) as binding_count FROM yalla_user_bindings;
