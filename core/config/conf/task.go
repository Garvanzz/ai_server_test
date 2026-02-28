package conf

type Task struct {
	Id            int32   `json:"Id"`
	Type          int32   `json:"Type"`
	TaskType      int32   `json:"TaskType"`
	Condition1    int32   `json:"Condition1"`
	Condition2    int32   `json:"Condition2"`
	ActivityValue int32   `json:"ActivityValue"`
	FrontTask     int32   `json:"frontTask"`
	BackTask      int32   `json:"backTask"`
	Reward        []ItemE `json:"Reward"`
	Param         []int32 `json:"Param"`
	Reset         bool    `json:"Reset"`
}

type TaskActivity struct {
	Id     int32   `json:"Id"`
	Type   int32   `json:"Type"`
	Value  int32   `json:"Value"`
	Reward []ItemE `json:"Reward"`
}
