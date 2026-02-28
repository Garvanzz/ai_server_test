package conf

type SuccinctSkill struct {
	Id          int32   `json:"Id"`
	Rate        int32   `json:"Rate"`
	Limit       []int32 `json:"Limit"`
	Weight      int32   `json:"Weight"`
	LinkSkillId int32   `json:"LinkSkillId"`
}
