package internal

import (
	"xfx/core/config"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
)

// 通告相关-抽卡-角色
func SyncNotice_DrawCardHero(ctx global.IPlayer, pl *model.Player, cardPoolType, typ, Id int32) {
	conf := config.Hero.All()[int64(Id)]
	//跑马灯
	confs := config.BroadCast.All()
	for _, v := range confs {
		if v.Type == define.HorseType_DrawCard {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.HorseType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.HorseType_Condition_Rate:
					_rate := params[index]
					if conf.Rate >= int32(_rate) {
						param = append(param, conf.Rate)
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_HeroId:
					if typ != 1 {
						pass = false
						break
					}
					_heroId := params[index]
					if Id == int32(_heroId) || int32(_heroId) == 0 {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_CardPoolType:
					_cardTyp := params[index]
					if cardPoolType == int32(_cardTyp) {
						param = append(param, cardPoolType)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			global.SyncHorse(ctx, pl, v, param)
		}
	}

	//聊天传闻
	confs_chuanwen := config.ChatChuanWen.All()
	for _, v := range confs_chuanwen {
		if v.Type == define.ChatChuanwenType_DrawCard {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.ChatChuanwenType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.ChatChuanwenType_Condition_Rate:
					_rate := params[index]
					if conf.Rate >= _rate {
						param = append(param, conf.Rate)
					} else {
						pass = false
					}
					break
				case define.ChatChuanwenType_Condition_HeroId:
					if typ != 1 {
						pass = false
						break
					}
					_heroId := params[index]
					if Id == _heroId || int32(_heroId) == 0 {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				case define.ChatChuanwenType_Condition_CardPool:
					_cardTyp := params[index]
					if cardPoolType == _cardTyp {
						param = append(param, cardPoolType)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}
			//发送
			SyncChatSend(ctx, pl, define.ChatTypeChuanwen, 0, "", param, v.Id, 0, nil)
		}
	}
}

// 通告相关-抽卡-宠物
func SyncNotice_DrawCardPet(ctx global.IPlayer, pl *model.Player, cardPoolType, typ, Id int32) {
	conf := config.Hero.All()[int64(Id)]
	//跑马灯
	confs := config.BroadCast.All()
	for _, v := range confs {
		if v.Type == define.HorseType_Pet {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.HorseType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.HorseType_Condition_Rate:
					_rate := params[index]
					if conf.Rate >= _rate {
						param = append(param, conf.Rate)
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_PetId:
					if typ != 6 {
						pass = false
						break
					}
					_heroId := params[index]
					if Id == int32(_heroId) || int32(_heroId) == 0 {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_CardPoolType:
					_cardTyp := params[index]
					if cardPoolType == _cardTyp {
						param = append(param, cardPoolType)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			global.SyncHorse(ctx, pl, v, param)
		}
	}

	//聊天传闻
	confs_chuanwen := config.ChatChuanWen.All()
	for _, v := range confs_chuanwen {
		if v.Type == define.ChatChuanwenType_DrawCardPet {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.ChatChuanwenType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.ChatChuanwenType_Condition_Rate:
					_rate := params[index]
					if conf.Rate >= _rate {
						param = append(param, conf.Rate)
					} else {
						pass = false
					}
					break
				case define.ChatChuanwenType_Condition_HeroId:
					if typ != 6 {
						pass = false
						break
					}
					_heroId := params[index]
					if Id == _heroId || int32(_heroId) == 0 {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				case define.ChatChuanwenType_Condition_CardPool:
					_cardTyp := params[index]
					if cardPoolType == _cardTyp {
						param = append(param, cardPoolType)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			SyncChatSend(ctx, pl, define.ChatTypeChuanwen, 0, "", param, v.Id, 0, nil)
		}
	}
}

// 通告相关-抽卡-鉴宝【藏品】
func SyncNotice_DrawCardGem(ctx global.IPlayer, pl *model.Player, cardPoolType, typ, Id int32) {
	conf := config.Hero.All()[int64(Id)]
	//跑马灯
	confs := config.BroadCast.All()
	for _, v := range confs {
		if v.Type == define.HorseType_Pet {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.HorseType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.HorseType_Condition_Rate:
					_rate := params[index]
					if conf.Rate >= _rate {
						param = append(param, conf.Rate)
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_CollectId:
					if typ != 5 {
						pass = false
						break
					}
					_heroId := params[index]
					if Id == int32(_heroId) || int32(_heroId) == 0 {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_CardPoolType:
					_cardTyp := params[index]
					if cardPoolType == _cardTyp {
						param = append(param, cardPoolType)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			global.SyncHorse(ctx, pl, v, param)
		}
	}
}

// 通告相关-章节变化
func SyncNotice_CharperChange(ctx global.IPlayer, pl *model.Player, Id int32) {
	//跑马灯
	confs := config.BroadCast.All()
	for _, v := range confs {
		if v.Type == define.HorseType_CharperChange {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.HorseType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.HorseType_Condition_CharperId:
					_charperId := params[index]
					if Id == _charperId {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			global.SyncHorse(ctx, pl, v, param)
		}
	}

	//聊天传闻
	confs_chuanwen := config.ChatChuanWen.All()
	for _, v := range confs_chuanwen {
		if v.Type == define.ChatChuanwenType_CharperChange {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.ChatChuanwenType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.ChatChuanwenType_Condition_CharperIndex:
					_charperId := params[index]
					if Id == _charperId {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			SyncChatSend(ctx, pl, define.ChatTypeChuanwen, 0, "", param, v.Id, 0, nil)
		}
	}
}

// 通告相关-商城购买
func SyncNotice_ShopBuy(ctx global.IPlayer, pl *model.Player, Id int32) {
	//跑马灯
	confs := config.BroadCast.All()
	for _, v := range confs {
		if v.Type == define.HorseType_ShopBuy {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.HorseType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.HorseType_Condition_ShopId:
					_shopId := params[index]
					if Id == _shopId {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			global.SyncHorse(ctx, pl, v, param)
		}
	}
}

// 通告相关-排名更新
func SyncNotice_RankUpdate(ctx global.IPlayer, pl *model.Player, rankType int, rankIndex int64) {
	//跑马灯
	confs := config.BroadCast.All()
	for _, v := range confs {
		if v.Type == define.HorseType_RankUpdate {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.HorseType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.HorseType_Condition_RankIndex:
					_index := params[index]
					if int64(_index) >= rankIndex {
						param = append(param, int32(rankIndex))
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_RankType:
					_type := params[index]
					if _type == int32(rankType) {
						param = append(param, int32(rankType))
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			global.SyncHorse(ctx, pl, v, param)
		}
	}
}

// 通告相关-获得功法
func SyncNotice_AddMagic(ctx global.IPlayer, pl *model.Player, Id int32) {
	//聊天传闻
	confs_chuanwen := config.ChatChuanWen.All()
	conf := config.HeroMagic.All()[int64(Id)]
	for _, v := range confs_chuanwen {
		if v.Type == define.ChatChuanwenType_AddMagic {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.ChatChuanwenType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.ChatChuanwenType_Condition_Rate:
					_rate := params[index]
					if conf.Rate >= int32(_rate) {
						param = append(param, conf.Rate)
					} else {
						pass = false
					}
					break
				case define.ChatChuanwenType_Condition_MagicId:
					_magicId := params[index]
					if Id == int32(_magicId) || int32(_magicId) == 0 {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			SyncChatSend(ctx, pl, define.ChatTypeChuanwen, 0, "", param, v.Id, 0, nil)
		}
	}
}

// 通告相关-获得坐骑
func SyncNotice_AddMount(ctx global.IPlayer, pl *model.Player, Id int32) {
	//聊天传闻
	confs_chuanwen := config.ChatChuanWen.All()
	conf := config.Mount.All()[int64(Id)]
	for _, v := range confs_chuanwen {
		if v.Type == define.ChatChuanwenType_AddMagic {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.ChatChuanwenType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.ChatChuanwenType_Condition_Rate:
					_rate := params[index]
					if conf.Rate >= int32(_rate) {
						param = append(param, conf.Rate)
					} else {
						pass = false
					}
					break
				case define.ChatChuanwenType_Condition_MountId:
					_mountId := params[index]
					if Id == int32(_mountId) || int32(_mountId) == 0 {
						param = append(param, Id)
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			SyncChatSend(ctx, pl, define.ChatTypeChuanwen, 0, "", param, v.Id, 0, nil)
		}
	}
}
