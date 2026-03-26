/*
目的：
- 新增 server_process 表，将进程管理职责从 game_server 中解耦
- game_server 只保留路由/展示数据（ip/port/logic_server_id/server_state 等）
- server_process 统一管理所有进程生命周期：login_server / main_server / game_server
- 支持 GM 后台对 login_server 进行启停管理（之前完全缺失）
- 将 build_server.go 中的硬编码路径迁移到数据库配置

适用版本：004_pay_order_add_entry_server_id 之后

回滚方案：DROP TABLE IF EXISTS server_process;
*/

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


/*
存量数据迁移（可选）：将 game_server 中 group_id=0 的进程记录迁入 server_process

INSERT INTO server_process (
    server_type, server_ref_id, server_name, manage_mode,
    process_bin_name, start_command, work_dir, sort_order
)
SELECT
    CASE WHEN group_id = 0 THEN 3 ELSE 2 END AS server_type,
    id                                        AS server_ref_id,
    server_name,
    IFNULL(NULLIF(manage_mode, ''), 'manual') AS manage_mode,
    IFNULL(NULLIF(process_name, ''), exe_name) AS process_bin_name,
    IFNULL(NULLIF(start_command, ''), exe_path) AS start_command,
    work_dir,
    id*10 AS sort_order
FROM game_server
WHERE process_name != '' OR exe_name != '' OR start_command != '' OR exe_path != '';

存量迁移后可以考虑清理 game_server 中的进程管理字段（manage_mode/process_name/start_command/work_dir/exe_name/exe_path）。
*/
