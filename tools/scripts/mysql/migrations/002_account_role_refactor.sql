/*
目的：
- 将 account 从“账号+角色混合表”拆分为“全局账号表”
- 新增 account_role，支持同一账号在多个入口服并存角色
- 为合服后的“入口不变、角色不变、逻辑服合并”提供结构基础

适用版本：
- 旧版 schema_full / 001_game_server_add_main_server_http_url 之后

回滚方案：
- 需依赖执行前数据库备份恢复；本脚本不提供自动回滚
*/

ALTER TABLE `account`
    ADD COLUMN IF NOT EXISTS `last_login_entry_server_id` INT NOT NULL DEFAULT 0 COMMENT '最近一次登录入口服ID' AFTER `chat_ban_reason`,
    ADD COLUMN IF NOT EXISTS `last_login_logic_server_id` INT NOT NULL DEFAULT 0 COMMENT '最近一次登录逻辑服ID' AFTER `last_login_entry_server_id`;

CREATE TABLE IF NOT EXISTS `account_role` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
    `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '账号主表ID',
    `uid` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '平台用户唯一标识',
    `entry_server_id` INT NOT NULL DEFAULT 0 COMMENT '入口服ID',
    `logic_server_id` INT NOT NULL DEFAULT 0 COMMENT '当前逻辑服ID',
    `origin_server_id` INT NOT NULL DEFAULT 0 COMMENT '来源服ID',
    `nick_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '角色昵称',
    `redis_id` BIGINT NOT NULL DEFAULT 0 COMMENT '玩家ID(dbId)',
    `system_mail_id` BIGINT NOT NULL DEFAULT 0 COMMENT '已处理的最大系统邮件ID（按角色）',
    `last_token` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '该角色最近登录 token',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `online_time` DATETIME NULL COMMENT '上线时间',
    `offline_time` DATETIME NULL COMMENT '下线时间',
    `last_login_time` DATETIME NULL COMMENT '最近登录时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uid_entry_server` (`uid`, `entry_server_id`),
    UNIQUE KEY `uk_redis_id` (`redis_id`),
    KEY `idx_account_id` (`account_id`),
    KEY `idx_logic_uid` (`logic_server_id`, `uid`),
    KEY `idx_logic_redis` (`logic_server_id`, `redis_id`),
    KEY `idx_entry_logic` (`entry_server_id`, `logic_server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家角色映射表';

INSERT INTO `account_role` (
    `account_id`, `uid`, `entry_server_id`, `logic_server_id`, `origin_server_id`, `nick_name`, `redis_id`, `system_mail_id`, `last_token`, `create_time`, `online_time`, `offline_time`, `last_login_time`
)
SELECT
    `id`,
    `uid`,
    CASE WHEN `origin_server_id` > 0 THEN `origin_server_id` ELSE `server_id` END,
    `server_id`,
    CASE WHEN `origin_server_id` > 0 THEN `origin_server_id` ELSE `server_id` END,
    `nick_name`,
    `redis_id`,
    `system_mail_id`,
    `last_token`,
    `create_time`,
    `online_time`,
    `offline_time`,
    `online_time`
FROM `account`
WHERE NOT EXISTS (
    SELECT 1
    FROM `account_role` r
    WHERE r.`uid` = `account`.`uid`
      AND r.`entry_server_id` = CASE WHEN `account`.`origin_server_id` > 0 THEN `account`.`origin_server_id` ELSE `account`.`server_id` END
);

UPDATE `account` a
INNER JOIN `account_role` r ON r.`account_id` = a.`id`
SET a.`last_login_entry_server_id` = r.`entry_server_id`,
    a.`last_login_logic_server_id` = r.`logic_server_id`
WHERE a.`last_login_entry_server_id` = 0
   OR a.`last_login_logic_server_id` = 0;
