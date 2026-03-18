# 合服活动与 Redis SOP

本文补充说明合服时与活动、排行榜、邮件、公会相关的 Redis 注意事项，默认基于当前项目的“入口服保留、逻辑服合并、MySQL 共享”设计。

## 1. 先确认 Redis 部署模式

- 若多个 `main_server` 共享同一套 Redis：
  - 玩家主数据、活动数据、排行榜数据天然可被目标逻辑服读取
  - 合服重点在于停服窗口、路由切换与业务表 `server_id` 迁移
- 若每个 `main_server` 使用独立 Redis：
  - 合服前必须额外执行 Redis 数据迁移或接受部分活动/排行榜重置
  - 否则来源服角色合服后会出现“DB 已迁移，但活动/排行榜状态仍留在旧 Redis”的问题

## 2. 关键 Redis Key 清单

### 2.1 角色与基础状态

- `AccountRole:<uid>:<entryServerId>`：入口服角色到 `playerId` 映射
- `Player:<playerId>` 以及各玩家子模块 Key：角色主数据

处理建议：

- 共享 Redis：无需迁移，仅验证目标逻辑服可读到来源角色数据
- 独立 Redis：必须迁移来源服玩家主数据到目标 Redis

### 2.2 活动状态

- `Activity:<actId>`：活动主数据
- `ActivityPlayer:<actId>`：活动参与者明细（Hash）

处理建议：

- 共享 Redis：保持现状即可
- 独立 Redis：
  - 常规活动建议迁移 `Activity:*` 与 `ActivityPlayer:*`
  - 若活动本身允许跨服重置，也可在合服窗口公告后重置

### 2.3 排行榜与竞技记录

- `rank_draw_hero:<actId>`
- `rank_recharge:<actId>`
- `rank_the_competition:<actId>_<group>`
- `rank_gofish:<actId>`
- `rank_arena:<actId>`
- `rank_tianti:<actId>`
- `rank_arena_record:<actId>_<playerId>`
- `rank_tianti_record:<actId>_<playerId>`

处理建议：

- 共享 Redis：
  - 排行榜会自然并入同一逻辑世界
  - 需在合服公告中说明榜单竞争人数扩大
- 独立 Redis：
  - 运营若要求“保留进度”，则需导出来源服 ZSET/HASH 并并入目标 Redis
  - 运营若接受“新赛季/重置”，则在合服窗口删除对应 Key 并重新开榜

### 2.4 邮件与公会运行态

- `systemMailId:<logicServerId>`
- `dailyMail:<logicServerId>`
- `GuildManager:<logicServerId>`
- `player_guild:<playerId>`
- `guild_chat_history:<guildId>`

处理建议：

- `systemMailId:*` / `dailyMail:*`：现在已按逻辑服隔离，合服后保留目标逻辑服即可
- `GuildManager:*`：按逻辑服隔离；独立 Redis 场景下建议迁移目标逻辑服对应 Key 或重建
- `player_guild:*`：若独立 Redis，需与玩家主数据一起迁移
- `guild_chat_history:*`：可选迁移；若不保留历史可直接清理

## 3. 推荐合服执行顺序

1. 停服并冻结登录
2. 备份 MySQL 与 Redis
3. 确认来源服、目标服 Redis 是否共享
4. 若 Redis 独立：先迁移玩家主数据、活动数据、排行榜 Key、公会缓存 Key
5. 执行 GM 合服工单（MySQL `server_id` 迁移 + `account_role.logic_server_id` 更新 + 路由切换）
6. 校验：
   - 来源入口服角色可登录到目标逻辑服
   - 活动面板、竞技场、天梯、排行可正常打开
   - 系统邮件、个人邮件、公会、好友链路正常
7. 开服后持续观察活动榜与邮件投诉

## 4. 独立 Redis 场景下的最低可行方案

若时间窗口不足，推荐优先级：

1. 必迁：玩家主数据、`player_guild:*`、`AccountRole:*`
2. 高优先：`ActivityPlayer:*`、排行榜 Key
3. 可选：`guild_chat_history:*`

如果第 2 类做不到，需提前公告“部分跨服活动/排行进度将在合服后重置”。

## 5. GM / Admin 辅助能力

- GM 接口：
  - `/gm/merge/redis-check`：返回 Redis 模式检查清单、关键 Key 模式、执行步骤
  - `/gm/merge/redis-script`：返回迁移脚本模板路径与命令示例
- Admin 页面：`合服管理` 中可直接执行 Redis 检查并查看模板命令
- 模板脚本：`tools/scripts/redis/merge_redis_migration_template.ps1`

说明：

- 当前脚本是“安全模板”，默认只输出建议命令与核对步骤，不会直接对生产 Redis 执行危险复制
- 真正执行时，请先替换 `SourceRedis/TargetRedis`，并结合你们现网的备份、网络与权限策略落地
