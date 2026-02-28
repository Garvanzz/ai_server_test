package global

import "xfx/proto/proto_public"

func ToCommonPlayerByRobot(Id int64) *proto_public.CommonPlayerInfo {
	ret := new(proto_public.CommonPlayerInfo)
	ret.PlayerId = Id
	ret.IsRobot = true
	return ret
}
