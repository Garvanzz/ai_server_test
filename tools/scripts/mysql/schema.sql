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
  nick_name        VARCHAR(64)  NOT NULL DEFAULT '',
  create_time      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  online_time      DATETIME     NULL,
  offline_time     DATETIME     NULL,
  device_id        VARCHAR(128) NOT NULL DEFAULT '',
  is_white_acc     INT          NOT NULL DEFAULT 0,
  login_ban        BIGINT       NOT NULL DEFAULT 0,
  login_ban_reason VARCHAR(255) NOT NULL DEFAULT '',
  platform         INT          NOT NULL DEFAULT 0,
  redis_id         BIGINT       NOT NULL DEFAULT 0,
  last_token       VARCHAR(128) NOT NULL DEFAULT '',
  system_mail_id    BIGINT      NOT NULL DEFAULT 0,
  chat_ban         BIGINT       NOT NULL DEFAULT 0,
  chat_ban_reason  VARCHAR(255) NOT NULL DEFAULT '',
  server_id        INT          NOT NULL DEFAULT 0,
  UNIQUE KEY uk_uid (uid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '玩家账号';

CREATE TABLE IF NOT EXISTS game_server (
  id                   BIGINT       NOT NULL PRIMARY KEY,
  channel              INT          NOT NULL DEFAULT 0,
  ip                   VARCHAR(64)  NOT NULL DEFAULT '',
  port                 INT          NOT NULL DEFAULT 0,
  server_state         TINYINT      NOT NULL DEFAULT 0,
  open_server_time     BIGINT       NOT NULL DEFAULT 0,
  stop_server_time     BIGINT       NOT NULL DEFAULT 0,
  group_id             INT          NOT NULL DEFAULT 0,
  server_name          VARCHAR(64)  NOT NULL DEFAULT '',
  exe_name             VARCHAR(128) NOT NULL DEFAULT '',
  exe_path             VARCHAR(255) NOT NULL DEFAULT '',
  main_server_http_url VARCHAR(128) NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '区服表';

CREATE TABLE IF NOT EXISTS server_group (
    id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(64) NOT NULL DEFAULT '',
    sort_order INT         NOT NULL DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '区服组表';

-- ---------------------------------------------------------------------------
-- 邮件
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS sys_mail_info (
  id          BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  mail_infos  JSON        DEFAULT NULL,
  items       JSON        DEFAULT NULL,
  create_time BIGINT      NOT NULL DEFAULT 0,
  expire_time BIGINT      NOT NULL DEFAULT 0,
  cfg_id      INT         NOT NULL DEFAULT 0,
  params      JSON        DEFAULT NULL,
  sender_name VARCHAR(64) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '系统邮件';

CREATE TABLE IF NOT EXISTS admin_mail (
  id           BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  type         INT         NOT NULL DEFAULT 0,
  player_ids   JSON        DEFAULT NULL,
  db_ids       JSON        DEFAULT NULL,
  cn_title     VARCHAR(255) NOT NULL DEFAULT '',
  cn_content   TEXT,
  en_title     VARCHAR(255) NOT NULL DEFAULT '',
  en_content   TEXT,
  items        JSON        DEFAULT NULL,
  effect_time  DATETIME    NOT NULL,
  create_time  DATETIME    NOT NULL,
  creator_name VARCHAR(64) NOT NULL DEFAULT '',
  status       TINYINT     NOT NULL DEFAULT 0,
  sender_name  VARCHAR(64) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '后台/延迟邮件';

CREATE TABLE IF NOT EXISTS player_mail_info (
  id                BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
  mail_infos        JSON        DEFAULT NULL,
  open_time         BIGINT      NOT NULL DEFAULT 0,
  create_time       BIGINT      NOT NULL DEFAULT 0,
  items             JSON        DEFAULT NULL,
  got_item          TINYINT     NOT NULL DEFAULT 0,
  cfg_id            INT         NOT NULL DEFAULT 0,
  params            JSON        DEFAULT NULL,
  expire_time       BIGINT      NOT NULL DEFAULT 0,
  sys_id            BIGINT      NOT NULL DEFAULT 0,
  account_id        VARCHAR(64) NOT NULL DEFAULT '',
  db_id             BIGINT      NOT NULL DEFAULT 0,
  type              INT         NOT NULL DEFAULT 0,
  sender_name       VARCHAR(64) NOT NULL DEFAULT '',
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
  id                  BIGINT   NOT NULL PRIMARY KEY,
  notice_board        TEXT,
  guild_name          VARCHAR(64) NOT NULL DEFAULT '',
  banner              INT      NOT NULL DEFAULT 0,
  banner_color        INT      NOT NULL DEFAULT 0,
  level_limit         INT      NOT NULL DEFAULT 0,
  master              INT      NOT NULL DEFAULT 0,
  ignore_level_limit  INT      NOT NULL DEFAULT 0,
  max_member_count    INT      NOT NULL DEFAULT 0,
  cur_member_count    INT      NOT NULL DEFAULT 0,
  apply_need_approval INT     NOT NULL DEFAULT 0,
  level               INT      NOT NULL DEFAULT 0,
  exp                 INT      NOT NULL DEFAULT 0,
  member_data         JSON    DEFAULT NULL,
  growth              INT      NOT NULL DEFAULT 0,
  reduce_time         INT      NOT NULL DEFAULT 0,
  add_suc_rare        INT      NOT NULL DEFAULT 0,
  yuanchi             TEXT,
  title               INT      NOT NULL DEFAULT 0
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
CREATE TABLE IF NOT EXISTS pay_order (
  id            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_id      VARCHAR(64)  NOT NULL DEFAULT '',
  game_user_id  VARCHAR(64)  NOT NULL DEFAULT '',
  amount        DECIMAL(10,2) NOT NULL DEFAULT 0,
  product_id    VARCHAR(64)  NOT NULL DEFAULT '',
  status        TINYINT      NOT NULL DEFAULT 0,
  created_at    DATETIME     NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '支付订单';

CREATE TABLE IF NOT EXISTS pay_cache_order (
  id            BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  order_id      VARCHAR(64)  NOT NULL DEFAULT '',
  game_user_id  VARCHAR(64)  NOT NULL DEFAULT '',
  data          JSON         DEFAULT NULL,
  created_at    DATETIME     NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '支付缓存订单';

-- ---------------------------------------------------------------------------
-- 登录服 / GM 服
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS hot_update (
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
