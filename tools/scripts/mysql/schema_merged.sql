SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================
-- 1. 账户库表 (account库) - login_server & gm_server 共用
-- ============================================================

-- 玩家账号表 (account)
-- 对应结构体: core/model/account.go Account
CREATE TABLE IF NOT EXISTS `account` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `uid`              VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '用户uid',
    `account`          VARCHAR(128) NOT NULL DEFAULT '' COMMENT '账号',
    `password`         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码',
    `type`             INT          NOT NULL DEFAULT 0  COMMENT '账号类型',
    `nick_name`        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '昵称',
    `create_time`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `online_time`      DATETIME     NULL COMMENT '上线时间',
    `offline_time`     DATETIME     NULL COMMENT '下线时间',
    `device_id`        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '设备ID',
    `is_white_acc`     INT          NOT NULL DEFAULT 0  COMMENT '白名单账号 0不是 1是',
    `login_ban`        BIGINT       NOT NULL DEFAULT 0  COMMENT '登录封禁 0未封禁 其他是封禁结束时间戳',
    `login_ban_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '登录封禁原因',
    `platform`         INT          NOT NULL DEFAULT 0  COMMENT '平台 1pc 2ios 3安卓',
    `redis_id`         BIGINT       NOT NULL DEFAULT 0  COMMENT '玩家ID(dbId)',
    `last_token`       VARCHAR(512) NOT NULL DEFAULT '' COMMENT '上次使用token',
    `system_mail_id`   BIGINT       NOT NULL DEFAULT 0  COMMENT '系统邮件ID',
    `chat_ban`         BIGINT       NOT NULL DEFAULT 0  COMMENT '聊天封禁 0未封禁 其他是封禁结束时间戳',
    `chat_ban_reason`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '聊天封禁原因',
    `server_id`        INT          NOT NULL DEFAULT 0  COMMENT '服务器ID',
    UNIQUE KEY `uk_uid` (`uid`),
    UNIQUE KEY `uk_account` (`account`),
    KEY `idx_server_id` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家账号表';

-- 区服组表 (server_group)
-- 对应结构体: core/model/server.go ServerGroup (带xorm标签)
CREATE TABLE IF NOT EXISTS `server_group` (
    `id`         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name`       VARCHAR(64) NOT NULL DEFAULT '' COMMENT '区服组名称',
    `sort_order` INT         NOT NULL DEFAULT 0  COMMENT '排序',
    KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='区服组表';

-- 游戏服表 (game_server)
-- 对应结构体: core/model/server.go ServerItem (带xorm标签)
CREATE TABLE IF NOT EXISTS `game_server` (
    `id`                   BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `channel`              INT          NOT NULL DEFAULT 0  COMMENT '渠道',
    `group_id`             INT          NOT NULL DEFAULT 0  COMMENT '区服组ID (0=游戏服进程, >0=区服)',
    `ip`                   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'IP地址',
    `port`                 INT          NOT NULL DEFAULT 0  COMMENT '端口',
    `main_server_http_url` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '大厅服HTTP地址，GM转发用',
    `server_state`         INT          NOT NULL DEFAULT 0  COMMENT '服务器状态: 0正常 1拥挤 2爆满 3维护 4未开服 5停服',
    `open_server_time`     BIGINT       NOT NULL DEFAULT 0  COMMENT '开服时间',
    `stop_server_time`     BIGINT       NOT NULL DEFAULT 0  COMMENT '停服时间',
    `server_name`          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '服务器名称',
    `exe_name`             VARCHAR(128) NOT NULL DEFAULT '' COMMENT '可执行文件名',
    `exe_path`             VARCHAR(512) NOT NULL DEFAULT '' COMMENT '可执行文件路径',
    KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='游戏服表';

-- 热更新表 (hot_update)
-- 对应结构体: core/model/hot_update.go HotUpdateItem / gm_server/dto/server.go HotUpdateItem (带xorm标签)
CREATE TABLE IF NOT EXISTS `hot_update` (
    `id`           BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `channel`      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道',
    `channel_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道名称',
    `version`      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '版本号',
    UNIQUE KEY `uk_channel` (`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='热更新配置表';

-- 公告表 (notice)
-- 对应结构体: core/model/notice.go NoticeItem
CREATE TABLE IF NOT EXISTS `notice` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `channel`     INT          NOT NULL DEFAULT 0  COMMENT '渠道 0表示全渠道',
    `server_id`   INT          NOT NULL DEFAULT 0  COMMENT '服务器ID 0表示全服',
    `title`       VARCHAR(256) NOT NULL DEFAULT '' COMMENT '标题',
    `content`     TEXT         COMMENT '内容',
    `expire_time` BIGINT       NOT NULL DEFAULT 0  COMMENT '过期时间',
    `effect_time` BIGINT       NOT NULL DEFAULT 0  COMMENT '生效时间',
    KEY `idx_channel_server` (`channel`, `server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公告表';

-- ============================================================
-- 2. GM 后台专用表 (account库)
-- ============================================================

-- GM管理员表 (admin)
-- 对应结构体: gm_server/dto/gm.go GmAccount
CREATE TABLE IF NOT EXISTS `admin` (
    `id`         BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_name`  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录账号',
    `password`   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码',
    `token`      VARCHAR(512) NOT NULL DEFAULT '' COMMENT 'token',
    `permission` INT          NOT NULL DEFAULT 0  COMMENT '权限 1=admin+editor 2=admin 3=editor',
    `name`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '显示名称',
    UNIQUE KEY `uk_user_name` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='GM管理员表';

-- ============================================================
-- 3. 支付相关表 (account库)
-- ============================================================

-- 支付完成订单表 (pay_order)
-- 对应结构体: core/model/recharge.go RechargeOrder
CREATE TABLE IF NOT EXISTS `pay_order` (
    `id`             BIGINT        NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `order_id`       VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '订单ID',
    `amount`         DECIMAL(10,2) NOT NULL DEFAULT 0   COMMENT '金额',
    `product_id`     VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '商品ID',
    `product_name`   VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '商品名称',
    `user_id`        VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '用户ID',
    `game_user_id`   VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '游戏用户ID',
    `server_id`      VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '服务器ID',
    `payment_time`   VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '支付时间',
    `channel_number` VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '渠道号',
    UNIQUE KEY `uk_order_id` (`order_id`),
    KEY `idx_game_user_id` (`game_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付完成订单表';

-- 支付缓存订单表 (pay_cache_order)
-- 对应结构体: core/model/recharge.go RechargeOrder (同一结构体用于两个表)
CREATE TABLE IF NOT EXISTS `pay_cache_order` (
    `id`             BIGINT        NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `order_id`       VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '订单ID',
    `amount`         DECIMAL(10,2) NOT NULL DEFAULT 0   COMMENT '金额',
    `product_id`     VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '商品ID',
    `product_name`   VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '商品名称',
    `user_id`        VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '用户ID',
    `game_user_id`   VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '游戏用户ID',
    `server_id`      VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '服务器ID',
    `payment_time`   VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '支付时间',
    `channel_number` VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '渠道号',
    UNIQUE KEY `uk_order_id` (`order_id`),
    KEY `idx_game_user_id` (`game_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付缓存订单表';

-- ============================================================
-- 4. 游戏数据库表 (各服独立数据库) - main_server 使用
-- ============================================================

-- 好友申请表 (friend_apply)
-- 对应结构体: core/model/friend.go FriendApply
CREATE TABLE IF NOT EXISTS `friend_apply` (
    `id`         INT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `player_id`  BIGINT       NOT NULL DEFAULT 0  COMMENT '申请发起人ID',
    `target_id`  BIGINT       NOT NULL DEFAULT 0  COMMENT '目标玩家ID',
    `msg`        VARCHAR(256) NOT NULL DEFAULT '' COMMENT '申请消息',
    `create_time` BIGINT      NOT NULL DEFAULT 0  COMMENT '创建时间',
    UNIQUE KEY `uk_player_target` (`player_id`, `target_id`),
    KEY `idx_target_id` (`target_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友申请表';

-- 好友黑名单表 (friend_block)
-- 对应结构体: core/model/friend.go FriendBlock
CREATE TABLE IF NOT EXISTS `friend_block` (
    `id`         INT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `player_id`  BIGINT       NOT NULL DEFAULT 0  COMMENT '玩家ID',
    `target_id`  BIGINT       NOT NULL DEFAULT 0  COMMENT '被拉黑玩家ID',
    `msg`        VARCHAR(256) NOT NULL DEFAULT '' COMMENT '备注',
    UNIQUE KEY `uk_player_target` (`player_id`, `target_id`),
    KEY `idx_player_id` (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友黑名单表';

-- 系统邮件表 (sys_mail_info)
-- 对应结构体: core/model/mail.go SysMailInfo
CREATE TABLE IF NOT EXISTS `sys_mail_info` (
    `id`          BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `mail_infos`  JSON        DEFAULT NULL COMMENT '邮件内容 {lang: {title, content}}',
    `items`       JSON        DEFAULT NULL COMMENT '道具附件 [{itemId, itemType, itemNum}]',
    `create_time` BIGINT      NOT NULL DEFAULT 0 COMMENT '创建时间',
    `expire_time` BIGINT      NOT NULL DEFAULT 0 COMMENT '过期时间',
    `cfg_id`      INT         NOT NULL DEFAULT 0 COMMENT '配置id',
    `params`      JSON        DEFAULT NULL COMMENT '参数',
    `sender_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '发送者名字'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统邮件表';

-- GM邮件表/后台邮件表 (admin_mail)
-- 对应结构体: core/model/mail.go GMMailInfo
CREATE TABLE IF NOT EXISTS `admin_mail` (
    `id`           BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `creator_name` VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建者',
    `create_time`  BIGINT       NOT NULL DEFAULT 0  COMMENT '创建时间',
    `effect_time`  BIGINT       NOT NULL DEFAULT 0  COMMENT '生效时间',
    `cn_title`     VARCHAR(256) NOT NULL DEFAULT '' COMMENT '中文标题',
    `cn_content`   TEXT         COMMENT '中文内容',
    `player_ids`   JSON         DEFAULT NULL COMMENT '玩家ID列表 [1,2,3]',
    `db_ids`   JSON         DEFAULT NULL COMMENT '玩家ID列表 [1,2,3]',
    `status`       INT          NOT NULL DEFAULT 0  COMMENT '状态',
    `type`         INT          NOT NULL DEFAULT 0  COMMENT '类型 1系统 2个人',
    `sender_name`  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '发送者名字',
    `items`        JSON         DEFAULT NULL COMMENT '附件物品 [{id, num, type}]'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='GM邮件表/后台延迟邮件';

-- 玩家邮件表 (player_mail_info)
-- 对应结构体: core/model/mail.go PlayerMailInfo
CREATE TABLE IF NOT EXISTS `player_mail_info` (
    `id`                BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `mail_infos`        JSON        DEFAULT NULL COMMENT '邮件内容',
    `open_time`         BIGINT      NOT NULL DEFAULT 0 COMMENT '开启时间',
    `create_time`       BIGINT      NOT NULL DEFAULT 0 COMMENT '创建时间',
    `items`             JSON        DEFAULT NULL COMMENT '附件',
    `got_item`          TINYINT     NOT NULL DEFAULT 0 COMMENT '是否领取奖励 0否 1是',
    `cfg_id`            INT         NOT NULL DEFAULT 0 COMMENT '配置id',
    `params`            JSON        DEFAULT NULL COMMENT '参数',
    `expire_time`       BIGINT      NOT NULL DEFAULT 0 COMMENT '过期时间',
    `sys_id`            BIGINT      NOT NULL DEFAULT 0 COMMENT '系统邮件id',
    `account_id`        VARCHAR(64) NOT NULL DEFAULT '' COMMENT '玩家account_id',
    `db_id`             BIGINT      NOT NULL DEFAULT 0 COMMENT '玩家id',
    `type`              INT         NOT NULL DEFAULT 0 COMMENT '邮件类型 0默认普通邮件 1联盟邮件',
    `sender_name`       VARCHAR(64) NOT NULL DEFAULT '' COMMENT '发送者名字',
    `is_has_attachment` TINYINT     NOT NULL DEFAULT 0 COMMENT '是否有附件',
    KEY `idx_account_id` (`account_id`),
    KEY `idx_db_id` (`db_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家邮件表';

-- ============================================================
-- 5. 公会系统表 (游戏数据库)
-- ============================================================

-- 公会表 (guild)
-- 说明: define/mysql.go 中定义了 GuildTable，但未找到对应 model 结构体
--       以下定义基于 schema.sql 和常见公会字段
CREATE TABLE IF NOT EXISTS `guild` (
    `id`             BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `notice_board`   TEXT        COMMENT '公告板',
    `guild_name`     VARCHAR(64) NOT NULL DEFAULT '' COMMENT '公会名称',
    `banner`         INT         NOT NULL DEFAULT 0 COMMENT '旗帜',
    `banner_color`   INT         NOT NULL DEFAULT 0 COMMENT '旗帜颜色',
    `level_limit`    INT         NOT NULL DEFAULT 0 COMMENT '等级限制',
    `master`         INT         NOT NULL DEFAULT 0 COMMENT '会长ID',
    `ignore_level_limit` INT     NOT NULL DEFAULT 0 COMMENT '忽略等级限制',
    `max_member_count` INT      NOT NULL DEFAULT 0 COMMENT '最大成员数',
    `cur_member_count` INT      NOT NULL DEFAULT 0 COMMENT '当前成员数',
    `apply_need_approval` INT   NOT NULL DEFAULT 0 COMMENT '申请需要审批',
    `level`          INT         NOT NULL DEFAULT 0 COMMENT '公会等级',
    `exp`            INT         NOT NULL DEFAULT 0 COMMENT '公会经验',
    `member_data`    JSON        DEFAULT NULL COMMENT '成员数据',
    `growth`         INT         NOT NULL DEFAULT 0 COMMENT '成长值',
    `reduce_time`    INT         NOT NULL DEFAULT 0 COMMENT '减少时间',
    `add_suc_rare`   INT         NOT NULL DEFAULT 0 COMMENT '增加成功稀有度',
    `yuanchi`        TEXT        COMMENT '元池',
    `title`          INT         NOT NULL DEFAULT 0 COMMENT '称号'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公会表';

-- 公会申请表 (guild_apply)
-- 说明: define/mysql.go 中定义，未找到对应 model 结构体
CREATE TABLE IF NOT EXISTS `guild_apply` (
    `id`          BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `guild_id`    BIGINT NOT NULL DEFAULT 0 COMMENT '帮会ID',
    `player_id`   BIGINT NOT NULL DEFAULT 0 COMMENT '玩家ID',
    `expiration`  BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间',
    `create_time` BIGINT NOT NULL DEFAULT 0 COMMENT '申请时间',
    UNIQUE KEY `uk_guild_player` (`guild_id`, `player_id`),
    KEY `idx_expiration` (`expiration`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='帮会申请表';

-- 公会日志表 (guild_log)
-- 说明: define/mysql.go 中定义
CREATE TABLE IF NOT EXISTS `guild_log` (
    `id`          BIGINT   NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `guild_id`    BIGINT   NOT NULL DEFAULT 0 COMMENT '公会ID',
    `create_time` BIGINT   NOT NULL DEFAULT 0 COMMENT '发生时间',
    `action`      INT      NOT NULL DEFAULT 0 COMMENT '事件类型',
    `player_id`   JSON     DEFAULT NULL COMMENT '玩家ID [1,2,3]',
    `content`     JSON     DEFAULT NULL COMMENT '参数列表 ["arg1","arg2"]',
    KEY `idx_guild_id` (`guild_id`),
    KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公会日志表';

SET FOREIGN_KEY_CHECKS = 1;