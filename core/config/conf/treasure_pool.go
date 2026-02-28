package conf

type TreasurePool struct {
	Id     int32 `json:"Id"`
	Title  int32 `json:"Title"`
	Type   int32 `json:"Type"`
	Value  int32 `json:"Value"`
	Rate   int32 `json:"Rate"`
	Num    int32 `json:"Num"`
	Weight int32 `json:"Weight"`
}
