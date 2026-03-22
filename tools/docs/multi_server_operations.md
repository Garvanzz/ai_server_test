# 多服与合服运作使用文档（main/login/gm）

本文基于当前项目代码与 `tools/scripts/mysql/schema_full.sql` 进行整理，目标是给出一份可直接执行的多服运作与合服说明，并记录已发现问题。

## 1. 适用范围

- 适用服务：`main_server`（游戏服）、`login_server`（登录服）、`gm_server`（后台）
- 适用数据库脚本目录：`tools/scripts/mysql`
- 当前核心设计：
  - MySQL 共享（逻辑服通过 `server_id` 隔离）
  - 区服入口与逻辑服解耦（`game_server.id` vs `game_server.logic_server_id`）
  - 合服通过路由切换 + 数据 `server_id` 迁移实现

---

## 2. 数据库脚本与初始化

目录（当前仓库实际路径）：

- `tools/scripts/mysql/schema_full.sql`：全量建表（目标态）
- `tools/scripts/mysql/reset_drop_all_tables.sql`：删表重建脚本（DROP TABLE）
- `tools/scripts/mysql/reset_truncate_all_tables.sql`：清空数据脚本（TRUNCATE，保留表结构）
- `tools/scripts/mysql/migrations/*.sql`：增量迁移脚本（按编号执行）
- `tools/scripts/mysql/migration_conventions.sql`：迁移规范说明入口（不放可执行 SQL）

初始化建议流程：

1. 全新环境：执行 `schema_full.sql`
2. 重建环境：先执行 `reset_drop_all_tables.sql`，再执行 `schema_full.sql`
3. 仅清空数据：执行 `reset_truncate_all_tables.sql`
4. 增量变更：在 `migrations/` 下新增 `NNN_<description>.sql`，按编号顺序执行

---

## 3. 多服核心概念（必须统一）

- `entry_server_id`（入口服）：客户端选服 ID，对应 `game_server.id`
- `logic_server_id`（逻辑服）：业务真实归属服，对应 `game_server.logic_server_id`
- `server_id`（业务服字段）：业务数据所属逻辑服 ID
- `origin_server_id`（来源服）：首次来源服，便于审计与补偿

关键原则：

1. 业务表查询/更新必须带 `server_id`
2. 合服后入口服可以保留，`logic_server_id` 指向目标服
3. 历史来源必须尽可能保留 `origin_server_id`

---

## 4. 表结构规划总览（按职责）

### 4.1 登录/区服路由（account 库）

- `account`：账号与角色映射（含 `server_id`, `origin_server_id`）
- `server_group`：区服分组展示元数据
- `game_server`：区服入口、逻辑路由、合服状态（核心）
- `hot_update`：热更版本配置
- `notice`：公告（`server_id=0` 全服）

### 4.2 GM 与支付（account 库）

- `admin`：GM 管理员账号
- `pay_order` / `pay_cache_order`：支付订单（含 `server_id`）

### 4.3 游戏业务共享表（main_server 使用）

- 邮件：`sys_mail_info`、`admin_mail`、`player_mail_info`
- 公会：`guild`、`guild_apply`、`guild_log`
- 好友：`friend_apply`、`friend_block`

### 4.4 合服工单与审计

- `merge_plan`：合服计划主表
- `merge_server_map`：来源服处理明细
- `merge_conflict_log`：冲突日志（如公会重名）

---

## 5. 三服协作流程

### 5.1 玩家登录与进服

1. 客户端调用 `login_server /servers/list` 获取入口服列表（返回含 `logicServerId`）
2. 客户端选择入口服后调用 `login_server /auth/login`
3. 登录服根据 `game_server.id` 解析目标 `logic_server_id`，并将账号 `server_id` 更新为逻辑服
4. 返回 `token + uid + serverId(逻辑服) + entryServerId(入口服)`
5. 客户端携带 token 连接 `main_server`，由 main_server 校验登录服 Redis token 后加载角色

补充约定：login_server 新接口字段统一使用 camelCase，例如区服列表返回 `serverList`、热更检查返回 `status/url/version`；当前仍保留 `ServerList`、`Status/Url/Version` 兼容旧客户端。

### 5.2 GM 管理入口

- GM 接口统一在 `gm_server`，关键路由：
  - 区服/进程：`/gm/servers/*`
  - 合服：`/gm/merge/*`
  - 邮件/公告等转发：由 `gm_server` 按 `game_server -> logic_server_id -> main_server_http_url` 路由

### 5.3 main_server 数据访问约定

- 已落地方向：大量业务 SQL 已显式带 `server_id`
- 运行要求：所有按服数据读写严格限定当前逻辑服，禁止跨服条件缺失

---

## 6. 合服标准操作（建议 SOP）

### 6.1 前置检查

1. 冻结变更：停服窗口执行
2. 备份：MySQL 全量备份 + Redis 快照
3. GM 预检查：`/gm/merge/precheck`（目前重点检查公会重名）
4. 核对目标/来源服是否都存在于 `game_server`

### 6.2 执行流程

1. 创建计划：`/gm/merge/plan/create`
2. 执行计划：`/gm/merge/execute`
3. 执行动作（当前实现）：
   - 重名公会改名（追加 `_S{source}`）
   - 将来源逻辑服业务表 `server_id` 更新为目标逻辑服
   - 将来源入口服 `game_server.logic_server_id` 指向目标服，`merge_state=2`
4. 验证：
   - `game_server` 路由与状态
   - 业务表按目标 `server_id` 可查询
   - 登录、发邮件、好友、公会等核心链路抽检

### 6.3 回滚策略（当前）

- 现有 `/gm/merge/rollback` 仅回滚 `game_server` 路由状态
- 不会自动回滚已迁移业务数据
- 因此必须依赖“备份恢复”作为真实数据回滚方案

---

## 7. 运维检查清单

- 开新服前：`game_server` 配置完整（`id/group_id/logic_server_id/main_server_http_url`）
- 合服前：冲突预检通过，备份已完成
- 合服后：
  - 登录服返回 `logicServerId` 正确
  - main_server 按目标逻辑服读写数据
  - gm_server 转发到目标 `main_server_http_url`
  - 支付、邮件、公会、好友抽检通过

---

## 8. 已发现问题与风险记录

以下为本次梳理中已确认的风险点，建议纳入后续修复计划。

1. 合服回滚不完整
   - 现状：`gm_server/logic/merge.go` 的回滚仅恢复 `game_server.logic_server_id/merge_state`
   - 风险：业务表已改 `server_id` 时，逻辑路由与真实数据不一致

2. 合服执行缺少事务包裹
   - 现状：执行合服为多条独立 SQL，部分失败时只能记录 `partial failed`
   - 风险：出现半成功状态，人工修复成本高

3. 合服迁移表覆盖不全（邮件侧）
   - 现状：当前迁移列表未覆盖 `sys_mail_info`、`admin_mail`
   - 风险：来源服定向系统邮件/延迟邮件在合服后可能无法被目标服加载（main_server 仅加载 `server_id in (0, 当前服)`）

4. 账号多服模型与唯一键语义需再次确认
   - 现状：表注释描述“uid 可在多逻辑服拥有角色”，但 `account` 仍有 `uk_account(account)` 全局唯一，`login_server` 按 `account` 单行登录
   - 风险：如果业务目标是真正“一账号多服并存角色”，当前模型与流程需要补充角色维度或放宽约束

---

## 9. 建议的下一步落地

1. 先补齐脚本安全性
   - 保持 `reset_drop_all_tables.sql`（重建）、`reset_truncate_all_tables.sql`（清数）、`migrations/`（增量）三类脚本边界清晰
   - `tools/docs` 仅保留说明文档，不存放可执行 SQL

2. 补强合服可靠性
   - 合服核心步骤加事务或分阶段幂等机制
   - 回滚方案明确为“DB/Redis 备份恢复”，并写成标准应急手册

3. 补齐邮件迁移策略
   - 明确 `sys_mail_info/admin_mail` 的合服行为（迁移或保留规则）并代码实现一致

4. 明确账号多服产品规则
   - 若要一账号多服并存角色，需补角色映射模型和登录流程
   - 若保持一账号单角色，需更新表注释与文档，避免歧义
