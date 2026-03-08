package equip

import (
	"errors"
	"xfx/pkg/utils"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_equip"
)

// 请求洗练数据
func ReqInitEquipSuccinct(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitEquipSuccinct) {
	res := &proto_equip.S2CInitEquipSuccinct{}
	res.Level = int32(pl.Equip.Succinct.Level)
	res.Exp = pl.Equip.Succinct.Exp
	res.UseIndex = int32(pl.Equip.Succinct.UseIndex)
	res.Succincts = model.ToSuccinctIndexProto(pl.Equip.Succinct.SuccinctIndexs)
	res.Awards = pl.Equip.Succinct.LevelAward
	res.CacheSuccincts = model.ToSuccinctCacheProto(pl.Equip.Succinct.CacheSuccinctIndexs)
	ctx.Send(res)
}

// 请求洗练
func ReqEquipSuccinctSkill(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SSuccinctSkill) {
	res := &proto_equip.S2CSuccinctSkill{}

	configs := config.Succinct.All()
	conf := conf2.Succinct{}
	for _, v := range configs {
		if v.Level == int32(pl.Equip.Succinct.Level) {
			conf = v
			break
		}
	}

	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断道具
	costs := make(map[int32]int32)
	costs[conf.Cost[0].ItemId] = conf.Cost[0].ItemNum
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	skillId, err := GetSuccinctDropSkill(conf.Weight, req.EquipIndex)
	if err != nil {
		log.Error("洗练技能失败:%v", err.Error())
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOSKILL
		ctx.Send(res)
		return
	}

	//删除道具
	internal.SubItems(ctx, pl, costs)

	if pl.Equip.Succinct.CacheSuccinctIndexs == nil {
		pl.Equip.Succinct.CacheSuccinctIndexs = make(map[int]*model.CacheSuccinctIndex)
	}

	pl.Equip.Succinct.CacheSuccinctIndexs[int(req.Index)] = &model.CacheSuccinctIndex{
		Index:      int(req.Index),
		SkillId:    skillId,
		EquipIndex: int(req.EquipIndex),
	}

	//增加经验
	pl.Equip.Succinct.Exp += 1
	if pl.Equip.Succinct.Exp >= conf.Exp {
		pl.Equip.Succinct.Level += 1
		pl.Equip.Succinct.Exp = pl.Equip.Succinct.Exp - conf.Exp
	}

	res.Level = int32(pl.Equip.Succinct.Level)
	res.Exp = pl.Equip.Succinct.Exp
	res.UseIndex = int32(pl.Equip.Succinct.UseIndex)
	res.CacheSuccincts = model.ToSuccinctCacheProto(pl.Equip.Succinct.CacheSuccinctIndexs)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// GetDrop 获取掉落
func GetSuccinctDropSkill(weight []int32, equipIndex int32) (int32, error) {
	//获取表
	rare := utils.WeightIndex(weight)
	skillConf := config.SuccinctSkill.All()
	skillWights := make([]int32, 0)
	skillIds := make([]int32, 0)
	for _, v := range skillConf {
		if v.Rate == int32(rare)+1 {
			if len(v.Limit) <= 0 {
				skillWights = append(skillWights, v.Weight)
				skillIds = append(skillIds, v.Id)
			} else if utils.ContainsInt32(v.Limit, equipIndex) == true {
				skillWights = append(skillWights, v.Weight)
				skillIds = append(skillIds, v.Id)
			}
		}
	}

	if len(skillWights) <= 0 {
		return 0, errors.New("GetSuccinctDropSkill config is null")
	}

	index := utils.WeightIndex(skillWights)
	skillId := skillIds[index]
	return skillId, nil
}

// 使用洗练
func ReqUseSuccinct(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUseSuccinct) {
	res := &proto_equip.S2CUseSuccinct{}

	//判断是不是当前方案
	if int32(pl.Equip.Succinct.UseIndex) != req.Index {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIPINDEX
		ctx.Send(res)
		return
	}

	//判断缓存里面是否有
	if _, ok := pl.Equip.Succinct.CacheSuccinctIndexs[int(req.Index)]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOSKILL
		ctx.Send(res)
		return
	}

	suc := pl.Equip.Succinct.CacheSuccinctIndexs[int(req.Index)]
	if int32(suc.EquipIndex) != req.EquipIndex {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOSKILL
		ctx.Send(res)
		return
	}

	if suc.SkillId != req.SkillId {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOSKILL
		ctx.Send(res)
		return
	}

	if pl.Equip.Succinct.SuccinctIndexs == nil {
		pl.Equip.Succinct.SuccinctIndexs = make(map[int]*model.SuccinctIndex)
	}

	if _, ok := pl.Equip.Succinct.SuccinctIndexs[int(req.Index)]; !ok {
		pl.Equip.Succinct.SuccinctIndexs[int(req.Index)] = new(model.SuccinctIndex)
		pl.Equip.Succinct.SuccinctIndexs[int(req.Index)].SkillId = make(map[int32]int32)
	}

	pl.Equip.Succinct.SuccinctIndexs[int(req.Index)].Index = int(req.Index)
	pl.Equip.Succinct.SuccinctIndexs[int(req.Index)].SkillId[req.EquipIndex] = req.SkillId

	//删除缓存里面
	delete(pl.Equip.Succinct.CacheSuccinctIndexs, int(req.Index))

	res.Succincts = model.ToSuccinctIndexProto(pl.Equip.Succinct.SuccinctIndexs)
	res.CacheSuccincts = model.ToSuccinctCacheProto(pl.Equip.Succinct.CacheSuccinctIndexs)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 删除洗练
func ReqDeleteSuccinct(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SDeleteSuccinct) {
	res := &proto_equip.S2CDeleteSuccinct{}

	//判断是不是当前方案
	if int32(pl.Equip.Succinct.UseIndex) != req.Index {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIPINDEX
		ctx.Send(res)
		return
	}

	//判断缓存里面是否有
	if _, ok := pl.Equip.Succinct.CacheSuccinctIndexs[int(req.Index)]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOSKILL
		ctx.Send(res)
		return
	}

	suc := pl.Equip.Succinct.CacheSuccinctIndexs[int(req.Index)]
	if int32(suc.EquipIndex) != req.EquipIndex {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOSKILL
		ctx.Send(res)
		return
	}

	if suc.SkillId != req.SkillId {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOSKILL
		ctx.Send(res)
		return
	}

	//删除缓存里面
	delete(pl.Equip.Succinct.CacheSuccinctIndexs, int(req.Index))

	res.CacheSuccincts = model.ToSuccinctCacheProto(pl.Equip.Succinct.CacheSuccinctIndexs)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 切换方案
func ReqCutSuccinct(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUseSuccinctScheme) {
	res := &proto_equip.S2CUseSuccinctScheme{}

	//判断是不是当前方案
	if int32(pl.Equip.Succinct.UseIndex) == req.Index {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	pl.Equip.Succinct.UseIndex = int(req.Index)
	res.Index = req.Index
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 方案改名
func ReqSuccinctChangeName(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SSuccincChangeName) {
	res := &proto_equip.S2CSuccincChangeName{}

	//判断是不是当前方案
	if int32(pl.Equip.Succinct.UseIndex) != req.Index {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIPINDEX
		ctx.Send(res)
		return
	}

	//判断缓存里面是否有
	if _, ok := pl.Equip.Succinct.SuccinctIndexs[int(req.Index)]; !ok {
		pl.Equip.Succinct.SuccinctIndexs[int(req.Index)] = new(model.SuccinctIndex)
		pl.Equip.Succinct.SuccinctIndexs[int(req.Index)].Index = int(req.Index)
		pl.Equip.Succinct.SuccinctIndexs[int(req.Index)].SkillId = make(map[int32]int32)
	}
	pl.Equip.Succinct.SuccinctIndexs[int(req.Index)].Name = req.Name
	res.Index = req.Index
	res.Name = req.Name
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 领取奖励
func ReqGetSuccinctAward(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SGetSuccinctLevelAward) {
	res := &proto_equip.S2CGetSuccinctLevelAward{}

	conf := config.Succinct.All()[int64(req.Id)]
	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断等级
	if conf.Level > int32(pl.Equip.Succinct.Level) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//判断领取没有
	if utils.ContainsInt32(pl.Equip.Succinct.LevelAward, req.Id) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	pl.Equip.Succinct.LevelAward = append(pl.Equip.Succinct.LevelAward, req.Id)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	res.Awards = pl.Equip.Succinct.LevelAward
	ctx.Send(res)
}
