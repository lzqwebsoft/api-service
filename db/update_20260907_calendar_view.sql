-- --------------------------------------------------------
-- Update SQL for Calendar View Menu & Permissions
-- Date: 2026-09-07
-- --------------------------------------------------------

USE `api_service`;

-- 1. Add `CalendarView` menu under `Calendar` (parent_id = 7)
INSERT INTO `admin_menus` (`id`, `parent_id`, `name`, `path`, `component`, `title`, `icon`, `is_hide`, `keep_alive`, `is_hide_tab`, `is_full_page`, `fixed_tab`, `sort_order`)
VALUES (17, 7, 'CalendarView', 'view', '/calendar/view', 'menus.calendar.view', 'ri:calendar-event-line', 0, 1, 0, 0, 0, 0)
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `path` = VALUES(`path`), `component` = VALUES(`component`), `title` = VALUES(`title`), `icon` = VALUES(`icon`);

-- 2. Grant permissions to Super Admin (role_id = 1) and Administrator (role_id = 2)
INSERT IGNORE INTO `admin_role_menus` (`role_id`, `menu_id`) VALUES (1, 17), (2, 17);
