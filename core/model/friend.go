package model

// 好友申请
type FriendApply struct {
	Id       int32
	PlayerId int64 // 申请发起人
	TargetId int64
	Msg      string
}

// 好友赠礼, //Redis每天0点删除
type FriendGift struct {
	IsCanGet bool //能够领取
	IsAlGet  bool //已经领取
	IsSend   bool
}

// 黑名单
type FriendBlock struct {
	Id       int32
	PlayerId int64
	TargetId int64
	Msg      string
}
