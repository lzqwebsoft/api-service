-- --------------------------------------------------------
-- Update SQL for Admin Session Expiration Index
-- Date: 2026-07-30
-- --------------------------------------------------------

USE `api_service`;

-- 1. Add index on `refresh_expires_at` in `admin_sessions` table to accelerate background session cleanup queries
ALTER TABLE `admin_sessions` ADD INDEX `idx_refresh_expires_at` (`refresh_expires_at`);
