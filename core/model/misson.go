package model

import (
	"xfx/proto/proto_mission"
)

type Mission struct {
	Box        *MissionItem //宝箱
	Lingyu     *MissionItem //灵玉
	ClimbTower *MissionItem //爬塔
}

type MissionItem struct {
	Stage        int32
	ChallengeNum int32
	Time         int64
}

type BattleReportBack_Mission struct {
	Stage int32
	Typ   int32
	Data  interface{}
}

func ToMissionProtoByMisson(v *Mission) map[int32]*proto_mission.MissionStageOption {
	opts := make(map[int32]*proto_mission.MissionStageOption)
	opts[int32(proto_mission.MissionType_Box)] = &proto_mission.MissionStageOption{
		Type:           proto_mission.MissionType_Box,
		Stage:          v.Box.Stage,
		ChallengeCount: v.Box.ChallengeNum,
	}
	opts[int32(proto_mission.MissionType_Lingyu)] = &proto_mission.MissionStageOption{
		Type:           proto_mission.MissionType_Lingyu,
		Stage:          v.Lingyu.Stage,
		ChallengeCount: v.Lingyu.ChallengeNum,
	}
	opts[int32(proto_mission.MissionType_ClimbTower)] = &proto_mission.MissionStageOption{
		Type:           proto_mission.MissionType_ClimbTower,
		Stage:          v.ClimbTower.Stage,
		ChallengeCount: v.ClimbTower.ChallengeNum,
	}
	return opts
}
