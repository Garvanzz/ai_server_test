# 滚服 + 共享 MySQL 改造：注意事项与隐性风险

本文档列出全服共享 MySQL、一服一 Redis 改造后需要遵守的约定和容易出现的隐性问题，便于排查与合服操作。

---

## 一、已完成的代码改造

- **core/db**：`NewConnect` 只建本服 Redis；每服引擎的 `Mysql` 统一指向 `CommonEngine.Mysql`；仅 `CommonEngine` 在 `Close()` 时关闭 MySQL。
- **core/model.ServerItem**：新增可选 `RedisHost`；`MysqlAddr` 仅保留兼容/展示，连接不再使用。
- **Redis 地址**：优先使用 `ServerItem.RedisHost:RedisPort`，为空时使用 env 的 Redis Host 与 `ServerItem.RedisPort`。

---

## 二、必须遵守的约定（防隐性设计问题）

### 2.1 按服表必须带 `server_id`

共享 MySQL 后，**所有“按服”存储的表**都必须：

1. **表结构**：有 `server_id` 字段（或等价含义字段），并建索引（至少包含 `server_id`）。
2. **写入**：`INSERT` 时必带当前服 ID（`server_id = GetEnv().ID` 或业务当前服）。
3. **查询/更新/删除**：所有 `SELECT/UPDATE/DELETE` 都必须带 `WHERE server_id = ?`（当前服），否则会**串服**（读到或改到其他服数据）。

当前需要重点核对/加 `server_id` 的表：

| 表名 | 说明 | 当前是否有 server_id | 建议 |
|------|------|----------------------|------|
| servergroup | 服列表，全局 | 无（按 id 查） | 保持现状 |
| account | 账号，按服 | 有 ServerId | 所有按 server_id 的查询必须带 WHERE |
| guild | 帮派 | **需加** | 加 server_id，所有语句带 server_id |
| guild_log | 帮派日志 | **需加** | 加 server_id，所有语句带 server_id |
| player_mail_info | 玩家邮件 | **需加** | 加 server_id，所有语句带 server_id |
| sys_mail_info | 系统邮件 | **需加** | 加 server_id（若按服发） |
| admin_mail | 后台/延迟邮件 | **需加** | 加 server_id，所有语句带 server_id |
| friend_apply | 好友申请 | **需加** | 加 server_id，所有语句带 server_id |
| friend_block | 黑名单 | **需加** | 加 server_id，所有语句带 server_id |
| pay_cache_order / pay_order | 支付订单 | 若有按服维度 | 按服则加 server_id |

**隐性风险**：漏写 `server_id` 或 WHERE 条件会直接导致**跨服读写**，数据错乱且难排查。建议：  
- 对上述表做一次全局搜索（`rdb.Mysql.Table("xxx")` / `define.XXX`），确认每条 SQL 都带 `server_id`。  
- 在代码评审中强制要求：凡“按服表”必带 server_id。

---

### 2.2 唯一约束与合服

合服时会对被合服的数据执行 `UPDATE xxx SET server_id = 目标服 WHERE server_id IN (被合服列表)`。若表上有**唯一约束**（如 `guild.guild_name`、`account.uid+server_id` 等），合服后可能冲突，必须在合服脚本中单独处理，例如：

- 帮派名：被合服服内帮派名加后缀（如 `_s2`）、或先重名校验再合并。
- 账号：若唯一键为 `(uid, server_id)`，合服只改 server_id 即可；若唯一键仅为 `uid`，则不能简单合并，需业务规则（如一个 uid 只保留一个服的数据）。

**注意**：  
- 合服前必须列出所有带 UNIQUE/PRIMARY 的按服表，逐条确认合服后的唯一约束策略。  
- 合服脚本建议先在从库/备份上跑一遍，再在正式环境执行。

---

### 2.3 ID 分段（合服不撞号）

玩家 ID 已采用 `serverId * PlayerIdBase + localId`，合服后不重算、不冲突。  
其他会入 MySQL 且可能参与合服的实体（帮派、邮件、活动等），建议也采用**分段 ID**，例如：

- `全局唯一 ID = serverId * 1e9 + 本服自增`，或  
- 雪花算法中嵌入 serverId。

这样合服时只改 `server_id` 或路由，不必改 ID，也不会出现 ID 冲突。  
**隐性风险**：若某表用“单服自增”且未分段，合服后多服数据在一起会 ID 冲突或覆盖，必须在改造时统一为分段 ID 或全局唯一方案。

---

### 2.4 连接池与连接数

- 共享 MySQL 后，**总连接数** = 所有 main_server 进程数 × 每进程 `MaxOpenConns`。
- 若多节点部署（如 10 个 main_server 进程、每进程 1200），则总连接数可达 12k，需将 MySQL `max_connections` 调大到 ≥ 总连接数，并留余量。
- 建议：多节点时适当降低每进程 `MaxOpenConns`（如 500～800），避免单机 MySQL 连接数过高。

---

### 2.5 不要对“共享 MySQL 引用”做特殊操作

- `Engines[serverId].Mysql` 与 `CommonEngine.Mysql` 是**同一引用**。  
- 除 `Close()` 外，不要对 `Engines[i].Mysql` 做“只关某个服”之类的关闭或替换操作，否则会影响所有服。  
- 关闭逻辑已约束为：仅 `CommonEngine.ownMysql == true` 时在 `Close()` 中关闭 MySQL。

---

### 2.6 Redis 与 servergroup 配置

- **新服开服**：在 servergroup 中新增一条记录，配置该服的 **Redis**（`RedisHost`/`RedisPort`）；**不要**再为该服配置独立 MySQL 地址，MySQL 统一用 env 的 `CommonAddr`。
- **合服**：  
  - Redis：用独立脚本把被合服 Redis 的数据迁到目标服 Redis（注意 key 前缀、DB 号、是否改 key 名）。  
  - MySQL：只做 `UPDATE ... SET server_id = 目标服 WHERE server_id IN (...)` 及唯一约束处理，不合并多个 MySQL 实例。

---

### 2.7 事务与跨表

- 共享 MySQL 后，一次事务可能涉及多服数据（若误用 server_id）。  
- 务必保证：**一个事务内只操作当前服的数据**（所有表都带 `server_id = 当前服`），避免跨服更新和锁混乱。

---

### 2.8 MysqlAddr 的兼容与配置

- `ServerItem.MysqlAddr` 已**弃用**（main_server 建连不再使用），仅保留供 GM/登录服展示或兼容。  
- 新服、新环境可不再配置 `mysql_addr`；若配置了也不会被 main_server 用于连接。  
- 若有脚本或后台仍依赖该字段，可保留填值，不影响当前改造。

---

### 2.9 servergroup 表结构建议（可选）

为与“只配 Redis、不配 MySQL”一致，建议 servergroup 表：

- 保留：`id`, `server_group`, `redis_host`, `redis_port`, `server_state`, `server_name`, 等业务字段。  
- 不再依赖：`mysql_addr`（可保留列做兼容，但不作为连接依据）。  
- MySQL 连接串只从 env 的 `Mysql.CommonAddr` 读取。

---

## 三、改造与合服检查清单

- [ ] 所有按服表已加 `server_id` 字段并在业务中写入/条件使用。  
- [ ] 所有按服表的 SELECT/UPDATE/DELETE 均带 `WHERE server_id = ?`。  
- [ ] 合服相关表的唯一约束已梳理，合服脚本中有对应处理（改名/合并规则等）。  
- [ ] 帮派/邮件等实体 ID 采用分段或全局唯一方案，合服后无 ID 冲突。  
- [ ] 多节点部署下 MySQL `max_connections` ≥ 节点数 × 每进程 MaxOpenConns，且已评估连接池大小。  
- [ ] 新服只加 servergroup + Redis，不新开 MySQL。  
- [ ] 合服流程文档已更新：Redis 迁移步骤 + MySQL 的 server_id 更新与唯一约束处理。

---

## 四、小结

- **核心风险**：漏写 `server_id` 导致串服；唯一约束未处理导致合服失败；未分段 ID 导致合服后冲突。  
- **建议**：按上表逐表加 `server_id` 并全量检查 SQL，合服前在测试环境完整走一遍合服流程，并保留回滚方案（如 DB 备份、Redis 快照）。
