package conf

type Learning struct {
	Id          int32 `json:"Id"`
	Rare        int32 `json:"Rare"`
	Type        int32 `json:"Type"`
	SkillType   int32 `json:"SkillType"`
	LinkSkillId int32 `json:"LinkSkillId"`
}

type LearningCompose struct {
	Id        int32   `json:"Id"`
	Condition []int32 `json:"Condition"`
	Value     int32   `json:"Value"`
}
