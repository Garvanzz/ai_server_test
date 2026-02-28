package conf

type Collection struct {
	Id            int32 `json:"Id"`
	Rate          int32 `json:"Rate"`
	ActiveFragNum int32 `json:"ActiveFragNum"`
	LinkSkillId   int32 `json:"LinkSkillId"`
	LinkBuffId    int32 `json:"LinkBuffId"`
	Fragment      int32 `json:"Fragment"`
}
