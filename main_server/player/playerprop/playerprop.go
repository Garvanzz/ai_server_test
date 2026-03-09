package playerprop

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_player"
)

func Init(pl *model.Player) {
	pl.PlayerProp = new(model.PlayerProp)
	pl.PlayerProp.Titles = []int32{}
	pl.PlayerProp.HeadFrames = []int32{
		define.PlayerHeadFrame,
	}
	pl.PlayerProp.Bubbles = []int32{
		define.PlayerBubbleID,
	}
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.PlayerProp)
	if err != nil {
		log.Error("player[%v],save playerProp marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerProp, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerProp, pl.Id))
	if err != nil {
		log.Error("player[%v],load playerProp error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.PlayerProp)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],playerProp base unmarshal error:%v", pl.Id, err)
	}
	pl.PlayerProp = m
}

// ReqInitPlayerProp 获取初始人物道具
func ReqInitPlayerProp(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SGetPlayerProp) {
	res := &proto_player.S2CGetPlayerProp{}
	res.Titles = pl.PlayerProp.Titles
	res.HeadFrames = pl.PlayerProp.HeadFrames
	res.Bubbles = pl.PlayerProp.Bubbles
	ctx.Send(res)
}
