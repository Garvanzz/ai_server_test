package dto

type GmCreateMergePlanReq struct {
	Name            string `json:"name"`
	TargetServerId  int    `json:"targetServerId"`
	SourceServerIds []int  `json:"sourceServerIds"`
	Remark          string `json:"remark"`
}

type GmExecuteMergePlanReq struct {
	PlanId int64 `json:"planId"`
}

type GmPrecheckMergeReq struct {
	TargetServerId  int   `json:"targetServerId"`
	SourceServerIds []int `json:"sourceServerIds"`
}
