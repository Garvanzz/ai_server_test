package magic

import (
	"encoding/json"
	"fmt"
	"sort"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_magic"
)

func Init(pl *model.Player) {
	pl.Magic = new(model.Magic)
	pl.Magic.Ids = make(map[int32]*model.MagicItem, 0)
	pl.Magic.LineUp = []int32{0, 0, 0, 0, 0, 0}
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Magic)
	if err != nil {
		log.Error("player[%v],save magic marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerMagic, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerMagic, pl.Id))
	if err != nil {
		log.Error("player[%v],load stage error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Magic)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load magic unmarshal error:%v", pl.Id, err)
	}

	pl.Magic = m
}

// ReqmagicList 请求法术
func ReqMagicInit(ctx global.IPlayer, pl *model.Player, req *proto_magic.C2SInitMagic) {
	ctx.Send(&proto_magic.S2CInitMagic{
		Option: model.ToMagicProto(pl.Magic),
	})
}

// ReqmagicUp 请求法术升级
func ReqMagicUpLevel(ctx global.IPlayer, pl *model.Player, req *proto_magic.C2SUpLevelMagic) {
	res := &proto_magic.S2CUpLevelMagic{}
	log.Debug(" 请求法术升级:%v", req.Id)
	//获取等级对应的法术
	if _, ok := pl.Magic.Ids[req.Id]; !ok {
		res.Code = proto_magic.ERRORCODEMAGIC_ERROR_NOMAGIC
		ctx.Send(res)
		return
	}

	magic := pl.Magic.Ids[req.Id]

	confs := config.HeroMagicLevel.All()
	conf := conf2.HeroMagicLevel{}
	for _, v := range confs {
		if v.Level == magic.Level+1 && v.MagicId == magic.Id {
			conf = v
			break
		}
	}
	log.Debug(" 请求法术升级:%v, %v", magic.Num, conf.UplevelCost)
	if conf.Id <= 0 {
		res.Code = proto_magic.ERRORCODEMAGIC_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}
	//判断个数
	if magic.Num < conf.UplevelCost {
		res.Code = proto_magic.ERRORCODEMAGIC_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	magic.Num -= conf.UplevelCost
	magic.Level += 1
	pl.Magic.Ids[req.Id] = magic
	res.Code = proto_magic.ERRORCODEMAGIC_ERR_Ok
	res.Magics = model.ToMagicProto(pl.Magic).Magics
	ctx.Send(res)
}

// ReqmagicUp 请求法术一键升级
func ReqMagicOneKeyUpLevel(ctx global.IPlayer, pl *model.Player, req *proto_magic.C2SOneKeyUplevel) {
	res := &proto_magic.S2COneKeyUplevel{}
	log.Debug("请求一键升级")
	confs := config.HeroMagicLevel.All()
	for id, _ := range pl.Magic.Ids {
		magic := pl.Magic.Ids[id]

		level := magic.Level
		for i := 0; ; i++ {
			conf := conf2.HeroMagicLevel{}
			for _, v := range confs {
				if v.Level == level+1 && v.MagicId == magic.Id {
					conf = v
					break
				}
			}
			log.Debug("请求一键升级:conf:%v", conf)
			if conf.Id <= 0 {
				continue
			}
			//判断个数
			if magic.Num < conf.UplevelCost {
				magic.Level = level
				break
			}

			magic.Num -= conf.UplevelCost
			level++
		}

		pl.Magic.Ids[id] = magic
	}

	res.Magics = model.ToMagicProto(pl.Magic).Magics
	ctx.Send(res)
}

// Reqmagicwear 请求法术装配
func ReqMagicWear(ctx global.IPlayer, pl *model.Player, req *proto_magic.C2SWearMagic) {
	res := &proto_magic.S2CWearMagic{}

	//获取等级对应的法术
	if _, ok := pl.Magic.Ids[req.Id]; !ok {
		res.Code = proto_magic.ERRORCODEMAGIC_ERROR_NOMAGIC
		ctx.Send(res)
		return
	}

	//判断类型
	conf := config.HeroMagic.All()[int64(req.Id)]
	if req.Index >= 1 && req.Index <= 3 {
		if conf.Type != 1 {
			res.Code = proto_magic.ERRORCODEMAGIC_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}
	} else if req.Index >= 4 && req.Index <= 6 {
		if conf.Type != 2 {
			res.Code = proto_magic.ERRORCODEMAGIC_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}
	}

	pl.Magic.LineUp[req.Index-1] = req.Id
	res.Code = proto_magic.ERRORCODEMAGIC_ERR_Ok
	res.LineUp = pl.Magic.LineUp
	ctx.Send(res)
}

// Reqmagicwear 请求法术卸下
func ReqMagicXiexia(ctx global.IPlayer, pl *model.Player, req *proto_magic.C2SXieXiaMagic) {
	res := &proto_magic.S2CXieXiaMagic{}

	//获取等级对应的法术
	if _, ok := pl.Magic.Ids[req.Id]; !ok {
		res.Code = proto_magic.ERRORCODEMAGIC_ERROR_NOMAGIC
		ctx.Send(res)
		return
	}

	if pl.Magic.LineUp[req.Index] != req.Id {
		res.Code = proto_magic.ERRORCODEMAGIC_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	pl.Magic.LineUp[req.Index] = 0
	res.Code = proto_magic.ERRORCODEMAGIC_ERR_Ok
	res.LineUp = pl.Magic.LineUp
	ctx.Send(res)
}

// Reqmagicwear 请求法术一键装配
func ReqMagicOneKeyWear(ctx global.IPlayer, pl *model.Player, req *proto_magic.C2SOneKeyWear) {
	res := &proto_magic.S2COneKeyWear{}

	fashu := make([]*model.MagicItem, 0)
	shentong := make([]*model.MagicItem, 0)
	confs := config.HeroMagic.All()
	for _, v := range pl.Magic.Ids {
		conf := confs[int64(v.Id)]
		if conf.Type == 1 {
			fashu = append(fashu, v)
		} else if conf.Type == 2 {
			shentong = append(shentong, v)
		}
	}

	//排序
	sort.Slice(fashu, func(i, j int) bool {
		confi := confs[int64(i)]
		confj := confs[int64(j)]
		if confi.Rate == confj.Rate {
			return fashu[i].Level > fashu[j].Level
		} else {
			return confi.Rate > confj.Rate
		}
	})

	sort.Slice(shentong, func(i, j int) bool {
		confi := confs[int64(i)]
		confj := confs[int64(j)]
		if confi.Rate == confj.Rate {
			return shentong[i].Level > shentong[j].Level
		} else {
			return confi.Rate > confj.Rate
		}
	})

	//取前3个
	for i := 1; i <= 3; i++ {
		if len(fashu) >= i {
			pl.Magic.LineUp[i-1] = fashu[i-1].Id
		} else {
			pl.Magic.LineUp[i-1] = 0
		}
	}

	for i := 1; i <= 3; i++ {
		if len(shentong) >= i {
			pl.Magic.LineUp[i-1+3] = shentong[i-1].Id
		} else {
			pl.Magic.LineUp[i-1+3] = 0
		}
	}

	res.LineUp = pl.Magic.LineUp
	ctx.Send(res)
}
