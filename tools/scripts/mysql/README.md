# MySQL 脚本规范

## 目录职责

- `schema_full.sql`：全量建表目标态（新环境初始化）
- `reset_drop_all_tables.sql`：删表重建（DROP TABLE）
- `reset_truncate_all_tables.sql`：清空业务数据（TRUNCATE，保留表结构）
- `migrations/NNN_<description>.sql`：增量迁移脚本（按编号顺序执行）
- `migration_conventions.sql`：迁移规范说明入口，不放可执行 SQL

## 执行建议

1. 新环境：`schema_full.sql`
2. 重建：`reset_drop_all_tables.sql` -> `schema_full.sql`
3. 清数：`reset_truncate_all_tables.sql`
4. 增量：依次执行 `migrations/` 编号脚本

## 编写要求

- 优先编写幂等脚本（可重复执行）
- 文件头写明：迁移目的、适用版本、回滚方案
- 不在 `tools/docs` 放置可执行 SQL，文档与脚本分离
