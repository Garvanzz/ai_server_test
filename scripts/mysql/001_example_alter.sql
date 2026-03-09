-- 示例：增量迁移（改字段时使用）
-- 使用方式：在 000_schema.sql 中改好表定义后，在此新增 ALTER，按编号顺序执行
-- 本文件仅为示例，实际执行前可删除或重命名为 002_xxx.sql 写真实 ALTER

-- 示例：给 account 表增加字段（若已存在请勿执行）
-- ALTER TABLE account ADD COLUMN example_field VARCHAR(64) NOT NULL DEFAULT '' COMMENT '示例字段';
