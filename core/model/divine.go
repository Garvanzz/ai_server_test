package model

import "xfx/proto/proto_equip"

// 领悟心得
type Divine struct {
	Divines  map[int32]map[int32]*DivineOption //领悟
	Learning map[int32]*LearningOption         //心得
}

type DivineOption struct {
	Id    int32
	Level int32
	Sid   int32 //心得Id
}

type LearningOption struct {
	Id  int32
	Num int32
}

func ToDivineProto(maps map[int32]map[int32]*DivineOption) map[int32]*proto_equip.DivineIndexItem {
	m := make(map[int32]*proto_equip.DivineIndexItem, 0)
	for k, v := range maps {
		item := new(proto_equip.DivineIndexItem)
		item.Index = k
		n := make(map[int32]*proto_equip.DivineItem, 0)
		for key, l := range v {
			n[key] = &proto_equip.DivineItem{
				ID:    l.Id,
				Level: l.Level,
				SId:   l.Sid,
			}
		}
		item.DivineItems = n
		m[k] = item
	}

	return m
}

func ToDivineSingleProto(opt *DivineOption) map[int32]*proto_equip.DivineItem {
	m := make(map[int32]*proto_equip.DivineItem, 0)
	m[opt.Id] = &proto_equip.DivineItem{
		ID:    opt.Id,
		Level: opt.Level,
		SId:   opt.Sid,
	}

	return m
}

func ToLearningProto(maps map[int32]*LearningOption) map[int32]*proto_equip.LearningOption {
	m := make(map[int32]*proto_equip.LearningOption, 0)
	for _, v := range maps {
		m[v.Id] = &proto_equip.LearningOption{
			Id:  v.Id,
			Num: v.Num,
		}
	}

	return m
}
