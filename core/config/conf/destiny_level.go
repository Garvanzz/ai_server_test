package conf

type DestinyLevel struct {
	Id             int32   `json:"Id"`
	FrontId        int32   `json:"FrontId"`
	Level          int32   `json:"Level"`
	CostItem       []ItemE `json:"CostItem"`
	TeamAttribute  int32   `json:"TeamAttribute"`
	AttributeValue int32   `json:"AttributeValue"`
}

type DestinyStage struct {
	Id                 int32   `json:"Id"`
	CostItem           []ItemE `json:"CostItem"`
	SelfAttribute      int32   `json:"SelfAttribute"`
	SelfAttributeValue int32   `json:"SelfAttributeValue"`
}
