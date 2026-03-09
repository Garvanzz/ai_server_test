package global

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_public"
)

// GetPlayerInfo 获取玩家装配的装备, 正在使用的装备/坐骑/神兵/背饰
func GetPlayerEquipBindInfo(dbId int64) (map[int32]*proto_public.CommonPlayerEquipInfo, int32, int32, int32) {
	equips := make(map[int32]*proto_public.CommonPlayerEquipInfo)
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerEquip, dbId))
	if err != nil {
		log.Error("player[%v],load bag error:%v", dbId, err)
		return equips, 0, 0, 0
	}

	if reply == nil {
		return equips, 0, 0, 0
	}

	dst := new(model.Equip)
	err = json.Unmarshal(reply.([]byte), &dst)
	if err != nil {
		log.Error("player[%v],load Equip unmarshal error:%v", dbId, err)
		return equips, 0, 0, 0
	}

	//正在使用的装备
	for _, v := range dst.Equips {
		if v.IsUse {
			equips[v.Id] = &proto_public.CommonPlayerEquipInfo{
				PlayerId: dbId,
				Id:       v.Id,
				Level:    v.Level,
				Index:    v.Index,
				CId:      v.CId,
			}
		}
	}

	//正在使用的坐骑
	mountId := dst.Mount.UseId

	//正在使用的神兵
	weaponryId := dst.Weaponry.UseId

	//正在使用的背饰
	braceId := int32(0)
	for _, v := range dst.Brace.BraceItems {
		if v.IsUse {
			braceId = v.Id
			break
		}
	}

	return equips, mountId, weaponryId, braceId
}

// GetPlayerCollectInfo 获取玩家藏品
func GetPlayerCollectInfo(dbId int64) map[int32]*proto_public.CommonPlayerCollectionInfo {
	collects := make(map[int32]*proto_public.CommonPlayerCollectionInfo)
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerCollection, dbId))
	if err != nil {
		log.Error("player[%v],load Collection error:%v", dbId, err)
		return collects
	}

	if reply == nil {
		return collects
	}

	dst := new(model.Collection)
	err = json.Unmarshal(reply.([]byte), &dst)
	if err != nil {
		log.Error("player[%v],load Equip unmarshal error:%v", dbId, err)
		return collects
	}

	//正在使用的装备
	for _, v := range dst.Collections {
		collects[v.Id] = &proto_public.CommonPlayerCollectionInfo{
			Star: v.Star,
			Id:   v.Id,
		}
	}

	return collects
}

// GetPlayerEquip 获取玩家装备
func GetPlayerEquip(dbId int64) *model.Equip {
	equip := new(model.Equip)
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerEquip, dbId))
	if err != nil {
		log.Error("player[%v],load equip error:%v", dbId, err)
		return equip
	}

	if reply == nil {
		return equip
	}

	err = json.Unmarshal(reply.([]byte), &equip)
	if err != nil {
		log.Error("player[%v],load equip unmarshal error:%v", dbId, err)
		return equip
	}
	return equip
}

// GetPlayerDestiny 获取玩家天命
func GetPlayerDestiny(dbId int64) *model.Destiny {
	destiny := new(model.Destiny)
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerDestiny, dbId))
	if err != nil {
		log.Error("player[%v],load destiny error:%v", dbId, err)
		return destiny
	}

	if reply == nil {
		return destiny
	}

	err = json.Unmarshal(reply.([]byte), &destiny)
	if err != nil {
		log.Error("player[%v],load destiny unmarshal error:%v", dbId, err)
		return destiny
	}
	return destiny
}
