package player

import (
	"xfx/core/db"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/messages"
	"xfx/proto/proto_player"
)

func dispatch(ctx global.IPlayer, pl *model.Player, msg any) any {
	switch msg := msg.(type) {
	case *messages.DispatchMessage: // 返回客户端消息
		ctx.Send(msg.Content)
	case *db.RedisRet:
		OnRet(ctx, pl, msg)

	//人物
	case *proto_player.C2SChangeName:
		ReqChangePlayerName(ctx, pl, msg)
	case *proto_player.C2SChangeTitle: //改称号
		ReqChangeTitle(ctx, pl, msg)
	case *proto_player.C2SChangeHead: //改头像
		ReqChangeHead(ctx, pl, msg)
	case *proto_player.C2SChangeHeadFrame: //改头像框
		ReqChangeHeadFrame(ctx, pl, msg)
	case *proto_player.C2STransformJob: //改职业
		ReqTransformJob(ctx, pl, msg)
	case *proto_player.C2SChangeSex: //改性别
		ReqChangeSex(ctx, pl, msg)
	case *proto_player.C2SChangeBubble: //改泡泡
		ReqChangeBubble(ctx, pl, msg)
	case *proto_player.C2SGetPlayerById:
		ReqGetPlayerInfoById(ctx, pl, msg) //获取个人信息

	case *messages.SysMessage: // TODO: 系统消息
	case *messages.GetPlayerDataMessage: //获取玩家数据
		return fromProps(pl)
	default:
		//return game.Process(ctx, pl, msg)
	}
	return nil
}
