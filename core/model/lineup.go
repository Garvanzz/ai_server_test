package model

import (
	"xfx/proto/proto_lineup"
)

type LineUp struct {
	LineUps map[int32]*LineUpOption
}

type LineUpOption struct {
	Type   int32
	HeroId []int32
}

func ToLineUpProto(maps map[int32]*LineUpOption) map[int32]*proto_lineup.CardLineUpMap {
	m := make(map[int32]*proto_lineup.CardLineUpMap, 0)
	for k, v := range maps {
		m[k] = &proto_lineup.CardLineUpMap{
			Type:   k,
			HeroId: v.HeroId,
		}
	}

	return m
}
