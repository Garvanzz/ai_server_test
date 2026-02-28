package base

import (
	"encoding/json"
	"fmt"
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

func Init(pl *model.Player, uid string) {
	pl.Base = new(model.PlayerBase)
	pl.Base.Name = fmt.Sprintf("ID%s", uid)
	pl.Base.CreateTime = time.Now().Unix()
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Base)
	if err != nil {
		log.Error("player[%v],save base marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("Load base error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerBase, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load base error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerBase, pl.Id))
	if err != nil {
		log.Error("player[%v],load base error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl, pl.Uid)
		return
	}

	m := new(model.PlayerBase)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load base unmarshal error:%v", pl.Id, err)
	}
	pl.Base = m
}
