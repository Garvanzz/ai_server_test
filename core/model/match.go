package model

// 匹配信息
//type MatchOpt struct {
//	PlayerId int64
//	Rank     int32
//}

type MatchTeam struct {
	Id          int32
	Type        int32
	IsGroup     bool
	AverageRank int32
}
