package internal

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_chat"
	"xfx/proto/proto_public"
)

// SyncChatSend 同步聊天
func SyncChatSend(ctx global.IPlayer, pl *model.Player, typ, id int32, content string, value []int32, cid int32, msgType int32, attachment *proto_public.AttachmentOption) (bool, proto_public.CommonErrorCode) {
	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqChatInfo error, no this server:%v", err)
		return false, proto_public.CommonErrorCode_ERR_MYSQLERROR
	}

	account := new(model.Account)
	_, err = db.CommonEngine.Mysql.Where("uid = ?", pl.Uid).Get(account)
	if err != nil {
		log.Error("check new mail error:%v", err)
		return false, proto_public.CommonErrorCode_ERR_MYSQLERROR
	}

	//先看是否封禁
	if account.ChatBan != 0 && account.ChatBan > utils.Now().Unix() {
		return false, proto_public.CommonErrorCode_ERR_LIMITCHAT
	}

	push := new(model.ChatInfo)
	push.DbId = pl.Id
	push.Content = content
	push.Time = utils.Now().Unix()
	push.Value = value
	push.Cid = cid
	push.Type = msgType

	//附件
	if push.Type == 2 {
		data, err := ChatSendTransaction(ctx, pl, attachment)
		if err != nil {
			log.Error("ChatSendTransaction error : %v", err)
			return false, proto_public.CommonErrorCode_ERR_ParamTypeError
		}
		push.AttachmentData = data
	}

	temp, err := json.Marshal(push)
	if err != nil {
		log.Error("marshal chat message error : %v", err)
		return false, proto_public.CommonErrorCode_ERR_MARSHALERR
	}

	if typ == define.ChatTypeWorld {
		rdb.RedisExec("RPUSH", define.ChatWorld, string(temp))
	} else if typ == define.ChatTypeZudui {
		rdb.RedisExec("RPUSH", fmt.Sprintf("%s:%d", define.ChatZudui, id), string(temp))
	} else if typ == define.ChatTypeGuild {
		//判断帮派ID
		rdb.RedisExec("RPUSH", fmt.Sprintf("%s:%d", define.ChatGuild, id), string(temp))
	} else if typ == define.ChatTypeChuanwen {
		rdb.RedisExec("RPUSH", define.ChatChuanwen, string(temp))
	} else if typ == define.ChatTypeKuafu {
		gdb := db.CommonEngine
		gdb.RedisExec("RPUSH", define.ChatKuafu, string(temp))
	}

	//推送消息
	ctx.Send(&proto_chat.PushChatInfo{
		Type: typ,
		Id:   id,
		MsgInfo: &proto_public.ChatInfo{
			DbId:       pl.Id,
			Uid:        pl.Uid,
			Name:       pl.Base.Name,
			FaceId:     int32(pl.GetProp(define.PlayerPropFaceId)),
			FaceSlotId: int32(pl.GetProp(define.PlayerPropFaceSlotId)),
			Time:       push.Time,
			Info:       push.Content,
			Value:      push.Value,
			Cid:        push.Cid,
			MsgType:    push.Type,
		}})
	return true, proto_public.CommonErrorCode_ERR_OK
}
