// Package serverdb 提供业界常见的、逻辑自洽的服务器与 DB 管理：
// - 连接配置仅来自部署层（Config），不存于业务表
// - 服列表仅存元数据（名称、状态、客户端连接地址），不含 Redis/MySQL 地址
// - 方案二：一进程一服一 Redis，共享 MySQL；合服后数据合并到单 Redis，单服处理
package serverdb

import "time"

// PlayerIdBase 玩家 ID 分段基数，合服不撞号：全局玩家 ID = serverId*PlayerIdBase + 本服自增
const PlayerIdBase = 1000000000

// 服状态（与业务约定一致，便于迁移）
const (
	ServerStateNormal      = 0 // 正常
	ServerStateYongji      = 1 // 拥挤
	ServerStateBaoMan      = 2 // 爆满
	ServerStateMaintenance = 3 // 维护
	ServerStateNoOpen      = 4 // 未开服
	ServerStateStop        = 5 // 停服
)

// Config 连接配置（仅部署层：env/配置文件），不存数据库。
// 一进程对应一服时：Redis 为本服唯一实例，Mysql 为全服共享。
type Config struct {
	// ServerId 本进程所属服 ID
	ServerId int

	// Redis 本服 Redis（方案二：一服一 Redis，合服后数据合并到此）
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	// Mysql 全服共享 MySQL 地址
	MysqlAddr string

	// 连接池可选；为 0 使用默认
	RedisMaxIdle     int
	RedisMaxActive   int
	RedisIdleTimeout time.Duration
	MysqlMaxIdle     int
	MysqlMaxOpen     int
	MysqlMaxLifetime time.Duration
}
