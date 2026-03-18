-- 为 game_server 增加轻量运行时管理字段
-- 兼容老字段 exe_name/exe_path，支持重复执行

SET @db_name = DATABASE();

SET @manage_mode_exists = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'game_server'
    AND COLUMN_NAME = 'manage_mode'
);

SET @process_name_exists = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'game_server'
    AND COLUMN_NAME = 'process_name'
);

SET @start_command_exists = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'game_server'
    AND COLUMN_NAME = 'start_command'
);

SET @work_dir_exists = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'game_server'
    AND COLUMN_NAME = 'work_dir'
);

SET @ddl_manage_mode = IF(
  @manage_mode_exists = 0,
  'ALTER TABLE `game_server` ADD COLUMN `manage_mode` VARCHAR(32) NOT NULL DEFAULT ''manual'' COMMENT ''运行管理模式 manual/local_command'' AFTER `server_name`;',
  'SELECT ''skip: game_server.manage_mode already exists'';'
);
PREPARE stmt_manage_mode FROM @ddl_manage_mode;
EXECUTE stmt_manage_mode;
DEALLOCATE PREPARE stmt_manage_mode;

SET @ddl_process_name = IF(
  @process_name_exists = 0,
  'ALTER TABLE `game_server` ADD COLUMN `process_name` VARCHAR(128) NOT NULL DEFAULT '''' COMMENT ''进程名'' AFTER `manage_mode`;',
  'SELECT ''skip: game_server.process_name already exists'';'
);
PREPARE stmt_process_name FROM @ddl_process_name;
EXECUTE stmt_process_name;
DEALLOCATE PREPARE stmt_process_name;

SET @ddl_start_command = IF(
  @start_command_exists = 0,
  'ALTER TABLE `game_server` ADD COLUMN `start_command` VARCHAR(512) NOT NULL DEFAULT '''' COMMENT ''启动命令'' AFTER `process_name`;',
  'SELECT ''skip: game_server.start_command already exists'';'
);
PREPARE stmt_start_command FROM @ddl_start_command;
EXECUTE stmt_start_command;
DEALLOCATE PREPARE stmt_start_command;

SET @ddl_work_dir = IF(
  @work_dir_exists = 0,
  'ALTER TABLE `game_server` ADD COLUMN `work_dir` VARCHAR(512) NOT NULL DEFAULT '''' COMMENT ''工作目录'' AFTER `start_command`;',
  'SELECT ''skip: game_server.work_dir already exists'';'
);
PREPARE stmt_work_dir FROM @ddl_work_dir;
EXECUTE stmt_work_dir;
DEALLOCATE PREPARE stmt_work_dir;

UPDATE `game_server`
SET
  `process_name` = CASE
    WHEN TRIM(IFNULL(`process_name`, '')) = '' THEN TRIM(IFNULL(`exe_name`, ''))
    ELSE `process_name`
  END,
  `start_command` = CASE
    WHEN TRIM(IFNULL(`start_command`, '')) = '' THEN TRIM(IFNULL(`exe_path`, ''))
    ELSE `start_command`
  END,
  `manage_mode` = CASE
    WHEN TRIM(IFNULL(`manage_mode`, '')) <> '' THEN `manage_mode`
    WHEN TRIM(IFNULL(`process_name`, '')) <> '' OR TRIM(IFNULL(`start_command`, '')) <> '' THEN 'local_command'
    WHEN TRIM(IFNULL(`exe_name`, '')) <> '' OR TRIM(IFNULL(`exe_path`, '')) <> '' THEN 'local_command'
    ELSE 'manual'
  END;
