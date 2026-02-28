package internal

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_item"
)

// PushResPassportScoreChange
func PushResPassportScoreChange(ctx global.IPlayer, pl *model.Player, items map[int32]int32) {
	if len(items) == 0 {
		log.Debug("add nil items, player id : %v", pl.Id)
		return
	}
	if len(items) > 0 {
		ctx.Send(&proto_item.PushPopReward{Items: global.PassportScoreFormatWithMap(items)})
	}
}
