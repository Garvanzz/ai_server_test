package openbox

import (
	"encoding/json"
	"fmt"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_openbox"
)

func Init(pl *model.Player) {
	pl.OpenBox = new(model.OpenBox)
	//默认1级
	pl.OpenBox.Level = 1
	pl.OpenBox.NextScoreBox = define.OPENBOX_MONEY
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.OpenBox)
	if err != nil {
		log.Error("player[%v],save openbox marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save openbox error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerOpenBox, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save openbox error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerOpenBox, pl.Id))
	if err != nil {
		log.Error("player[%v],load bag error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.OpenBox)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load openbox unmarshal error:%v", pl.Id, err)
	}

	pl.OpenBox = m
}

// ReqInit 请求开箱子数据
func ReqInitOpenBox(ctx global.IPlayer, pl *model.Player, req *proto_openbox.C2SInitOpenBox) {
	resp := new(proto_openbox.S2CInitOpenBox)

	ref := RefreshBoxLevelUpTime(pl)
	if ref == false {
		ctx.Send(resp)
		return
	}

	resp.Box = &proto_openbox.BoxOption{
		Level:           pl.OpenBox.Level,
		Exp:             pl.OpenBox.Exp,
		Score:           pl.OpenBox.Score,
		LastUpLevelTime: pl.OpenBox.LastUpTime,
		IsUpLevel:       pl.OpenBox.IsUpLevelBox,
		NextScoreBox:    pl.OpenBox.NextScoreBox,
	}
	ctx.Send(resp)
}

// 刷新宝箱的时间
func RefreshBoxLevelUpTime(pl *model.Player) bool {
	if pl.OpenBox.IsUpLevelBox {
		boxdropConf, ok := config.BoxLevelDrop.Find(int64(pl.OpenBox.Level))
		if !ok {
			return false
		}

		if utils.Now().Unix()-pl.OpenBox.LastUpTime >= int64(boxdropConf.UpTime*3600) {
			pl.OpenBox.Level += 1
			pl.OpenBox.IsUpLevelBox = false
			pl.OpenBox.LastUpTime = 0
		}
	}

	return true
}

// ReqInit 请求升级
func ReqUpLevelBox(ctx global.IPlayer, pl *model.Player, req *proto_openbox.C2SUpBoxLevel) {
	resp := new(proto_openbox.S2CUpBoxLevel)

	ref := RefreshBoxLevelUpTime(pl)
	if ref == false {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	boxdropConf, ok := config.BoxLevelDrop.Find(int64(pl.OpenBox.Level))
	if !ok {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	if pl.OpenBox.IsUpLevelBox {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	if pl.OpenBox.Exp < boxdropConf.UpExp {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_ITEMNOENGTH
		ctx.Send(resp)
		return
	}

	pl.OpenBox.LastUpTime = utils.Now().Unix()
	pl.OpenBox.IsUpLevelBox = true
	resp.Box = &proto_openbox.BoxOption{
		Level:           pl.OpenBox.Level,
		Exp:             pl.OpenBox.Exp,
		Score:           pl.OpenBox.Score,
		LastUpLevelTime: pl.OpenBox.LastUpTime,
		IsUpLevel:       pl.OpenBox.IsUpLevelBox,
		NextScoreBox:    pl.OpenBox.NextScoreBox,
	}
	resp.Code = proto_openbox.ERRORCODEOPENBOX_ERR_Ok
	ctx.Send(resp)
}

// ReqInit 请求积分换宝箱
func ReqSocreBuyBox(ctx global.IPlayer, pl *model.Player, req *proto_openbox.C2SScoreBox) {
	resp := new(proto_openbox.S2CScoreBox)

	ref := RefreshBoxLevelUpTime(pl)
	if ref == false {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	boxdropConf, ok := config.BoxLevelDrop.Find(int64(pl.OpenBox.Level))
	if !ok {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	awards := make([]conf2.ItemE, 0)
	nextBox := pl.OpenBox.NextScoreBox
	score := pl.OpenBox.Score

	//全部兑换
	if req.IsAll {
		for i := 0; ; i++ {
			price := boxdropConf.ScoreBuy[nextBox-1]
			if price > score {
				pl.OpenBox.NextScoreBox = nextBox
				pl.OpenBox.Score = score
				break
			}

			if nextBox == define.OPENBOX_MONEY {
				awards = append(awards, conf2.ItemE{
					ItemId:   define.ItemIdBoxMoney,
					ItemType: define.ItemTypeItem,
					ItemNum:  1,
				})
			} else if nextBox == define.OPENBOX_STORE {
				awards = append(awards, conf2.ItemE{
					ItemId:   define.ItemIdBoxStore,
					ItemType: define.ItemTypeItem,
					ItemNum:  1,
				})
			} else if nextBox == define.OPENBOX_EQUIP {
				awards = append(awards, conf2.ItemE{
					ItemId:   define.ItemIdBoxEquip,
					ItemType: define.ItemTypeItem,
					ItemNum:  1,
				})
			} else if nextBox == define.OPENBOX_MAGIC {
				awards = append(awards, conf2.ItemE{
					ItemId:   define.ItemIdBoxMagick,
					ItemType: define.ItemTypeItem,
					ItemNum:  1,
				})
			}

			nextBox = int32(utils.WeightIndex(boxdropConf.ScoreRefWeight)) + 1
			score -= price
		}
	} else {
		price := boxdropConf.ScoreBuy[nextBox-1]
		if price > score {
			resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_ITEMNOENGTH
			ctx.Send(resp)
			return
		}

		if nextBox == define.OPENBOX_MONEY {
			awards = append(awards, conf2.ItemE{
				ItemId:   define.ItemIdBoxMoney,
				ItemType: define.ItemTypeItem,
				ItemNum:  1,
			})
		} else if nextBox == define.OPENBOX_STORE {
			awards = append(awards, conf2.ItemE{
				ItemId:   define.ItemIdBoxStore,
				ItemType: define.ItemTypeItem,
				ItemNum:  1,
			})
		} else if nextBox == define.OPENBOX_EQUIP {
			awards = append(awards, conf2.ItemE{
				ItemId:   define.ItemIdBoxEquip,
				ItemType: define.ItemTypeItem,
				ItemNum:  1,
			})
		} else if nextBox == define.OPENBOX_MAGIC {
			awards = append(awards, conf2.ItemE{
				ItemId:   define.ItemIdBoxMagick,
				ItemType: define.ItemTypeItem,
				ItemNum:  1,
			})
		}

		nextBox = int32(utils.WeightIndex(boxdropConf.ScoreRefWeight)) + 1
		score -= price
		pl.OpenBox.NextScoreBox = nextBox
		pl.OpenBox.Score = score
	}

	if len(awards) <= 0 {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	bag.AddAward(ctx, pl, awards, false)
	resp.Box = &proto_openbox.BoxOption{
		Level:           pl.OpenBox.Level,
		Exp:             pl.OpenBox.Exp,
		Score:           pl.OpenBox.Score,
		LastUpLevelTime: pl.OpenBox.LastUpTime,
		IsUpLevel:       pl.OpenBox.IsUpLevelBox,
		NextScoreBox:    pl.OpenBox.NextScoreBox,
	}
	resp.Code = proto_openbox.ERRORCODEOPENBOX_ERR_Ok
	ctx.Send(resp)
}

// ReqOpenBox 请求开箱子
func ReqOpenBox(ctx global.IPlayer, pl *model.Player, req *proto_openbox.C2SOpenBox) {
	resp := new(proto_openbox.S2COpenBox)

	if req.Count > config.Global.Get().UseItemMaxCount {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_OUTLIMIT
		ctx.Send(resp)
		return
	}

	//使用限制
	if req.Count > 100 {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_OUTLIMIT
		ctx.Send(resp)
		return
	}

	ref := RefreshBoxLevelUpTime(pl)
	if ref == false {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	//判断道具
	costs := make(map[int32]int32)
	var itemConfig conf2.Item
	itemConfigs := config.Item.All()
	if req.Type == define.OPENBOX_MONEY {
		costs[define.ItemIdBoxMoney] = req.Count
		itemConfig = itemConfigs[int64(define.ItemIdBoxMoney)]
	} else if req.Type == define.OPENBOX_STORE {
		costs[define.ItemIdBoxStore] = req.Count
		itemConfig = itemConfigs[int64(define.ItemIdBoxStore)]
	} else if req.Type == define.OPENBOX_EQUIP {
		costs[define.ItemIdBoxEquip] = req.Count
		itemConfig = itemConfigs[int64(define.ItemIdBoxEquip)]
	} else if req.Type == define.OPENBOX_MAGIC {
		costs[define.ItemIdBoxMagick] = req.Count
		itemConfig = itemConfigs[int64(define.ItemIdBoxMagick)]
	}

	//判断道具是否足够
	if !internal.CheckItemsEnough(pl, costs) {
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_ITEMNOENGTH
		ctx.Send(resp)
		return
	}

	if itemConfig.Type == define.BagItemTypeBox && itemConfig.UseValue <= 0 {
		log.Error("道具不可使用:%v", req.Type)
		resp.Code = proto_openbox.ERRORCODEOPENBOX_ERROR_PARAMERROR
		ctx.Send(resp)
		return
	}

	//删除道具
	internal.SubItems(ctx, pl, costs)

	awards := make([]conf2.ItemE, 0)
	score := pl.OpenBox.Score
	exp := pl.OpenBox.Exp
	for i := int32(0); i < req.Count; i++ {
		award, _score, _exp := GetBoxDrop(itemConfig.UseValue, pl.OpenBox.Level, pl)
		score += _score
		exp += _exp
		awards = append(awards, award...)
	}

	pl.OpenBox.Score = score
	pl.OpenBox.Exp = exp
	bag.AddAward(ctx, pl, awards, false)

	//任务
	task.Dispatch(ctx, pl, define.TaskOpenBoxTime, req.Count, 0, true)
	if req.Type == define.OPENBOX_MAGIC {
		task.Dispatch(ctx, pl, define.TaskOpenMagicBoxTime, req.Count, 0, true)
	}

	itemids := make([]*proto_openbox.ItemOption, 0)
	for _, v := range awards {
		itemids = append(itemids, &proto_openbox.ItemOption{
			Type:  v.ItemType,
			Value: v.ItemId,
			Num:   v.ItemNum,
		})
	}

	resp.Items = itemids
	resp.Box = &proto_openbox.BoxOption{
		Level:           pl.OpenBox.Level,
		Exp:             pl.OpenBox.Exp,
		Score:           pl.OpenBox.Score,
		LastUpLevelTime: pl.OpenBox.LastUpTime,
		IsUpLevel:       pl.OpenBox.IsUpLevelBox,
		NextScoreBox:    pl.OpenBox.NextScoreBox,
	}
	ctx.Send(resp)
}

// GetDrop 获取掉落
func GetBoxDrop(typ int32, level int32, pl *model.Player) (l []conf2.ItemE, score, exp int32) {
	//获取等级
	boxdropConf, ok := config.BoxLevelDrop.Find(int64(level))
	if !ok {
		log.Error("GetBoxdropConf conf not found %v", typ)
		return nil, 0, 0
	}

	//对权重进行计算
	if typ == define.OPENBOX_MONEY {
		//金币
		moneynum := boxdropConf.Money[0]
		//浮动值
		rangNum := utils.RandInt(0, boxdropConf.Money[0]*config.Global.Get().MoneyBoxRange/100)
		l = append(l, conf2.ItemE{
			ItemId:   define.ItemIdMoney,
			ItemType: define.ItemTypeItem,
			ItemNum:  moneynum + rangNum,
		})
		score += boxdropConf.GetScore[0]
		exp += boxdropConf.Exp[0]
	} else if typ == define.OPENBOX_STORE {
		//突破石
		l = append(l, conf2.ItemE{
			ItemId:   define.ItemIdTupoStore,
			ItemType: define.ItemTypeItem,
			ItemNum:  boxdropConf.Store,
		})

		//金币
		moneynum := boxdropConf.Money[1]
		//浮动值
		rangNum := utils.RandInt(0, boxdropConf.Money[1]*config.Global.Get().MoneyBoxRange/100)
		l = append(l, conf2.ItemE{
			ItemId:   define.ItemIdMoney,
			ItemType: define.ItemTypeItem,
			ItemNum:  moneynum + rangNum,
		})

		score += boxdropConf.GetScore[1]
		exp += boxdropConf.Exp[1]
	} else if typ == define.OPENBOX_EQUIP {
		equipFunc := func() {
			//获取装备表
			rare := utils.WeightIndex(boxdropConf.EquipWeight)
			equipConf := config.Equip.All()
			equipId := int32(0)
			heroConf, _ := config.Hero.Find(int64(pl.GetProp(define.PlayerPropHeroId)))
			for _, v := range equipConf {
				if v.Rate == int32(rare+1) {
					//主角匹配职业
					if heroConf.Job != 0 && v.Index == 1 && v.HeroAttId != heroConf.Id {
						continue
					} else {
						equipId = v.Id
						break
					}
				}
			}

			if equipId > 0 {
				l = append(l, conf2.ItemE{
					ItemId:   equipId,
					ItemType: define.ItemTypeEquip,
					ItemNum:  1,
				})
			}
		}

		//双倍奖励
		for i := 0; i < 2; i++ {
			equipFunc()
		}

		//金币
		moneynum := boxdropConf.Money[2]
		//浮动值
		rangNum := utils.RandInt(0, boxdropConf.Money[2]*config.Global.Get().MoneyBoxRange/100)
		l = append(l, conf2.ItemE{
			ItemId:   define.ItemIdMoney,
			ItemType: define.ItemTypeItem,
			ItemNum:  moneynum + rangNum,
		})
		score += boxdropConf.GetScore[2]
		exp += boxdropConf.Exp[2]

	} else if typ == define.OPENBOX_MAGIC {
		magicFunc := func() {
			//获取法术表
			rare := utils.WeightIndex(boxdropConf.MagicWeight)
			magicConf := config.HeroMagic.All()
			magicId := int32(0)
			for _, v := range magicConf {
				if v.Rate == int32(rare+1) {
					magicId = v.Id
					break
				}
			}

			if magicId > 0 {
				l = append(l, conf2.ItemE{
					ItemId:   magicId,
					ItemType: define.ItemTypeMagic,
					ItemNum:  1,
				})
			}
		}

		//双倍奖励
		for i := 0; i < 2; i++ {
			magicFunc()
		}

		//金币
		moneynum := boxdropConf.Money[3]
		//浮动值
		rangNum := utils.RandInt(0, boxdropConf.Money[3]*config.Global.Get().MoneyBoxRange/100)
		l = append(l, conf2.ItemE{
			ItemId:   define.ItemIdMoney,
			ItemType: define.ItemTypeItem,
			ItemNum:  moneynum + rangNum,
		})
		score += boxdropConf.GetScore[3]
		exp += boxdropConf.Exp[3]
	}
	return l, score, exp
}
