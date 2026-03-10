package dto

// 1 获取关卡列表信息。 字段： 周目     章节id    关卡id。 经验  通关状态 （通关。未通关，正在通关） 挑战boss是否成功
type GmReqGetStageInfo struct {
	Uid      string
	ServerId int
}

type GmRespGetStageInfo struct {
	Cycle   int32
	Chapter int32
	Stage   int32
	Exp     int32
	State   string // 1通关 2未通关 3正在通关
}

// 2 设置关卡信息，传入 周目 章节和关卡id。可设置字段  经验  通关状态   挑战boss是否成功。
type GmReqSetStageInfo struct {
	Uid      string
	ServerId int
	Cycle    int32
	Chapter  int32
	StageId  int32
	Exp      int32
	State    int32 // 1通关 2 挑战boss成功
}

// 3 添加关卡信息， 传入 周目 章节 关卡id 通关状态 挑战boss是否成功。 如果传入关卡 之前的关卡信息没有也需要生成 例如我传入的10008 服务器只有10002。那中间的也要生成
type GmAddStageInfo struct {
	Uid      string
	ServerId int
	Cycle    int32
	Chapter  int32
	StageId  int32
	Exp      int32
	State    int32 // 1通关 2 挑战boss成功
}

// 4 删除关卡信息，传入周目 章节 关卡id
type GmDeleteStageInfo struct {
	Uid          string
	ServerId     int
	IsDelCycle   bool
	IsDelChapter bool
	Cycle        int32
	Chapter      int32
	Stage        int32
}
