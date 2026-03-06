package match

import (
	"time"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
)

var Module = func() module.Module {
	return Mgr
}

var Mgr *Manager

func init() {
	Mgr = new(Manager)
	Mgr.topPkMatch = new(Match)
	Mgr.rankMatch = new(Match)
	Mgr.arenaMatch = new(Match)
}

type Manager struct {
	modules.BaseModule
	topPkMatch *Match // 巅峰决斗
	arenaMatch *Match //竞技场
	rankMatch  *Match // 天梯
}

func (m *Manager) OnInit(app module.App) {
	m.BaseModule.OnInit(app)

	m.topPkMatch.init()
	m.rankMatch.init()
	m.arenaMatch.init()

	m.Register("StartMatch", m.OnStartMatch)
	m.Register("CancelMatch", m.OnCancelMatch)
}

func (m *Manager) GetType() string { return define.ModuleMatch }

func (m *Manager) OnTick(delta time.Duration) {
	m.rankMatch.update()
	m.topPkMatch.update()
	m.arenaMatch.update()
}

func (m *Manager) OnMessage(msg interface{}) interface{} {
	log.Debug("* Match message %v", msg)
	return nil
}

// OnStartMatch 开始匹配
func (m *Manager) OnStartMatch(team *model.MatchTeam) bool {
	log.Debug("开始匹配")

	if team.Type == define.MatchModRank {
		return m.rankMatch.startMatch(team)
	} else if team.Type == define.MatchModTopPk {
		return m.topPkMatch.startMatch(team)
	} else if team.Type == define.MatchModArena {
		return m.arenaMatch.startMatch(team)
	}
	return false
}

// OnCancelMatch 取消匹配
func (m *Manager) OnCancelMatch(mod, roomId int32) bool {
	log.Debug("取消匹配")

	if mod == define.MatchModRank {
		return m.rankMatch.cancelMatch(roomId)
	} else if mod == define.MatchModTopPk {
		return m.topPkMatch.cancelMatch(roomId)
	} else if mod == define.MatchModArena {
		return m.arenaMatch.cancelMatch(roomId)
	}
	return true
}

// 通知匹配到的队列
func notifyTeam(team1, team2 *model.MatchTeam) {
	// TODO:改成异步
	invoke.RoomClient(Mgr).MatchToRoomInfo(team1.Id, team2.Id)

	// TODO:同步匹配信息 (排位加了 匹配没加)
	//res := &proto_room.PushMatchTeam{
	//	IsStart: false,
	//	Time:    int32(10),
	//}
	//for i := 0; i < len(team1.Players); i++ {
	//	l.context.Invoke("Login", "castAgent", team1.Players[i].Uid, &messages.DispatchMessage{
	//		Content: res,
	//	})
	//}
}
