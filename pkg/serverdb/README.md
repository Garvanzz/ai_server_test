# serverdb — 逻辑自洽的服务器与 DB 管理

本包按**业界常见做法 + 方案二**实现：

- **连接配置**：仅来自部署层（`Config` / env），不存于业务表。
- **服列表**：仅存元数据（名称、状态、客户端连接地址等），**不含** Redis/MySQL 地址。
- **方案二**：一进程一服一 Redis，共享 MySQL；合服时数据合并到单 Redis，合服后按单服处理。

## 设计原则

| 项目         | 做法 |
|--------------|------|
| Redis 地址   | 仅来自 `Config.RedisAddr`（env），不存 server 表 |
| MySQL 地址   | 仅来自 `Config.MysqlAddr`（env），不存 server 表 |
| 服列表表     | 只存 id、名称、组、状态、开服时间、**客户端连接地址**等，供选服/GM 展示 |
| 一进程       | 一个 Redis + 一个 MySQL 连接池；`GetEngine(serverId)` 仅当 serverId == 本进程服 ID 返回引擎 |

## 使用方式

### 1. 用 Config 启动（连接信息来自 env 或配置）

```go
cfg := serverdb.Config{
    ServerId:       1,
    RedisAddr:      "127.0.0.1:6379",
    RedisPassword:  "",
    RedisDB:        0,
    MysqlAddr:      "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4",
}
m := serverdb.NewManager(cfg)
if err := m.Start(); err != nil {
    log.Fatal(err)
}
defer m.Close()

// 可选：从 MySQL 加载服列表元数据（表内无 Redis/MySQL 字段）
_ = m.LoadServerList("server_list") // 或 "servergroup" 兼容现有表

serverdb.SetGlobal(m)
```

### 2. 业务侧取引擎

```go
// 本进程唯一引擎（单服）
eng, err := serverdb.GetEngine(env.ServerId)

// 按玩家 ID（仅本服玩家有数据时返回）
eng, err := serverdb.GetEngineByPlayerId(playerId)
```

### 3. 与现有 core/db 的对应关系

| core/db              | serverdb                          |
|----------------------|-----------------------------------|
| `db.GetEngine(serverId)`       | `serverdb.GetEngine(serverId)`，仅 serverId == Config.ServerId 有值 |
| `db.GetEngineByPlayerId(id)`   | `serverdb.GetEngineByPlayerId(id)` |
| `db.CommonEngine.Mysql`        | `eng, _ := serverdb.DefaultEngine(); eng.Mysql` |
| `rdb.RedisExec(...)`          | `eng.RedisExec(...)`              |

## 表结构

- 推荐表名：`server_list`（见 `schema.sql`），仅元数据列。
- 若沿用现有 `servergroup`：可传 `LoadServerList("servergroup")`，表中**不要**再存 redis_port/mysql_addr，连接只从 Config 读。

## 合服（方案二）

1. **Redis**：把被合服 Redis 数据迁移到目标服 Redis（脚本 SCAN + 导入）。
2. **MySQL**：共享库下按服表做 `UPDATE xxx SET server_id = 目标服 WHERE server_id IN (...)` 及唯一约束处理。
3. **进程**：目标服进程只连一个 Redis（env 里该服 Redis 地址）+ 共享 MySQL；不再按 playerId 路由到多 Redis。

## 常量

- `PlayerIdBase`：玩家 ID 分段基数（1e9），合服不撞号。
- `ServerStateNormal` 等：服状态，与业务约定一致。
