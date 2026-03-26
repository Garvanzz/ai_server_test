-- 迁移 006：从 game_server 表移除进程管理相关字段
-- 背景：进程管理数据已迁移到 server_process 表，game_server 仅保留路由/展示字段。
-- 本迁移需在 005_add_server_process_table.sql 完成（含数据迁移）之后执行。
--
-- 移除字段：
--   exe_name       (原始进程文件名，由 schema 初始版本引入)
--   exe_path       (原始进程路径，由 schema 初始版本引入)
--   manage_mode    (管理模式，由 migration 003 引入)
--   process_name   (进程名，由 migration 003 引入)
--   start_command  (启动命令，由 migration 003 引入)
--   work_dir       (工作目录，由 migration 003 引入)
--
-- 注意：迁移 005 中已生成数据迁移参考脚本，请在执行本迁移前确认 server_process 数据完整。

ALTER TABLE `game_server`
    DROP COLUMN IF EXISTS `exe_name`,
    DROP COLUMN IF EXISTS `exe_path`,
    DROP COLUMN IF EXISTS `manage_mode`,
    DROP COLUMN IF EXISTS `process_name`,
    DROP COLUMN IF EXISTS `start_command`,
    DROP COLUMN IF EXISTS `work_dir`;
