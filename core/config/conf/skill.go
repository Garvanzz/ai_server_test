package conf

type Skill struct {
	Id          int32     `json:"Id"`
	Type        int32     `json:"Type"`
	TargetType  int32     `json:"TargetType"`
	TargetLimit int32     `json:"TargetLimit"`
	TargetNum   int32     `json:"TargetNum"`
	Effect      []int32   `json:"Effect"`
	EffectParam [][]int32 `json:"EffectParam"`
	Cd          int32     `json:"Cd"`
	InitCd      int32     `json:"InitCd"`
}

type SkillUpLevel struct {
	Id         int32    `json:"Id"`
	SkillCost  []ItemE  `json:"SkillCost"`
	SkillType  []int32  `json:"SkillType"`
	SkillValue []string `json:"SkillValue"`
}
