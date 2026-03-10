# MySQL 使用点梳理与全服共享性能分析

## 一、MySQL 使用点汇总（main_server + core）

以下仅统计 **main_server** 与 **core/db** 内对 MySQL 的访问（不含 login_server/gm_server 等独立进程）。

### 1.1 按表分类

| 表名 | 所在模块 | 操作类型 | 触发场景 | 大致频率 |
|------|----------|----------|----------|----------|
| **servergroup** | core/db, logic/activity | Get(by id), Find(by server_group) | 进程启动、活动加载时 | 极低（启动/偶发） |
| **account** | player/chat, player/internal/chat, player/mail, logic/login, player/friend | Get(uid), Update(uid/nick_name/redis_id/sys_mail_id), ForUpdate+Get+Update | 聊天校验、封禁检查、登录更新、邮件已读进度、好友推荐随机 | 中（登录 1 次/人；聊天/邮件/好友按请求） |
| **player_mail_info** | logic/mail, player/mail | Find(全表/条件), Insert, Update(got_item), Delete(id) | 邮件列表、打开/领奖/删除、系统/个人发信 | 中（按邮件相关请求） |
| **sys_mail_info** | logic/mail | Find(全表), Insert | 邮件模块启动加载、发系统邮件 | 低（启动 1 次 + 发系统邮件时） |
| **admin_mail** | logic/mail | Find(status=1), Get(id), Update(status) | 延迟邮件 OnTick、发送时 | 低（按延迟邮件条数） |
| **guild**（帮派） | logic/guild | Find(全服), Insert, Update(id), Exist(guild_name), Delete(id) | 加载帮派列表、创建/更新/删帮派 | 低（创建/管理时） |
| **guild_log** | logic/guild | Insert, Find(guild_id)+Delete | 写帮派日志、查日志、删日志 | 低 |
| **friend_apply** | player/friend | Exist, Count, Insert, Find, Get, Delete | 好友申请/处理/列表 | 中（按好友操作） |
| **friend_block** | player/friend | Find, Exist, Insert, Delete | 黑名单 | 低 |
| **pay_cache_order** | player/shop | Insert, Get(order_id), Delete | 支付下单/查询/清理 | 低（有支付时） |
| **pay_order** | player/shop | Insert | 支付落单 | 低 |

### 1.2 按调用路径分类（读/写）

**CommonEngine.Mysql（当前“公共”库，多服共享后仍是同一实例）：**

- account：Get / Update / ForUpdate+Get+Update（聊天、登录、邮件进度、好友推荐）
- pay_cache_order / pay_order：Insert、Get、Delete（支付）

**rdb.Mysql（当前按 serverId 取到的引擎；改为共享 MySQL 后仍走同一实例）：**

- servergroup：Get、Find（启动/活动）
- sys_mail_info / admin_mail / player_mail_info：Find、Insert、Update、Delete（邮件）
- guild / guild_log：Find、Insert、Update、Exist、Delete（帮派）
- friend_apply / friend_block：各类查询与增删（好友）

### 1.3 高频与潜在热点

- **相对高频**：account 的 Get/Update（登录、聊天、邮件、好友）、player_mail_info 的 Find/Update/Insert/Delete、friend_apply 的 Exist/Find/Insert。
- **单次较重**：sys_mail_info 全表 Find（仅启动一次）、servergroup Find（按组拉全量，仅启动/偶发）。
- **索引敏感**：account(uid)、player_mail_info(db_id/sys_id/id)、guild(guild_name, id)、friend_apply(player_id, target_id) 等，需保证有合适索引，避免全表扫描。

---

## 二、全服共享 MySQL 的性能与容量

### 2.1 共享后的变化

- **连接数**：当前若每服一个 MySQL，则每进程内每个 Engine 一个连接池（如 maxOpen=1200）。改为全服共享后，**同一进程内**可共用一个池（例如只保留 CommonEngine.Mysql 一个池），单进程连接数会**下降**。但若有 **多台 main_server 进程**（多节点部署），则总连接数 = **节点数 × 每进程 MaxOpenConns**，需要 MySQL `max_connections` ≥ 该值。
- **QPS**：所有服的请求最终打到**同一实例**，总 QPS = Σ 每服 MySQL QPS。

### 2.2 QPS 是什么？怎么和「在线人数」对应？

**QPS** = 每秒打到 MySQL 的请求次数（Queries Per Second）。数值越大，数据库越忙。

不用记 QPS 具体数字也没关系，可以用**全服总同时在线人数**来粗判压力，换算关系如下。

**换算思路**：每个在线玩家在峰值时段，平均每秒会触发一定次数的 MySQL 操作（登录更新、聊天查账号、邮件/好友读写等）。当前业务下，按代码路径粗算：

- 每人每分钟约 5～15 次游戏内操作（点按钮、聊天、领邮件等），其中会落到 MySQL 的约占 5%～15%。
- 折合下来：**每 1000 同时在线 ≈ 约 150～250 次 MySQL/秒**（峰值、经验值）。

即：

**「每 1000 同时在线 ≈ 约 200 MySQL QPS」**（取中值，用于下表；实际以你压测为准。）

这样你就可以**只用“同时在线人数”**来对照下面的结论表，不必自己算 QPS。

### 2.3 单服 MySQL QPS 量级（经验估计）

若按单服峰值在线 1000 人、每 1000 在线约 200 QPS 计，则单服约 **200 MySQL QPS**。多服共享时：

- 5 服（约 5k 在线）→ 约 1k QPS  
- 10 服（约 1w 在线）→ 约 2k QPS  
- 20 服（约 2w 在线）→ 约 4k QPS  
- 50 服（约 5w 在线）→ 约 10k QPS  

（实际与业务比例、缓存使用、是否有批量写等有关，需压测校准。）

### 2.4 何时可能出性能问题（数量级）

- **单实例 MySQL**（8C16G～16C32G、SSD、索引合理、无大事务）：  
  - 简单主键/索引查询 + 轻量更新：约 **5k～2w QPS** 较常见；再往上要看磁盘和锁。  
  - 用在线人数粗算：约 **2 万同时在线** 对应约 4k QPS 量级，一般没问题；**4～5 万同时在线** 对应约 8k～10k QPS，单实例压力较大，需考虑读写分离或分库。

- **连接数**：  
  - 若每进程 MaxOpenConns=1200，10 个 main_server 进程即 12k 连接，需 MySQL `max_connections` 调大（如 16k+）并注意连接复用与空闲超时。

- **锁与热点**：  
  - account 表按 uid 更新（登录、邮件进度、昵称等）可能产生行锁竞争；若同一 uid 并发高（多端/重试），可能成为热点行，但通常**到较高在线、数千 QPS 级别**才明显，且可通过缓存或队列缓解。

### 2.5 结论（用「同时在线人数」衡量）

下面用**全服总同时在线人数**作为主要指标（不再强调 QPS），便于直接判断。

| 全服总同时在线（约） | 对应 MySQL 压力（约） | 共享单实例是否容易出问题 |
|----------------------|------------------------|----------------------------|
| **5 千以内**         | 约 1k QPS 以内         | 一般不会                  |
| **5 千～1 万**       | 约 1k～2k QPS          | 一般不会                  |
| **1 万～2 万**       | 约 2k～4k QPS          | 通常没问题，建议做好索引与监控 |
| **2 万～4 万**      | 约 4k～8k QPS         | 有压力，需优化 + 监控，必要时读写分离 |
| **4 万～5 万以上**  | 约 8k～10k+ QPS       | 单实例易成瓶颈，需读写分离/分库/缓存等 |

**简单记**：  
- **2 万同时在线以内**：全服共享一个 MySQL 通常还能扛得住，注意索引和连接数即可。  
- **超过 2 万、向 5 万靠近**：就要开始考虑拆分（读写分离、分库）或加缓存，避免单实例成为瓶颈。

（若你单服在线很高、或某些功能特别吃 MySQL，可按实际压测把「每 1000 在线」对应的 QPS 调大，再按比例换算上表。）

---

## 三、建议（在共享 MySQL 前提下）

1. **索引**：account(uid)、player_mail_info(db_id, sys_id)、guild(server_id, guild_name)、friend_apply(player_id, target_id) 等按查询条件建好索引，避免全表扫描。
2. **连接池**：共享 MySQL 时，单进程内只保留一个 MySQL 池；多节点部署时控制每进程 MaxOpenConns（如 500～800），使总连接数 < MySQL max_connections。
3. **监控**：对 MySQL QPS、连接数、慢查询、锁等待做监控；若全服**同时在线**接近或超过 2 万（约 4k QPS 量级），建议提前评估读写分离或分库。
4. **热点**：若 account 等表出现单行热点，可对高频读做 Redis 缓存或写合并，降低对 MySQL 的实时 QPS。

上述使用点与量级可作为“全服共享 MySQL 是否会有性能问题、大概在什么数量级会出问题”的参考；精确结论建议用实际服数与压测再校准一次。
