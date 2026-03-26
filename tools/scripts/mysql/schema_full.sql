SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

/*
======================================================================
ai_server_test - MySQL 全量建表脚本（合服彻底版）
----------------------------------------------------------------------
设计目标：
1) 所有业务数据统一在共享 MySQL。
2) 业务上“按服隔离”的表统一使用 server_id（逻辑服ID）。
3) 保留 origin_server_id（来源服ID）用于追溯、审计、补偿。
4) game_server 增加 logic_server_id，实现入口服与逻辑服解耦。
5) 无兼容负担：本脚本是目标态，不考虑历史字段兼容。

说明：
- 本脚本默认用于全新环境初始化。
- 若要重建，请先执行 reset_drop_all_tables.sql 再执行本文件。
- 所有 JSON 字段均使用 MySQL JSON 类型。
======================================================================
*/

-- ============================================================
-- 0. 公共约定
-- ============================================================
-- server_id: 当前逻辑服ID（业务读写都必须带）
-- origin_server_id: 数据来源服ID（首次写入时设置，后续不改）


-- ============================================================
-- 1) 登录/区服/平台相关（account 库）
-- ============================================================

-- 1.1 玩家账号与角色映射
-- 说明：
-- - 一个 uid 可在多个逻辑服拥有角色，因此唯一键使用 (uid, server_id)
-- - account 字段保留全局唯一，避免同名账号注册冲突
-- 说明：
-- - account 为全局唯一账号主体，uid 全局唯一，不含服务器/角色字段
-- - account_role 为每入口服下的角色映射，支持一账号多服多角色
CREATE TABLE IF NOT EXISTS `account` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `uid`              VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '平台用户唯一标识（全局唯一）',
    `account`          VARCHAR(128) NOT NULL DEFAULT '' COMMENT '账号名',
    `password`         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码',
    `type`             INT          NOT NULL DEFAULT 0 COMMENT '账号类型',
    `create_time`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `online_time`      DATETIME     NULL COMMENT '上线时间',
    `offline_time`     DATETIME     NULL COMMENT '下线时间',
    `device_id`        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '设备ID',
    `is_white_acc`     TINYINT      NOT NULL DEFAULT 0 COMMENT '白名单账号 0否 1是',
    `login_ban`        BIGINT       NOT NULL DEFAULT 0 COMMENT '登录封禁结束时间戳, 0未封禁',
    `login_ban_reason` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '登录封禁原因',
    `platform`         INT          NOT NULL DEFAULT 0 COMMENT '平台 1pc 2ios 3安卓',
    `chat_ban`         BIGINT       NOT NULL DEFAULT 0 COMMENT '聊天封禁结束时间戳, 0未封禁',
    `chat_ban_reason`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '聊天封禁原因',
    `last_login_entry_server_id` INT NOT NULL DEFAULT 0 COMMENT '最近一次登录入口服ID',
    `last_login_logic_server_id` INT NOT NULL DEFAULT 0 COMMENT '最近一次登录逻辑服ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uid` (`uid`),
    UNIQUE KEY `uk_account` (`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家全局账号表';


-- 1.2 玩家角色映射表
-- 说明：
-- - 一个 uid 可在多个入口服拥有角色，唯一键使用 (uid, entry_server_id)
-- - entry_server_id：客户端选服ID，合服后永不改变
-- - logic_server_id：实际业务归属服，合服时只需改此字段
-- - origin_server_id：首次创建时的服ID，写入后不再修改，用于审计追溯
CREATE TABLE IF NOT EXISTS `account_role` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `account_id`       BIGINT       NOT NULL DEFAULT 0 COMMENT '账号主表ID',
    `uid`              VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '平台用户唯一标识',
    `entry_server_id`  INT          NOT NULL DEFAULT 0 COMMENT '入口服ID（客户端选服ID，合服后不变）',
    `logic_server_id`  INT          NOT NULL DEFAULT 0 COMMENT '当前逻辑服ID（合服后指向目标服）',
    `origin_server_id` INT          NOT NULL DEFAULT 0 COMMENT '来源服ID（首次创建时写入，后续不改）',
    `nick_name`        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '角色昵称',
    `redis_id`         BIGINT       NOT NULL DEFAULT 0 COMMENT '玩家ID(dbId)，对应Redis中的Player数据',
    `system_mail_id`   BIGINT       NOT NULL DEFAULT 0 COMMENT '已处理的最大系统邮件ID（按角色）',
    `last_token`       VARCHAR(512) NOT NULL DEFAULT '' COMMENT '该角色最近登录 token',
    `create_time`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `online_time`      DATETIME     NULL COMMENT '上线时间',
    `offline_time`     DATETIME     NULL COMMENT '下线时间',
    `last_login_time`  DATETIME     NULL COMMENT '最近登录时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uid_entry_server` (`uid`, `entry_server_id`),
    UNIQUE KEY `uk_redis_id` (`redis_id`),
    KEY `idx_account_id` (`account_id`),
    KEY `idx_logic_uid` (`logic_server_id`, `uid`),
    KEY `idx_logic_redis` (`logic_server_id`, `redis_id`),
    KEY `idx_entry_logic` (`entry_server_id`, `logic_server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家角色映射表（支持一账号多服多角色）';


-- 1.3 区服组（仅用于前端分组展示）
CREATE TABLE IF NOT EXISTS `server_group` (
    `id`          BIGINT      NOT NULL AUTO_INCREMENT COMMENT '分组ID',
    `name`        VARCHAR(64) NOT NULL DEFAULT '' COMMENT '分组名称',
    `sort_order`  INT         NOT NULL DEFAULT 0 COMMENT '排序',
    `group_type`  TINYINT     NOT NULL DEFAULT 0 COMMENT '分组类型 0常规 1推荐 2历史',
    `is_visible`  TINYINT     NOT NULL DEFAULT 1 COMMENT '是否展示 0否 1是',
    PRIMARY KEY (`id`),
    KEY `idx_sort_visible` (`sort_order`, `is_visible`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='区服分组表';


-- 1.3 区服路由表（核心）
-- 说明：
-- - id 是“入口服ID”（客户端选服ID）
-- - logic_server_id 是“逻辑服ID”（业务实际归属）
-- - 合服后可把被合服入口服的 logic_server_id 指向目标服
CREATE TABLE IF NOT EXISTS `game_server` (
    `id`                   BIGINT       NOT NULL AUTO_INCREMENT COMMENT '入口服ID',
    `channel`              INT          NOT NULL DEFAULT 0 COMMENT '渠道',
    `group_id`             INT          NOT NULL DEFAULT 0 COMMENT '区服组ID（0=进程服，>0=展示服）',
    `logic_server_id`      BIGINT       NOT NULL DEFAULT 0 COMMENT '逻辑服ID（未合服时=自身id）',
    `merge_state`          TINYINT      NOT NULL DEFAULT 0 COMMENT '合服状态 0正常 1待合服 2已合服 3回滚中',
    `merge_time`           BIGINT       NOT NULL DEFAULT 0 COMMENT '合服生效时间戳',
    `ip`                   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '入口IP',
    `port`                 INT          NOT NULL DEFAULT 0 COMMENT '入口端口',
    `main_server_http_url` VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'main_server HTTP地址',
    `server_state`         INT          NOT NULL DEFAULT 0 COMMENT '服务器状态: 0正常 1拥挤 2爆满 3维护 4未开服 5停服',
    `open_server_time`     BIGINT       NOT NULL DEFAULT 0 COMMENT '开服时间戳',
    `stop_server_time`     BIGINT       NOT NULL DEFAULT 0 COMMENT '停服时间戳',
    `server_name`          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '服务器名称',
    PRIMARY KEY (`id`),
    KEY `idx_group_id` (`group_id`),
    KEY `idx_logic_server_id` (`logic_server_id`),
    KEY `idx_merge_state` (`merge_state`),
    KEY `idx_channel_group` (`channel`, `group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='区服路由与配置表';


-- 1.4 热更新配置
CREATE TABLE IF NOT EXISTS `hot_update` (
    `id`           BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    `channel`      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道编码',
    `channel_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '渠道名称',
    `version`      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '版本号',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_channel` (`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='热更新配置表';


-- 1.5 公告
-- 说明：server_id = 0 表示全服公告；否则表示指定逻辑服公告
CREATE TABLE IF NOT EXISTS `notice` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `channel`     INT          NOT NULL DEFAULT 0 COMMENT '渠道，0表示全渠道',
    `server_id`   INT          NOT NULL DEFAULT 0 COMMENT '逻辑服ID，0表示全服',
    `title`       VARCHAR(256) NOT NULL DEFAULT '' COMMENT '标题',
    `content`     TEXT         COMMENT '内容',
    `expire_time` BIGINT       NOT NULL DEFAULT 0 COMMENT '过期时间戳',
    `effect_time` BIGINT       NOT NULL DEFAULT 0 COMMENT '生效时间戳',
    PRIMARY KEY (`id`),
    KEY `idx_channel_server` (`channel`, `server_id`),
    KEY `idx_effect_expire` (`effect_time`, `expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公告表';


-- ============================================================
-- 2) GM 管理后台（account 库）
-- ============================================================

CREATE TABLE IF NOT EXISTS `admin` (
    `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `user_name`  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录账号',
    `password`   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码',
    `token`      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '登录token',
    `permission` INT          NOT NULL DEFAULT 0 COMMENT '权限 1=admin+editor 2=admin 3=editor',
    `name`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '显示名称',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_name` (`user_name`),
    KEY `idx_token` (`token`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='GM管理员表';


-- ============================================================
-- 3) 支付相关（account 库）
-- ============================================================

CREATE TABLE IF NOT EXISTS `pay_order` (
    `id`              BIGINT        NOT NULL AUTO_INCREMENT COMMENT '主键',
    `order_id`        VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '订单ID',
    `amount`          DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '金额',
    `product_id`      VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '商品ID',
    `product_name`    VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '商品名称',
    `user_id`         VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '平台用户ID',
    `game_user_id`    VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '游戏用户ID(uid)',
    `server_id`       INT           NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `entry_server_id` INT           NOT NULL DEFAULT 0 COMMENT '充值入口服ID（审计/补偿追溯用）',
    `payment_time`    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '支付时间文本',
    `channel_number`  VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '渠道号',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_id` (`order_id`),
    KEY `idx_server_game_user` (`server_id`, `game_user_id`),
    KEY `idx_server_payment_time` (`server_id`, `payment_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付完成订单表';

CREATE TABLE IF NOT EXISTS `pay_cache_order` (
    `id`              BIGINT        NOT NULL AUTO_INCREMENT COMMENT '主键',
    `order_id`        VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '订单ID',
    `amount`          DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '金额',
    `product_id`      VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '商品ID',
    `product_name`    VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '商品名称',
    `user_id`         VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '平台用户ID',
    `game_user_id`    VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '游戏用户ID(uid)',
    `server_id`       INT           NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `entry_server_id` INT           NOT NULL DEFAULT 0 COMMENT '充值入口服ID（审计/补偿追溯用）',
    `payment_time`    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '支付时间文本',
    `channel_number`  VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '渠道号',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_order_id` (`order_id`),
    KEY `idx_server_game_user` (`server_id`, `game_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付缓存订单表';


-- ============================================================
-- 4) 邮件系统（游戏共享库）
-- ============================================================

-- 4.1 系统邮件
-- 说明：
-- - server_id=0 表示全服系统邮件
-- - server_id>0 表示指定逻辑服系统邮件
CREATE TABLE IF NOT EXISTS `sys_mail_info` (
    `id`               BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_id`        INT         NOT NULL DEFAULT 0 COMMENT '逻辑服ID，0为全服',
    `origin_server_id` INT         NOT NULL DEFAULT 0 COMMENT '来源服ID',
    `mail_infos`       JSON        DEFAULT NULL COMMENT '邮件内容 {lang:{title,content}}',
    `items`            JSON        DEFAULT NULL COMMENT '附件 [{itemId,itemType,itemNum}]',
    `create_time`      BIGINT      NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `expire_time`      BIGINT      NOT NULL DEFAULT 0 COMMENT '过期时间戳',
    `cfg_id`           INT         NOT NULL DEFAULT 0 COMMENT '配置ID',
    `params`           JSON        DEFAULT NULL COMMENT '扩展参数',
    `sender_name`      VARCHAR(64) NOT NULL DEFAULT '' COMMENT '发送者名称',
    PRIMARY KEY (`id`),
    KEY `idx_server_expire` (`server_id`, `expire_time`),
    KEY `idx_server_create` (`server_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统邮件表';


-- 4.2 GM延迟邮件
CREATE TABLE IF NOT EXISTS `admin_mail` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_id`        INT          NOT NULL DEFAULT 0 COMMENT '逻辑服ID，0=全服',
    `origin_server_id` INT          NOT NULL DEFAULT 0 COMMENT '来源服ID',
    `creator_name`     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '创建者',
    `create_time`      BIGINT       NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `effect_time`      BIGINT       NOT NULL DEFAULT 0 COMMENT '生效时间戳',
    `cn_title`         VARCHAR(256) NOT NULL DEFAULT '' COMMENT '中文标题',
    `cn_content`       TEXT         COMMENT '中文内容',
    `en_title`         VARCHAR(256) NOT NULL DEFAULT '' COMMENT '英文标题',
    `en_content`       TEXT         COMMENT '英文内容',
    `player_ids`       JSON         DEFAULT NULL COMMENT '账号ID或UID列表',
    `db_ids`           JSON         DEFAULT NULL COMMENT '玩家DBID列表',
    `status`           INT          NOT NULL DEFAULT 0 COMMENT '状态 0待生效 1待发 2已发',
    `type`             INT          NOT NULL DEFAULT 0 COMMENT '类型 1系统 2个人',
    `sender_name`      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '发送者名称',
    `items`            JSON         DEFAULT NULL COMMENT '附件 [{id,num,type}]',
    PRIMARY KEY (`id`),
    KEY `idx_server_status_effect` (`server_id`, `status`, `effect_time`),
    KEY `idx_creator_time` (`creator_name`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='GM延迟邮件表';


-- 4.3 玩家邮件
CREATE TABLE IF NOT EXISTS `player_mail_info` (
    `id`                BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_id`         INT         NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `origin_server_id`  INT         NOT NULL DEFAULT 0 COMMENT '来源服ID',
    `mail_infos`        JSON        DEFAULT NULL COMMENT '邮件内容',
    `open_time`         BIGINT      NOT NULL DEFAULT 0 COMMENT '开启时间戳',
    `create_time`       BIGINT      NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    `items`             JSON        DEFAULT NULL COMMENT '附件',
    `got_item`          TINYINT     NOT NULL DEFAULT 0 COMMENT '是否已领取附件 0否 1是',
    `cfg_id`            INT         NOT NULL DEFAULT 0 COMMENT '配置ID',
    `params`            JSON        DEFAULT NULL COMMENT '扩展参数',
    `expire_time`       BIGINT      NOT NULL DEFAULT 0 COMMENT '过期时间戳',
    `sys_id`            BIGINT      NOT NULL DEFAULT 0 COMMENT '系统邮件ID（个人邮件为0）',
    `account_id`        VARCHAR(64) NOT NULL DEFAULT '' COMMENT '玩家账号/UID',
    `db_id`             BIGINT      NOT NULL DEFAULT 0 COMMENT '玩家ID',
    `type`              INT         NOT NULL DEFAULT 0 COMMENT '邮件类型 0普通 1联盟',
    `sender_name`       VARCHAR(64) NOT NULL DEFAULT '' COMMENT '发送者名称',
    `is_has_attachment` TINYINT     NOT NULL DEFAULT 0 COMMENT '是否包含交易附件',
    PRIMARY KEY (`id`),
    KEY `idx_server_db_id` (`server_id`, `db_id`, `id`),
    KEY `idx_server_sys_id` (`server_id`, `sys_id`),
    KEY `idx_server_account_id` (`server_id`, `account_id`),
    KEY `idx_server_expire` (`server_id`, `expire_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='玩家邮件表';


-- ============================================================
-- 5) 公会系统（游戏共享库）
-- ============================================================

CREATE TABLE IF NOT EXISTS `guild` (
    `id`                 BIGINT      NOT NULL AUTO_INCREMENT COMMENT '公会ID',
    `server_id`          INT         NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `origin_server_id`   INT         NOT NULL DEFAULT 0 COMMENT '来源服ID',
    `notice_board`       TEXT        COMMENT '公告板',
    `guild_name`         VARCHAR(64) NOT NULL DEFAULT '' COMMENT '公会名称',
    `banner`             INT         NOT NULL DEFAULT 0 COMMENT '旗帜',
    `banner_color`       INT         NOT NULL DEFAULT 0 COMMENT '旗帜颜色',
    `level_limit`        INT         NOT NULL DEFAULT 0 COMMENT '入会等级限制',
    `master`             BIGINT      NOT NULL DEFAULT 0 COMMENT '会长玩家ID',
    `ignore_level_limit` INT         NOT NULL DEFAULT 0 COMMENT '是否忽略等级限制',
    `max_member_count`   INT         NOT NULL DEFAULT 0 COMMENT '最大成员数',
    `cur_member_count`   INT         NOT NULL DEFAULT 0 COMMENT '当前成员数',
    `apply_need_approval` INT        NOT NULL DEFAULT 0 COMMENT '申请是否需审批',
    `level`              INT         NOT NULL DEFAULT 0 COMMENT '公会等级',
    `exp`                INT         NOT NULL DEFAULT 0 COMMENT '公会经验',
    `member_data`        JSON        DEFAULT NULL COMMENT '成员数据',
    `growth`             INT         NOT NULL DEFAULT 0 COMMENT '成长值',
    `reduce_time`        INT         NOT NULL DEFAULT 0 COMMENT '减少时长',
    `add_suc_rare`       INT         NOT NULL DEFAULT 0 COMMENT '增加成功率',
    `yuanchi`            TEXT        COMMENT '元池数据',
    `title`              INT         NOT NULL DEFAULT 0 COMMENT '主题/称号',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_server_guild_name` (`server_id`, `guild_name`),
    KEY `idx_server_id` (`server_id`),
    KEY `idx_server_master` (`server_id`, `master`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公会表';


CREATE TABLE IF NOT EXISTS `guild_apply` (
    `id`          BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_id`   INT    NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `guild_id`    BIGINT NOT NULL DEFAULT 0 COMMENT '公会ID',
    `player_id`   BIGINT NOT NULL DEFAULT 0 COMMENT '玩家ID',
    `expiration`  BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间戳',
    `create_time` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_server_guild_player` (`server_id`, `guild_id`, `player_id`),
    KEY `idx_server_expiration` (`server_id`, `expiration`),
    KEY `idx_server_player` (`server_id`, `player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公会申请表';


CREATE TABLE IF NOT EXISTS `guild_log` (
    `id`          BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_id`   INT    NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `guild_id`    BIGINT NOT NULL DEFAULT 0 COMMENT '公会ID',
    `create_time` BIGINT NOT NULL DEFAULT 0 COMMENT '发生时间戳',
    `action`      INT    NOT NULL DEFAULT 0 COMMENT '事件类型',
    `player_id`   JSON   DEFAULT NULL COMMENT '玩家ID数组',
    `content`     JSON   DEFAULT NULL COMMENT '参数数组',
    PRIMARY KEY (`id`),
    KEY `idx_server_guild_time` (`server_id`, `guild_id`, `create_time`),
    KEY `idx_server_time` (`server_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公会日志表';


-- ============================================================
-- 6) 好友系统（游戏共享库）
-- ============================================================

CREATE TABLE IF NOT EXISTS `friend_apply` (
    `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_id`   INT          NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `player_id`   BIGINT       NOT NULL DEFAULT 0 COMMENT '申请人ID',
    `target_id`   BIGINT       NOT NULL DEFAULT 0 COMMENT '目标玩家ID',
    `msg`         VARCHAR(256) NOT NULL DEFAULT '' COMMENT '申请消息',
    `create_time` BIGINT       NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_server_player_target` (`server_id`, `player_id`, `target_id`),
    KEY `idx_server_target` (`server_id`, `target_id`, `id`),
    KEY `idx_server_create` (`server_id`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友申请表';


CREATE TABLE IF NOT EXISTS `friend_block` (
    `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_id`  INT          NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `player_id`  BIGINT       NOT NULL DEFAULT 0 COMMENT '玩家ID',
    `target_id`  BIGINT       NOT NULL DEFAULT 0 COMMENT '被拉黑玩家ID',
    `msg`        VARCHAR(256) NOT NULL DEFAULT '' COMMENT '备注',
    `create_time` BIGINT      NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_server_player_target` (`server_id`, `player_id`, `target_id`),
    KEY `idx_server_player` (`server_id`, `player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友黑名单表';


-- ============================================================
-- 7) 合服任务与审计（GM 工单化）
-- ============================================================

CREATE TABLE IF NOT EXISTS `merge_plan` (
    `id`               BIGINT       NOT NULL AUTO_INCREMENT COMMENT '计划ID',
    `name`             VARCHAR(128) NOT NULL DEFAULT '' COMMENT '计划名称',
    `target_server_id` INT          NOT NULL DEFAULT 0 COMMENT '目标逻辑服ID',
    `source_server_ids` JSON        DEFAULT NULL COMMENT '来源入口服ID数组',
    `status`           TINYINT      NOT NULL DEFAULT 0 COMMENT '状态 0待执行 1执行中 2成功 3失败 4已回滚',
    `operator`         VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '操作人',
    `start_time`       BIGINT       NOT NULL DEFAULT 0 COMMENT '执行开始时间戳',
    `end_time`         BIGINT       NOT NULL DEFAULT 0 COMMENT '执行结束时间戳',
    `rollback_time`    BIGINT       NOT NULL DEFAULT 0 COMMENT '回滚时间戳',
    `remark`           VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    PRIMARY KEY (`id`),
    KEY `idx_status_start` (`status`, `start_time`),
    KEY `idx_target_server` (`target_server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='合服计划表';


CREATE TABLE IF NOT EXISTS `merge_server_map` (
    `id`               BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
    `plan_id`          BIGINT      NOT NULL DEFAULT 0 COMMENT '计划ID',
    `source_server_id` INT         NOT NULL DEFAULT 0 COMMENT '来源入口服ID',
    `target_server_id` INT         NOT NULL DEFAULT 0 COMMENT '目标逻辑服ID',
    `state`            TINYINT     NOT NULL DEFAULT 0 COMMENT '状态 0待处理 1成功 2失败',
    `err_msg`          VARCHAR(512) NOT NULL DEFAULT '' COMMENT '错误信息',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_plan_source` (`plan_id`, `source_server_id`),
    KEY `idx_target_state` (`target_server_id`, `state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='合服映射明细表';


CREATE TABLE IF NOT EXISTS `merge_conflict_log` (
    `id`              BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `plan_id`         BIGINT       NOT NULL DEFAULT 0 COMMENT '计划ID',
    `server_id`       INT          NOT NULL DEFAULT 0 COMMENT '逻辑服ID',
    `conflict_type`   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '冲突类型 guild_name/player_name/data_error',
    `biz_key`         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '业务键，如公会名',
    `old_value`       VARCHAR(512) NOT NULL DEFAULT '' COMMENT '原值',
    `new_value`       VARCHAR(512) NOT NULL DEFAULT '' COMMENT '新值',
    `resolved`        TINYINT      NOT NULL DEFAULT 0 COMMENT '是否已解决 0否 1是',
    `created_at`      BIGINT       NOT NULL DEFAULT 0 COMMENT '创建时间戳',
    PRIMARY KEY (`id`),
    KEY `idx_plan_type` (`plan_id`, `conflict_type`),
    KEY `idx_server_resolved` (`server_id`, `resolved`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='合服冲突日志表';


-- ============================================================
-- 8) 进程管理表（与 game_server 展示/路由表解耦）
-- ============================================================
-- 设计目标：
-- - game_server 只负责客户端选服路由（展示、ip/port、logic_server_id）
-- - server_process 统一管理所有服务进程的生命周期（login/main/game/battle）
-- - 支持 login_server 进程管理（之前完全缺失）
-- - 将 build 相关配置（仓库地址、编译目录）从代码硬编码迁到数据库
-- - server_ref_id 关联 game_server.id，便于启动后更新 server_state；login_server 为 0
CREATE TABLE IF NOT EXISTS `server_process` (
    `id`                BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
    `server_type`       TINYINT      NOT NULL DEFAULT 0 COMMENT '进程类型 1=login_server 2=main_server 3=game_server/battle',
    `server_ref_id`     BIGINT       NOT NULL DEFAULT 0 COMMENT '关联 game_server.id（main/game 进程对应的展示服ID，login 为 0）',
    `server_name`       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '进程显示名称',
    `manage_mode`       VARCHAR(32)  NOT NULL DEFAULT 'manual' COMMENT '管理模式 manual=手动 local_command=本地命令',
    `process_bin_name`  VARCHAR(128) NOT NULL DEFAULT '' COMMENT '进程二进制名称（用于 pgrep/tasklist 检测）',
    `start_command`     VARCHAR(512) NOT NULL DEFAULT '' COMMENT '启动命令（完整可执行路径或 shell 命令）',
    `work_dir`          VARCHAR(512) NOT NULL DEFAULT '' COMMENT '工作目录（启动时的 cwd）',
    `http_health_url`   VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'HTTP 健康检查地址（连通性检测，如 http://ip:port）',
    `build_repo_url`    VARCHAR(512) NOT NULL DEFAULT '' COMMENT '代码仓库 URL（空=不支持在线编译）',
    `build_source_dir`  VARCHAR(512) NOT NULL DEFAULT '' COMMENT '编译源码目录（go build 执行目录）',
    `build_output_dir`  VARCHAR(512) NOT NULL DEFAULT '' COMMENT '编译产物复制目标目录',
    `build_output_name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '编译产物文件名',
    `sort_order`        INT          NOT NULL DEFAULT 0 COMMENT '排序',
    `remark`            VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    PRIMARY KEY (`id`),
    KEY `idx_server_type` (`server_type`),
    KEY `idx_server_ref` (`server_ref_id`),
    KEY `idx_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='进程管理表（统一管理 login/main/game 所有服务进程）';


-- ============================================================
-- 9) 合服后推荐执行模板（注释，不自动执行）
-- ============================================================
/*
本模板演示将「入口服 @src」并入「逻辑服 @dst」的完整步骤。
使用前替换 @src / @dst 为实际服ID，例如 SET @src=2, @dst=1。

────────────────────────────────────────────────────────────
Step 0: 停服维护，执行前做全量数据库备份
────────────────────────────────────────────────────────────

────────────────────────────────────────────────────────────
Step 1: 预检查——扫描潜在冲突，写入 merge_conflict_log
────────────────────────────────────────────────────────────
-- 1.1 公会重名检查
SELECT g1.guild_name, g1.server_id AS src_server, g2.server_id AS dst_server
FROM guild g1
JOIN guild g2 ON g1.guild_name = g2.guild_name AND g2.server_id = @dst
WHERE g1.server_id = @src;

-- 1.2 有冲突时先处理重名（追加来源服后缀，并发邮件通知玩家改名）
-- UPDATE guild
--   SET guild_name = CONCAT(guild_name, '_', @src)
-- WHERE server_id = @src
--   AND guild_name IN (SELECT guild_name FROM guild WHERE server_id = @dst);

────────────────────────────────────────────────────────────
Step 2: 路由重定向（ip / port / http_url 必须同步更新）
────────────────────────────────────────────────────────────
UPDATE game_server gs_src
JOIN   game_server gs_dst ON gs_dst.id = @dst
SET    gs_src.logic_server_id      = @dst,
       gs_src.ip                   = gs_dst.ip,
       gs_src.port                 = gs_dst.port,
       gs_src.main_server_http_url = gs_dst.main_server_http_url,
       gs_src.merge_state          = 2,
       gs_src.merge_time           = UNIX_TIMESTAMP()
WHERE  gs_src.id = @src;

────────────────────────────────────────────────────────────
Step 3: account_role 逻辑服路由更新（此步骤之前模板遗漏）
────────────────────────────────────────────────────────────
UPDATE account_role
SET    logic_server_id = @dst
WHERE  logic_server_id = @src;

────────────────────────────────────────────────────────────
Step 4: 全业务表 server_id 迁移
────────────────────────────────────────────────────────────
UPDATE guild
SET    server_id        = @dst,
       origin_server_id = IF(origin_server_id = 0, @src, origin_server_id)
WHERE  server_id = @src;

UPDATE guild_apply      SET server_id = @dst WHERE server_id = @src;
UPDATE guild_log        SET server_id = @dst WHERE server_id = @src;

UPDATE player_mail_info
SET    server_id        = @dst,
       origin_server_id = IF(origin_server_id = 0, @src, origin_server_id)
WHERE  server_id = @src;

UPDATE sys_mail_info
SET    server_id        = @dst,
       origin_server_id = IF(origin_server_id = 0, @src, origin_server_id)
WHERE  server_id = @src AND server_id != 0;  -- server_id=0 为全服邮件，不迁移

UPDATE admin_mail
SET    server_id        = @dst,
       origin_server_id = IF(origin_server_id = 0, @src, origin_server_id)
WHERE  server_id = @src;

UPDATE friend_apply     SET server_id = @dst WHERE server_id = @src;
UPDATE friend_block     SET server_id = @dst WHERE server_id = @src;

UPDATE pay_order        SET server_id = @dst WHERE server_id = @src;
UPDATE pay_cache_order  SET server_id = @dst WHERE server_id = @src;

────────────────────────────────────────────────────────────
Step 5: 排行榜 Redis 数据合并（在 Redis 侧执行，以 rank_power 为例）
────────────────────────────────────────────────────────────
-- ZUNIONSTORE rank_power_{@dst}    2  rank_power_{@src}    rank_power_{@dst}    AGGREGATE MAX
-- ZUNIONSTORE rank_recharge_{@dst} 2  rank_recharge_{@src} rank_recharge_{@dst} AGGREGATE MAX
-- （其余榜单同理，参考 core/define/rank.go 中的 RankTypeToKey）
-- 合并后删除源服 key：DEL rank_power_{@src} rank_recharge_{@src} ...

────────────────────────────────────────────────────────────
Step 6: 更新合服计划状态
────────────────────────────────────────────────────────────
-- UPDATE merge_plan SET status = 2, end_time = UNIX_TIMESTAMP() WHERE id = <plan_id>;

注意：
- Step 2（路由重定向）执行后客户端即可进入新服，务必在维护窗口内完成全部步骤。
- 若中途失败，将 game_server.merge_state 改为 3（回滚中）并回滚路由。
- 合服完成后保留旧入口服记录（merge_state=2），用于历史追溯。
*/

SET FOREIGN_KEY_CHECKS = 1;
