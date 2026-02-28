package conf

type CollectionUpStar struct {
	Id              int32   `json:"Id"`
	CollectionId    int32   `json:"CollectionId"`
	Star            int32   `json:"Star"`
	NeedFragNum     int32   `json:"NeedFragNum"`
	UpStarCondition []ItemE `json:"UpStarCondition"`
}
