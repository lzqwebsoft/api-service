-- --------------------------------------------------------
-- Update SQL for File Management & Download Statistics
-- Date: 2026-07-27
-- --------------------------------------------------------

USE `api_service`;

-- 1. Create `app_files` table
CREATE TABLE IF NOT EXISTS `app_files` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `app_record_id` INT NOT NULL,
    `version` VARCHAR(50) NOT NULL DEFAULT '',
    `file_name` VARCHAR(255) NOT NULL,
    `download_url` TEXT NOT NULL,
    `file_size` BIGINT DEFAULT 0,
    `description` TEXT,
    `download_count` INT DEFAULT 0,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,  -- 1: Enabled, 0: Disabled
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY `idx_app_record_id` (`app_record_id`),
    CONSTRAINT `fk_app_files_app` FOREIGN KEY (`app_record_id`) REFERENCES `apps` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Create `file_download_logs` table
CREATE TABLE IF NOT EXISTS `file_download_logs` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `file_id` INT NOT NULL,
    `ip` VARCHAR(50) DEFAULT '',
    `ip_location` VARCHAR(100) DEFAULT '',
    `user_agent` TEXT,
    `referer` TEXT,
    `channel` VARCHAR(50) DEFAULT '',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    KEY `idx_file_id` (`file_id`),
    CONSTRAINT `fk_download_logs_file` FOREIGN KEY (`file_id`) REFERENCES `app_files` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Add `Files` menu and grant permissions to Admin & User roles
INSERT INTO `admin_menus` (`id`, `parent_id`, `name`, `path`, `component`, `title`, `icon`, `is_hide`, `keep_alive`, `is_hide_tab`, `is_full_page`, `fixed_tab`, `sort_order`)
VALUES (16, 3, 'Files', 'files', '/token/files', 'menus.token.files', 'ri:file-download-line', 0, 1, 0, 0, 0, 5)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `path` = VALUES(`path`), `component` = VALUES(`component`), `title` = VALUES(`title`), `icon` = VALUES(`icon`);

INSERT IGNORE INTO `admin_role_menus` (`role_id`, `menu_id`) VALUES (1, 16), (2, 16);

