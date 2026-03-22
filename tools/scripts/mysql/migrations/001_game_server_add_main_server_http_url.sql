-- 迁移目标：为 game_server 增加 main_server_http_url 字段
-- 说明：脚本可重复执行；字段已存在时自动跳过

SET @db_name = DATABASE();

SET @column_exists = (
  SELECT COUNT(*)
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
    AND TABLE_NAME = 'game_server'
    AND COLUMN_NAME = 'main_server_http_url'
);

SET @ddl = IF(
  @column_exists = 0,
  'ALTER TABLE `game_server` ADD COLUMN `main_server_http_url` VARCHAR(512) NOT NULL DEFAULT '''' COMMENT ''大厅服 HTTP 地址，GM 转发用'';',
  'SELECT ''skip: game_server.main_server_http_url already exists'';'
);

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
