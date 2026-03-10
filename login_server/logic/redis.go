package logic

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

var RedisPool *redis.Pool

func InitRedis(dbAddr, password string, dbNum int) {
	RedisPool = &redis.Pool{
		MaxIdle:     400,
		MaxActive:   2000,
		IdleTimeout: 60 * 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", dbAddr,
				redis.DialPassword(password),
				redis.DialDatabase(dbNum))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

func RedisExec(cmd string, args ...interface{}) (reply interface{}, err error) {
	conn := RedisPool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}
