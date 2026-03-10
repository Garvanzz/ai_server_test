package db

import (
	"time"
	"xfx/core/db"
	"xfx/gm_server/conf"
	"xfx/pkg/env"

	"github.com/gomodule/redigo/redis"

	"xorm.io/xorm"
)

var AccountDb *xorm.Engine

func init() {
	AccountDb = db.NewMysqlEngine(&env.Mysql{
		CommonAddr: conf.Server.AccountAddr,
	})
}

type RedisEngine struct {
	Redis *redis.Pool
}

func InitRedis(dbAddr string, password string) *RedisEngine {
	_redisEngine := new(RedisEngine)
	_redisEngine.Start(dbAddr, password)
	return _redisEngine
}

func (db *RedisEngine) Start(dbAddr string, password string) {
	db.Redis = &redis.Pool{
		MaxIdle:     400,
		MaxActive:   2000,
		IdleTimeout: 60 * 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", dbAddr, redis.DialPassword(password))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

func (db *RedisEngine) RedisExec(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn := db.Redis.Get()
	defer conn.Close()

	return conn.Do(cmd, args...)
}

func (db *RedisEngine) getConn() redis.Conn {
	return db.Redis.Get()
}
