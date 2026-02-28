package conf

type Drop struct {
	Id          int64   `json:"Id"`
	Type        int64   `json:"type"`
	Num         int64   `json:"num"`
	Probability []int64 `json:"probability"`
	Weight      []int64 `json:"weight"`
	Rewards     []ItemE `json:"rewards"`
}
