-- 服列表表（仅元数据，不含 Redis/MySQL 连接信息）
-- 连接配置仅来自部署层（env/Config），本表只供客户端选服、GM 展示
-- 若沿用现有 servergroup 表，可只保留/使用元数据列，去掉 redis_port、mysql_addr 等

CREATE TABLE IF NOT EXISTS server_list (
  id               BIGINT       NOT NULL PRIMARY KEY COMMENT '服ID',
  serverName       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '服名称',
  server_group     INT          NOT NULL DEFAULT 0 COMMENT '服组',
  serverState      TINYINT      NOT NULL DEFAULT 0 COMMENT '0正常 1拥挤 2爆满 3维护 4未开服 5停服',
  openServerTime   BIGINT       NOT NULL DEFAULT 0 COMMENT '开服时间戳',
  stopServerTime   BIGINT       NOT NULL DEFAULT 0 COMMENT '停服时间戳',
  channel          INT          NOT NULL DEFAULT 0 COMMENT '渠道',
  ip               VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '客户端连接 IP',
  port             INT          NOT NULL DEFAULT 0 COMMENT '客户端连接端口',
  loginServerUrl   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '登录服地址'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服列表(仅元数据)';
