package welfare

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
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_welfare"
)

func Init(pl *model.Player) {
	pl.Welfare = new(model.Welfare)
	pl.Welfare.DaySign = new(model.DaySign)
	pl.Welfare.MonthCard = make(map[int32]*model.MonthCard)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Welfare)
	if err != nil {
		log.Error("player[%v],save Welfare marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerWelfare, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerWelfare, pl.Id))
	if err != nil {
		log.Error("player[%v],load task error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Welfare)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load task unmarshal error:%v", pl.Id, err)
	}
	pl.Welfare = m

	// TODO:load new tasks
}

// 签到
func ReqDaySignInit(ctx global.IPlayer, pl *model.Player, req *proto_welfare.C2SDaySign) {
	res := new(proto_welfare.S2CDaySign)

	//判断一下时间
	if pl.Welfare.DaySign.FirstDayTime == 0 {
		pl.Welfare.DaySign.FirstDayTime = utils.Now().Unix()
	} else {
		if utils.DaysBetweenTwoTimeUnix(pl.Welfare.DaySign.FirstDayTime, utils.Now().Unix()) > 30 {
			pl.Welfare.DaySign.FirstDayTime = utils.Now().Unix()
			pl.Welfare.DaySign.IsDaySign = false
			pl.Welfare.DaySign.Day = make([]int32, 0)
			pl.Welfare.DaySign.SignTime = 0
			pl.Welfare.DaySign.AccDay = make([]int32, 0)
		}
	}

	if utils.DaysBetweenTwoTimeUnix(pl.Welfare.DaySign.SignTime, utils.Now().Unix()) >= 1 {
		pl.Welfare.DaySign.IsDaySign = false
	}

	if pl.Welfare.DaySign.Day == nil {
		pl.Welfare.DaySign.Day = make([]int32, 0)
	}
	if pl.Welfare.DaySign.AccDay == nil {
		pl.Welfare.DaySign.AccDay = make([]int32, 0)
	}

	res.IsSign = pl.Welfare.DaySign.IsDaySign
	res.AccDay = pl.Welfare.DaySign.AccDay
	res.Day = pl.Welfare.DaySign.Day
	res.CurDay = utils.DaysBetweenTwoTimeUnix(pl.Welfare.DaySign.FirstDayTime, utils.Now().Unix()) + 1
	log.Debug("sign init date: %v", res)
	ctx.Send(res)
}

// 签到奖励
func ReqSignAward(ctx global.IPlayer, pl *model.Player, req *proto_welfare.C2SDayAward) {
	res := new(proto_welfare.S2CDayAward)

	curday := utils.DaysBetweenTwoTimeUnix(pl.Welfare.DaySign.FirstDayTime, utils.Now().Unix())
	if req.Id > curday+1 {
		res.Code = proto_welfare.ERRORGAMECODE_ERR_NOTSUP
		ctx.Send(res)
		return
	}

	conf, _ := config.DaySign.Find(int64(req.Id))
	//补签
	if req.SignType == define.SignType_BuSignIn {
		have := utils.ContainsInt32(pl.Welfare.DaySign.Day, req.Id)
		if have {
			res.Code = proto_welfare.ERRORGAMECODE_ERR_NOTSUP
			ctx.Send(res)
			return
		}

		//消耗道具
		costItems := make(map[int32]int32)
		costItems[config.Global.Get().SupplementSign[0].ItemId] = config.Global.Get().SupplementSign[0].ItemNum
		if !internal.CheckItemsEnough(pl, costItems) {
			res.Code = proto_welfare.ERRORGAMECODE_ERR_ITEMNOTENOUGH
			ctx.Send(res)
			return
		}

		//扣道具
		internal.SubItems(ctx, pl, costItems)
		pl.Welfare.DaySign.Day = append(pl.Welfare.DaySign.Day, req.Id)
	} else if req.SignType == define.SignType_Normal {

		if pl.Welfare.DaySign.SignTime != 0 {
			if utils.CheckIsSameDayBySec(pl.Welfare.DaySign.SignTime, utils.Now().Unix(), 0) {
				res.Code = proto_welfare.ERRORGAMECODE_ERR_ALSIGN
				ctx.Send(res)
				return
			}
		}

		//签到
		if pl.Welfare.DaySign.IsDaySign == true {
			res.Code = proto_welfare.ERRORGAMECODE_ERR_ALSIGN
			ctx.Send(res)
			return
		}

		have := utils.ContainsInt32(pl.Welfare.DaySign.Day, req.Id)
		if have {
			res.Code = proto_welfare.ERRORGAMECODE_ERR_NOTSUP
			ctx.Send(res)
			return
		}

		pl.Welfare.DaySign.IsDaySign = true
		pl.Welfare.DaySign.SignTime = utils.Now().Unix()
		pl.Welfare.DaySign.Day = append(pl.Welfare.DaySign.Day, req.Id)
	} else if req.SignType == define.SignType_AccSignIn {
		have := utils.ContainsInt32(pl.Welfare.DaySign.AccDay, req.Id)
		if have {
			res.Code = proto_welfare.ERRORGAMECODE_ERR_NOTSUP
			ctx.Send(res)
			return
		}

		pl.Welfare.DaySign.AccDay = append(pl.Welfare.DaySign.AccDay, req.Id)
	}

	//发奖励
	bag.AddAward(ctx, pl, conf.Reward, true)
	res.Code = proto_welfare.ERRORGAMECODE_ERR_Ok
	res.IsSign = pl.Welfare.DaySign.IsDaySign
	res.CurDay = utils.DaysBetweenTwoTimeUnix(pl.Welfare.DaySign.FirstDayTime, utils.Now().Unix()) + 1
	res.Day = pl.Welfare.DaySign.Day
	res.AccDay = pl.Welfare.DaySign.AccDay

	ctx.Send(res)
}

// 月卡
func ReqMonthCardInit(ctx global.IPlayer, pl *model.Player, req *proto_welfare.C2SMonthCardInit) {
	res := new(proto_welfare.S2CMonthCardInit)
	updateMonthCardState(pl)
	res.Option = model.ToWelfareMonthCardProto(pl.Welfare.MonthCard)
	ctx.Send(res)
}

// 刷新月卡数据
func updateMonthCardState(pl *model.Player) {
	for key, v := range pl.Welfare.MonthCard {
		//判断下过期没
		if key == define.MonthCard_GemAppraisal {
			//判断有没有购买鉴宝月卡

			reply, err := db.RedisExec("HGET", define.GemAppraisal_MonthCard, pl.Id)
			if err != nil {
				log.Error("[%v],load getExchangeAddNum error:%v", pl.Id, err)
				continue
			}

			if reply == nil {
				delete(pl.Welfare.MonthCard, key)
				continue
			}

			if !utils.CheckIsSameDayBySec(v.GetTime, utils.Now().Unix(), 0) {
				pl.Welfare.MonthCard[key].IsGet = false
			}
		}
	}
}

// 领取月卡
func ReqMonthCardGetAward(ctx global.IPlayer, pl *model.Player, req *proto_welfare.C2SGetMonthCard) {
	res := new(proto_welfare.S2CGetMonthCard)

	//鉴宝
	if req.Type == define.MonthCard_GemAppraisal {
		reply, err := db.RedisExec("HGET", define.GemAppraisal_MonthCard, pl.Id)
		if err != nil {
			log.Error("[%v],load ReqMonthCardGetAward error:%v", pl.Id, err)
			res.Code = proto_welfare.ERRORGAMECODE_ERR_NOTSUP
			ctx.Send(res)
			return
		}

		if reply == nil {
			log.Error("ReqMonthCardGetAward error, no this server:%v", err)
			res.Code = proto_welfare.ERRORGAMECODE_ERR_NOBUY
			ctx.Send(res)
			return
		}

		m := new(model.GemAppraisalMonthCard)
		err = json.Unmarshal(reply.([]byte), &m)
		if err != nil {
			log.Error("player[%v],load ReqMonthCardGetAward unmarshal error:%v", pl.Id, err)
			res.Code = proto_welfare.ERRORGAMECODE_ERR_NOTSUP
			ctx.Send(res)
			return
		}

		//判断生效没
		if m.GetDay >= m.EffectDay {
			res.Code = proto_welfare.ERRORGAMECODE_ERR_NOTSUP
			ctx.Send(res)
			return
		}

		monthCard := new(model.MonthCard)
		if _, ok := pl.Welfare.MonthCard[req.Type]; !ok {
			monthCard = pl.Welfare.MonthCard[req.Type]
		}

		if monthCard.GetTime <= 0 {
			monthCard.GetTime = utils.Now().Unix()
			monthCard.IsGet = true
		} else {
			//判断是否领取
			if utils.CheckIsSameDayBySec(monthCard.GetTime, utils.Now().Unix(), 0) {
				res.Code = proto_welfare.ERRORGAMECODE_ERR_ALGET
				ctx.Send(res)
				return
			}

			monthCard.IsGet = true
			monthCard.GetTime = utils.Now().Unix()
		}

		//玩家数据
		pl.Welfare.MonthCard[req.Type] = monthCard

		//鉴宝月卡表
		m.GetDay += 1
		m.GetTime = utils.Now().Unix()
		js, _ := json.Marshal(m)
		db.RedisExec("HSET", define.GemAppraisal_MonthCard, pl.Id, js)

		res.Option = &proto_welfare.MonthCardOption{
			IsGet: monthCard.IsGet,
		}
		//奖励
		confs := config.MonthCard.All()
		conf := conf2.MonthCard{}
		for _, v := range confs {
			if v.Type == req.Type {
				conf = v
				break
			}
		}
		if conf.Id > 0 && len(conf.Reward) > 0 {
			bag.AddAward(ctx, pl, conf.Reward, true)
		}
	}

	res.Code = proto_welfare.ERRORGAMECODE_ERR_Ok
	ctx.Send(res)
}

// 功能开启
func ReqFuncOpenInit(ctx global.IPlayer, pl *model.Player, req *proto_welfare.C2SFunctionOpenInit) {
	res := new(proto_welfare.S2CFunctionOpenInit)
	if pl.Welfare.FunctionOpen == nil {
		pl.Welfare.FunctionOpen = make([]string, 0)
	}
	res.OpenAward = pl.Welfare.FunctionOpen
	ctx.Send(res)
}

// 功能开启奖励
func ReqFuncOpenAward(ctx global.IPlayer, pl *model.Player, req *proto_welfare.C2SFunctionAward) {
	res := new(proto_welfare.S2CFunctionAward)
	if utils.ContainsString(pl.Welfare.FunctionOpen, req.Mod) {
		res.OpenAward = pl.Welfare.FunctionOpen
		ctx.Send(res)
		return
	}

	//判断条件
	var conf conf2.FunctionOpen
	confs := config.FunctionOpen.All()
	for _, v := range confs {
		if v.Type == req.Mod {
			conf = v
			break
		}
	}

	if conf.Id <= 0 {
		log.Error("没有找到配置:%v", req.Mod)
		res.OpenAward = pl.Welfare.FunctionOpen
		ctx.Send(res)
		return
	}

	state := internal.FuncOpenJudgeLogic(conf, pl)
	if !state {
		log.Error("功能开启不满足条件:%v", req.Mod)
		res.OpenAward = pl.Welfare.FunctionOpen
		ctx.Send(res)
		return
	}
	log.Debug("*********456")
	//奖励
	award := conf.Reward
	bag.AddAward(ctx, pl, award, true)

	pl.Welfare.FunctionOpen = append(pl.Welfare.FunctionOpen, req.Mod)
	res.OpenAward = pl.Welfare.FunctionOpen
	ctx.Send(res)
}
