# 滚服 + 合服 数据库设计说明

## 一、当前 core/db 设计概览

### 1.1 结构

- **CommonEngine**：进程启动时用 env 配置建一个「公共」引擎
  - **MySQL**：`env.Mysql.CommonAddr`，单实例
  - **Redis**：`env.Redis`（单地址），用于全局/公共
  - 用途：查 `servergroup` 表拿本节点所属服务器列表、账号表等

- **Engines**：`map[serverId]*CDBEngine`
  - 从 `servergroup` 表查出本节点所在 `server_group` 下的所有 `ServerItem`
  - 对每个 ServerItem：**每个服一条 MySQL 地址 + 一个 Redis 端口**
  - `NewConnect(v, app)`：Redis 用 `v.RedisPort`（同一 host 不同 port），MySQL 用 `v.MysqlAddr`（**每服一个 MySQL 地址**）

- **ID 设计**
  - 玩家 ID：`GetPlayerId(serverId)` → Redis `INCR playerId`，再 `serverId * PlayerIdBase + localId`（已分服，合服不撞号）
  - `GetEngineByPlayerId(playerId)`：`serverId = playerId / PlayerIdBase`，再取对应 Engine

### 1.2 现状小结

| 维度     | 当前做法                         |
|----------|----------------------------------|
| Redis    | 每服一个 Redis（不同 port）     |
| MySQL    | 每服一个 MySQL 地址（MysqlAddr）|
| 玩家 ID  | 已分服：serverId*1e9 + 自增     |
| 合服     | 需合并多套 MySQL + 多套 Redis   |

你的目标：**一服一 Redis、合服时主要合并 Redis；MySQL 不想每服一个，希望共享且合服好做。**

---

## 二、滚服 + 合服 常见做法

### 2.1 Redis：一服一 Redis，合服 = 合并 Redis

- **滚服**：新服 = 新 Redis 实例（或新 DB 号），热数据、玩家缓存、排行榜等都在该 Redis。
- **合服**：
  - 方案 A：把被合服 Redis 的数据导到目标服 Redis（脚本 SCAN + 迁移，必要时 key 加前缀/改名）。
  - 方案 B：多 Redis 用 DB 号区分时，合服后只保留目标服 DB，其它数据先导入再下线。
- **前提**：所有会跨服或合服后共存的 ID，**从第一天就按 serverId 分段**（你已用 `serverId*1e9+local`），这样合服后 ID 不冲突。

### 2.2 MySQL：共享库 + server_id 分表，合服 = 改 server_id

- **常见**：**一个（或少量）MySQL 实例**，多服共用。
  - **全局表**：账号、服列表、订单、GM 邮件等，不按服分，所有服读写同一批表。
  - **按服分表**：帮派、邮件、活动记录等，表里带 `server_id`，同一张表存多服数据，查询带 `WHERE server_id = ?`。
- **合服**：
  - Redis：按上面做数据迁移/合并。
  - MySQL：不需要合并多个库，只需把被合服的行的 `server_id` 更新为目标服，例如：  
    `UPDATE guild SET server_id = 1 WHERE server_id IN (2,3);`  
    若有唯一约束（如 guild_name），需在合服脚本里处理（改名或加后缀等）。
- **好处**：部署简单（一套 MySQL）、备份/运维简单、合服不动库结构，只改数据和路由。

### 2.3 ID 设计（合服不撞号）

- **玩家 ID**：已满足 `serverId * 1e9 + 自增`，合服后保留原 ID 即可。
- **其它实体**（帮派、邮件、活动等）若也进 MySQL 且可能合服：
  - 推荐：**全局唯一 ID = f(serverId, 自增)**，例如 `serverId * 1e9 + localId`，或雪花算法里带 serverId。
  - 这样合服时只改 `server_id` 或路由，不必改 ID，也不会冲突。

---

## 三、推荐方案（对齐你的需求）

### 3.1 原则

1. **一服一 Redis**：热数据、玩家缓存、排行榜等只写本服 Redis；合服时只做 Redis 数据合并。
2. **MySQL 共享**：全服（或同一集群）共用一个 MySQL 实例；按服分的表统一加 `server_id`，合服只改 `server_id` 和业务上的唯一约束。
3. **ID 从根上分服**：玩家 ID 已分服；其它需要合服的表（帮派、邮件等）也用「serverId + 自增」生成唯一 ID，合服不重算 ID。
4. **合服流程简单**：Redis 迁移脚本 + MySQL 若干条 UPDATE（+ 唯一约束处理），无需拆库合库。

### 3.2 架构示意

```
                    ┌─────────────────────────────────────────┐
                    │           共享 MySQL（一个实例）           │
                    │  全局表：account, servergroup, pay...    │
                    │  分服表：guild, mail, ... (带 server_id)  │
                    └─────────────────────────────────────────┘
                                          │
         ┌────────────────────────────────┼────────────────────────────────┐
         │                                │                                │
    服 1 进程                        服 2 进程                        服 3 进程
         │                                │                                │
    Redis-1 (port1)                Redis-2 (port2)                Redis-3 (port3)
    玩家/排行榜/缓存                  玩家/排行榜/缓存                  玩家/排行榜/缓存
```

- 每个进程：`GetEngine(serverId)` 得到 **本服的 Redis + 共享的 MySQL**。
- 新增服：在 servergroup 里加一条，配新 Redis 的 host/port，**不新增 MySQL**。

### 3.3 与当前 core/db 的差异与改造点

| 项目       | 当前 core/db              | 推荐做法                     |
|------------|---------------------------|------------------------------|
| Redis      | 每服一个（不同 port）     | 保持，一服一 Redis           |
| MySQL      | 每服一个地址 (MysqlAddr)  | **改为全服共用一个**         |
| Engine 含义| 每服独立 Redis + 独立 MySQL | 每服 **独立 Redis + 共享 MySQL** |
| servergroup| 存 RedisPort + MysqlAddr  | 只存 **Redis 地址/port**，MySQL 用 env 公共配置 |
| 合服       | 要合并多 MySQL + 多 Redis | 只合并多 Redis；MySQL 只 UPDATE server_id |

### 3.4 具体改造建议（core/db）

1. **ServerItem 模型**
   - 保留：`Id`, `RedisPort`（或改为 `RedisHost`+`RedisPort`，支持每服不同 Redis 实例）。
   - 去掉或弃用：`MysqlAddr`（不再按服配 MySQL）。

2. **Start() / NewConnect()**
   - **CommonEngine**：继续用 env 的 MySQL +（若需要）一个全局 Redis。
   - 从 **servergroup** 读出的每条 ServerItem，只根据 Item 配置建 **Redis 连接**；MySQL **统一用 CommonEngine.Mysql**（或从 env 再拿一次连接串，但逻辑上「一个进程一个 MySQL 连接池」）。
   - 即：`Engines[serverId].Redis` = 该服专属 Redis；`Engines[serverId].Mysql` = CommonEngine.Mysql（或同一配置的引擎引用）。

3. **GetEngine(serverId)**
   - 返回：该服的 Redis + **共享的 MySQL**，业务侧调用方式不变（仍用 `rdb.Mysql` / `rdb.RedisExec`），只是底层 MySQL 指向同一实例。

4. **表与 ID**
   - 所有「按服」的 MySQL 表：加 `server_id` 字段（若还没有），写入时用当前服 id；读取时 `WHERE server_id = ?`。
   - 帮派 ID、邮件 ID 等：改为 `serverId*1e9+localId`（或你们现有的 ID 生成方式），保证合服后不撞号。

5. **合服流程（简要）**
   - Redis：写脚本从被合服 Redis 把 key 迁到目标服 Redis（注意 key 是否带 server 前缀、是否需要改）。
   - MySQL：`UPDATE xxx SET server_id = 目标服 WHERE server_id IN (被合服列表)`；处理重名等唯一约束（如帮派名加后缀或合并规则）。
   - 下线被合服节点，目标服进程可只连「目标服 Redis + 共享 MySQL」。

这样：**一服一 Redis、合服主要合并 Redis；MySQL 共享、合服只改 server_id，无需每服一个 MySQL，也方便做合服。**

---

## 四、可选：servergroup 表结构建议

便于以后只配 Redis、不配 MySQL，例如：

```text
id                 - 服 ID（与 PlayerIdBase 对齐：serverId）
server_group       - 组（同组可同进程多服）
redis_host         - 该服 Redis 地址（可默认同 env，新服再改）
redis_port         - 该服 Redis 端口
server_state       - 状态
server_name        - 展示名
...
（不再需要 mysql_addr）
```

MySQL 连接串只从 env 的 `Mysql.CommonAddr` 读一次，所有服共用。

---

## 五、core/db 代码级改造清单

1. **`core/db/db.go`**
   - `NewConnect(v model.ServerItem, app module.App)`：不再用 `v.MysqlAddr` 建 MySQL；改为 `dbEngine.Mysql = CommonEngine.Mysql`（或从 app 拿共享引擎），只根据 `v` 建 Redis（如 `v.RedisHost`+`v.RedisPort`，若无 RedisHost 则用 env 的 Redis host）。
   - 若 ServerItem 当前只有 `RedisPort` 且 Redis 同 host：可保留 `Redis: NewRedisPool(fmt.Sprintf("%s:%d", host, v.RedisPort), ...)`，`Mysql` 赋值为 `CommonEngine.Mysql`。

2. **`core/model` 的 ServerItem**
   - 保留 `RedisPort`（或扩展为 RedisHost + RedisPort）。
   - 若不再按服连 MySQL，可标记 `MysqlAddr` 为废弃或删除；配置表 `servergroup` 不再填 mysql_addr。

3. **业务侧**
   - 所有用 `rdb.Mysql` 的地方不变，只是底层指向同一 MySQL；需要「按服」的表确保有 `server_id` 且写入/查询带本服 id。
   - 合服时：先执行 Redis 合并脚本，再执行 MySQL 的 `UPDATE ... SET server_id = ? WHERE server_id IN (...)` 及唯一约束处理。

4. **ID 与 GetEngineByPlayerId**
   - 保持 `GetPlayerId(serverId)`、`GetEngineByPlayerId(playerId)` 逻辑不变；其它需合服的实体 ID 建议同样采用「serverId * 1e9 + 自增」或等价分段规则。

---

## 六、总结

- **滚服**：新服 = 新 Redis + servergroup 一条新记录，**不新开 MySQL**。
- **合服**：合并多个 Redis 的数据到目标服 Redis；MySQL 只做 `server_id` 更新和唯一约束处理。
- **core/db 改造**：Engines 中每个 serverId 只绑定「本服 Redis + 共享 MySQL」；ServerItem 不再使用 MysqlAddr，MySQL 统一用 CommonEngine（或同一配置），即可在保持现有调用方式的前提下，满足「一服一 Redis、MySQL 共享、合服简单」的设计目标。

文档中的「五、core/db 代码级改造清单」可直接作为改造步骤对照实现。
