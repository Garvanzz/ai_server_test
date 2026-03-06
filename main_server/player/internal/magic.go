package internal

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/proto/proto_magic"
)

// AddMagic 添加装备 key = id, value = 等级
func AddMagic(ctx global.IPlayer, pl *model.Player, Id, num int32) {
	magic := pl.Magic.Ids

	if _, ok := magic[Id]; ok {
		pl.Magic.Ids[Id].Num += num
	} else {
		pl.Magic.Ids[Id] = &model.MagicItem{
			Id:    Id,
			Level: 0,
			Num:   num,
		}

		//通告相关
		SyncNotice_AddMagic(ctx, pl, Id)
	}

	ctx.Send(&proto_magic.PushMagicChange{Option: model.ToMagicProto(pl.Magic)})
}
