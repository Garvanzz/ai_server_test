package room

import (
	"xfx/core/define"
	"xfx/proto/proto_public"
)

//
//// 结算
//func (mgr *Manager) OnSettleGame(settle *proto_public.S2SGameSettleInfo) {
//	var sle = &proto_room.S2CGameSettle{
//		ToGameOver: &proto_room.ToGameOver{},
//	}
//	sle.GetToGameOver().FaildGameOver = &proto_room.FaildGameOver{}
//	sle.GetToGameOver().SuccessGameOver = &proto_room.SuccessGameOver{}
//	sle.GetToGameOver().Time = settle.GetGameTime()
//
//	var getGroupReport = func(reports map[uint64]*proto_public.GameReportPlayer, settles map[uint64]*proto_public.GameSettlePlayer, group int32) []*proto_room.GameOverPlayer {
//		var overs = make([]*proto_room.GameOverPlayer, 0)
//		for _, v := range reports {
//			data, _ := mgr.OnGetRoomByPlayerId(v.GetPlayerId())
//			if data != nil { // TODO: data.GetGroup() == group
//				overs = append(overs, &proto_room.GameOverPlayer{
//					CommonPlayerInfo: &proto_public.CommonPlayerInfo{
//						//PlayerId: data.GetPlayerId(),
//						Name: data.GetName(),
//						//FaceId:   data.GetFaceId(),
//						//Level:    data.GetLevel(),
//					},
//					Damage:          v.GetDamage(),
//					WithStandDamage: v.GetWithStandDamage(),
//					Heal:            v.GetHeal(),
//					Cards:           settles[uint64(v.GetPlayerId())].GetCards(),
//					Level:           int32(mgr.OnCalculateRare(v.GetPlayerId(), reports, group == settle.GetWinGroup() && v.GetPlayerId() == mgr.OnMvpOrSvp(int64(settle.RoomId), group, reports))),
//					KillNum:         v.GetKillNum(),
//					RankScore:       20,
//					AwardOptions:    make([]*proto_public.AwardOption, 0),
//				})
//			}
//		}
//		return overs
//	}
//
//	for i := 0; i < len(mgr.rooms[settle.GetRoomId()].Players); i++ {
//		player := mgr.rooms[settle.GetRoomId()].Players[i]
//		//判断自己是否胜利
//		// TODO:sle.GetToGameOver().IsWin = player.GetGroup() == settle.GetWinGroup()
//		//败方
//		var arr = getGroupReport(settle.GetGameReportPlayer(), settle.GetGameSettlePlayer(), settle.GetFaildGroup())
//		sle.GetToGameOver().GetFaildGameOver().GameOverPlayer = arr
//		sle.GetToGameOver().GetFaildGameOver().Svp = mgr.OnMvpOrSvp(int64(settle.GetRoomId()), settle.GetFaildGroup(), settle.GetGameReportPlayer())
//
//		var num = int32(0)
//		var grs = mgr.OnGetSettleGroup(settle.GetFaildGroup(), int64(settle.GetRoomId()), settle.GetGameReportPlayer())
//		for _, v := range grs {
//			num += v.GetKillNum()
//		}
//		sle.GetToGameOver().GetFaildGameOver().KillNum = num
//
//		//胜方
//		var arrs = getGroupReport(settle.GetGameReportPlayer(), settle.GetGameSettlePlayer(), settle.GetWinGroup())
//		sle.GetToGameOver().GetSuccessGameOver().GameOverPlayer = arrs
//		sle.GetToGameOver().GetSuccessGameOver().Mvp = mgr.OnMvpOrSvp(int64(settle.GetRoomId()), settle.GetWinGroup(), settle.GetGameReportPlayer())
//
//		num = int32(0)
//		grs = mgr.OnGetSettleGroup(settle.GetWinGroup(), int64(settle.GetRoomId()), settle.GetGameReportPlayer())
//		for _, v := range grs {
//			num += v.GetKillNum()
//		}
//		sle.GetToGameOver().GetSuccessGameOver().KillNum = num
//
//		//通知客户端
//		invoke.Dispatch(mgr, player.PlayerId, sle)
//	}
//}

// TODO:获取数据根据组
func (mgr *Manager) OnGetSettleGroup(Group int32, roomId int64, reports map[uint64]*proto_public.GameReportPlayer) map[int64]*proto_public.GameReportPlayer {
	//var arr map[int64]*proto_public.GameReportPlayer
	//for _, v := range reports {
	//data, _ := mgr.OnGetRoomByPlayerId(v.GetPlayerId())
	//if data.Group == Group {
	//	arr[v.GetPlayerId()] = v
	//}
	//}
	//return arr

	return nil
}

// 是否是mvp/svp
func (mgr *Manager) OnMvpOrSvp(RoomId int64, Group int32, reports map[uint64]*proto_public.GameReportPlayer) int64 {
	var arr map[uint64]*proto_public.GameReportPlayer

	// TODO:
	//for _, v := range reports {
	//	data, _ := mgr.OnGetRoomByPlayerId(v.GetPlayerId())
	//	data.GetIsGame()
	//	if data.GetGroup() == Group {
	//		arr[uint64(v.GetPlayerId())] = v
	//	}
	//}

	var arrs map[uint64]float64
	for id, _ := range arr {
		var score = mgr.OnCalculateScore(int64(id), arr)
		arrs[id] = score
	}

	var bigsco float64 = 0
	var bigid uint64 = 0
	for id, sco := range arrs {
		if sco > bigsco {
			bigid = id
		}
	}

	return int64(bigid)
}

// 评级
func (mgr *Manager) OnCalculateRare(playerId int64, reports map[uint64]*proto_public.GameReportPlayer, isFirst bool) int {
	var totalScore = mgr.OnCalculateScore(playerId, reports)

	if totalScore > 0.9 && isFirst {
		return define.SETTLE_RARE_SSSS
	} else if totalScore > 0.7 {
		return define.SETTLE_RARE_SSS
	} else if totalScore > 0.5 {
		return define.SETTLE_RARE_SS
	} else if totalScore > 0.3 {
		return define.SETTLE_RARE_S
	}
	return define.SETTLE_RARE_Null
}

func (mgr *Manager) OnCalculateScore(playerId int64, reports map[uint64]*proto_public.GameReportPlayer) float64 {
	//计算公式
	var alldamage = int64(0)
	var allWithStandDamage = int64(0)
	var allheal = int64(0)
	var allkill = int32(0)

	var report *proto_public.GameReportPlayer
	for _, v := range reports {
		if v.GetPlayerId() == playerId {
			report = v
		}
		alldamage += v.GetDamage()
		allWithStandDamage += v.GetWithStandDamage()
		allheal += v.GetHeal()
		allkill += v.GetKillNum()
	}

	damageScore := float64(report.GetDamage()) / float64(alldamage)
	damageTakenScore := float64(report.GetWithStandDamage()) / float64(allWithStandDamage)
	healingScore := float64(report.GetHeal()) / float64(allheal)
	killsScore := float64(report.GetKillNum()) / float64(allkill)

	// 权重设定
	w1, w2, w3, w4 := 0.4, 0.2, 0.3, 0.1

	// 计算总分
	totalScore := w1*damageScore + w2*damageTakenScore + w3*healingScore + w4*killsScore
	return totalScore
}
