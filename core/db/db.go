package db

import (
	"errors"
	"fmt"
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

type CDBEngine struct {
	Redis *redis.Pool
	Mysql *xorm.Engine
}

var (
	Engine          *CDBEngine
	LoginRedisPool  *redis.Pool // 登录服 Redis（与 login_server 共用，仅 token 等）
	currentServerId int
	asyncGo         *Go
)

func Start(app module.App) {
	currentServerId = app.GetEnv().ID
	Engine = new(CDBEngine)
	Engine.Redis = NewRedisPool(app.GetEnv().Redis.Host, app.GetEnv().Redis.Password, app.GetEnv().Redis.DbNum, app.GetEnv().Redis)
	Engine.Mysql = NewMysqlEngine(app.GetEnv().Mysql)

	if cfg := app.GetEnv().LoginRedis; cfg != nil && cfg.Host != "" {
		LoginRedisPool = NewRedisPool(cfg.Host, cfg.Password, cfg.DbNum, cfg)
	}

	asyncGo = NewGo(app.System())
	asyncGo.start()
}

func Close() {
	asyncGo.stop()
	if LoginRedisPool != nil {
		LoginRedisPool.Close()
		LoginRedisPool = nil
	}
	if Engine != nil {
		if Engine.Redis != nil {
			Engine.Redis.Close()
		}
		if Engine.Mysql != nil {
			_ = Engine.Mysql.Close()
		}
		Engine = nil
	}
}

func NewMysqlEngine(cfg *env.Mysql) *xorm.Engine {
	_engine, err := xorm.NewEngine("mysql", cfg.CommonAddr)
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
		fmt.Println("数据库地址:", cfg.CommonAddr, err)
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

func GetEngine() (*CDBEngine, error) {
	if Engine == nil {
		return nil, errors.New("db: not started")
	}
	return Engine, nil
}

// RedisExec 同步执行 Redis 命令
func RedisExec(cmd string, args ...interface{}) (reply interface{}, err error) {
	if Engine == nil || Engine.Redis == nil {
		return nil, errors.New("db: not started")
	}
	conn := Engine.Redis.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

// RedisLoginExec 使用登录服 Redis 执行命令
func RedisLoginExec(cmd string, args ...interface{}) (reply interface{}, err error) {
	if LoginRedisPool == nil {
		return nil, errors.New("db: login redis not configured")
	}
	conn := LoginRedisPool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

// GetEquipId 获取唯一装备id
func GetEquipId() (id int, err error) {
	return redis.Int(RedisExec("INCRBY", "equipId", 1))
}

// GetActivityId 获取唯一活动id
func GetActivityId() (id int, err error) {
	return redis.Int(RedisExec("INCRBY", "activityId", 1))
}

// GetDelayMailId 获取延时邮件id
func GetDelayMailId() (id int, err error) {
	return redis.Int(RedisExec("INCRBY", "delayMailId", 1))
}

// GetPlayerId 获取唯一玩家id
func GetPlayerId() (id int64, err error) {
	_id, err := redis.Int(RedisExec("INCRBY", "playerId", 1))
	if err != nil {
		return 0, err
	}
	// 区号 + ID
	return int64(currentServerId*define.PlayerIdBase + _id), nil
}

// GetRoomId 获取房间id
func GetRoomId() (id int64, err error) {
	return redis.Int64(RedisExec("INCRBY", "roomId", 1))
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

// RedisAsyncExec 异步执行 Redis 命令
func RedisAsyncExec(pid module.PID, opType int, params []int64, cmd string, args ...interface{}) {
	if Engine == nil || Engine.Redis == nil {
		log.Error("db: RedisAsyncExec not started")
		return
	}
	err := asyncGo.submitJob(Engine.Redis, cmd, args, func(res any, err error) {
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
