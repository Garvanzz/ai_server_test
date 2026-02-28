package model

// PlayerBase 玩家基础信息
type PlayerBase struct {
	Name       string
	CreateTime int64
}

type PlayerProp struct {
	HeadFrames []int32
	Titles     []int32
	Bubbles    []int32
}
