-- 清空：删除当前库中所有业务表（执行后需再执行 000_schema.sql 建表）
-- 执行前请确认库与权限，建议先在测试库执行
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS account;
DROP TABLE IF EXISTS admin_mail;
DROP TABLE IF EXISTS player_mail_info;
DROP TABLE IF EXISTS sys_mail_info;
DROP TABLE IF EXISTS servergroup;
DROP TABLE IF EXISTS paycacheorder;
DROP TABLE IF EXISTS payorder;
DROP TABLE IF EXISTS friend_apply;
DROP TABLE IF EXISTS friend_block;
DROP TABLE IF EXISTS guild;
DROP TABLE IF EXISTS guild_apply;
DROP TABLE IF EXISTS guild_log;
DROP TABLE IF EXISTS hotupdate;
DROP TABLE IF EXISTS notice;
DROP TABLE IF EXISTS gameserver;
DROP TABLE IF EXISTS admin;

SET FOREIGN_KEY_CHECKS = 1;
