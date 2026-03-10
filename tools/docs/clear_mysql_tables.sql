-- =============================================================================
-- 当前项目用到的 MySQL 表清空脚本
-- 来源：main_server / login_server / gm_server / core 中所有 .Table(...) 引用
-- 用途：清空业务数据，保留表结构（TRUNCATE）
-- 执行前请确认数据库与账号，建议先在测试库执行
-- =============================================================================

-- 关闭外键检查，避免清空顺序问题（若表有外键）
SET FOREIGN_KEY_CHECKS = 0;

-- 核心 / 主服
TRUNCATE TABLE account;
TRUNCATE TABLE admin_mail;
TRUNCATE TABLE player_mail_info;
TRUNCATE TABLE sys_mail_info;
TRUNCATE TABLE servergroup;
TRUNCATE TABLE paycacheorder;
TRUNCATE TABLE payorder;
TRUNCATE TABLE friend_apply;
TRUNCATE TABLE friend_block;
TRUNCATE TABLE guild;
TRUNCATE TABLE guild_apply;
TRUNCATE TABLE guild_log;

-- 登录服 / GM 服
TRUNCATE TABLE hotupdate;
TRUNCATE TABLE notice;
TRUNCATE TABLE gameserver;
TRUNCATE TABLE admin;

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- =============================================================================
-- 表名与代码定义对照（core/define、login_server/define、gm_server/define）
-- =============================================================================
-- account          -> AccountTable        玩家账号
-- admin            -> Admin               GM 后台账号（gm_server）
-- admin_mail       -> AdminMailTable      后台/延迟邮件
-- player_mail_info -> PlayerMailInfoTable 玩家邮件
-- sys_mail_info    -> SysMailInfo         系统邮件
-- servergroup      -> ServerGroup         服列表/服组
-- paycacheorder    -> PayCacheOrder       支付缓存订单
-- payorder         -> PayOrder           支付订单
-- friend_apply     -> FriendApply         好友申请
-- friend_block     -> FriendBlock         黑名单
-- guild            -> GuildTable         帮派
-- guild_apply      -> GuildApplyTable     帮派申请
-- guild_log        -> GuildLogTable       帮派日志
-- hotupdate        -> HotUpdate           热更（login_server/gm_server）
-- notice           -> Notice              公告（login_server/gm_server）
-- gameserver       -> GameServer          游戏服（gm_server）
-- =============================================================================
