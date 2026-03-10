-- 为 game_server 表增加 main_server HTTP 地址字段，供 GM 多区服转发使用
-- 执行后请为各区服配置对应大厅服地址，例如：http://192.168.1.10:9505

ALTER TABLE game_server
ADD COLUMN main_server_http_url VARCHAR(512) NOT NULL DEFAULT '' COMMENT '大厅服 HTTP 地址，GM 转发用' AFTER login_server_url;
