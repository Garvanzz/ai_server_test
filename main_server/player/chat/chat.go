package chat

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_chat"
	"xfx/proto/proto_public"
)

// ReqChatInfo 请求所有聊天消息 1/跨服 2/世界 3/组队 4/私聊 5/帮会 6/传闻
func ReqChatInfo(ctx global.IPlayer, pl *model.Player, req *proto_chat.C2SChatInfo) {
	resp := new(proto_chat.S2CChatInfo)
	resp.MsgInfo = make([]*proto_public.ChatInfo, 0)

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqChatInfo error, no this server:%v", err)
		ctx.Send(resp)
		return
	}

	resp.Type = req.Type
	if req.Type == define.ChatTypeKuafu {
		gdb := db.CommonEngine
		reply, err := gdb.RedisExec("LRANGE", define.ChatKuafu, 0-define.ChatMsgLen, -1)
		if err != nil {
			log.Error("load chat world db err : %v", err)
		}
		resp.MsgInfo = chatToProto(ctx, unmarshalChatInfo(reply))
	} else if req.Type == define.ChatTypeWorld { // 世界频道
		reply, err := rdb.RedisExec("LRANGE", define.ChatWorld, 0-define.ChatMsgLen, -1)
		if err != nil {
			log.Error("load chat world db err : %v", err)
		}
		resp.MsgInfo = chatToProto(ctx, unmarshalChatInfo(reply))
	} else if req.Type == define.ChatTypeGuild { // 帮会频道
		reply, err := rdb.RedisExec("LRANGE", fmt.Sprintf("%s:%d", define.ChatGuild, req.Id), 0-define.ChatMsgLen, -1)
		if err != nil {
			log.Error("load chat room db err : %v", err)
		}
		resp.MsgInfo = chatToProto(ctx, unmarshalChatInfo(reply))
	} else if req.Type == define.ChatTypePrivate { // 私聊频道

	} else if req.Type == define.ChatTypeZudui { // 组队
		reply, err := rdb.RedisExec("LRANGE", fmt.Sprintf("%s:%d", define.ChatZudui, req.Id), 0-define.ChatMsgLen, -1)
		if err != nil {
			log.Error("load chat room db err : %v", err)
		}
		resp.MsgInfo = chatToProto(ctx, unmarshalChatInfo(reply))
	} else if req.Type == define.ChatTypeChuanwen { // 传闻
		reply, err := rdb.RedisExec("LRANGE", define.ChatChuanwen, 0-define.ChatMsgLen, -1)
		if err != nil {
			log.Error("load chat room db err : %v", err)
		}
		resp.MsgInfo = chatToProto(ctx, unmarshalChatInfo(reply))
	}

	ctx.Send(resp)
}

// ReqSendChat 请求发送聊天
func ReqSendChat(ctx global.IPlayer, pl *model.Player, req *proto_chat.C2SSendChatMsg) {
	res := &proto_chat.S2CSendChatMsg{}

	if req.MsgType == 2 && req.Type != define.ChatTypeWorld {
		log.Error("chat msg is attachment,but type:%d", req.Type)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	state, code := internal.SyncChatSend(ctx, pl, req.Type, req.Id, req.Content, nil, 0, req.MsgType, req.AttachmentInfo)
	if !state {
		res.Code = code
		ctx.Send(res)
		return
	}

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 请求私聊信息
func ReqPrivateChatData(ctx global.IPlayer, pl *model.Player, req *proto_chat.C2SGetPrivateChatData) {
	resp := new(proto_chat.S2CGetPrivateChatData)
	resp.List = make([]*proto_public.ChatInfo, 0)

	//判断有没有这个人
	account := new(model.Account)
	account.RedisId = req.Id
	//account.ServerId = pl.ServerId TODO：玩家id 已经加了服务器id
	has, err := db.CommonEngine.Mysql.Table(define.AccountTable).Get(&account)
	if err != nil {
		log.Error("chat db error:%v", err)
		ctx.Send(resp)
		return
	}

	if !has {
		log.Error("chat db error, no this account:%v", err)
		ctx.Send(resp)
		return
	}

	var key string
	if pl.Id > req.Id {
		key = fmt.Sprintf("%s%d-%d", define.PrivateChatDataKey, req.Id, pl.Id)
	} else {
		key = fmt.Sprintf("%s%d-%d", define.PrivateChatDataKey, pl.Id, req.Id)
	}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqChatInfo error, no this server:%v", err)
		ctx.Send(resp)
		return
	}

	// 不能超过最大长度
	rdb.RedisExec("LTRIM", key, 0-define.MsgMaxLen, -1)

	// TODO:这里使用异步的获取方式
	reply, err := rdb.RedisExec("LRANGE", key, 0, -1)
	if err != nil {
		log.Error("get private data from redis error:%v", err)
		ctx.Send(resp)
		return
	}

	resp.List = make([]*proto_public.ChatInfo, 0)

	res, _ := reply.([]interface{})
	for i := 0; i < len(res); i++ {
		temp := new(proto_public.ChatInfo)
		err := json.Unmarshal(res[i].([]byte), &temp)
		if err != nil {
			log.Error("unmarshal private chat history error :%v", err)
			ctx.Send(resp)
			return
		}
		resp.List = append(resp.List, temp)
	}
	ctx.Send(resp)
}

// 请求发送私聊信息
func ReqSendPrivateChatData(ctx global.IPlayer, pl *model.Player, req *proto_chat.C2SSendPrivateChat) {
	resp := new(proto_chat.S2CSendPrivateChat)

	//判断有没有这个人
	account := new(model.Account)
	account.RedisId = req.Id
	//account.ServerId = pl.ServerId
	has, err := db.CommonEngine.Mysql.Table(define.AccountTable).Get(&account)
	if err != nil {
		log.Error("chat db error:%v", err)
		ctx.Send(resp)
		return
	}

	if !has {
		log.Error("chat db error, no this account:%v", err)
		ctx.Send(resp)
		return
	}

	var key string
	if pl.Id > req.Id {
		key = fmt.Sprintf("%s%d-%d", define.PrivateChatDataKey, req.Id, pl.Id)
	} else {
		key = fmt.Sprintf("%s%d-%d", define.PrivateChatDataKey, pl.Id, req.Id)
	}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqChatInfo error, no this server:%v", err)
		ctx.Send(resp)
		return
	}

	// 私聊信息
	talkData := &proto_public.ChatInfo{
		DbId:       pl.Id,
		Uid:        pl.Uid,
		Name:       pl.Base.Name,
		FaceId:     int32(pl.GetProp(define.PlayerPropFaceId)),
		FaceSlotId: int32(pl.GetProp(define.PlayerPropFaceSlotId)),
		Info:       req.Content,
	}

	b, err := json.Marshal(talkData)
	if err != nil {
		log.Error("send private chat data json marshal error:%v", err)
		return
	}

	rdb.RedisExec("RPUSH", key, string(b))

	// 不能超过最大长度
	rdb.RedisExec("LTRIM", key, 0-define.MsgMaxLen, -1)

	// 设置过期时间
	rdb.RedisExec("EXPIRE", key, define.PrivateChatExpiration)

	// TODO:推送给聊天对象
	//推送消息
	ctx.Send(&proto_chat.PushChatInfo{
		Type: 4,
		Id:   int32(req.Id),
		MsgInfo: &proto_public.ChatInfo{
			DbId:       pl.Id,
			Uid:        pl.Uid,
			Name:       pl.Base.Name,
			FaceId:     int32(pl.GetProp(define.PlayerPropFaceId)),
			FaceSlotId: int32(pl.GetProp(define.PlayerPropFaceSlotId)),
			Time:       utils.Now().Unix(),
			Info:       req.Content,
		}})

	return
}

func chatToProto(ctx global.IPlayer, msg []*model.ChatInfo) []*proto_public.ChatInfo {
	res := make([]*proto_public.ChatInfo, 0, len(msg))

	for i := 0; i < len(msg); i++ {
		chatInfo := new(proto_public.ChatInfo)
		playerInfo := global.GetPlayerInfo(msg[i].DbId)
		chatInfo.DbId = playerInfo.Id
		chatInfo.Uid = playerInfo.Uid
		chatInfo.Name = playerInfo.Name
		chatInfo.FaceId = playerInfo.FaceId
		chatInfo.FaceSlotId = playerInfo.FaceSlotId
		chatInfo.Time = msg[i].Time
		chatInfo.Info = msg[i].Content
		chatInfo.Value = msg[i].Value
		chatInfo.Cid = msg[i].Cid
		chatInfo.MsgType = msg[i].Type

		//附件
		if chatInfo.MsgType == 2 && chatInfo.AttachmentInfo != nil {
			// 查询订单
			order := invoke.TransactionClient(ctx).GetOrder(chatInfo.AttachmentInfo.Id)
			if order == nil {
				continue
			}
		}
		res = append(res, chatInfo)
	}
	return res
}

func unmarshalChatInfo(reply any) []*model.ChatInfo {
	chatInfos := make([]*model.ChatInfo, 0)

	if reply == nil {
		return chatInfos
	}

	res, _ := reply.([]interface{})
	for i := 0; i < len(res); i++ {
		temp := new(model.ChatInfo)
		err := json.Unmarshal(res[i].([]byte), &temp)
		if err != nil {
			log.Error("unmarshal chat history error :%v", err)
			return chatInfos
		}
		chatInfos = append(chatInfos, temp)
	}

	return chatInfos
}
