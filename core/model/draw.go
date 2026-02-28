package model

import (
	"xfx/proto/proto_draw"
)

// 抽角色
type DrawHero struct {
	BdCount     int32
	ToDayIsFree bool
	Level       int32
	LastTime    int64
	Exp         int32
	LevelAwards []int32
	BigCard     []int32
	Pools       map[int32]*DrawPool
}

type DrawPool struct {
	PoolId      int32
	StageAwards []int32
	StarTime    int64
	DrawNum     int32
}

type RecruitResp struct {
	BdNum int32
	Ids   []int32
}

func ToDrawCardProto(maps *DrawHero) *proto_draw.DrawOption {
	pools := make(map[int32]*proto_draw.DrawPoolOption)
	for _, v := range maps.Pools {
		pools[v.PoolId] = &proto_draw.DrawPoolOption{
			PoolId:        v.PoolId,
			PoolStartTime: v.StarTime,
			StageAwards:   v.StageAwards,
			DrawNum:       v.DrawNum,
		}
	}
	return &proto_draw.DrawOption{
		BdCount:     maps.BdCount,
		ToDayIsFree: maps.ToDayIsFree,
		Level:       maps.Level,
		Exp:         maps.Exp,
		Awards:      maps.LevelAwards,
		Pools:       pools,
	}
}
