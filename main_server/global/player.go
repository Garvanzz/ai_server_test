package global

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"runtime"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/agent"
	"xfx/pkg/log"
)

type IPlayer interface {
	Self() agent.PID
	Cast(pid agent.PID, msg any)
	Call(pid agent.PID, msg any) (any, error)
	Invoke(mod, fn string, args ...any) (any, error)
	InvokeP(pid agent.PID, fn string, args ...any) (any, error)
	Send(msg any)
	Watch(pid agent.PID)
	Unwatch(pid agent.PID)
	OnSave(isSync bool)
	Stop()
}

// GetPlayerInfo 获取玩家基础信息
func GetPlayerInfo(dbId int64) *model.PlayerInfo {
	rdb, err := db.GetEngineByPlayerId(dbId)
	if err != nil {
		log.Error("GetPlayerInfo error, no this server:%v", err)
		return nil
	}
	values, err := redis.Values(rdb.RedisExec("hgetall", fmt.Sprintf("%s:%d", define.Player, dbId)))
	if err != nil {
		log.Error("GetPlayerInfo get redis player %d error:%v", dbId, err)
		return nil
	}

	dst := new(model.PlayerInfo)
	err = redis.ScanStruct(values, dst)
	if err != nil {
		log.Error("GetPlayerInfo scan struct error:%v", err)
		return nil
	}

	return dst
}

func StackTrace() string {
	buf := make([]byte, 4096)
	l := runtime.Stack(buf, false)
	return fmt.Sprintf("%s", buf[:l])
}
