-- ---------------------------------------------------------------------------
-- GM 后台专用表
-- 管理人员登录、鉴权等
-- ---------------------------------------------------------------------------
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- GM 后台管理员表（登录账号、token、权限）
CREATE TABLE IF NOT EXISTS admin (
  id          BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_name   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录账号',
  password    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '密码',
  token       VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'token',
  permission  INT          NOT NULL DEFAULT 0  COMMENT '权限，1=admin+editor 2=admin 3=editor',
  name        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '显示名称',
  UNIQUE KEY uk_user_name (user_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='GM 后台管理员';

SET FOREIGN_KEY_CHECKS = 1;
