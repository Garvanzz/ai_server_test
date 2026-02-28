package conf

type Stage struct {
	Id            int32   `json:"Id"`
	Front         int32   `json:"front"`
	Next          int32   `json:"next"`
	StageExp      int32   `json:"StageExp"`
	Chapter       int32   `json:"Chapter"`
	Boss          int32   `json:"boss"`
	StageAward    []ItemE `json:"StageAward"`
	BossAward     []ItemE `json:"BossAward"`
	BossDropAward int32   `json:"BossDropAward"`
}
