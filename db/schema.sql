CREATE DATABASE IF NOT EXISTS `api_service` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `api_service`;

CREATE TABLE IF NOT EXISTS `apps` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `app_id` VARCHAR(50) NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `version` VARCHAR(50) NOT NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,    -- 1: Active, 0: Inactive
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_appid_version` (`app_id`, `version`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `tokens` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token` VARCHAR(255) NOT NULL UNIQUE,
    `app_record_id` INT NOT NULL,
    `platform` VARCHAR(20) NOT NULL,
    `is_revoked` TINYINT(1) NOT NULL DEFAULT 0,  -- 1: Revoked, 0: Active
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY `idx_token` (`token`),
    CONSTRAINT `fk_tokens_app` FOREIGN KEY (`app_record_id`) REFERENCES `apps` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_users` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE,
    `password_hash` VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `admin_sessions` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `session_token` VARCHAR(255) NOT NULL UNIQUE,
    `username` VARCHAR(50) NOT NULL,
    `expires_at` TIMESTAMP NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY `idx_session_token` (`session_token`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `token_blacklist` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token` VARCHAR(255) NOT NULL,
    `platform` VARCHAR(20) NOT NULL,
    `version` VARCHAR(50) NOT NULL,
    `user_uuid` VARCHAR(100) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_token_user` (`token`, `user_uuid`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `token_access_logs` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token` VARCHAR(255) NOT NULL,
    `platform` VARCHAR(20) NOT NULL,
    `version` VARCHAR(50) NOT NULL,
    `user_uuid` VARCHAR(100) NOT NULL,
    `ip` VARCHAR(45) NOT NULL,
    `api_path` VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `calendar_exception`  (
  `date` date NOT NULL COMMENT '日期',
  `region` varchar(50) NOT NULL DEFAULT '中国大陆' COMMENT '地区',
  `is_workday` tinyint(1) NOT NULL COMMENT '1=上班，0=休息',
  `holiday_name` varchar(100) DEFAULT NULL COMMENT '特定节假日名称',
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
