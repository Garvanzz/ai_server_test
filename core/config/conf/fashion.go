package conf

type Fashion struct {
	Id          int32           `json:"Id"`
	AttributeId map[int32]int32 `json:"AttributeId"`
	HandBookExp int32           `json:"HandBookExp"`
}

type Headwear struct {
	Id          int32           `json:"Id"`
	AttributeId map[int32]int32 `json:"AttributeId"`
	HandBookExp int32           `json:"HandBookExp"`
}
