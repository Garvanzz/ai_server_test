package conf

type BroadCast struct {
	Id        int32   `json:"Id"`
	Condition []int32 `json:"condition"`
	Type      int32   `json:"Type"`
	Param     []int32 `json:"param"`
	Priority  int32   `json:"priority"`
	Time      int32   `json:"time"`
	Scene     int32   `json:"Scene"`
	Channel   int32   `json:"channel"`
}

type ChatChuanWen struct {
	Id        int32   `json:"Id"`
	Condition []int32 `json:"condition"`
	Type      int32   `json:"Type"`
	Param     []int32 `json:"param"`
	Channel   int32   `json:"channel"`
}
