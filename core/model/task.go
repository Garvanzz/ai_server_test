package model

type PlayerTask struct {
	Version     int32                     // 数据版本
	Progress    map[int32]map[int32]int32 // 指标进度: taskType -> extraCondition -> value
	TaskLimit   map[int32]int32           // 指标限制计数
	Buckets     map[int32]map[int32]*Task // 任务桶: bucketType -> taskId -> state
	Points      map[int32]int32           // 活跃点: activityType -> point
	ClaimRecord map[int32]map[int32]bool  // 奖励领取记录: claimType -> rewardId -> claimed
	ResetAt     map[int32]int64           // 重置游标: domain -> unix
}

type Task struct {
	Id             int32
	InitialProcess int32
	TaskType       int32
	Condition      int32
	ExtraCondition int32
	ReceiveAward   bool
}
