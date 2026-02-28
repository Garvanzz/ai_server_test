package model

import "xfx/proto/proto_public"

type ChatInfo struct {
	DbId           int64
	Content        string
	Time           int64
	Value          []int32
	Cid            int32
	Type           int32 //1:正常发言 2 附件
	AttachmentData *AttachmentData
}

type PrivateChatList struct {
	List map[int64]*proto_public.CommonPlayerInfo
}
