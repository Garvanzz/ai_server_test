package friend

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_friend"
	"xfx/proto/proto_public"
)

// 请求添加好友
func ReqAddFriend(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SRequestAddFriend) {
	res := &proto_friend.S2CAddFriend{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("Load Equip error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	// 判断自己好友数量超上限没有
	count, err := rdb.RedisExec("scard", fmt.Sprintf("%s:%d", define.Friend, pl.Id))
	if err != nil {
		log.Error("reqAddFriend redis error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	if count.(int64) >= int64(config.Global.Get().MaxFriendNum) {
		log.Error("reqAddFriend num limit:%v,%v", count, int64(config.Global.Get().MaxFriendNum))
		res.Code = proto_friend.CommonErrorCode_ERR_FriendNumLimit
		ctx.Send(res)
		return
	}

	targetRedisId := req.Pid

	if targetRedisId == pl.Id {
		log.Error("reqAddFriend add self")
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	// 检查是否已经是好友
	isFriend, err := rdb.RedisExec("sismember", fmt.Sprintf("%s:%d", define.Friend, pl.Id), targetRedisId)
	if err != nil {
		log.Error("remove friend db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	if isFriend.(int64) == 1 {
		log.Error("add friend has been friend:%v", targetRedisId)
		res.Code = proto_friend.CommonErrorCode_ERR_IsFriended
		ctx.Send(res)
		return
	}

	// 判断对方好友数量超上限没有
	count, err = rdb.RedisExec("scard", fmt.Sprintf("%s:%d", define.Friend, targetRedisId))
	if err != nil {
		log.Error("reqAddFriend redis error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	if count.(int64) >= int64(config.Global.Get().MaxFriendNum) {
		res.Code = proto_friend.CommonErrorCode_ERR_FriendNumLimit
		ctx.Send(res)
		return
	}
	conn := rdb.Mysql
	// 检查是否已经有好友申请
	exist, err := conn.Table(define.FriendApply).Where("player_id = ? AND target_id = ?", pl.Id, targetRedisId).Exist()
	if err != nil {
		log.Error("add friend db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	if exist {
		log.Debug("reqAddFriend has send apply:%v,uid:%v", err, req.Pid)
		res.Code = proto_friend.CommonErrorCode_ERR_IsFriended
		ctx.Send(res)
		return
	}

	// 判断条数
	applyCount, err := conn.Table(define.FriendApply).Where("target_id = ?", targetRedisId).Count()
	if err != nil {
		log.Error("reqAddFriend db error3:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	if applyCount >= int64(config.Global.Get().MaxFriendNum) {
		res.Code = proto_friend.CommonErrorCode_ERR_FriendNumLimit
		ctx.Send(res)
		return
	}

	ApplyId, _ := redis.Int(rdb.RedisExec("INCRBY", "friendApplyId", 1))
	// 发送申请
	apply := &model.FriendApply{
		Id:       int32(ApplyId),
		PlayerId: pl.Id,
		TargetId: targetRedisId,
	}

	_, err = conn.Table(define.FriendApply).Insert(apply)
	if err != nil {
		log.Error("insert friend apply error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	res.Code = proto_friend.CommonErrorCode_ERR_OK
	ctx.Send(res)

	// 推送给对方
	invoke.Dispatch(ctx, pl.Id, &proto_friend.PushNewApply{})
}

// 请求查找好友
func ReqFindFriend(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SFindFriend) {
	res := &proto_friend.S2CFindFriend{}
	if len(req.IdOrName) == 0 {
		log.Error("reqAddFriend uid and dbId is empty:%v,%v", req.IdOrName)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("Load Equip error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	//目前只判断uid,后面改uid+昵称
	reply, err := rdb.RedisExec("get", fmt.Sprintf("%s:%s", define.Account, req.IdOrName))
	if reply == nil || err != nil {
		log.Error("reply null")
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	}

	dbId, err := redis.Int64(reply, nil)
	if err != nil {
		log.Error("dbId player id error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	}

	if dbId == pl.Id {
		log.Error("reqAddFriend add self")
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	r := new(proto_friend.FriendOption)
	r.Info = new(proto_public.CommonPlayerInfo)
	playerInfo := global.GetPlayerInfo(dbId)
	r.Info.PlayerId = playerInfo.Id
	r.Info.Name = playerInfo.Name
	r.Info.Level = playerInfo.Level
	r.Info.FaceId = playerInfo.FaceId
	r.Info.FaceSlotId = playerInfo.FaceSlotId
	r.Info.Power = 0

	//是否在线
	isOnline := invoke.LoginClient(ctx).IsOnline(dbId)
	r.Online = isOnline
	if r.Online == false {
		r.LastOnlineTime = time.Now().Unix() - playerInfo.OfflineTime
	}

	//是否申请
	conn := rdb.Mysql
	// 检查是否已经有好友申请
	exist, err := conn.Table(define.FriendApply).Where("player_id = ? AND target_id = ?", pl.Id, playerInfo.Id).Exist()
	if err != nil {
		log.Error("add friend db error:%v", err)
		r.IsApply = false
	} else {
		if exist {
			r.IsApply = true
		} else {
			r.IsApply = false
		}
	}

	res.FriendOption = []*proto_friend.FriendOption{r}
	res.Code = proto_friend.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 请求好友列表
func ReqFriendList(ctx global.IPlayer, pl *model.Player) {
	res := &proto_friend.S2CFriendList{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqFriendList error, no this server:%v", err)
		ctx.Send(res)
		return
	}

	data, err := getFriendList(rdb, ctx, pl)
	if err != nil {
		log.Error("ReqFriendList err: %v", err)
		ctx.Send(res)
		return
	}

	res.FriendOption = data
	ctx.Send(res)
}

// 获取好友列表
func getFriendList(rdb *db.CDBEngine, ctx global.IPlayer, pl *model.Player) ([]*proto_friend.FriendOption, error) {
	reply, err := rdb.RedisExec("smembers", fmt.Sprintf("%s:%d", define.Friend, pl.Id))
	if err != nil {
		return nil, err
	}

	vals := reply.([]interface{})
	data := make([]*proto_friend.FriendOption, 0)
	for i := 0; i < len(vals); i++ {
		str := string(vals[i].([]byte))
		dbId, _ := strconv.ParseInt(str, 10, 64)
		r := new(proto_friend.FriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		playerInfo := global.GetPlayerInfo(dbId)
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		r.Info.Power = 0

		//是否在线
		isOnline := invoke.LoginClient(ctx).IsOnline(dbId)
		r.Online = isOnline
		if r.Online == false {
			r.LastOnlineTime = time.Now().Unix() - playerInfo.OfflineTime
		}

		//判断状态
		reply, err = rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), dbId)
		if err != nil {
			return nil, err
		}

		m := new(model.FriendGift)
		if reply == nil {
			r.SendGiftState = define.FriendGiftState_Send
		} else {
			err = json.Unmarshal(reply.([]byte), &m)
			if err != nil {
				return nil, err
			}

			//优先显示可以领取，然后是可以发送，然后是已经领取，最后是无
			if m.IsCanGet {
				r.SendGiftState = define.FriendGiftState_OtherCanGet
			} else {
				if m.IsSend {
					if m.IsAlGet {
						r.SendGiftState = define.FriendGiftState_OtherAlGet
					} else {
						r.SendGiftState = define.FriendGiftState_Null
					}
				} else {
					r.SendGiftState = define.FriendGiftState_Send
				}
			}
		}

		data = append(data, r)
	}
	return data, nil
}

// 请求申请列表
func ReqFriendApplyList(ctx global.IPlayer, pl *model.Player) {
	res := &proto_friend.S2CFriendApplyList{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqFriendList error, no this server:%v", err)
		ctx.Send(res)
		return
	}

	conn := rdb.Mysql
	list := make([]*model.FriendApply, 0)
	err = conn.Table(define.FriendApply).Where("target_id = ?", pl.Id).Limit(int(int64(config.Global.Get().MaxFriendNum))).Find(&list)
	if err != nil {
		log.Error("reqFriendApplyList db error:%v", err)
		ctx.Send(res)
		return
	}

	ret := make([]*proto_friend.ApplyFriendOption, 0)
	for _, v := range list {
		playerInfo := global.GetPlayerInfo(v.PlayerId)
		r := new(proto_friend.ApplyFriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		r.Id = v.Id
		r.Message = v.Msg
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		r.Info.Power = 0

		//是否在线
		isOnline := invoke.LoginClient(ctx).IsOnline(playerInfo.Id)
		r.Online = isOnline
		if r.Online == false {
			r.LastOnlineTime = time.Now().Unix() - playerInfo.OfflineTime
		}

		ret = append(ret, r)
	}
	res.Apply = ret
	ctx.Send(res)
}

// 删除好友
func ReqRemoveFriend(ctx global.IPlayer, pl *model.Player, msg *proto_friend.C2SDeleteFriend) {
	res := &proto_friend.S2CDeleteFriend{}
	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqFriendList error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}
	exist, err := rdb.RedisExec("sismember", fmt.Sprintf("%s:%d", define.Friend, pl.Id), msg.PlayerId)
	if err != nil {
		log.Error("remove friend db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	}

	if exist.(int64) == 0 {
		log.Error("remove friend not exist:%v", msg.PlayerId)
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	}

	// 删除自己的好友
	_, err = rdb.RedisExec("srem", fmt.Sprintf("%s:%d", define.Friend, pl.Id), msg.PlayerId)
	if err != nil {
		log.Error("remove friend db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	}

	// 删除对方的好友
	_, err = rdb.RedisExec("srem", fmt.Sprintf("%s:%d", define.Friend, msg.PlayerId), pl.Id)
	if err != nil {
		log.Error("remove friend db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	}

	res.Code = proto_friend.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 请求处理好友请求
func ReqDealFriendApply(ctx global.IPlayer, pl *model.Player, msg *proto_friend.C2SreqDealFriendApply) {
	res := &proto_friend.S2CDealFriendApply{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqDealFriendApply error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	// 检查是否已经有好友申请
	conn := rdb.Mysql

	apply := new(model.FriendApply)
	_, err = conn.Table(define.FriendApply).Where("id = ?", msg.ApplyId).Get(apply)
	if err != nil {
		log.Error("reqDealFriendApply db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	if apply.Id == 0 {
		log.Error("reqDealFriendApply apply not exist:%v", msg.ApplyId)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	// delete apply
	n, err := conn.Table(define.FriendApply).Where("id = ?", apply.Id).Delete()
	if err != nil || n != 1 {
		log.Error("reqDealFriendApply delete error:%v,%v", err, n)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	if msg.Action == true { // 同意
		if apply.PlayerId == pl.Id {
			log.Error("reqDealFriendApply add self:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
			ctx.Send(res)
			return
		}

		// 判断自己好友数量超上限没有
		count, err := rdb.RedisExec("scard", fmt.Sprintf("%s:%d", define.Friend, pl.Id))
		if err != nil {
			log.Error("reqDealFriendApply redis error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		if count.(int64) >= int64(config.Global.Get().MaxFriendNum) {
			log.Error("reqDealFriendApply num limit:%v,%v", count, int64(config.Global.Get().MaxFriendNum))
			res.Code = proto_friend.CommonErrorCode_ERR_ApplyNumLimit
			ctx.Send(res)
			return
		}

		// 判断对方好友数量超上限没有
		count, err = rdb.RedisExec("scard", fmt.Sprintf("%s:%d", define.Friend, apply.PlayerId))
		if err != nil {
			log.Error("reqDealFriendApply redis error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		if count.(int64) >= int64(config.Global.Get().MaxFriendNum) {
			res.Code = proto_friend.CommonErrorCode_ERR_OtherFriendNumLimit
			ctx.Send(res)
			return
		}

		// 然后添加到对方和自己的set
		_, err = rdb.RedisExec("sadd", fmt.Sprintf("%s:%d", define.Friend, apply.PlayerId), pl.Id)
		if err != nil {
			log.Error("reqDealFriendApply redis error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		_, err = rdb.RedisExec("sadd", fmt.Sprintf("%s:%d", define.Friend, pl.Id), apply.PlayerId)
		if err != nil {
			log.Error("reqDealFriendApply redis error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}
	}

	//同步列表
	list := make([]*model.FriendApply, 0)
	err = conn.Table(define.FriendApply).Where("target_id = ?", pl.Id).Limit(int(int64(config.Global.Get().MaxFriendNum))).Find(&list)
	if err != nil {
		log.Error("syncFriendApplyList db error:%v", err)
		ctx.Send(res)
		return
	}

	ret := make([]*proto_friend.ApplyFriendOption, 0)
	for _, v := range list {
		playerInfo := global.GetPlayerInfo(v.PlayerId)
		r := new(proto_friend.ApplyFriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		r.Id = v.Id
		r.Message = v.Msg
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		ret = append(ret, r)
	}
	res.Apply = ret
	res.Code = proto_friend.CommonErrorCode_ERR_OK

	ctx.Send(res)
}

// 请求好友赠送
func ReqFriendGift(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SReqFriendGift) {
	res := &proto_friend.S2CFriendGift{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("Load Equip error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	// 检查是否已经是好友
	isFriend, err := rdb.RedisExec("sismember", fmt.Sprintf("%s:%d", define.Friend, pl.Id), req.Pid)
	if err != nil {
		log.Error("remove friend db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	if isFriend.(int64) <= 0 {
		log.Error("add friend has been friend:%v", req.Pid)
		res.Code = proto_friend.CommonErrorCode_ERR_IsFriended
		ctx.Send(res)
		return
	}

	//自己这边
	reply, err := rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), req.Pid)
	if err != nil {
		log.Error("db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	m := new(model.FriendGift)
	if reply == nil {
		m.IsSend = true
	} else {
		err = json.Unmarshal(reply.([]byte), &m)
		if err != nil {
			log.Error("Unmarshal error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		if m.IsSend {
			res.Code = proto_friend.CommonErrorCode_ERR_IsApplyed
			ctx.Send(res)
			return
		}
		m.IsSend = true
	}

	js, _ := json.Marshal(m)
	rdb.RedisExec("HSET", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), req.Pid, js)

	// 设置过期时间
	_, err = rdb.RedisExec("EXPIREAT", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), utils.TimestampTodayMillisecond())
	if err != nil {
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	//对方
	reply, err = rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.Friend_Gift, req.Pid), pl.Id)
	if err != nil {
		log.Error("db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	m = new(model.FriendGift)
	if reply != nil {
		err = json.Unmarshal(reply.([]byte), &m)
		if err != nil {
			log.Error("Unmarshal error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}
	}

	m.IsCanGet = true

	js, _ = json.Marshal(m)
	rdb.RedisExec("HSET", fmt.Sprintf("%s:%d", define.Friend_Gift, req.Pid), pl.Id, js)

	// 设置过期时间
	_, err = rdb.RedisExec("EXPIREAT", fmt.Sprintf("%s:%d", define.Friend_Gift, req.Pid), utils.TimestampTodayMillisecond())
	if err != nil {
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	respdata, err := getFriendList(rdb, ctx, pl)
	if err != nil {
		log.Error("ReqFriendList err: %v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	res.FriendOption = respdata
	res.Code = proto_friend.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 请求好友赠送领取
func ReqGetFriendGift(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SReqGetFriendGift) {
	res := &proto_friend.S2CGetFriendGift{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("Load ReqGetFriendGift error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	// 检查是否已经是好友
	isFriend, err := rdb.RedisExec("sismember", fmt.Sprintf("%s:%d", define.Friend, pl.Id), req.Pid)
	if err != nil {
		log.Error("remove friend db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ADDFriendFaild
		ctx.Send(res)
		return
	}

	if isFriend.(int64) <= 0 {
		log.Error("add friend has been friend:%v", req.Pid)
		res.Code = proto_friend.CommonErrorCode_ERR_IsFriended
		ctx.Send(res)
		return
	}

	reply, err := rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), req.Pid)
	if err != nil {
		log.Error("db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	m := new(model.FriendGift)
	if reply == nil {
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	} else {
		err = json.Unmarshal(reply.([]byte), &m)
		if err != nil {
			log.Error("Unmarshal error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		if m.IsAlGet {
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		if !m.IsCanGet {
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}
	}

	//发送奖励
	award := config.Global.Get().FriendGift
	bag.AddAward(ctx, pl, award, true)

	m.IsAlGet = true
	m.IsCanGet = false
	js, _ := json.Marshal(m)
	rdb.RedisExec("HSET", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), req.Pid, js)

	// 设置过期时间
	_, err = rdb.RedisExec("EXPIREAT", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), utils.TimestampTodayMillisecond())
	if err != nil {
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	respdata, err := getFriendList(rdb, ctx, pl)
	if err != nil {
		log.Error("ReqFriendList err: %v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	res.FriendOption = respdata
	res.Code = proto_friend.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 一键请求好友赠送和领取
func ReqOneKeyFriendGift(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SOneKeyFriendGift) {
	res := &proto_friend.S2COneKeyFriendGift{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("Load Equip error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	reply, err := rdb.RedisExec("smembers", fmt.Sprintf("%s:%d", define.Friend, pl.Id))
	if err != nil {
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	vals := reply.([]interface{})
	data := make([]*proto_friend.FriendOption, 0)
	awards := make([]conf.ItemE, 0)
	for i := 0; i < len(vals); i++ {
		str := string(vals[i].([]byte))
		dbId, _ := strconv.ParseInt(str, 10, 64)
		//赠送 --------------------------------
		//自己这边
		replyself, err := rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), dbId)
		if err != nil {
			log.Error("db error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		m := new(model.FriendGift)
		if replyself == nil {
			m.IsSend = true
		} else {
			err = json.Unmarshal(replyself.([]byte), &m)
			if err != nil {
				log.Error("Unmarshal error:%v", err)
				res.Code = proto_friend.CommonErrorCode_ERR_DBERR
				ctx.Send(res)
				return
			}

			if m.IsSend {
				continue
			}
			m.IsSend = true
		}

		//领取
		if m.IsCanGet {
			a := config.Global.Get().FriendGift
			awards = append(awards, a...)
			m.IsCanGet = false
		}

		js, _ := json.Marshal(m)
		rdb.RedisExec("HSET", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), dbId, js)

		// 设置过期时间
		_, err = rdb.RedisExec("EXPIREAT", fmt.Sprintf("%s:%d", define.Friend_Gift, pl.Id), utils.TimestampTodayMillisecond())
		if err != nil {
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		//对方
		reply, err = rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.Friend_Gift, dbId), pl.Id)
		if err != nil {
			log.Error("db error:%v", err)
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		m = new(model.FriendGift)
		if reply != nil {
			err = json.Unmarshal(reply.([]byte), &m)
			if err != nil {
				log.Error("Unmarshal error:%v", err)
				res.Code = proto_friend.CommonErrorCode_ERR_DBERR
				ctx.Send(res)
				return
			}
		}

		m.IsCanGet = true

		js, _ = json.Marshal(m)
		rdb.RedisExec("HSET", fmt.Sprintf("%s:%d", define.Friend_Gift, dbId), pl.Id, js)

		// 设置过期时间
		_, err = rdb.RedisExec("EXPIREAT", fmt.Sprintf("%s:%d", define.Friend_Gift, dbId), utils.TimestampTodayMillisecond())
		if err != nil {
			res.Code = proto_friend.CommonErrorCode_ERR_DBERR
			ctx.Send(res)
			return
		}

		r := new(proto_friend.FriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		playerInfo := global.GetPlayerInfo(dbId)
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId

		//是否在线
		isOnline := invoke.LoginClient(ctx).IsOnline(dbId)
		r.Online = isOnline
		if !r.Online {
			r.LastOnlineTime = time.Now().Unix() - playerInfo.OfflineTime
		}

		json.Unmarshal(replyself.([]byte), &m)
		//优先显示可以领取，然后是可以发送，然后是已经领取，最后是无
		if m.IsCanGet {
			r.SendGiftState = define.FriendGiftState_OtherCanGet
		} else {
			if m.IsSend {
				if m.IsAlGet {
					r.SendGiftState = define.FriendGiftState_OtherAlGet
				} else {
					r.SendGiftState = define.FriendGiftState_Null
				}
			} else {
				r.SendGiftState = define.FriendGiftState_Send
			}
		}

		data = append(data, r)
	}

	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, true)
	}

	res.FriendOption = data
	res.Code = proto_friend.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 请求黑名单列表
func ReqBlockFriendList(ctx global.IPlayer, pl *model.Player) {
	res := &proto_friend.S2CReqInitBlockFriend{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqFriendList error, no this server:%v", err)
		ctx.Send(res)
		return
	}

	conn := rdb.Mysql
	list := make([]*model.FriendBlock, 0)
	err = conn.Table(define.FriendBlock).Where("player_id = ?", pl.Id).Limit(int(int64(config.Global.Get().MaxFriendNum))).Find(&list)
	if err != nil {
		log.Error("reqFriendBlockList db error:%v", err)
		ctx.Send(res)
		return
	}

	ret := make([]*proto_friend.FriendOption, 0)
	for _, v := range list {
		playerInfo := global.GetPlayerInfo(v.TargetId)
		r := new(proto_friend.FriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		r.Info.Power = 0

		//是否在线
		isOnline := invoke.LoginClient(ctx).IsOnline(playerInfo.Id)
		r.Online = isOnline
		if !r.Online {
			r.LastOnlineTime = time.Now().Unix() - playerInfo.OfflineTime
		}

		ret = append(ret, r)
	}
	res.FriendOption = ret
	ctx.Send(res)
}

// 拉入黑名单
func ReqBlockFriend(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SReqBlockFriend) {
	res := &proto_friend.S2CRespBlockFriend{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqFriendList error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	conn := rdb.Mysql
	block := new(model.FriendBlock)
	has, err := conn.Table(define.FriendBlock).Where("target_id = ? AND player_id ", req.PlayerId, pl.Id).Exist(&block)
	if err != nil {
		log.Error("reqFriendBlockList db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	if has {
		log.Error("reqFriendBlockList db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_ALBlock
		ctx.Send(res)
		return
	}

	//从别人的好友中删除
	_, err = rdb.RedisExec("SREM", fmt.Sprintf("%s:%d", define.Friend, req.PlayerId), pl.Id)
	if err != nil {
		log.Error("别人的好友中删除 redis error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	block.TargetId = req.PlayerId
	block.PlayerId = pl.Id
	_, err = conn.Table(define.FriendBlock).Insert(&block)
	if err != nil {
		log.Error("reqFriendBlock db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	list := make([]*model.FriendBlock, 0)
	err = conn.Table(define.FriendBlock).Where("player_id = ?", pl.Id).Limit(int(int64(config.Global.Get().MaxFriendNum))).Find(&list)
	if err != nil {
		log.Error("reqFriendBlockList db error:%v", err)
		ctx.Send(res)
		return
	}

	ret := make([]*proto_friend.FriendOption, 0)
	for _, v := range list {
		playerInfo := global.GetPlayerInfo(v.TargetId)
		r := new(proto_friend.FriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		ret = append(ret, r)
	}
	res.FriendOption = ret
	ctx.Send(res)
}

// 解除黑名单
func ReqUnLockBlockFriend(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SReqUnlockBlockFriend) {
	res := &proto_friend.S2CRespUnlockBlockFriend{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqUnLockBlockFriend error, no this server:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	conn := rdb.Mysql
	block := new(model.FriendBlock)
	has, err := conn.Table(define.FriendBlock).Where("target_id = ? AND player_id ", req.PlayerId, pl.Id).Exist(&block)
	if err != nil {
		log.Error("reqFriendBlockList db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	if !has {
		log.Error("reqFriendBlockList db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_NOPlayer
		ctx.Send(res)
		return
	}

	block.TargetId = req.PlayerId
	block.PlayerId = pl.Id
	_, err = conn.Table(define.FriendBlock).Delete(&block)
	if err != nil {
		log.Error("reqFriendBlock db error:%v", err)
		res.Code = proto_friend.CommonErrorCode_ERR_DBERR
		ctx.Send(res)
		return
	}

	list := make([]*model.FriendBlock, 0)
	err = conn.Table(define.FriendBlock).Where("player_id = ?", pl.Id).Limit(int(int64(config.Global.Get().MaxFriendNum))).Find(&list)
	if err != nil {
		log.Error("reqFriendBlockList db error:%v", err)
		ctx.Send(res)
		return
	}

	ret := make([]*proto_friend.FriendOption, 0)
	for _, v := range list {
		playerInfo := global.GetPlayerInfo(v.TargetId)
		r := new(proto_friend.FriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		ret = append(ret, r)
	}

	res.FriendOption = ret
	res.Code = proto_friend.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 请求推荐好友
func ReqTuijianFriend(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SReqTuijianFriend) {
	res := &proto_friend.S2CReSPTuijianFriend{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqUnLockBlockFriend error, no this server:%v", err)
		ctx.Send(res)
		return
	}

	// 获取列表
	reply, err := rdb.RedisExec("SMEMBERS", fmt.Sprintf("%s:%d", define.Friend_Recommend, pl.Id))
	if err != nil {
		log.Error("remove friend db error:%v", err)
		ctx.Send(res)
		return
	}

	data := make([]*proto_friend.FriendOption, 0)
	var list []int64

	vals := reply.([]interface{})
	if len(vals) > 0 {
		for i := 0; i < len(vals); i++ {
			str := string(vals[i].([]byte))
			dbId, _ := strconv.ParseInt(str, 10, 64)
			list = append(list, dbId)
		}
	} else {
		list, err = RefreshFriendTuijian(pl)
		if err != nil {
			log.Error("remove friend db error:%v", err)
			ctx.Send(res)
			return
		}

		//保存
		args := make([]interface{}, 0, len(list)+1)
		args = append(args, fmt.Sprintf("%s:%d", define.Friend_Recommend, pl.Id))
		for _, item := range list {
			args = append(args, item)
		}

		rdb.RedisExec("sadd", args...)
	}

	for i := 0; i < len(list); i++ {
		r := new(proto_friend.FriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		playerInfo := global.GetPlayerInfo(list[i])
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		r.Info.Power = 0

		//是否在线
		isOnline := invoke.LoginClient(ctx).IsOnline(playerInfo.Id)
		r.Online = isOnline
		if !r.Online {
			r.LastOnlineTime = time.Now().Unix() - playerInfo.OfflineTime
		}

		//是否申请
		conn := rdb.Mysql
		// 检查是否已经有好友申请
		exist, err := conn.Table(define.FriendApply).Where("player_id = ? AND target_id = ?", pl.Id, playerInfo.Id).Exist()
		if err != nil {
			log.Error("add friend db error:%v", err)
			continue
		}

		if exist {
			r.IsApply = true
		} else {
			r.IsApply = false
		}

		data = append(data, r)
	}

	res.FriendOption = data
	ctx.Send(res)
}

// 请求刷新推荐好友
func ReqRefreshTuijianFriend(ctx global.IPlayer, pl *model.Player, req *proto_friend.C2SReqRefreshFriend) {
	res := &proto_friend.S2CReqRefreshFriend{}

	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqUnLockBlockFriend error, no this server:%v", err)
		ctx.Send(res)
		return
	}

	// 获取列表
	reply, err := rdb.RedisExec("SMEMBERS", fmt.Sprintf("%s:%d", define.Friend_Recommend, pl.Id))
	if err != nil {
		log.Error("remove friend db error:%v", err)
		ctx.Send(res)
		return
	}

	data := make([]*proto_friend.FriendOption, 0)
	var list []int64

	vals := reply.([]interface{})
	if len(vals) > 0 {
		//先移除之前的list
		_, err = rdb.RedisExec("UNLINK", fmt.Sprintf("%s:%d", define.Friend_Recommend, pl.Id))
		if err != nil {
			log.Error("remove friend db error:%v", err)
			ctx.Send(res)
			return
		}
	}

	list, err = RefreshFriendTuijian(pl)
	if err != nil {
		log.Error("remove friend db error:%v", err)
		ctx.Send(res)
		return
	}

	//保存
	args := make([]interface{}, 0, len(list)+1)
	args = append(args, fmt.Sprintf("%s:%d", define.Friend_Recommend, pl.Id))
	for _, item := range list {
		args = append(args, item)
	}

	rdb.RedisExec("sadd", args...)

	for i := 0; i < len(list); i++ {
		r := new(proto_friend.FriendOption)
		r.Info = new(proto_public.CommonPlayerInfo)
		playerInfo := global.GetPlayerInfo(list[i])
		r.Info.PlayerId = playerInfo.Id
		r.Info.Name = playerInfo.Name
		r.Info.Level = playerInfo.Level
		r.Info.FaceId = playerInfo.FaceId
		r.Info.FaceSlotId = playerInfo.FaceSlotId
		r.Info.Power = 0

		//是否在线
		isOnline := invoke.LoginClient(ctx).IsOnline(playerInfo.Id)
		r.Online = isOnline
		if !r.Online {
			r.LastOnlineTime = time.Now().Unix() - playerInfo.OfflineTime
		}

		//是否申请
		conn := rdb.Mysql
		// 检查是否已经有好友申请
		exist, err := conn.Table(define.FriendApply).Where("player_id = ? AND target_id = ?", pl.Id, playerInfo.Id).Exist()
		if err != nil {
			log.Error("add friend db error:%v", err)
			continue
		}

		if exist {
			r.IsApply = true
		} else {
			r.IsApply = false
		}

		data = append(data, r)
	}

	res.FriendOption = data
	ctx.Send(res)
}

// 刷新好友推荐
func RefreshFriendTuijian(pl *model.Player) ([]int64, error) {
	//获取玩家列表
	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("RefreshFriendTuijian error, no this server:%v", err)
		return nil, err
	}

	reply, err := rdb.RedisExec("smembers", fmt.Sprintf("%s:%d", define.Friend, pl.Id))
	if err != nil {
		return nil, err
	}

	vals := reply.([]interface{})
	ids := make([]int64, 0)
	for i := 0; i < len(vals); i++ {
		str := string(vals[i].([]byte))
		dbId, _ := strconv.ParseInt(str, 10, 64)
		ids = append(ids, dbId)
	}

	//排除自己
	ids = append(ids, pl.Id)

	// 将 []int64 转换为 []interface{}
	idInterfaces := make([]interface{}, len(ids))
	for i, id := range ids {
		idInterfaces[i] = id
	}

	//排除好友
	account := make([]*model.Account, 0)
	query := db.CommonEngine.Mysql.Table("account")
	//Where("server_id = ?", pl.ServerId) // TODO:

	if len(idInterfaces) > 0 {
		query = query.NotIn("redis_id", idInterfaces) // 使用 xorm 的 NotIn 方法
	}

	err = query.OrderBy("RAND()").Limit(5).Find(&account)
	if err != nil {
		log.Error("check new mail error:%v", err)
		return nil, err
	}

	res := make([]int64, 0)
	for _, v := range account {
		res = append(res, v.RedisId)
	}

	return res, nil
}
