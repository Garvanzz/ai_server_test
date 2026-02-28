package define

// 匹配类型
const (
	MatchModTopPk = 1 //限时巅峰决斗
	MatchModArena = 2 //竞技场
	MatchModRank  = 3 //天梯
)

const (
	RoomStateNormal = iota + 1
	RoomStateMatch
	RoomStateGame
)
