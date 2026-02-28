package conf

type Uproar struct {
	Id           int64   `json:"Id"`
	BossId       int32   `json:"bossId"`
	ChallengeNum []int32 `json:"challengeNum"`
}

type UproarFrequency struct {
	Id              int64   `json:"Id"`
	Frequency       int32   `json:"frequency"`
	OneLevelAward   []ItemE `json:"OneLevelAward"`
	TwoLevelAward   []ItemE `json:"TwoLevelAward"`
	ThreeLevelAward []ItemE `json:"ThreeLevelAward"`
	FourLevelAward  []ItemE `json:"FourLevelAward"`
	FiveLevelAward  []ItemE `json:"FiveLevelAward"`
	SexLevelAward   []ItemE `json:"SexLevelAward"`
	SevenLevelAward []ItemE `json:"SevenLevelAward"`
	EightLevelAward []ItemE `json:"EightLevelAward"`
	NineLevelAward  []ItemE `json:"NineLevelAward"`
}
