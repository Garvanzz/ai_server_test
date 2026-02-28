package conf

type Shop struct {
	Id          int32   `json:"Id"`
	Type        int32   `json:"Type"`
	Sort        int32   `json:"Sort"`
	LimitType   int32   `json:"LimitType"`
	LimitNum    int32   `json:"LimitNum"`
	IsBuy       bool    `json:"IsBuy"`
	RechargeId  int32   `json:"RechargeId"`
	UnLock      int32   `json:"UnLock"`
	UnLockValue int32   `json:"UnLockValue"`
	CostItem    []ItemE `json:"CostItem"`
	GetItem     []ItemE `json:"GetItem"`
	FirstAward  []ItemE `json:"FirstAward"`
}
