package db

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/env"
	"xfx/pkg/log"
	"xfx/pkg/module"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
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
	Rs    *redsync.Redsync
}

var (
	Engines      map[int]*CDBEngine // 各个区服连接
	CommonEngine *CDBEngine         // 公用数据库
	asyncGo      *Go
)

func init() {
	Engines = make(map[int]*CDBEngine)
}

// Start 初始服务器列表
func Start(app module.App) {
	//redis
	CommonEngine = new(CDBEngine)
	CommonEngine.Redis = NewRedisPool(app.GetEnv().Redis.Host, app.GetEnv().Redis.Password, app.GetEnv().Redis.DbNum, app.GetEnv().Redis)
	CommonEngine.Rs = redsync.New(redigo.NewPool(CommonEngine.Redis))
	CommonEngine.Mysql = NewMysqlEngine(app.GetEnv().Mysql.CommonAddr, app.GetEnv().Mysql)

	serverItem := new(model.ServerItem)
	ok, err := CommonEngine.Mysql.Table(define.ServerGroup).Where("id = ?", app.GetEnv().ID).Get(serverItem)
	if !ok || err != nil {
		panic("mysql数据库连接失败")
	}

	serverItems := make([]model.ServerItem, 0)
	err = CommonEngine.Mysql.Table(define.ServerGroup).Where("server_group = ?", serverItem.ServerGroup).Find(&serverItems)
	if err != nil {
		fmt.Println(err)
		panic("mysql数据库连接失败")
	}

	// 连接对应的数据库
	for _, v := range serverItems {
		//判断服务器状态
		if v.ServerState != define.ServerStateNormal && v.ServerState != define.ServerStateYongji && v.ServerState != define.ServerStateBaoMan {
			continue
		}

		Engines[int(v.Id)] = NewConnect(v, app)
	}

	asyncGo = NewGo(app.System())
	asyncGo.start()
}

func Close() {
	asyncGo.stop()

	for _, _engine := range Engines {
		_engine.Close()
	}

	CommonEngine.Close()
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

func NewConnect(v model.ServerItem, app module.App) *CDBEngine {
	dbEngine := new(CDBEngine)
	dbEngine.Redis = NewRedisPool(fmt.Sprintf("172.16.1.50:%d", v.RedisPort), app.GetEnv().Redis.Password, app.GetEnv().Redis.DbNum, app.GetEnv().Redis)
	dbEngine.Rs = redsync.New(redigo.NewPool(dbEngine.Redis))
	dbEngine.Mysql = NewMysqlEngine(v.MysqlAddr, app.GetEnv().Mysql)
	return dbEngine
}

// GetEngine 根据服务器id获取数据库
func GetEngine(serverId int) (*CDBEngine, error) {
	_engine, ok := Engines[serverId]
	if !ok {
		return nil, errors.New("no this server:" + strconv.Itoa(serverId))
	}

	return _engine, nil
}

// GetEngineByPlayerId 通过玩家id获取数据库
func GetEngineByPlayerId(playerId int64) (*CDBEngine, error) {
	serverId := playerId / define.PlayerIdBase
	_engine, ok := Engines[int(serverId)]
	if !ok {
		return nil, errors.New(fmt.Sprintf("get server id error:%d", serverId))
	}

	return _engine, nil
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

// GetRedisMutex 获取redis锁
func (c *CDBEngine) GetRedisMutex(key string) *redsync.Mutex {
	return c.Rs.NewMutex(key)
}

func (c *CDBEngine) Close() {
	if c.Redis != nil {
		c.Redis.Close()
	}
	if c.Mysql != nil {
		c.Mysql.Close()
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
