package conf

type IdleBox struct {
	Id                int32           `json:"Id"`
	StageRange        []int32         `json:"StageRange"`        //推图区间
	StageReward       []ItemE         `json:"StageReward"`       // 推图奖励
	AddStageRewardNum map[int32]int32 `json:"AddStageRewardNum"` //满小时后奖励数量提升
	TowerReward       []ItemE         `json:"TowerReward"`       // 爬塔奖励
	TowerRange        []int32         `json:"TowerRange"`        //爬塔区间
	AddTowerRewardNum map[int32]int32 `json:"AddTowerRewardNum"` //满小时后奖励数量提升
}
