/*
目的：
- 为 pay_order / pay_cache_order 增加 entry_server_id 字段
- 记录充值时玩家使用的入口服ID，用于合服后的审计与 GM 补偿追溯
- 合服后 server_id 统一指向目标逻辑服，但 entry_server_id 保留原始入口信息

适用版本：
- 003_game_server_runtime_fields 之后

回滚方案：
- ALTER TABLE pay_order       DROP COLUMN entry_server_id;
- ALTER TABLE pay_cache_order DROP COLUMN entry_server_id;
*/

SET @db_name = DATABASE();

-- pay_order
SET @col_exists_pay = (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @db_name
      AND TABLE_NAME   = 'pay_order'
      AND COLUMN_NAME  = 'entry_server_id'
);
SET @ddl_pay = IF(
    @col_exists_pay = 0,
    'ALTER TABLE `pay_order` ADD COLUMN `entry_server_id` INT NOT NULL DEFAULT 0 COMMENT ''充值入口服ID（审计/补偿追溯用）'' AFTER `server_id`;',
    'SELECT ''skip: pay_order.entry_server_id already exists'';'
);
PREPARE stmt FROM @ddl_pay;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- pay_cache_order
SET @col_exists_cache = (
    SELECT COUNT(*)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = @db_name
      AND TABLE_NAME   = 'pay_cache_order'
      AND COLUMN_NAME  = 'entry_server_id'
);
SET @ddl_cache = IF(
    @col_exists_cache = 0,
    'ALTER TABLE `pay_cache_order` ADD COLUMN `entry_server_id` INT NOT NULL DEFAULT 0 COMMENT ''充值入口服ID（审计/补偿追溯用）'' AFTER `server_id`;',
    'SELECT ''skip: pay_cache_order.entry_server_id already exists'';'
);
PREPARE stmt FROM @ddl_cache;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
