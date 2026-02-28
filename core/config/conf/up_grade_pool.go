package conf

type ShenjiPool struct {
	Id     int32 `json:"Id"`
	Type   int32 `json:"Type"`
	Value  int32 `json:"Value"`
	Rate   int32 `json:"Rate"`
	Num    int32 `json:"Num"`
	Weight int32 `json:"Weight"`
}
