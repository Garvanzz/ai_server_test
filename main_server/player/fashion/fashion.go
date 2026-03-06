package fashion

import (
	"encoding/json"
	"fmt"
	"xfx/core/common"
	"xfx/core/config"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_fashion"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Fashion = new(model.Fashion)
	pl.Fashion.FashionHandbookIds = make([]int32, 0)
	pl.Fashion.FashionItems = make(map[int32]*model.FashionItem)
	pl.Fashion.HeadWearHandbookIds = make([]int32, 0)
	pl.Fashion.HeadWear = make(map[int32]*model.HeadWearItem)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Fashion)
	if err != nil {
		log.Error("player[%v],save Fashion marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save Fashion error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerFashion, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load Fashion error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerFashion, pl.Id))
	if err != nil {
		log.Error("player[%v],load Fashion error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Fashion)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Fashion unmarshal error:%v", pl.Id, err)
	}

	pl.Fashion = m
}

// ReqInitFashion 请求时装数据
func ReqInitFashion(ctx global.IPlayer, pl *model.Player, req *proto_fashion.C2SInitFashion) {
	res := &proto_fashion.S2CInitFashion{}

	res.HandbookExp = pl.Fashion.FashionHandbookExp
	res.HandbookIds = pl.Fashion.FashionHandbookIds

	fashions := model.ToFashionProtoByFashion(pl.Fashion.FashionItems)
	res.Fashions = fashions
	ctx.Send(res)
}

// ReqInitFashion 请求时装使用
func ReqUseFashion(ctx global.IPlayer, pl *model.Player, req *proto_fashion.C2SUseFashion) {
	res := &proto_fashion.S2CUseFashion{}

	for _, v := range pl.Fashion.FashionItems {
		v.Use = false
	}

	if req.Id != 0 {
		if req.Wear {
			for _, v := range pl.Fashion.FashionItems {
				if v.Id == req.Id {
					v.Use = true
					break
				}
			}
		}
		//} else {
		//	//默认
		//	for _, v := range pl.Fashion.FashionItems {
		//		if v.Id == define.DefaultFashionId {
		//			v.Use = true
		//		}
		//	}
		//}
	}

	fashions := model.ToFashionProtoByFashion(pl.Fashion.FashionItems)
	res.Fashions = fashions
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 时装图集奖励
func FashionHandbookAward(ctx global.IPlayer, pl *model.Player, req *proto_fashion.C2SGetFashionHandBookAward) {
	res := &proto_fashion.S2CGetFashionHandBookAward{}
	//算出等级
	level := int32(0)
	ids := make([]int32, 0)
	confs := config.HandbookAward.All()
	for _, v := range confs {
		if v.Type == define.HandbookAwardType_Fashion {
			// 如果玩家经验大于等于配置所需经验，则更新等级
			if pl.Fashion.FashionHandbookExp >= v.Exp {
				ids = append(ids, v.Id)
				// 假设配置是按等级顺序的，取满足条件的最高等级
				if v.Level > level {
					level = v.Level
				}
			}
		}
	}

	//判断等级是否超出限制
	for _, v := range req.Id {
		if !common.IsHaveValueIntArray(ids, v) {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}

		//判断有没有领取的
		if pl.Fashion.FashionHandbookIds != nil {
			if common.IsHaveValueIntArray(pl.Fashion.FashionHandbookIds, v) {
				res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
				ctx.Send(res)
				return
			}
		}
	}

	if pl.Fashion.FashionHandbookIds == nil {
		pl.Fashion.FashionHandbookIds = make([]int32, 0)
	}
	pl.Fashion.FashionHandbookIds = append(pl.Fashion.FashionHandbookIds, req.Id...)
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqInitHeadWear 请求头饰数据
func ReqInitHeadWear(ctx global.IPlayer, pl *model.Player, req *proto_fashion.C2SInitHeadWear) {
	res := &proto_fashion.S2CInitHeadWear{}

	res.HandbookExp = pl.Fashion.HeadWearHandbookExp
	res.HandbookIds = pl.Fashion.HeadWearHandbookIds

	fashions := model.ToHeadWearProtoByHeadWear(pl.Fashion.HeadWear)
	res.Headwear = fashions
	ctx.Send(res)
}

// ReqUseHeadWear 请求头饰使用
func ReqUseHeadWear(ctx global.IPlayer, pl *model.Player, req *proto_fashion.C2SUseHeadWear) {
	res := &proto_fashion.S2CUseHeadWear{}

	for _, v := range pl.Fashion.HeadWear {
		v.Use = false
	}

	if req.Id != 0 {
		if req.Wear {
			for _, v := range pl.Fashion.HeadWear {
				if v.Id == req.Id {
					v.Use = true
					break
				}
			}
		} else {
			//默认
			for _, v := range pl.Fashion.HeadWear {
				if v.Id == define.DefaultHeadWearId {
					v.Use = true
				}
			}
		}
	}

	fashions := model.ToHeadWearProtoByHeadWear(pl.Fashion.HeadWear)
	log.Debug("***:%v", fashions)
	res.Headwear = fashions
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 头饰图集奖励
func HeadWearHandbookAward(ctx global.IPlayer, pl *model.Player, req *proto_fashion.C2SGetHeadWearHandBookAward) {
	res := &proto_fashion.S2CGetHeadWearHandBookAward{}
	//算出等级
	level := int32(0)
	ids := make([]int32, 0)
	confs := config.HandbookAward.All()
	for _, v := range confs {
		if v.Type == define.HandbookAwardType_HeadWear {
			// 如果玩家经验大于等于配置所需经验，则更新等级
			if pl.Fashion.HeadWearHandbookExp >= v.Exp {
				ids = append(ids, v.Id)
				// 假设配置是按等级顺序的，取满足条件的最高等级
				if v.Level > level {
					level = v.Level
				}
			}
		}
	}

	//判断等级是否超出限制
	for _, v := range req.Id {
		if !common.IsHaveValueIntArray(ids, v) {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}

		//判断有没有领取的
		if pl.Fashion.HeadWearHandbookIds != nil {
			if common.IsHaveValueIntArray(pl.Fashion.HeadWearHandbookIds, v) {
				res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
				ctx.Send(res)
				return
			}
		}
	}

	if pl.Fashion.HeadWearHandbookIds == nil {
		pl.Fashion.HeadWearHandbookIds = make([]int32, 0)
	}
	pl.Fashion.HeadWearHandbookIds = append(pl.Fashion.HeadWearHandbookIds, req.Id...)
	res.HandbookIds = pl.Fashion.HeadWearHandbookIds
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}
