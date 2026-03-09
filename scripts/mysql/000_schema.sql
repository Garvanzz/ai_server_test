-- 完整建表（唯一真相）：用于清空后全量建表；改字段时先改本文件再写增量迁移
-- 执行前可先执行 clear.sql
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- 账号与服列表
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS account (
  id               BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  uid              VARCHAR(64)  NOT NULL DEFAULT '',
  account          VARCHAR(128) NOT NULL DEFAULT '',
  password         VARCHAR(128) NOT NULL DEFAULT '',
  type             INT          NOT NULL DEFAULT 0,
  nickName         VARCHAR(64)  NOT NULL DEFAULT '',
  createTime       DATETIME     NOT NULL,
  onlineTime       DATETIME     NOT NULL,
  offlineTime      DATETIME     NOT NULL,
  deviceId         VARCHAR(128) NOT NULL DEFAULT '',
  isWhiteAcc       INT          NOT NULL DEFAULT 0,
  loginBan         BIGINT       NOT NULL DEFAULT 0,
  loginBanReason   VARCHAR(255) NOT NULL DEFAULT '',
  platform         INT          NOT NULL DEFAULT 0,
  redisId          BIGINT       NOT NULL DEFAULT 0,
  lastToken        VARCHAR(128) NOT NULL DEFAULT '',
  systemMailId     BIGINT       NOT NULL DEFAULT 0,
  chatBan          BIGINT       NOT NULL DEFAULT 0,
  chatBanReason    VARCHAR(255) NOT NULL DEFAULT '',
  serverId          INT          NOT NULL DEFAULT 0,
  UNIQUE KEY uk_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '玩家账号';

CREATE TABLE IF NOT EXISTS servergroup (
  id               BIGINT       NOT NULL PRIMARY KEY,
  channel          INT          NOT NULL DEFAULT 0,
  ip               VARCHAR(64)  NOT NULL DEFAULT '',
  port             INT          NOT NULL DEFAULT 0,
  redisHost        VARCHAR(128) NOT NULL DEFAULT '',
  redisPort        INT          NOT NULL DEFAULT 0,
  mysqlAddr        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '已弃用，仅兼容',
  serverState      TINYINT      NOT NULL DEFAULT 0,
  openServerTime   BIGINT       NOT NULL DEFAULT 0,
  stopServerTime   BIGINT       NOT NULL DEFAULT 0,
  serverName       VARCHAR(64)  NOT NULL DEFAULT '',
  loginServerUrl   VARCHAR(255) NOT NULL DEFAULT '',
  server_group     INT          NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '服列表(元数据)';

-- ---------------------------------------------------------------------------
-- 邮件
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sys_mail_info (
  id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  mail_infos JSON        DEFAULT NULL,
  items      JSON        DEFAULT NULL,
  createTime BIGINT      NOT NULL DEFAULT 0,
  expireTime BIGINT      NOT NULL DEFAULT 0,
  cfgId      INT         NOT NULL DEFAULT 0,
  params     JSON        DEFAULT NULL,
  senderName VARCHAR(64) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '系统邮件';

CREATE TABLE IF NOT EXISTS admin_mail (
  id          BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  type        INT         NOT NULL DEFAULT 0,
  playerIds   JSON        DEFAULT NULL,
  dbIds       JSON        DEFAULT NULL,
  cnTitle     VARCHAR(255) NOT NULL DEFAULT '',
  cnContent   TEXT,
  enTitle     VARCHAR(255) NOT NULL DEFAULT '',
  enContent   TEXT,
  items       JSON        DEFAULT NULL,
  effectTime  DATETIME    NOT NULL,
  createTime  DATETIME    NOT NULL,
  creatorName VARCHAR(64) NOT NULL DEFAULT '',
  status      TINYINT     NOT NULL DEFAULT 0,
  senderName  VARCHAR(64) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '后台/延迟邮件';

CREATE TABLE IF NOT EXISTS player_mail_info (
  id              BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  mail_infos      JSON        DEFAULT NULL,
  openTime        BIGINT      NOT NULL DEFAULT 0,
  createTime      BIGINT      NOT NULL DEFAULT 0,
  items           JSON        DEFAULT NULL,
  got_item        TINYINT     NOT NULL DEFAULT 0,
  cfgId           INT         NOT NULL DEFAULT 0,
  params          JSON        DEFAULT NULL,
  expireTime      BIGINT      NOT NULL DEFAULT 0,
  sys_id          BIGINT      NOT NULL DEFAULT 0,
  account_id      VARCHAR(64) NOT NULL DEFAULT '',
  db_id           BIGINT      NOT NULL DEFAULT 0,
  type            INT         NOT NULL DEFAULT 0,
  senderName      VARCHAR(64) NOT NULL DEFAULT '',
  is_has_attachment TINYINT   NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '玩家邮件';

-- ---------------------------------------------------------------------------
-- 好友
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS friend_apply (
  id        INT         NOT NULL AUTO_INCREMENT PRIMARY KEY,
  player_id BIGINT      NOT NULL DEFAULT 0,
  target_id BIGINT      NOT NULL DEFAULT 0,
  msg       VARCHAR(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '好友申请';

CREATE TABLE IF NOT EXISTS friend_block (
  id        INT         NOT NULL AUTO_INCREMENT PRIMARY KEY,
  player_id BIGINT      NOT NULL DEFAULT 0,
  target_id BIGINT      NOT NULL DEFAULT 0,
  msg       VARCHAR(255) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '黑名单';

-- ---------------------------------------------------------------------------
-- 帮派
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS guild (
  id                 BIGINT   NOT NULL PRIMARY KEY,
  notice_board       TEXT,
  guild_name         VARCHAR(64) NOT NULL DEFAULT '',
  banner             INT      NOT NULL DEFAULT 0,
  banner_color       INT      NOT NULL DEFAULT 0,
  level_limit        INT      NOT NULL DEFAULT 0,
  master             INT      NOT NULL DEFAULT 0,
  ignore_level_limit INT      NOT NULL DEFAULT 0,
  max_member_count   INT      NOT NULL DEFAULT 0,
  cur_member_count   INT      NOT NULL DEFAULT 0,
  apply_need_approval INT     NOT NULL DEFAULT 0,
  level              INT      NOT NULL DEFAULT 0,
  exp                INT      NOT NULL DEFAULT 0,
  member_data        JSON    DEFAULT NULL,
  growth             INT      NOT NULL DEFAULT 0,
  reduce_time        INT      NOT NULL DEFAULT 0,
  add_suc_rare       INT      NOT NULL DEFAULT 0,
  yuanchi            TEXT,
  title              INT      NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '帮派';

CREATE TABLE IF NOT EXISTS guild_apply (
  id        BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  guild_id  BIGINT NOT NULL DEFAULT 0,
  player_id BIGINT NOT NULL DEFAULT 0,
  state     TINYINT NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '帮派申请';

CREATE TABLE IF NOT EXISTS guild_log (
  id        BIGINT   NOT NULL AUTO_INCREMENT PRIMARY KEY,
  guild_id  BIGINT   NOT NULL DEFAULT 0,
  content   TEXT,
  timestamp BIGINT   NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '帮派日志';

-- ---------------------------------------------------------------------------
-- 支付
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payorder (
  id            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_id      VARCHAR(64)  NOT NULL DEFAULT '',
  game_user_id  VARCHAR(64)  NOT NULL DEFAULT '',
  amount        DECIMAL(10,2) NOT NULL DEFAULT 0,
  product_id    VARCHAR(64)  NOT NULL DEFAULT '',
  status        TINYINT      NOT NULL DEFAULT 0,
  created_at    DATETIME     NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '支付订单';

CREATE TABLE IF NOT EXISTS paycacheorder (
  id            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_id      VARCHAR(64)  NOT NULL DEFAULT '',
  game_user_id  VARCHAR(64)  NOT NULL DEFAULT '',
  data          JSON         DEFAULT NULL,
  created_at    DATETIME     NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '支付缓存订单';

-- ---------------------------------------------------------------------------
-- 登录服 / GM 服
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS hotupdate (
  id      BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  channel INT         NOT NULL DEFAULT 0,
  version VARCHAR(64) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '热更';

CREATE TABLE IF NOT EXISTS notice (
  id        BIGINT   NOT NULL AUTO_INCREMENT PRIMARY KEY,
  title     VARCHAR(128) NOT NULL DEFAULT '',
  content   TEXT,
  start_at  DATETIME NOT NULL,
  end_at    DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '公告';

CREATE TABLE IF NOT EXISTS gameserver (
  id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  serverName VARCHAR(64) NOT NULL DEFAULT '',
  ip         VARCHAR(64) NOT NULL DEFAULT '',
  port       INT         NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '游戏服';

CREATE TABLE IF NOT EXISTS admin (
  id        BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_name VARCHAR(64) NOT NULL DEFAULT '',
  password  VARCHAR(128) NOT NULL DEFAULT '',
  token     VARCHAR(128) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT 'GM后台账号';

SET FOREIGN_KEY_CHECKS = 1;
