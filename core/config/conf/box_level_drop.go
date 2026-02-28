package conf

type BoxLevelDrop struct {
	Id             int64   `json:"Id"`
	Rare           []int32 `json:"Rare"`
	EquipWeight    []int32 `json:"EquipWeight"`
	MagicWeight    []int32 `json:"MagicWeight"`
	ScoreRefWeight []int32 `json:"ScoreRefWeight"`
	Money          []int32 `json:"Money"`
	Store          int32   `json:"Store"`
	Exp            []int32 `json:"Exp"`
	GetScore       []int32 `json:"GetScore"`
	ScoreBuy       []int32 `json:"ScoreBuy"`
	UpExp          int32   `json:"UpExp"`
	UpTime         int32   `json:"UpTime"`
}
