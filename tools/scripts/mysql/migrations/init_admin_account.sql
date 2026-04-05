-- ============================================================
-- GM 管理后台初始化 - 默认管理员账户
-- ============================================================
-- 执行方式:
--   mysql -u game -p wenjiacheng < init_admin_account.sql
-- 
-- 或在 MySQL 命令行中直接粘贴本脚本内容
-- 
-- 说明:
--   如果 admin 表为空，将自动插入一个默认管理员账户
--   登录用户名: admin
--   登录密码: admin123
--   权限级别: 2 (admin 角色)

-- 检查是否已有账户存在
SELECT '检查 admin 表状态...' AS '操作';

-- 如果表为空，则插入默认管理员
INSERT INTO admin (user_name, password, permission, name)
SELECT 'admin', MD5('admin123'), 2, 'System Administrator'
WHERE NOT EXISTS (
    SELECT 1 FROM admin LIMIT 1
);

-- 显示操作结果
SELECT 
    CASE 
        WHEN ROW_COUNT() > 0 THEN '✓ 默认管理员账户已创建'
        ELSE '⚠ admin 表中已有账户，跳过初始化'
    END AS '初始化结果';

-- 列出所有管理员账户
SELECT 
    '=' AS '',
    '现有 GM 管理员账户列表:' AS ''
UNION ALL
SELECT 
    CONCAT('ID: ', id) AS '',
    CONCAT('用户名: ', user_name) AS ''
FROM admin
UNION ALL
SELECT
    CONCAT('权限级别: ', permission, ' (1=admin+editor, 2=admin, 3=editor)') AS '',
    CONCAT('显示名称: ', name) AS ''
FROM admin
UNION ALL
SELECT '=' AS '', '首次登录提示:' AS ''
UNION ALL
SELECT '', '用户名: admin' AS ''
UNION ALL
SELECT '', '密码: admin123' AS ''
UNION ALL
SELECT '', '登录后请立即修改密码!' AS '';
