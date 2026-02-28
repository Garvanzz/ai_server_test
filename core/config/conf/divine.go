package conf

type Divine struct {
	Id             int32   `json:"Id"`
	Slot           int32   `json:"Slot"`
	Type           int32   `json:"Type"`
	FrontNode      []int32 `json:"FrontNode"`
	BackNode       []int32 `json:"backNode"`
	Level          int32   `json:"Level"`
	LevelCost      []int32 `json:"LevelCost"`
	UnLockCost     int32   `json:"UnLockCost"`
	AttributeValue []int32 `json:"AttributeValue"`
}
