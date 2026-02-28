package conf

type Recharge struct {
	Id       int32 `json:"Id"`
	ItemId   int32 `json:"ItemId"`
	Platform int32 `json:"Platform"`
	Price    int32 `json:"Price"`
	Discount int32 `json:"Discount"`
	SrcPrice int32 `json:"SrcPrice"`
}
