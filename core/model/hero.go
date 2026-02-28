package model

import (
	"xfx/proto/proto_handbook"
	"xfx/proto/proto_hero"
)

// 图鉴
type Handbook struct {
	Handbooks      map[int32]*HandbookHero
	HandbookOption *HandbookOption
}

type HandbookOption struct {
	Level int32
	Exp   int32
	GetId []int32
}

type HandbookHero struct {
	Id       int32
	IsGetExp bool
	GetExp   int32
}

type Hero struct {
	Hero map[int32]*HeroOption
	Skin map[int32]*SkinOption
}

type HeroOption struct {
	Id          int32
	Star        int32
	Level       int32
	Stage       int32
	Exp         int32
	Skin        string
	Cultivation map[int32]int32 //修为
}

type SkinOption struct {
	Id    int32
	SrcId int32
}

func ToBagHeroProto(maps map[int32]*HeroOption) map[int32]*proto_hero.HeroOption {
	m := make(map[int32]*proto_hero.HeroOption, 0)
	for k, v := range maps {
		m[k] = &proto_hero.HeroOption{
			Id:          v.Id,
			Star:        v.Star,
			Level:       v.Level,
			Exp:         v.Exp,
			Stage:       v.Stage,
			Cultivation: v.Cultivation,
		}
	}

	return m
}

func ToBagHeroProtoByHero(v *HeroOption) *proto_hero.HeroOption {
	return &proto_hero.HeroOption{
		Id:          v.Id,
		Star:        v.Star,
		Level:       v.Level,
		Exp:         v.Exp,
		Stage:       v.Stage,
		Cultivation: v.Cultivation,
	}
}

func ToBagSkinProto(maps map[int32]*SkinOption) map[int32]*proto_hero.SkinOption {
	m := make(map[int32]*proto_hero.SkinOption, 0)
	for k, v := range maps {
		m[k] = &proto_hero.SkinOption{
			Id:    v.Id,
			SrcId: v.SrcId,
		}
	}

	return m
}

func ToHandBookHeroProtoByHandBook(v map[int32]*HandbookHero) map[int32]*proto_handbook.HandbookHero {
	equip := make(map[int32]*proto_handbook.HandbookHero)
	for l, b := range v {
		equip[l] = &proto_handbook.HandbookHero{
			Id:       b.Id,
			IsGetExp: b.IsGetExp,
			GetExp:   b.GetExp,
		}
	}

	return equip
}

func ToHandBookOptProtoByHandBook(v *HandbookOption) *proto_handbook.HandbookOption {
	return &proto_handbook.HandbookOption{
		Level: v.Level,
		Exp:   v.Exp,
		GetId: v.GetId,
	}
}
