CREATE DATABASE IF NOT EXISTS `api_service` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `api_service`;

CREATE TABLE IF NOT EXISTS `apps` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `app_id` VARCHAR(50) NOT NULL UNIQUE,
    `name` VARCHAR(100) NOT NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,    -- 1: Active, 0: Inactive
    `is_deleted` TINYINT(1) NOT NULL DEFAULT 0,   -- 1: Deleted, 0: Normal
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `tokens` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token` VARCHAR(255) NOT NULL UNIQUE,
    `app_record_id` INT NOT NULL,
    `platform` VARCHAR(20) NOT NULL,
    `version` VARCHAR(50) NOT NULL DEFAULT '',
    `version_operator` VARCHAR(10) NOT NULL DEFAULT '=',
    `is_revoked` TINYINT(1) NOT NULL DEFAULT 0,  -- 1: Revoked, 0: Active
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY `idx_token` (`token`),
    CONSTRAINT `fk_tokens_app` FOREIGN KEY (`app_record_id`) REFERENCES `apps` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_users` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE,
    `password_hash` VARCHAR(255) NOT NULL,
    `nickname` VARCHAR(100) DEFAULT '',
    `real_name` VARCHAR(100) DEFAULT '',
    `email` VARCHAR(100) DEFAULT '',
    `phone` VARCHAR(20) DEFAULT '',
    `gender` INT DEFAULT 1,         -- 1: 男, 0: 女, -1: 未知
    `avatar` VARCHAR(255) DEFAULT '',
    `address` VARCHAR(255) DEFAULT '',
    `description` TEXT,
    `status` INT DEFAULT 1,          -- 1: 在线, 2: 离线, 3: 异常, 4: 注销
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_sessions` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `access_token` VARCHAR(255) NOT NULL UNIQUE,
    `refresh_token` VARCHAR(255) NOT NULL UNIQUE,
    `user_id` INT NOT NULL,
    `access_expires_at` BIGINT NOT NULL,
    `refresh_expires_at` BIGINT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY `idx_access_token` (`access_token`),
    KEY `idx_refresh_token` (`refresh_token`),
    CONSTRAINT `fk_admin_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `admin_users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_roles` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL UNIQUE,
    `code` VARCHAR(50) NOT NULL UNIQUE,
    `description` VARCHAR(255) DEFAULT '',
    `enabled` TINYINT(1) DEFAULT 1,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_user_roles` (
    `user_id` INT NOT NULL,
    `role_id` INT NOT NULL,
    PRIMARY KEY (`user_id`, `role_id`),
    CONSTRAINT `fk_user_roles_user` FOREIGN KEY (`user_id`) REFERENCES `admin_users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_user_roles_role` FOREIGN KEY (`role_id`) REFERENCES `admin_roles` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_menus` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `parent_id` INT DEFAULT 0,
    `name` VARCHAR(100) NOT NULL,
    `path` VARCHAR(255) NOT NULL,
    `component` VARCHAR(255),
    `title` VARCHAR(100) NOT NULL,
    `icon` VARCHAR(100) DEFAULT '',
    `is_hide` TINYINT(1) DEFAULT 0,
    `keep_alive` TINYINT(1) DEFAULT 0,
    `is_hide_tab` TINYINT(1) DEFAULT 0,
    `is_full_page` TINYINT(1) DEFAULT 0,
    `fixed_tab` TINYINT(1) DEFAULT 0,
    `sort_order` INT DEFAULT 0,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_role_menus` (
    `role_id` INT NOT NULL,
    `menu_id` INT NOT NULL,
    PRIMARY KEY (`role_id`, `menu_id`),
    CONSTRAINT `fk_role_menus_role` FOREIGN KEY (`role_id`) REFERENCES `admin_roles` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_role_menus_menu` FOREIGN KEY (`menu_id`) REFERENCES `admin_menus` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_menu_auths` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `menu_id` INT NOT NULL,
    `title` VARCHAR(100) NOT NULL,
    `auth_mark` VARCHAR(100) NOT NULL,
    CONSTRAINT `fk_menu_auths_menu` FOREIGN KEY (`menu_id`) REFERENCES `admin_menus` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `token_blacklist` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token_id` INT NOT NULL,
    `user_uuid` VARCHAR(100) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_token_user` (`token_id`, `user_uuid`),
    CONSTRAINT `fk_token_blacklist_token` FOREIGN KEY (`token_id`) REFERENCES `tokens` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `token_access_logs` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token_id` INT NOT NULL,
    `user_uuid` VARCHAR(100) NOT NULL,
    `ip` VARCHAR(45) NOT NULL,
    `ip_location` VARCHAR(100) DEFAULT '' COMMENT 'IP归属地',
    `version` VARCHAR(50) NOT NULL DEFAULT '',
    `api_path` VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_token_access_logs_token` FOREIGN KEY (`token_id`) REFERENCES `tokens` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `calendar_exception`  (
  `date` date NOT NULL COMMENT '日期',
  `region` varchar(50) NOT NULL DEFAULT '中国大陆' COMMENT '地区',
  `is_workday` tinyint(1) NOT NULL COMMENT '1=上班，0=休息',
  `description` varchar(100) DEFAULT NULL COMMENT '事由，如“春节放假”或“劳动节调休上班”',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`date`,`region`),
  KEY `idx_year` ((year(`date`)))
) ENGINE=InnoDB COMMENT='放假调休例外表';

CREATE TABLE `holiday` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '节日名称',
  `type` varchar(50) NOT NULL COMMENT '类型：solar(固定节日), weekday(变动日期节日), industry(行业节日)',
  `month` int NOT NULL COMMENT '月份 (1-12)',
  `day` int NOT NULL DEFAULT '0' COMMENT '公历日期 (1-31)，如果是星期变动则为0',
  `week_number` int NOT NULL DEFAULT '0' COMMENT '第几个星期 (1-5)',
  `day_of_week` int NOT NULL DEFAULT '0' COMMENT '星期几 (1=周一, 7=周日)',
  `regions` varchar(100) NOT NULL DEFAULT 'cn' COMMENT '适用地区，逗号分隔，如 cn,hk,tw',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='标准节假日定义表';

CREATE TABLE IF NOT EXISTS `user_feedback` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token_id` INT NOT NULL COMMENT '关联 Token ID',
    `user_uuid` VARCHAR(100) DEFAULT '' COMMENT '用户 UUID',
    `content` TEXT NOT NULL COMMENT '反馈意见内容',
    `contact` VARCHAR(255) DEFAULT '' COMMENT '联系方式（邮箱/手机号/微信等）',
    `ip` VARCHAR(45) DEFAULT '' COMMENT '客户端IP',
    `ip_location` VARCHAR(255) DEFAULT '' COMMENT 'IP归属性地',
    `version` VARCHAR(50) NOT NULL DEFAULT '',
    `status` INT DEFAULT 0 COMMENT '处理状态 0:待处理 1:已处理',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT `fk_user_feedback_token` FOREIGN KEY (`token_id`) REFERENCES `tokens` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户意见反馈表';