package conf

type Weaponry struct {
	Id          int32 `json:"Id"`
	Rate        int32 `json:"Rate"`
	LinkSkillId int32 `json:"LinkSkillId"`
	LinkBuffId  int32 `json:"LinkBuffId"`
	HandBookExp int32 `json:"HandBookExp"`
}

type WeaponryLevel struct {
	Id          int32 `json:"Id"`
	WeaponryId  int32 `json:"WeaponryId"`
	Level       int32 `json:"Level"`
	CostNum     int32 `json:"CostNum"`
	AddSumAtk   int32 `json:"AddSumAtk"`
	AddSumDef   int32 `json:"AddSumDef"`
	AddSumHp    int32 `json:"AddSumHp"`
	AddSumForce int32 `json:"AddSumForce"`
	AddDef      int32 `json:"AddDef"`
	AddCrit     int32 `json:"AddCrit"`
}

type WeaponryStar struct {
	Id              int32   `json:"Id"`
	Stage           int32   `json:"Stage"`
	Star            int32   `json:"Star"`
	UpStarCondition []ItemE `json:"UpStarCondition"`
	AddAtk          int32   `json:"AddAtk"`
	AddDef          int32   `json:"AddDef"`
	AddHp           int32   `json:"AddHp"`
	AddForce        int32   `json:"AddForce"`
	AddCrit         int32   `json:"AddCrit"`
}
