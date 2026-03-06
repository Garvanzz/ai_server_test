package internal

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
)

// 添加材料
// AddGuildMaterical
func AddGuildMaterical(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	materials := make(map[int32]int32)
	materials[Id] = Num

	err := invoke.GuildClient(ctx).YuanchiAddMaterials(pl.ToContext(), materials)
	if err != nil {
		log.Error("invoke guild error:%v", err)
		return
	}

	//同步变化
}
