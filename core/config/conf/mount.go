package conf

type Mount struct {
	Id          int32   `json:"Id"`
	Rate        int32   `json:"Rate"`
	UnLock      int32   `json:"UnLock"`
	UnlockValue []int32 `json:"UnlockValue"`
	LinkSkillId int32   `json:"LinkSkillId"`
	LinkBuffId  int32   `json:"LinkBuffId"`
	AddSumAtk   int32   `json:"AddSumAtk"`
	AddSumDef   int32   `json:"AddSumDef"`
	AddSumHp    int32   `json:"AddSumHp"`
	AddSumForce int32   `json:"AddSumForce"`
	AddDef      int32   `json:"AddDef"`
	AddCrit     int32   `json:"AddCrit"`
}

type MountEnergy struct {
	Id               int32   `json:"Id"`
	Level            int32   `json:"Level"`
	UpLevelCondition []ItemE `json:"UpLevelCondition"`
	SuccessRate      int32   `json:"SuccessRate"`
	Weight           []int32 `json:"Weight"`
}

type MountEnergyAttribute struct {
	Id            int32 `json:"Id"`
	Level         int32 `json:"Level"`
	AttributeType int32 `json:"AttributeType"`
	AddAttribute1 int32 `json:"AddAttribute1"`
	AddAttribute2 int32 `json:"AddAttribute2"`
}

type MountLevel struct {
	Id          int32 `json:"Id"`
	MountId     int32 `json:"MountId"`
	Level       int32 `json:"Level"`
	CostNum     int32 `json:"CostNum"`
	AddSumAtk   int32 `json:"AddSumAtk"`
	AddSumDef   int32 `json:"AddSumDef"`
	AddSumHp    int32 `json:"AddSumHp"`
	AddSumForce int32 `json:"AddSumForce"`
	AddDef      int32 `json:"AddDef"`
	AddCrit     int32 `json:"AddCrit"`
}

type MountStage struct {
	Id              int32   `json:"Id"`
	Stage           int32   `json:"Stage"`
	Star            int32   `json:"Star"`
	UpStarCondition []ItemE `json:"UpStarCondition"`
	AddAtk          int32   `json:"AddAtk"`
	AddDef          int32   `json:"AddDef"`
	AddHp           int32   `json:"AddHp"`
	AddForce        int32   `json:"AddForce"`
}
