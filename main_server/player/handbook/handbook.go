package handbook

import (
	"encoding/json"
	"fmt"
	"xfx/pkg/utils"
	"xfx/core/config"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_handbook"
)

func Init(pl *model.Player) {
	pl.Handbook = new(model.Handbook)
	pl.Handbook.Handbooks = make(map[int32]*model.HandbookHero, 0)
	pl.Handbook.HandbookOption = new(model.HandbookOption)
	pl.Handbook.HandbookOption.Level = 0
	pl.Handbook.HandbookOption.Exp = 0
	pl.Handbook.HandbookOption.GetId = make([]int32, 0)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Handbook)
	if err != nil {
		log.Error("player[%v],save bag marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save handbook error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerHandbook, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load handbook error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerHandbook, pl.Id))
	if err != nil {
		log.Error("player[%v],load bag error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Handbook)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load bag unmarshal error:%v", pl.Id, err)
	}
	pl.Handbook = m
}

// 请求图鉴
func ReqHandBookInfo(ctx global.IPlayer, pl *model.Player, req *proto_handbook.C2SHandBookData) {
	resp := new(proto_handbook.S2CHandBookData)
	resp.Ids = model.ToHandBookHeroProtoByHandBook(pl.Handbook.Handbooks)
	resp.HandbookOption = model.ToHandBookOptProtoByHandBook(pl.Handbook.HandbookOption)
	ctx.Send(resp)
}

// 获取经验
func ReqHandBookGetExp(ctx global.IPlayer, pl *model.Player, req *proto_handbook.C2SGetHandBookExp) {
	resp := new(proto_handbook.S2CGetHandBookExp)
	if _, ok := pl.Handbook.Handbooks[req.Id]; !ok {
		resp.Code = proto_handbook.ERRORCODEHANDBOOK_ERROR_NOHERO
		ctx.Send(resp)
		return
	}

	if pl.Handbook.Handbooks[req.Id].IsGetExp == false || pl.Handbook.Handbooks[req.Id].GetExp <= 0 {
		resp.Code = proto_handbook.ERRORCODEHANDBOOK_ERROR_ALGET
		ctx.Send(resp)
		return
	}

	pl.Handbook.HandbookOption.Exp += pl.Handbook.Handbooks[req.Id].GetExp
	//判断经验
	confs := config.HandbookAward.All()

	level := int64(pl.Handbook.HandbookOption.Level)
	exp := pl.Handbook.HandbookOption.Exp
	curLevel := int32(level)
	for k := level; k < 20; k++ {
		conf := confs[k]
		curLevel = int32(k)
		if exp >= conf.Exp {
			exp -= conf.Exp
		} else {
			break
		}
	}

	pl.Handbook.HandbookOption.Exp = exp
	pl.Handbook.HandbookOption.Level = curLevel

	pl.Handbook.Handbooks[req.Id].IsGetExp = false
	pl.Handbook.Handbooks[req.Id].GetExp = 0

	resp.Code = proto_handbook.ERRORCODEHANDBOOK_ERR_Ok
	resp.Id = model.ToHandBookHeroProtoByHandBook(pl.Handbook.Handbooks)
	resp.HandbookOption = model.ToHandBookOptProtoByHandBook(pl.Handbook.HandbookOption)

	ctx.Send(resp)
}

// 领取奖励
func ReqHandBookAward(ctx global.IPlayer, pl *model.Player, req *proto_handbook.C2SGetHandBookAward) {
	resp := new(proto_handbook.S2CGetHandBookAward)

	confs := config.HandbookAward.All()
	for _, v := range req.Id {
		conf := confs[int64(v)]
		if conf.Id <= 0 {
			resp.Code = proto_handbook.ERRORCODEHANDBOOK_ERROR_CONFIGERROR
			ctx.Send(resp)
			return
		}
	}

	for _, k := range pl.Handbook.HandbookOption.GetId {
		if utils.ContainsInt32(req.Id, k) {
			resp.Code = proto_handbook.ERRORCODEHANDBOOK_ERROR_ALGET
			ctx.Send(resp)
			return
		}
	}

	for _, k := range req.Id {
		if k <= pl.Handbook.HandbookOption.Level {
			pl.Handbook.HandbookOption.GetId = append(pl.Handbook.HandbookOption.GetId, k)
		}
	}

	respchange := new(proto_handbook.PushHandBookChange)
	respchange.Ids = model.ToHandBookHeroProtoByHandBook(pl.Handbook.Handbooks)
	respchange.HandbookOption = model.ToHandBookOptProtoByHandBook(pl.Handbook.HandbookOption)
	ctx.Send(respchange)

	resp.Code = proto_handbook.ERRORCODEHANDBOOK_ERR_Ok
	ctx.Send(resp)
}
