CREATE DATABASE IF NOT EXISTS `api_service` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `api_service`;

CREATE TABLE IF NOT EXISTS `apps` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `app_id` VARCHAR(50) NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `version` VARCHAR(50) NOT NULL,
    `token_ttl` INT NOT NULL DEFAULT 3600,       -- Token lifespan in seconds
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,    -- 1: Active, 0: Inactive
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_appid_version` (`app_id`, `version`)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS `tokens` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `token` VARCHAR(255) NOT NULL UNIQUE,
    `app_record_id` INT NOT NULL,
    `expires_at` TIMESTAMP NOT NULL,
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

