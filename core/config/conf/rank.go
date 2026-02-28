package conf

type RankAward struct {
	Id    int32   `json:"Id"`
	Rank  []int32 `json:"Rank"`
	Type  int32   `json:"Type"`
	Award []ItemE `json:"Award"`
}
