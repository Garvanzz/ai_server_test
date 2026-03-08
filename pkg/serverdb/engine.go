package serverdb

import (
	"github.com/gomodule/redigo/redis"
	"xorm.io/xorm"
)

// Engine 单服数据引擎：本服 Redis + 共享 MySQL。
// 与 core/db.CDBEngine 字段兼容，便于业务迁移。
type Engine struct {
	Redis *redis.Pool
	Mysql *xorm.Engine
}

// RedisExec 同步执行 Redis 命令
func (e *Engine) RedisExec(cmd string, args ...interface{}) (interface{}, error) {
	if e == nil || e.Redis == nil {
		return nil, ErrNoEngine
	}
	conn := e.Redis.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

// GetPlayerId 生成唯一玩家 ID（serverId*PlayerIdBase + 本服自增），合服不撞号
func (e *Engine) GetPlayerId(serverId int) (int64, error) {
	reply, err := e.RedisExec("INCRBY", "playerId", 1)
	if err != nil {
		return 0, err
	}
	v, err := redis.Int(reply, err)
	if err != nil {
		return 0, err
	}
	return int64(serverId*PlayerIdBase + v), nil
}

// GetDelayMailId 延时邮件 ID
func (e *Engine) GetDelayMailId() (int, error) {
	reply, err := e.RedisExec("INCRBY", "delayMailId", 1)
	if err != nil {
		return 0, err
	}
	return redis.Int(reply, err)
}

// GetRoomId 房间 ID
func (e *Engine) GetRoomId() (int64, error) {
	reply, err := e.RedisExec("INCRBY", "roomId", 1)
	if err != nil {
		return 0, err
	}
	return redis.Int64(reply, err)
}

// GetEquipId 装备 ID
func (e *Engine) GetEquipId() (int, error) {
	reply, err := e.RedisExec("INCRBY", "equipId", 1)
	if err != nil {
		return 0, err
	}
	return redis.Int(reply, err)
}

// GetActivityId 活动 ID
func (e *Engine) GetActivityId() (int, error) {
	reply, err := e.RedisExec("INCRBY", "activityId", 1)
	if err != nil {
		return 0, err
	}
	return redis.Int(reply, err)
}

// Close 关闭连接（Redis 必关；Mysql 由 Manager 统一关一次）
func (e *Engine) Close(closeMysql bool) {
	if e == nil {
		return
	}
	if e.Redis != nil {
		e.Redis.Close()
	}
	if closeMysql && e.Mysql != nil {
		_ = e.Mysql.Close()
	}
}
