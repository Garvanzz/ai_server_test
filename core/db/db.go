package db

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"xfx/core/define"
	"xfx/pkg/env"
	"xfx/pkg/log"
	"xfx/pkg/module"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
)

type IDBSink interface {
	//OnRet(ret *CDBRet)
}

// CDBEngine 单服引擎：一进程一 Redis 一 MySQL，连接配置仅来自 env（方案二）
type CDBEngine struct {
	Redis *redis.Pool
	Mysql *xorm.Engine
}

var (
	// Engine 全局唯一数据引擎（单服：env 配置的 Redis + MySQL）
	Engine *CDBEngine
	// CommonEngine 保留别名，与现有代码兼容，指向 Engine
	CommonEngine *CDBEngine
	// currentServerId 本进程服 ID，Start 时从 env 写入；GetEngine 仅在此 ID 时返回引擎
	currentServerId int
	asyncGo         *Go
)

// Start 单服模式：仅用 env 建立本服 Redis 与共享 MySQL，不再读 servergroup 建多连接
func Start(app module.App) {
	currentServerId = app.GetEnv().ID
	Engine = new(CDBEngine)
	Engine.Redis = NewRedisPool(app.GetEnv().Redis.Host, app.GetEnv().Redis.Password, app.GetEnv().Redis.DbNum, app.GetEnv().Redis)
	Engine.Mysql = NewMysqlEngine(app.GetEnv().Mysql.CommonAddr, app.GetEnv().Mysql)
	CommonEngine = Engine

	asyncGo = NewGo(app.System())
	asyncGo.start()
}

func Close() {
	asyncGo.stop()
	if Engine != nil {
		if Engine.Redis != nil {
			Engine.Redis.Close()
		}
		if Engine.Mysql != nil {
			_ = Engine.Mysql.Close()
		}
		Engine = nil
		CommonEngine = nil
	}
}

func NewMysqlEngine(addr string, cfg *env.Mysql) *xorm.Engine {
	_engine, err := xorm.NewEngine("mysql", addr)
	if err != nil {
		panic(err)
	}
	_engine.Logger().SetLevel(xlog.LOG_OFF)
	_engine.ShowSQL(false)

	maxIdle := 240
	maxOpen := 1200
	maxLife := 14400 * time.Second
	if cfg != nil {
		if cfg.MaxIdleConns > 0 {
			maxIdle = cfg.MaxIdleConns
		}
		if cfg.MaxOpenConns > 0 {
			maxOpen = cfg.MaxOpenConns
		}
		if cfg.ConnMaxLifetime > 0 {
			maxLife = cfg.ConnMaxLifetime
		}
	}
	_engine.SetMaxIdleConns(maxIdle)
	_engine.SetMaxOpenConns(maxOpen)
	_engine.SetConnMaxLifetime(maxLife)

	err = _engine.Ping()
	if err != nil {
		fmt.Println("数据库地址:", addr)
		panic("mysql数据库连接失败")
	}
	return _engine
}

func NewRedisPool(host, password string, dataBase int, cfg *env.Redis) *redis.Pool {
	maxIdle := 200
	maxActive := 2000
	idleTimeout := 60 * 60 * time.Second
	connectTimeout := 5 * time.Second
	readTimeout := 3 * time.Second
	writeTimeout := 3 * time.Second

	if cfg != nil {
		if cfg.MaxIdle > 0 {
			maxIdle = cfg.MaxIdle
		}
		if cfg.MaxActive > 0 {
			maxActive = cfg.MaxActive
		}
		if cfg.IdleTimeout > 0 {
			idleTimeout = cfg.IdleTimeout
		}
		if cfg.ConnectTimeout > 0 {
			connectTimeout = cfg.ConnectTimeout
		}
		if cfg.ReadTimeout > 0 {
			readTimeout = cfg.ReadTimeout
		}
		if cfg.WriteTimeout > 0 {
			writeTimeout = cfg.WriteTimeout
		}
	}

	pool := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host,
				redis.DialPassword(password),
				redis.DialDatabase(dataBase),
				redis.DialConnectTimeout(connectTimeout),
				redis.DialReadTimeout(readTimeout),
				redis.DialWriteTimeout(writeTimeout),
			)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < 5*time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	client := pool.Get()
	defer client.Close()

	_, err := client.Do("ping")
	if err != nil {
		panic(err)
	}

	return pool
}

// GetEngine 根据服务器 id 取引擎。单服模式：仅当 serverId == 本进程 env.ID 时返回 Engine
func GetEngine(serverId int) (*CDBEngine, error) {
	if Engine == nil {
		return nil, errors.New("db: not started")
	}
	if serverId != currentServerId {
		return nil, errors.New("db: no this server:" + strconv.Itoa(serverId))
	}
	return Engine, nil
}

// GetEngineByPlayerId 通过玩家 id 取引擎。单服模式：仅有一个 Engine，直接返回
func GetEngineByPlayerId(playerId int64) (*CDBEngine, error) {
	if Engine == nil {
		return nil, errors.New("db: not started")
	}
	return Engine, nil
}

// GetEquipId 获取唯一装备id
func (c *CDBEngine) GetEquipId() (id int, err error) {
	return redis.Int(c.RedisExec("INCRBY", "equipId", 1))
}

// GetActivityId 获取唯一活动id
func (c *CDBEngine) GetActivityId() (id int, err error) {
	return redis.Int(c.RedisExec("INCRBY", "activityId", 1))
}

// GetDelayMailId TODO:Get延时邮件Id 获取延时邮件id
func (c *CDBEngine) GetDelayMailId() (id int, err error) {
	return redis.Int(c.RedisExec("INCRBY", "delayMailId", 1))
}

// GetPlayerId 获取唯一玩家id
func (c *CDBEngine) GetPlayerId(serverId int) (id int64, err error) {
	_id, err := redis.Int(c.RedisExec("INCRBY", "playerId", 1))
	if err != nil {
		return 0, err
	}

	//区号 + ID
	return int64(serverId*define.PlayerIdBase + _id), nil
}

// GetRoomId 获取房间id
func (c *CDBEngine) GetRoomId() (id int64, err error) {
	return redis.Int64(c.RedisExec("INCRBY", "roomId", 1))
}

// Close 关闭该引擎的 Redis 与 MySQL（单服时一般由 db.Close() 统一关闭全局 Engine）
func (c *CDBEngine) Close() {
	if c == nil {
		return
	}
	if c.Redis != nil {
		c.Redis.Close()
	}
	if c.Mysql != nil {
		_ = c.Mysql.Close()
	}
}

// RedisExec 同步直接redis命令
func (c *CDBEngine) RedisExec(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn := c.Redis.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

// RedisAsyncExec 异步执行redis 命令
func (c *CDBEngine) RedisAsyncExec(pid module.PID, opType int, params []int64, cmd string, args ...interface{}) {
	err := asyncGo.submitJob(c.Redis, cmd, args, func(res any, err error) {
		asyncGo.system.Cast(pid, &RedisRet{
			OpType: opType,
			Params: params,
			Reply:  res,
			Err:    err,
		})
	})
	if err != nil {
		log.Error("%v", err)
	}
}
