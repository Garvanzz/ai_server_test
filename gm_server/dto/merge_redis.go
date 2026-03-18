package dto

type GmRedisMergeCheckReq struct {
	TargetServerId  int    `json:"targetServerId"`
	SourceServerIds []int  `json:"sourceServerIds"`
	RedisMode       string `json:"redisMode"`
}
