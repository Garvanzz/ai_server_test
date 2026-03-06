package guild

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
)

// ReqGuildYuanchiInit 帮会元池初始
func ReqGuildYuanchiInit(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SInitYuanchi) {
	resp := new(proto_guild.S2CInitYuanchi)

	//获取自己的帮派数据
	info, err := invoke.GuildClient(ctx).GetYuanchiData(pl.ToContext())
	if err != nil {
		log.Error("invoke guild error:%v", err)
		ctx.Send(resp)
		return
	}

	if info == nil {
		log.Error("invoke guild error is nil")
		ctx.Send(resp)
		return
	}

	log.Debug("info:%v", info)

	resp = info
	ctx.Send(resp)
}

// ReqGuildRefiningLog 帮会炼制记录
func ReqGuildRefiningLog(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SReqRecord) {
	resp := new(proto_guild.S2CReqRecord)

	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save Destiny error, no this server:%v", err)
		ctx.Send(resp)
		return
	}

	key := fmt.Sprintf("guild_refing_history:%d", req.GuildId)
	result, err := rdb.RedisExec("LRANGE", key, 0, -1)
	if err != nil {
		ctx.Send(resp)
		return
	}

	items := result.([]interface{})
	log := new(model.GuildRefiningLog)
	for _, item := range items {
		json.Unmarshal(item.([]byte), log)
		resp.Records = append(resp.Records, &proto_guild.YuanchiRecord{
			Id:       log.Data.Id,
			Alltime:  log.Data.AllTime,
			Time:     log.Data.Time,
			Rare:     log.Data.Rate,
			IsSuc:    log.State,
			Showtime: log.Time,
		})
	}

	ctx.Send(resp)
}
