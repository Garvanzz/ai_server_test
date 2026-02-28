package model

type Award struct {
	Type  int32
	Value int32
	Num   int32
}

type OpenBox struct {
	Level        int32
	Exp          int32
	LastUpTime   int64
	IsUpLevelBox bool
	Score        int32
	NextScoreBox int32
}

func MakeAward(awardType int32, item int32, num int32) *Award {
	award := &Award{
		Type:  awardType,
		Value: item,
		Num:   num,
	}
	return award
}

type Bag struct {
	Items map[int32]int32
}
