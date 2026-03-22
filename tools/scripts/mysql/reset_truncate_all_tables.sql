-- 清空业务数据，保留表结构
-- 适用场景：联调/测试环境重置数据，不重建 schema

SET FOREIGN_KEY_CHECKS = 0;

TRUNCATE TABLE account;
TRUNCATE TABLE server_group;
TRUNCATE TABLE game_server;
TRUNCATE TABLE hot_update;
TRUNCATE TABLE notice;
TRUNCATE TABLE admin;

TRUNCATE TABLE pay_order;
TRUNCATE TABLE pay_cache_order;

TRUNCATE TABLE sys_mail_info;
TRUNCATE TABLE admin_mail;
TRUNCATE TABLE player_mail_info;

TRUNCATE TABLE guild;
TRUNCATE TABLE guild_apply;
TRUNCATE TABLE guild_log;
TRUNCATE TABLE friend_apply;
TRUNCATE TABLE friend_block;

TRUNCATE TABLE merge_plan;
TRUNCATE TABLE merge_server_map;
TRUNCATE TABLE merge_conflict_log;

SET FOREIGN_KEY_CHECKS = 1;
