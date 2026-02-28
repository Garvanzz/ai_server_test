package conf

type BraceAura struct {
	Id       int32   `json:"Id"`
	Level    int32   `json:"Level"`
	CostNum  []int32 `json:"CostNum"`
	AddAtk   int32   `json:"AddAtk"`
	AddDef   int32   `json:"AddDef"`
	AddHp    int32   `json:"AddHp"`
	AddForce int32   `json:"AddForce"`
}

type BraceAuraStage struct {
	Id       int32 `json:"Id"`
	Stage    int32 `json:"Stage"`
	AddAtk   int32 `json:"AddAtk"`
	AddDef   int32 `json:"AddDef"`
	AddHp    int32 `json:"AddHp"`
	AddForce int32 `json:"AddForce"`
}

type BraceTalent struct {
	Id            int32   `json:"Id"`
	Job           int32   `json:"Job"`
	Group         int32   `json:"Group"`
	FrontNode     []int32 `json:"FrontNode"`
	BackNode      []int32 `json:"BackNode"`
	TalentLevelId int32   `json:"TalentLevelId"`
	UnLockLevel   int32   `json:"UnLockLevel"`
}

type BraceTalentLevel struct {
	Id            int32   `json:"Id"`
	TalentLevelId int32   `json:"TalentLevelId"`
	AttId         int32   `json:"AttId"`
	Level         int32   `json:"Level"`
	CostItem      []ItemE `json:"CostItem"`
	AttValue      int32   `json:"AttValue"`
}

type Braces struct {
	Id              int32 `json:"Id"`
	Rate            int32 `json:"Rate"`
	UnLockAuraLevel int32 `json:"UnLockAuraLevel"`
	LinkSkillId     int32 `json:"LinkSkillId"`
	LinkBuffId      int32 `json:"LinkBuffId"`
	HandBookExp     int32 `json:"HandBookExp"`
}

type BracesLevel struct {
	Id          int32 `json:"Id"`
	BracesId    int32 `json:"BracesId"`
	Level       int32 `json:"Level"`
	CostNum     int32 `json:"CostNum"`
	AddSumAtk   int32 `json:"AddSumAtk"`
	AddSumDef   int32 `json:"AddSumDef"`
	AddSumHp    int32 `json:"AddSumHp"`
	AddSumForce int32 `json:"AddSumForce"`
	AddDef      int32 `json:"AddDef"`
	AddCrit     int32 `json:"AddCrit"`
}
