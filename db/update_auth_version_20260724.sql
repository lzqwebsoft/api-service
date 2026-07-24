-- --------------------------------------------------------
-- Update SQL for 2026-07-24 database migrations
-- --------------------------------------------------------

USE `api_service`;

-- 1. Modify `apps` table
-- Drop the original unique index on (app_id, version)
ALTER TABLE `apps` DROP INDEX `idx_appid_version`;

-- Drop the `version` column from `apps`
ALTER TABLE `apps` DROP COLUMN `version`;

-- Add unique index constraint on `app_id`
ALTER TABLE `apps` ADD UNIQUE KEY `idx_appid` (`app_id`);


-- 2. Modify `tokens` table
-- Add `version` constraint and comparison operator
ALTER TABLE `tokens` 
  ADD COLUMN `version` VARCHAR(50) NOT NULL DEFAULT '' AFTER `platform`,
  ADD COLUMN `version_operator` VARCHAR(10) NOT NULL DEFAULT '=' AFTER `version`;
-- Update existing tokens to default version constraint '1.0.0'
UPDATE `tokens` SET `version` = '1.0.0';

-- 3. Modify `token_access_logs` table
-- Add column to log client actual request version
ALTER TABLE `token_access_logs` 
  ADD COLUMN `version` VARCHAR(50) NOT NULL DEFAULT '' AFTER `ip_location`;
UPDATE `token_access_logs` SET `version` = '1.0.0';

-- 4. Modify `user_feedback` table
-- Add column to store actual client version during feedback submission
ALTER TABLE `user_feedback` 
  ADD COLUMN `version` VARCHAR(50) NOT NULL DEFAULT '' AFTER `ip_location`;
UPDATE `user_feedback` SET `version` = '1.0.0';