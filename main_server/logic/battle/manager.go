package battle

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/proto/proto_game"
	"xfx/proto/proto_player"
)

var battleId int64

// 大闹天宫
type BattleDanaotiangong struct {
	BattleId int64
	Stage    int32
}

// 副本
type BattleMission struct {
	BattleId int64
	Stage    int32
	typ      int32
}

// 主关卡boss
type BattleStageBoss struct {
	BattleId int64
	Stage    int32
	Chapter  int32
	Cycle    int32
}

// 玩家
type BattlePlayer struct {
	BattleId int64
	PlayerId int64
}

// 竞技场
type BattleArena struct {
	BattleId int64
	PlayerId int64
	ActId    int64
	Fuchou   bool
	ActCId   int64
}

// 天梯
type BattleTianti struct {
	BattleId int64
	PlayerId int64
	ActId    int64
	ActCid   int64
}

var Module = func() module.Module {
	return &Manager{
		battleIdMap:         make(map[int64]int),
		battleDanaotiangong: make(map[int64]*BattleDanaotiangong),
		battleMission:       make(map[int64]*BattleMission),
		battleStageBoss:     make(map[int64]*BattleStageBoss),
		battlePlayer:        make(map[int64]*BattlePlayer),
		battleArena:         make(map[int64]*BattleArena),
		battleTianti:        make(map[int64]*BattleTianti),
	}
}

type Manager struct {
	modules.BaseModule
	battleId            int64                          // 当前最大id
	battleIdMap         map[int64]int                  //key:Uid value:type
	battleDanaotiangong map[int64]*BattleDanaotiangong //大闹天宫
	battleMission       map[int64]*BattleMission       //副本
	battleStageBoss     map[int64]*BattleStageBoss     //关卡boss
	battlePlayer        map[int64]*BattlePlayer        //玩家
	battleArena         map[int64]*BattleArena         //竞技场
	battleTianti        map[int64]*BattleTianti        //天梯
}

func (m *Manager) OnInit(app module.App) {
	m.BaseModule.OnInit(app)

	m.Register("BattleDanaotiangong", m.BattleDanaotiangong)
	m.Register("ReqChallengeBattleReport", m.ReqChallengeBattleReport)
	m.Register("BattleMission", m.BattleMission)
	m.Register("BattleStageBoss", m.BattleStageBoss)
	m.Register("BattlePlayer", m.BattlePlayer)
	m.Register("BattleArena", m.BattleArena)
	m.Register("BattleTianti", m.BattleTianti)

	reply, err := db.RedisExec("get", "battleId")
	if err != nil {
		log.Error("load battleId err:%v", err)
		return
	}

	// 加载id
	if reply != nil {
		err = json.Unmarshal(reply.([]byte), &m.battleId)
		log.Debug("load battleId ：%+v", m.battleId)
		if err != nil {
			log.Error("battleId Id err:%v", err)
			return
		}
	}

	replyMap, err := db.RedisExec("get", "battleIdMap")
	if err != nil {
		log.Error("load replyMap err:%v", err)
		return
	}

	// 加载battleIdMap
	if replyMap != nil {
		err = json.Unmarshal(replyMap.([]byte), &m.battleIdMap)
		log.Debug("load battleIdMap ：%+v", m.battleIdMap)
		if err != nil {
			log.Error("battleIdMap Id err:%v", err)
			return
		}
	}

	reply_danao, err := db.RedisExec("get", "battleDanaotiangong")
	if err != nil {
		log.Error("load reply_danao err:%v", err)
		return
	}

	// 加载reply_danao
	if reply_danao != nil {
		err = json.Unmarshal(reply_danao.([]byte), &m.battleDanaotiangong)
		log.Debug("load reply_danao ：%+v", m.battleDanaotiangong)
		if err != nil {
			log.Error("reply_danao Id err:%v", err)
			return
		}
	}

}

func (m *Manager) OnStart(ctx module.Context) {
	m.BaseModule.OnStart(ctx)
	event.AddEventListener(define.EventTypePlayerOnline, m.Self())
	event.AddEventListener(define.EventTypePlayerOffline, m.Self())
}

func (m *Manager) GetType() string { return define.ModuleBattle }

func (m *Manager) OnTick(delta time.Duration) {

}

func (m *Manager) OnMessage(msg interface{}) interface{} {
	switch v := msg.(type) {
	case *event.Event:
		m.OnEvent(v)
	default:
		return nil
	}
	return nil
}

// OnEvent 事件回调
func (m *Manager) OnEvent(event *event.Event) {
	if event == nil {
		return
	}

	if event.M == nil {
		return
	}

	// 玩家基础信息
	ctx, ok := event.M["player"].(*proto_player.Context)
	if !ok {
		log.Error("activity event find no player data")
		return
	}

	switch event.Type {
	case define.EventTypePlayerOffline:
		m.BattlePlayerOffline(ctx)
	case define.EventTypePlayerOnline:
	}
}

func (m *Manager) OnDestroy() {
	m.OnSave()
}

func (m *Manager) OnSave() {
	data, err := json.Marshal(m.battleId)
	if err != nil {
		log.Error("battleId data error:", err)
		return
	}

	_, err = db.RedisExec("set", "battleId", data)
	if err != nil {
		log.Error("set battleId error: %v", err)
		return
	}
}

// 掉线
func (m *Manager) BattlePlayerOffline(ctx *proto_player.Context) {
	if _, ok := m.battleIdMap[ctx.Id]; !ok {
		return
	}
	log.Debug("掉线 处理战斗:$v", ctx.Id)
	delete(m.battleDanaotiangong, ctx.Id)
	delete(m.battleIdMap, ctx.Id)
	delete(m.battleMission, ctx.Id)
	delete(m.battleStageBoss, ctx.Id)
	delete(m.battlePlayer, ctx.Id)
	delete(m.battleArena, ctx.Id)
	delete(m.battleTianti, ctx.Id)
}

// 战报
func (m *Manager) ReqChallengeBattleReport(ctx *proto_player.Context, req *proto_game.C2SChallengeBattleReport) (model.ChallengeBattleReportBack, error) {
	if _, ok := m.battleIdMap[ctx.Id]; !ok {
		str := fmt.Sprintf("%v : is no battle, id : %v", ctx.Id, req.BattleId)
		return model.ChallengeBattleReportBack{
			Scene: define.BattleScene_None,
		}, errors.New(str)
	}

	scene := m.battleIdMap[ctx.Id]
	var backMsg interface{}
	//删除id
	delete(m.battleIdMap, ctx.Id)

	switch scene {
	case define.BattleScene_Danaotiangong:
		dd := m.battleDanaotiangong[ctx.Id]
		delete(m.battleDanaotiangong, ctx.Id)
		backMsg = model.BattleReportBack_Danaotiangong{
			Stage: dd.Stage,
			Data:  req,
		}

		return model.ChallengeBattleReportBack{
			Scene: define.BattleScene_Danaotiangong,
			Data:  backMsg,
		}, nil
	case define.BattleScene_Mission:
		dd := m.battleMission[ctx.Id]
		delete(m.battleMission, ctx.Id)
		backMsg = model.BattleReportBack_Mission{
			Stage: dd.Stage,
			Data:  req,
			Typ:   dd.typ,
		}

		return model.ChallengeBattleReportBack{
			Scene: define.BattleScene_Mission,
			Data:  backMsg,
		}, nil
	case define.BattleScene_StageBoss:
		dd := m.battleStageBoss[ctx.Id]
		delete(m.battleStageBoss, ctx.Id)
		backMsg = model.BattleReportBack_StageBoss{
			Stage:   dd.Stage,
			Data:    req,
			Chapter: dd.Chapter,
			Cycle:   dd.Cycle,
		}

		return model.ChallengeBattleReportBack{
			Scene: define.BattleScene_StageBoss,
			Data:  backMsg,
		}, nil
	case define.BattleScene_Player:
		pd := m.battlePlayer[ctx.Id]
		delete(m.battlePlayer, ctx.Id)
		backMsg = model.BattleReportBack_Player{
			Data:     req,
			PlayerId: pd.PlayerId,
		}

		return model.ChallengeBattleReportBack{
			Scene: define.BattleScene_Player,
			Data:  backMsg,
		}, nil
	case define.BattleScene_Arena:
		pd := m.battleArena[ctx.Id]
		delete(m.battleArena, ctx.Id)
		backMsg = model.BattleReportBack_Arena{
			Data:     req,
			PlayerId: pd.PlayerId,
			ActId:    pd.ActId,
			Fuchou:   pd.Fuchou,
			ActCId:   pd.ActCId,
		}
		return model.ChallengeBattleReportBack{
			Scene: define.BattleScene_Arena,
			Data:  backMsg,
		}, nil
	case define.BattleScene_Tianti:
		pd := m.battleTianti[ctx.Id]
		delete(m.battleTianti, ctx.Id)
		backMsg = model.BattleReportBack_Tianti{
			Data:     req,
			PlayerId: pd.PlayerId,
			ActId:    pd.ActId,
			ActCId:   pd.ActCid,
		}

		return model.ChallengeBattleReportBack{
			Scene: define.BattleScene_Tianti,
			Data:  backMsg,
		}, nil
	}

	str := fmt.Sprintf("%v : is no battle id : %v", ctx.Id, req.BattleId)
	return model.ChallengeBattleReportBack{
		Scene: define.BattleScene_None,
		Data:  backMsg,
	}, errors.New(str)
}

// 大闹天宫
func (m *Manager) BattleDanaotiangong(ctx *proto_player.Context, stageId int32) (int64, error) {
	if _, ok := m.battleIdMap[ctx.Id]; ok {
		str := fmt.Sprintf("%v : is battle danaotiangong", ctx.Id)
		return 0, errors.New(str)
	}

	//获取id
	m.battleId++
	m.battleIdMap[ctx.Id] = define.BattleScene_Danaotiangong
	bat := new(BattleDanaotiangong)
	bat.BattleId = m.battleId
	bat.Stage = stageId
	m.battleDanaotiangong[ctx.Id] = bat
	return bat.BattleId, nil
}

// 副本
func (m *Manager) BattleMission(ctx *proto_player.Context, typ, stageId int32) (int64, error) {
	if _, ok := m.battleIdMap[ctx.Id]; ok {
		str := fmt.Sprintf("%v : is battle mission", ctx.Id)
		return 0, errors.New(str)
	}

	//获取id
	m.battleId++
	m.battleIdMap[ctx.Id] = define.BattleScene_Mission
	bat := new(BattleMission)
	bat.BattleId = m.battleId
	bat.Stage = stageId
	bat.typ = typ
	m.battleMission[ctx.Id] = bat
	return bat.BattleId, nil
}

// 关卡boss
func (m *Manager) BattleStageBoss(ctx *proto_player.Context, cycle, stageId, chapter int32) (int64, error) {
	if _, ok := m.battleIdMap[ctx.Id]; ok {
		str := fmt.Sprintf("%v : is battle stageboss", ctx.Id)
		return 0, errors.New(str)
	}

	//获取id
	m.battleId++
	m.battleIdMap[ctx.Id] = define.BattleScene_StageBoss
	bat := new(BattleStageBoss)
	bat.BattleId = m.battleId
	bat.Stage = stageId
	bat.Chapter = chapter
	bat.Cycle = cycle
	m.battleStageBoss[ctx.Id] = bat
	return bat.BattleId, nil
}

// 玩家
func (m *Manager) BattlePlayer(ctx *proto_player.Context, playerId int64) (int64, error) {
	if _, ok := m.battleIdMap[ctx.Id]; ok {
		str := fmt.Sprintf("%v : is battle player", ctx.Id)
		return 0, errors.New(str)
	}

	//获取id
	m.battleId++
	m.battleIdMap[ctx.Id] = define.BattleScene_Player
	bat := new(BattlePlayer)
	bat.BattleId = m.battleId
	bat.PlayerId = playerId
	m.battlePlayer[ctx.Id] = bat
	return bat.BattleId, nil
}

// 竞技场
func (m *Manager) BattleArena(ctx *proto_player.Context, playerId int64, ActId int64, ActCid int64, Fuchou bool) (int64, error) {
	if _, ok := m.battleIdMap[ctx.Id]; ok {
		str := fmt.Sprintf("%v : is battle arena", ctx.Id)
		return 0, errors.New(str)
	}

	//获取id
	m.battleId++
	m.battleIdMap[ctx.Id] = define.BattleScene_Arena
	bat := new(BattleArena)
	bat.BattleId = m.battleId
	bat.PlayerId = playerId
	bat.ActId = ActId
	bat.ActCId = ActCid
	bat.Fuchou = Fuchou
	m.battleArena[ctx.Id] = bat
	return bat.BattleId, nil
}

// 天梯
func (m *Manager) BattleTianti(ctx *proto_player.Context, playerId int64, ActId int64, ActCid int64) (int64, error) {
	if _, ok := m.battleIdMap[ctx.Id]; ok {
		str := fmt.Sprintf("%v : is battle tianti", ctx.Id)
		return 0, errors.New(str)
	}

	//获取id
	m.battleId++
	m.battleIdMap[ctx.Id] = define.BattleScene_Tianti
	bat := new(BattleTianti)
	bat.BattleId = m.battleId
	bat.PlayerId = playerId
	bat.ActId = ActId
	bat.ActCid = ActCid
	m.battleTianti[ctx.Id] = bat
	return bat.BattleId, nil
}
