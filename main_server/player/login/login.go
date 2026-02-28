package login

import (
	"time"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/agent"
	"xfx/pkg/log"
	Proto_Player "xfx/proto/proto_player"
	Proto_Public "xfx/proto/proto_public"
)

// Login 登录
func Login(ctx global.IPlayer, pl *model.Player) {
	resp := new(Proto_Player.S2CLogin)
	resp.Timestamp = time.Now().Unix()

	// TODO: 服务器时间
	resp.EndTimeUnix = 0
	resp.ZoneOffset = time.Now().Unix()
	resp.State = Proto_Public.CommonState_Success

	player := new(Proto_Player.Player)
	player.Id = pl.Id
	player.Uid = pl.Uid
	player.Name = pl.Base.Name
	player.CreateTime = pl.Base.CreateTime
	player.Level = int32(pl.GetProp(define.PlayerPropLevel))
	player.HeadIcon = int32(pl.GetProp(define.PlayerPropFaceId))
	player.Exp = int32(pl.GetProp(define.PlayerPropExp))
	player.Title = int32(pl.GetProp(define.PlayerPropTitle))
	player.Rank = int32(pl.GetProp(define.PlayerPropRank))
	player.HeadFrame = int32(pl.GetProp(define.PlayerPropFaceSlotId))
	player.Job = int32(pl.GetProp(define.PlayerPropJob))
	player.Sex = int32(pl.GetProp(define.PlayerPropSex))
	player.Clan = int32(pl.GetProp(define.PlayerPropClan))
	player.HeroId = int32(pl.GetProp(define.PlayerPropHeroId))
	resp.Player = player
	log.Debug("登录回调：%v", resp)

	ctx.Send(resp)
}

func Replace(ctx global.IPlayer, pl *model.Player, session agent.PID) error {
	if pl.Cache.Session != nil {
		pl.Cache.Disconnect = true
		ctx.Cast(pl.Cache.Session, &Proto_Player.S2CKick{})
	}
	pl.Cache.Session = session
	return nil
}

func Logout(ctx global.IPlayer, pl *model.Player) {
	log.Debug("====================logout")

	event.DoEvent(define.EventTypePlayerOffline, map[string]any{
		"player": pl.ToContext(),
	})
	invoke.RoomClient(ctx).Disconnect(pl)
	ctx.OnSave(true)
	ctx.Stop()
}

func Disconnect(ctx global.IPlayer, pl *model.Player) {
	if !pl.Cache.Disconnect {
		pl.Cache.Session = nil
	}
	pl.Cache.Disconnect = false
}
