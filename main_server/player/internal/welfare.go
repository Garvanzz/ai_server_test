package internal

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_welfare"
)

// PushMonthCard 推送月卡
func PushMonthCard(ctx global.IPlayer, pl *model.Player, typ int32) {
	res := &proto_welfare.PushChangeMonthCard{}
	opt := new(proto_welfare.MonthCardOption)
	log.Debug("月卡数据:%v", pl.Welfare.MonthCard[typ])
	opt.IsGet = pl.Welfare.MonthCard[typ].IsGet
	res.Option = make(map[int32]*proto_welfare.MonthCardOption)
	res.Option[typ] = opt
	log.Debug("推送月卡:%v", res)
	ctx.Send(res)
}
