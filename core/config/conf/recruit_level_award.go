package conf

type RecruitLevelAward struct {
	Id     int32   `json:"Id"`
	Level  int32   `json:"Level"`
	Exp    int32   `json:"Exp"`
	Weight []int32 `json:"Weight"`
	Award  []ItemE `json:"Award"`
}
