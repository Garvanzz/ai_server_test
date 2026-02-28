package model

import (
	"xfx/proto/proto_destiny"
)

// 天命
type Destiny struct {
	Ids     []int32
	Level   int32
	SelfIds []int32
}

// 神机宝匣
type ShenjiDraw struct {
	Pools map[int32]*ShenjiDrawPool
}

type ShenjiDrawPool struct {
	PoolId         int32
	Num            int32
	PoolStartTime  int64
	BdNum          int32
	ShenjiRecords  []*ShenjiRecord
	LastRecordTime int64 //上一次清空记录时间戳
}

type ShenjiRecord struct {
	Id  int32
	Num int32
}

func ToDestinyShenjiProto(opt map[int32]*ShenjiDrawPool) map[int32]*proto_destiny.ShenjiPoolOption {
	m := make(map[int32]*proto_destiny.ShenjiPoolOption, 0)
	for k, v := range opt {
		rec := make([]*proto_destiny.ShenjiAwardRecord, 0)
		for i := 0; i < len(v.ShenjiRecords); i++ {
			rec = append(rec, &proto_destiny.ShenjiAwardRecord{
				Id:  v.ShenjiRecords[i].Id,
				Num: v.ShenjiRecords[i].Num,
			})
		}
		m[k] = &proto_destiny.ShenjiPoolOption{
			PoolId:        v.PoolId,
			PoolStartTime: v.PoolStartTime,
			Num:           v.Num,
		}
	}

	return m
}
