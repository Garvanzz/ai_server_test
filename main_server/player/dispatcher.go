package player

import (
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/messages"
	"xfx/main_server/player/activity"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/battle"
	"xfx/main_server/player/cdkey"
	"xfx/main_server/player/chat"
	"xfx/main_server/player/collection"
	"xfx/main_server/player/danaotiangong"
	"xfx/main_server/player/destiny"
	"xfx/main_server/player/divine"
	"xfx/main_server/player/draw"
	"xfx/main_server/player/equip"
	"xfx/main_server/player/fashion"
	"xfx/main_server/player/friend"
	"xfx/main_server/player/gemappraisal"
	"xfx/main_server/player/guild"
	"xfx/main_server/player/handbook"
	"xfx/main_server/player/hero"
	"xfx/main_server/player/paradise"
	"xfx/main_server/player/idle_box"
	"xfx/main_server/player/lineup"
	"xfx/main_server/player/login"
	"xfx/main_server/player/magic"
	"xfx/main_server/player/mail"
	"xfx/main_server/player/mission"
	"xfx/main_server/player/openbox"
	"xfx/main_server/player/pet"
	"xfx/main_server/player/playerprop"
	"xfx/main_server/player/rank"
	"xfx/main_server/player/room"
	"xfx/main_server/player/shenjidraw"
	"xfx/main_server/player/shop"
	"xfx/main_server/player/skill"
	"xfx/main_server/player/stage"
	"xfx/main_server/player/task"
	"xfx/main_server/player/transaction"
	"xfx/main_server/player/welfare"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_chat"
	"xfx/proto/proto_danaotiangong"
	"xfx/proto/proto_destiny"
	"xfx/proto/proto_draw"
	"xfx/proto/proto_equip"
	"xfx/proto/proto_fashion"
	"xfx/proto/proto_friend"
	"xfx/proto/proto_game"
	"xfx/proto/proto_guild"
	"xfx/proto/proto_handbook"
	"xfx/proto/proto_hero"
	"xfx/proto/proto_huaguoshan"
	"xfx/proto/proto_idlebox"
	"xfx/proto/proto_item"
	"xfx/proto/proto_lineup"
	"xfx/proto/proto_magic"
	"xfx/proto/proto_mail"
	"xfx/proto/proto_mission"
	"xfx/proto/proto_openbox"
	"xfx/proto/proto_pet"
	"xfx/proto/proto_player"
	"xfx/proto/proto_rank"
	"xfx/proto/proto_room"
	"xfx/proto/proto_shop"
	"xfx/proto/proto_skill"
	"xfx/proto/proto_stage"
	"xfx/proto/proto_task"
	"xfx/proto/proto_transaction"
	"xfx/proto/proto_welfare"
)

func dispatch(ctx global.IPlayer, pl *model.Player, _msg any) any {
	switch msg := _msg.(type) {
	case *messages.LoginSuccess: // 登录成功回调
		login.Login(ctx, pl)
	case *messages.LoginReplace: // 顶号
		return login.Replace(ctx, pl, msg.Session)
	case *messages.Logout: // 登出回调
		room.Logout(ctx, pl)
		login.Logout(ctx, pl)
	case *messages.Disconnect: // 断开连接
		room.Logout(ctx, pl)
		login.Logout(ctx, pl)
	case *messages.DispatchMessage: // 返回客户端消息
		ctx.Send(msg.Content)
	case *db.RedisRet:
		OnRet(ctx, pl, msg)
	//匹配
	case *proto_room.C2SRoomMatchGame:
		room.StartGameMatch(ctx, pl)
	case *proto_room.C2SMatchCancel:
		room.CancelGameMatch(ctx, pl)

	//关卡
	case *proto_stage.C2SInitStage:
		stage.ReqStageList(ctx, pl, msg)
	case *proto_stage.C2SKillEnemy:
		stage.ReqStageKillEnemy(ctx, pl, msg)
	case *messages.StageSettle:
		stage.SettleStageGame(ctx, pl, msg)
	case *proto_stage.C2SChallengeStageBossBattle:
		stage.ReqStageBossBattleChallenge(ctx, pl, msg)
	case *proto_stage.C2SGetFreePlayer:
		stage.GetStageFreePlayer(ctx, pl, msg)
	case *proto_stage.C2SUnLockHiddenStory:
		stage.UnlockHiddenStory(ctx, pl, msg)

	// Room
	case *proto_room.C2SCreateRoom: //创建房间
		room.CreateRoom(ctx, pl, msg)
	case *proto_room.C2SJoinRoom: //加入房间
		room.JoinRoom(ctx, pl, msg)
	case *proto_room.C2SExitRoom: //退出房间
		room.ExitRoom(ctx, pl, msg)
	case *proto_room.C2SDissolveRoom: //解散房间
		room.DissolveRoom(ctx, pl, msg)
	case *proto_room.C2SStartGame: //开始游戏
		room.StartGame(ctx, pl, msg)
	case *proto_room.C2SPermitLook: //允许观战
		room.PermitLook(ctx, pl, msg)
	case *proto_room.C2SGameReady: //准备
		room.ReadyGame(ctx, pl, msg)
	case *proto_room.C2SGameReadyCancel: //取消准备
		room.CancelReadyGame(ctx, pl, msg)
	case *proto_room.C2SRoomLineUp: //上阵
		room.LineUpGame(ctx, pl, msg)
	case *proto_room.C2SFindRoom: //搜索房间
		room.FindRoom(ctx, pl, msg)
	case *proto_room.C2SRangeJoinRoom: //随机加入房间
		room.RangleJoinRoom(ctx, pl, msg)
	case *proto_room.C2SRefreshRoom: //刷新房间
		room.RefreshRoomList(ctx, pl, msg)
	case *proto_room.C2SChangePassword: //修改密码
		room.ChangePassword(ctx, pl, msg)
	case *proto_room.C2SRoomOpen: //公开
		room.SetOpen(ctx, pl, msg)
	case *proto_room.C2SSetGroup: //设置阵营
		room.SetGroup(ctx, pl, msg)
	case *proto_room.C2SGetInviteList: //获取邀请列表
		room.GetInviteList(ctx, pl, msg)
	case *proto_room.C2SRoomInvite: //邀请
		room.RoomInvite(ctx, pl, msg)
	case *proto_room.C2SInviteBack: //邀请反馈
		room.RoomInviteBack(ctx, pl, msg)

	//好友
	case *proto_friend.C2SReqFriendList: //好友列表
		friend.ReqFriendList(ctx, pl)
	case *proto_friend.C2SreqFriendApplyList: //请求好友申请列表
		friend.ReqFriendApplyList(ctx, pl)
	case *proto_friend.C2SDeleteFriend: //删除好友
		friend.ReqRemoveFriend(ctx, pl, msg)
	case *proto_friend.C2SreqDealFriendApply: //处理申请
		friend.ReqDealFriendApply(ctx, pl, msg)
	case *proto_friend.C2SRequestAddFriend: //添加好友
		friend.ReqAddFriend(ctx, pl, msg)
	case *proto_friend.C2SReqFriendGift: //好友赠送
		friend.ReqFriendGift(ctx, pl, msg)
	case *proto_friend.C2SReqGetFriendGift: //好友赠送领取
		friend.ReqGetFriendGift(ctx, pl, msg)
	case *proto_friend.C2SOneKeyFriendGift: //一键领取和赠送
		friend.ReqOneKeyFriendGift(ctx, pl, msg)
	case *proto_friend.C2SReqBlockFriend: //黑名单
		friend.ReqBlockFriend(ctx, pl, msg)
	case *proto_friend.C2SReqInitBlockFriend: //黑名单列表
		friend.ReqBlockFriendList(ctx, pl)
	case *proto_friend.C2SReqUnlockBlockFriend: //解除黑名单
		friend.ReqUnLockBlockFriend(ctx, pl, msg)
	case *proto_friend.C2SReqTuijianFriend: //推荐好友
		friend.ReqTuijianFriend(ctx, pl, msg)
	case *proto_friend.C2SReqRefreshFriend: //刷新推荐好友
		friend.ReqRefreshTuijianFriend(ctx, pl, msg)
	case *proto_friend.C2SFindFriend: //查找好友
		friend.ReqFindFriend(ctx, pl, msg)

	// 邮件
	case *proto_mail.C2SMailList: // 获取邮件列表
		mail.ReqMailList(ctx, pl, msg)
	case *proto_mail.C2SOpenMail: // 打开邮件
		mail.ReqOpenMail(ctx, pl, msg)
	case *proto_mail.C2SDelMail: // 删除邮件
		mail.ReqDelMail(ctx, pl, msg)
	case *proto_mail.C2SDelAllMail: // 删除所有邮件
		mail.ReqDelAllMails(ctx, pl, msg)
	case *proto_mail.C2SCollectMailItem: // 收取邮件道具
		mail.ReqCollectMailItem(ctx, pl, msg)
	case *proto_mail.C2SCollectAllMailItems: // 一键读取
		mail.ReqCollectAllMailItems(ctx, pl, msg)

	// 背包
	case *proto_item.C2SBag: // 请求背包数据
		bag.ReqBag(ctx, pl, msg)
	case *proto_item.C2SUseItem: // 使用道具
		bag.ReqUseItem(ctx, pl, msg)
	case *proto_item.C2SCompositionItem: // 合成道具
		bag.ReqCompositionItem(ctx, pl, msg)
	case *proto_item.C2SSellItem: // 售卖道具
		bag.ReqSellItem(ctx, pl, msg)

	// 兑换码
	case *proto_item.C2SExchangeCDKey: // 兑换码兑换
		cdkey.ReqExchangeCDKey(ctx, pl, msg)
	case *proto_item.C2SInitCDKey: // 初始化兑换码数据
		cdkey.ReqInitCDKey(ctx, pl, msg)

		//开箱子
	case *proto_openbox.C2SInitOpenBox: // 初始开箱子
		openbox.ReqInitOpenBox(ctx, pl, msg)
	case *proto_openbox.C2SOpenBox: // 开箱子
		openbox.ReqOpenBox(ctx, pl, msg)
	case *proto_openbox.C2SUpBoxLevel: // 升级等级
		openbox.ReqUpLevelBox(ctx, pl, msg)
	case *proto_openbox.C2SScoreBox: // 积分换宝箱
		openbox.ReqSocreBuyBox(ctx, pl, msg)

	//装备
	case *proto_equip.C2SInitEquip:
		equip.ReqInitEquip(ctx, pl, msg)
	case *proto_equip.C2SWearEquip:
		equip.ReqWearEquip(ctx, pl, msg)
	case *proto_equip.C2SSellEquip:
		equip.ReqSellEquip(ctx, pl, msg)

	//坐骑
	case *proto_equip.C2SInitMount:
		equip.ReqInitMount(ctx, pl, msg)
	case *proto_equip.C2SUpLevelMount:
		equip.ReqLevelUpMount(ctx, pl, msg)
	case *proto_equip.C2SUseMount:
		equip.ReqUseMount(ctx, pl, msg)
	case *proto_equip.C2SChangeMountName:
		equip.ReqMountChangeName(ctx, pl, msg)
	case *proto_equip.C2SUpEnergyLevel:
		equip.ReqMountUpEnergy(ctx, pl, msg)
	case *proto_equip.C2SUpLevelMountItem:
		equip.ReqLevelUpMountItem(ctx, pl, msg)
	case *proto_equip.C2SGetMountHandBookAward:
		equip.MountHandbookAward(ctx, pl, msg)

		//神兵
	case *proto_equip.C2SInitWeaponry:
		equip.ReqInitWeaponry(ctx, pl, msg)
	case *proto_equip.C2SUpLevelWeaponry:
		equip.ReqLevelUpWeaponry(ctx, pl, msg)
	case *proto_equip.C2SUseWeaponry:
		equip.ReqUseWeaponry(ctx, pl, msg)
	case *proto_equip.C2SUpLevelWeaponryItem:
		equip.ReqLevelUpWeaponryItem(ctx, pl, msg)
	case *proto_equip.C2SGetWeaponryHandBookAward:
		equip.WeaponHandbookAward(ctx, pl, msg)

	//附魔
	case *proto_equip.C2SInitEquipEnchant:
		equip.ReqInitEquipEnchant(ctx, pl, msg)
	case *proto_equip.C2SUseEnchant:
		equip.ReqUseEnchant(ctx, pl, msg)
	case *proto_equip.C2SEnchantLevel:
		equip.ReqUpLevelEnchant(ctx, pl, msg)
	case *proto_equip.C2SOneKeyEnchantLevel:
		equip.ReqOneKeyUpLevelEnchant(ctx, pl, msg)

	//洗练
	case *proto_equip.C2SInitEquipSuccinct:
		equip.ReqInitEquipSuccinct(ctx, pl, msg)
	case *proto_equip.C2SSuccinctSkill:
		equip.ReqEquipSuccinctSkill(ctx, pl, msg)
	case *proto_equip.C2SUseSuccinct:
		equip.ReqUseSuccinct(ctx, pl, msg)
	case *proto_equip.C2SDeleteSuccinct:
		equip.ReqDeleteSuccinct(ctx, pl, msg)
	case *proto_equip.C2SUseSuccinctScheme:
		equip.ReqCutSuccinct(ctx, pl, msg)
	case *proto_equip.C2SSuccincChangeName:
		equip.ReqSuccinctChangeName(ctx, pl, msg)
	case *proto_equip.C2SGetSuccinctLevelAward:
		equip.ReqGetSuccinctAward(ctx, pl, msg)

	//背饰
	case *proto_equip.C2SInitBraces:
		equip.ReqInitEquipBrace(ctx, pl, msg)
	case *proto_equip.C2SLevelUpAura:
		equip.ReqEquipBraceAuraUpLevel(ctx, pl, msg)
	case *proto_equip.C2SGetAuraStageAward:
		equip.ReqEquipBraceAuraStageAward(ctx, pl, msg)
	case *proto_equip.C2SLevelUpBrace:
		equip.ReqEquipBraceUpLevel(ctx, pl, msg)
	case *proto_equip.C2SUseBrace:
		equip.ReqEquipBraceUse(ctx, pl, msg)
	case *proto_equip.C2SBraceIndexChangeName:
		equip.ReqEquipBraceChangeName(ctx, pl, msg)
	case *proto_equip.C2SLevelUpTalent:
		equip.ReqEquipBraceTalentUpLevel(ctx, pl, msg)
	case *proto_equip.C2STransformBraceIndex:
		equip.ReqTransformBraceIndex(ctx, pl, msg)
	case *proto_equip.C2SResetBraceTalent:
		equip.ReqEquipBraceTalentReSet(ctx, pl, msg)
	case *proto_equip.C2SGetBraceHandBookAward:
		equip.BraceHandbookAward(ctx, pl, msg)

	//天命
	case *proto_destiny.C2SReqDestinyInit:
		destiny.ReqInitDestiny(ctx, pl, msg)
	case *proto_destiny.C2SReqUnLockDestiny:
		destiny.ReqUnLockDestiny(ctx, pl, msg)
	case *proto_destiny.C2SReqUnlockSelfDestiny:
		destiny.ReqUnLockSelfDestiny(ctx, pl, msg)
	case *proto_destiny.C2SReqOneKeyUnLock:
		destiny.ReqOneKeyUnLockDestiny(ctx, pl, msg)

	//领悟
	case *proto_equip.C2SInitDivine:
		divine.ReqInitDivine(ctx, pl, msg)
	case *proto_equip.C2SReqDivineUpLevel:
		divine.ReqDivineLevelUp(ctx, pl, msg)
	case *proto_equip.C2SReqUnLockDivine:
		divine.ReqDivineUnLock(ctx, pl, msg)
	case *proto_equip.C2SReqWearDivine:
		divine.ReqDivineWear(ctx, pl, msg)
	case *proto_equip.C2SReqRemoveDivine:
		divine.ReqDivineRemove(ctx, pl, msg)
	case *proto_equip.C2SReqDivineCompose:
		divine.ReqLearnCompose(ctx, pl, msg)
	case *proto_equip.C2SReqOnekeyDivineCompose:
		divine.ReqOneKeyLearnCompose(ctx, pl, msg)
	case *proto_equip.C2SReqResetDivine:
		divine.ReqDivineReset(ctx, pl, msg)
	case *proto_equip.C2SOneKeyUpLevelDivine:
		divine.ReqOneKeyDivineLevelUp(ctx, pl, msg)

	//神机
	case *proto_destiny.C2SShenjiBoxInit:
		shenjidraw.ReqInitShenjiDraw(ctx, pl, msg)
	case *proto_destiny.C2SDrawShenji:
		shenjidraw.ReqShenjiDraw(ctx, pl, msg)
	case *proto_destiny.C2SReqGetDrawRecord:
		shenjidraw.ReqShenjiDrawRecord(ctx, pl, msg)

	//藏品
	case *proto_equip.C2SInitCollection:
		collection.ReqInitCollection(ctx, pl, msg)
	case *proto_equip.C2SCollectionUpStar:
		collection.ReqUpStarCollection(ctx, pl, msg)
	case *proto_equip.C2SCollectionWear:
		collection.ReqWearCollection(ctx, pl, msg)
	case *proto_equip.C2SCollectionRemove:
		collection.ReqRemoveCollection(ctx, pl, msg)
	case *proto_equip.C2SSetCollectionSlotHero:
		collection.ReqSetSlotHeroCollection(ctx, pl, msg)

	//鉴宝
	case *proto_draw.C2SGemAppraisalInit:
		gemappraisal.ReqInitGemAppraisal(ctx, pl, msg)
	case *proto_draw.C2SDrawGemAppraisal:
		gemappraisal.ReqDrawGemAppraisal(ctx, pl, msg)
	case *proto_draw.C2SGetGemAppraisalStageAward:
		gemappraisal.ReqGemAppraisalStageAward(ctx, pl, msg)

	//角色
	case *proto_hero.C2SInitHero:
		hero.ReqInitHero(ctx, pl, msg)
	case *proto_hero.C2SInitSkin:
		hero.ReqInitSkin(ctx, pl, msg)
	case *proto_hero.C2SHeroUpLevel:
		hero.ReqHeroUpLevel(ctx, pl, msg)
	case *proto_hero.C2SHeroUpStar:
		hero.ReqHeroUpStar(ctx, pl, msg)
	case *proto_hero.C2SHeroUpStage:
		hero.ReqHeroUpStage(ctx, pl, msg)
	case *proto_hero.C2SHeroUpCultivation:
		hero.ReqHeroUpCultivation(ctx, pl, msg)
	case *proto_hero.C2SReSetHero:
		hero.ReqReSetHero(ctx, pl, msg)
	case *proto_player.C2SPlayerPowerChange:
		ReqPlayerChangePower(ctx, pl, msg)

		//布阵
	case *proto_lineup.C2SInitLineUp:
		lineup.ReqInitLineUp(ctx, pl, msg)
	case *proto_lineup.C2SSetLineUp:
		lineup.ReqSetLineUp(ctx, pl, msg)
	case *proto_lineup.C2SSetLineUpAndReplace:
		lineup.ReqReplaceSetLineUp(ctx, pl, msg)

	//技能
	case *proto_skill.C2SInitSkill:
		skill.ReqSkillList(ctx, pl, msg)

	//法术
	case *proto_magic.C2SInitMagic:
		magic.ReqMagicInit(ctx, pl, msg)
	case *proto_magic.C2SUpLevelMagic:
		magic.ReqMagicUpLevel(ctx, pl, msg)
	case *proto_magic.C2SOneKeyUplevel:
		magic.ReqMagicOneKeyUpLevel(ctx, pl, msg)
	case *proto_magic.C2SWearMagic:
		magic.ReqMagicWear(ctx, pl, msg)
	case *proto_magic.C2SOneKeyWear:
		magic.ReqMagicOneKeyWear(ctx, pl, msg)
	case *proto_magic.C2SXieXiaMagic:
		magic.ReqMagicXiexia(ctx, pl, msg)

	//宠物
	case *proto_pet.C2SInitPet:
		pet.ReqInitPet(ctx, pl, msg)
	case *proto_pet.C2SUpLevelPet:
		pet.ReqPetLevelUp(ctx, pl, msg)
	case *proto_pet.C2SUpStagePet:
		pet.ReqPetStageUp(ctx, pl, msg)
	case *proto_pet.C2SUpStarPet:
		pet.ReqPetStarUp(ctx, pl, msg)
	case *proto_pet.C2SResetPet:
		pet.ReqPetReset(ctx, pl, msg)
	case *proto_pet.C2SPetDrawInit:
		pet.ReqInitPetDraw(ctx, pl, msg)
	case *proto_pet.C2SPetDrawCard:
		pet.ReqDrawPet(ctx, pl, msg)
	case *proto_pet.C2SGetPetDrawCardStageAward:
		pet.ReqpetStageAward(ctx, pl, msg)
	case *proto_pet.C2SGetPetRecord:
		pet.ReqPetRecorde(ctx, pl, msg)
	case *proto_pet.C2SPetCall:
		pet.ReqPetCall(ctx, pl, msg)
	case *proto_pet.C2SDispatchPet:
		pet.ReqPetDispatchPet(ctx, pl, msg)
	case *proto_pet.C2SGetPetHandBookExp:
		pet.ReqPetHandBookGetExp(ctx, pl, msg)
	case *proto_pet.C2SGetPetHandBookAward:
		pet.ReqPetHandBookAward(ctx, pl, msg)
	case *proto_pet.C2SUnderstandGift:
		pet.ReqPetUnderstandGift(ctx, pl, msg)
	case *proto_pet.C2SPointUnderstandGift:
		pet.ReqPetUnderstandPointGift(ctx, pl, msg)
	case *proto_pet.C2SPetXilianGift:
		pet.ReqPetXilianGift(ctx, pl, msg)
	case *proto_pet.C2SPetSureXilianGift:
		pet.ReqPetSureXilianGift(ctx, pl, msg)
	case *proto_pet.C2SUnderstandSkill:
		pet.ReqPetUnderstandSkill(ctx, pl, msg)
	case *proto_pet.C2SPointUnderstandSkill:
		pet.ReqPetUnderstandPointSkill(ctx, pl, msg)
	case *proto_pet.C2SForgetSkill:
		pet.ReqPetRemovePointSkill(ctx, pl, msg)
	case *proto_pet.C2SPetForge:
		pet.ReqPetEquipStrengthen(ctx, pl, msg)
	case *proto_pet.C2SPetBreakdown:
		pet.ReqPetEquipBreakdown(ctx, pl, msg)
	case *proto_pet.C2SPetWearEquip:
		pet.ReqPetEquipWear(ctx, pl, msg)
	case *proto_pet.C2SCatchPetBattle:
		pet.ReqPetCatch(ctx, pl, msg)

	//招募
	case *proto_draw.C2SDrawInit:
		drawhero.ReqInitDraw(ctx, pl, msg)
	case *proto_draw.C2SDrawCard:
		drawhero.ReqDrawCard(ctx, pl, msg)
	case *proto_draw.C2SGetDrawCardLevelAward:
		drawhero.ReqHeroDrawLevelAward(ctx, pl, msg)
	case *proto_draw.C2SGetDrawCardStageAward:
		drawhero.ReqHeroDrawStageAward(ctx, pl, msg)

	//商城
	case *proto_shop.C2SShopData: //请求数据
		shop.ReqShopData(ctx, pl, msg)
	case *proto_shop.C2SBuyShop: //购买
		shop.ReqShopBuyData(ctx, pl, msg)
	case *proto_shop.C2SGMBuyShop: //GM购买
		shop.ReqGMShopBuyData(ctx, pl, msg)
	case *proto_shop.C2SReqRechargeBackAward: //充值领取奖励
		shop.ReqBackShopBuyAward(ctx, pl, msg)

	//聊天
	case *proto_chat.C2SChatInfo:
		chat.ReqChatInfo(ctx, pl, msg)
	case *proto_chat.C2SSendChatMsg:
		chat.ReqSendChat(ctx, pl, msg)
	//私聊
	case *proto_chat.C2SGetPrivateChatData:
		chat.ReqPrivateChatData(ctx, pl, msg)
	case *proto_chat.C2SSendPrivateChat:
		chat.ReqSendPrivateChatData(ctx, pl, msg)

	//图鉴
	case *proto_handbook.C2SHandBookData:
		handbook.ReqHandBookInfo(ctx, pl, msg)
	case *proto_handbook.C2SGetHandBookExp:
		handbook.ReqHandBookGetExp(ctx, pl, msg)
	case *proto_handbook.C2SGetHandBookAward:
		handbook.ReqHandBookAward(ctx, pl, msg)

		//福利
		//签到
	case *proto_welfare.C2SDaySign:
		welfare.ReqDaySignInit(ctx, pl, msg)
	case *proto_welfare.C2SDayAward:
		welfare.ReqSignAward(ctx, pl, msg)
	case *proto_welfare.C2SMonthCardInit:
		welfare.ReqMonthCardInit(ctx, pl, msg)
	case *proto_welfare.C2SGetMonthCard:
		welfare.ReqMonthCardGetAward(ctx, pl, msg)
	case *proto_welfare.C2SFunctionOpenInit:
		welfare.ReqFuncOpenInit(ctx, pl, msg)
	case *proto_welfare.C2SFunctionAward:
		welfare.ReqFuncOpenAward(ctx, pl, msg)

	//人物
	case *proto_player.C2SChangeName:
		ReqChangePlayerName(ctx, pl, msg)
	case *proto_player.C2SChangeTitle: //改称号
		ReqChangeTitle(ctx, pl, msg)
	case *proto_player.C2SChangeHead: //改头像
		ReqChangeHead(ctx, pl, msg)
	case *proto_player.C2SChangeHeadFrame: //改头像框
		ReqChangeHeadFrame(ctx, pl, msg)
	case *proto_player.C2STransformJob: //改职业
		ReqTransformJob(ctx, pl, msg)
	case *proto_player.C2SChangeSex: //改性别
		ReqChangeSex(ctx, pl, msg)
	case *proto_player.C2SChangeBubble: //改泡泡
		ReqChangeBubble(ctx, pl, msg)
	case *proto_player.C2SGetPlayerById:
		ReqGetPlayerInfoById(ctx, pl, msg) //获取个人信息
	//人物属性道具
	case *proto_player.C2SGetPlayerProp:
		playerprop.ReqInitPlayerProp(ctx, pl, msg)

	// 帮会
	case *proto_guild.C2SSetGuildRule: // 设置帮会信息
		guild.ReqSetGuildRule(ctx, pl, msg)
	case *proto_guild.C2SImpeachMaster: // 弹劾会长
		guild.ReqImpeachMaster(ctx, pl, msg)
	case *proto_guild.C2SLeaveGuild: // 离开帮会
		guild.ReqLeaveGuild(ctx, pl, msg)
	case *proto_guild.C2SKickOutMember: // 帮会踢人
		guild.ReqKickOutMember(ctx, pl, msg)
	case *proto_guild.C2SDealApply: // 处理请求
		guild.ReqDealApply(ctx, pl, msg)
	case *proto_guild.C2SAssignPosition: // 任命职位
		guild.ReqAssignPosition(ctx, pl, msg)
	case *proto_guild.C2SCreateGuild: // 创建帮会
		guild.ReqCreateGuild(ctx, pl, msg)
	case *proto_guild.C2SGuildEvent: // 获取帮会日志
		guild.ReqGuildEvents(ctx, pl, msg)
	case *proto_guild.C2SJoinGuild: // 加入帮会
		guild.ReqJoinGuild(ctx, pl, msg)
	case *proto_guild.C2SGetMemberList: // 请求帮会成员列表
		guild.ReqMemberList(ctx, pl, msg)
	case *proto_guild.C2SGetApply: // 获取帮会申请列表
		guild.ReqGuildApplyList(ctx, pl, msg)
	case *proto_guild.C2SSearchByName: // 根据名字搜索帮会
		guild.ReqSearchGuildByName(ctx, pl, msg)
	case *proto_guild.C2SGuildByPage: // 根据页数获取帮会信息
		guild.ReqGuildListByPage(ctx, pl, msg)
	case *proto_guild.C2SPlayerGuildDetail: // 获取玩家帮会数据
		guild.ReqPlayerGuildDetail(ctx, pl, msg)
	case *proto_guild.C2SDissolveGuild: // 解散帮会
		guild.ReqDissolveGuild(ctx, pl, msg)
	case *proto_guild.C2SGetGuildInfo: // 获取帮会数据
		guild.ReqGuildData(ctx, pl, msg)
	case *proto_guild.C2SChangeGuildName: //帮会改名
		guild.ReqGuildChangeName(ctx, pl, msg)
	case *proto_guild.C2SGuildSign: //帮会签到
		guild.ReqGuildSign(ctx, pl, msg)
	case *proto_guild.C2SBuildGuildMap: //帮会建造
		guild.ReqGuildBuild(ctx, pl, msg)
	case *proto_guild.C2SGetBuildMapInfo: //获取帮会信息
		guild.ReqGuildBuildMapInfo(ctx, pl, msg)
	case *proto_guild.C2SGuildPray:
		guild.ReqGuildPray(ctx, pl, msg) //请求祈福

	//帮会元池
	case *proto_guild.C2SInitYuanchi:
		guild.ReqGuildYuanchiInit(ctx, pl, msg)
	case *proto_guild.C2SReqRecord:
		guild.ReqGuildRefiningLog(ctx, pl, msg)

	// 任务
	case *proto_task.C2SGetTasks: // 请求任务数据
		task.ReqTaskData(ctx, pl, msg)
	case *proto_task.C2SGetReward: // 领奖任务奖励
		task.ReqReceiveReward(ctx, pl, msg)
	case *proto_task.C2SGetActivePointReward: // 领奖活跃点奖励
		task.ReqReceiveActivePointReward(ctx, pl, msg)

		// 挂机宝箱
	case *proto_idlebox.C2SAddTime: // 请求挂机宝箱加时
		idle_box.ReqAddTime(ctx, pl, msg)
	case *proto_idlebox.C2SGetIdleBoxData: // 获取挂机宝箱数据
		idle_box.ReqGetIdleBoxData(ctx, pl, msg)
	case *proto_idlebox.C2SReceiveAward: // 领奖奖励
		idle_box.ReqReceiveReward(ctx, pl, msg)

	// 活动
	case *proto_activity.C2SActivityStatus: // 获取活动状态列表
		activity.ReqActivityStatus(ctx, pl, msg)
	case *proto_activity.C2SActivityData: // 获取活动数据
		activity.ReqActivityData(ctx, pl, msg)
	case *proto_activity.C2SActivityDataList: // 获取多个活动数据
		activity.ReqActivityDataList(ctx, pl, msg)
	case *proto_activity.C2SActivityAward: // 领取活动奖励
		activity.ReqGetActivityAward(ctx, pl, msg)
	case *proto_activity.C2SActivityBuy: // 活动购买
		activity.ReqActivityBuy(ctx, pl, msg)
	case *proto_activity.C2STheCompetitionChooseGroup: //巅峰对决选择阵营
		activity.ReqActivityTheCompetitionChooseGroupId(ctx, pl, msg)
	case *proto_activity.C2STheCompetitionStake: //巅峰对决押注
		activity.ReqActivityTheCompetitionStake(ctx, pl, msg)
	case *proto_activity.C2SGetFundAward:
		activity.ReqGetActivityFundAward(ctx, pl, msg)
	case *proto_activity.C2SArenaRefreshBattlePlayer: //竞技场刷新对手
		activity.ReqActivityArenaRefreshBattlePlayer(ctx, pl, msg)
	case *proto_activity.C2SArenaSetLineUp: //竞技场布阵阵容
		activity.ReqActivityTheArenaSetLineUp(ctx, pl, msg)
	case *proto_activity.C2SArenaGetPlayerLineUp: //竞技场获取布阵
		activity.ReqActivityArenaGetPlayerLineUp(ctx, pl, msg)
	case *proto_activity.C2SArenaReqRecord:
		activity.ReqActivityArenaBattleRecord(ctx, pl, msg)
	case *proto_activity.C2SArenaBattle:
		activity.ReqActivityArenaBattle(ctx, pl, msg)
	case *proto_activity.C2SLadderRaceSetLineUp: //天梯阵容
		activity.ReqActivityLadderRaceSetLineUp(ctx, pl, msg)
	case *proto_activity.C2SLadderRaceGetPlayerLineUp: //天梯获取玩家阵容
		activity.ReqActivityLadderRaceGetPlayerLineUp(ctx, pl, msg)
	case *proto_activity.C2SLadderRaceReqRecord:
		activity.ReqActivityTiantiBattleRecord(ctx, pl, msg)
	case *proto_activity.C2SLadderRaceBattle:
		activity.ReqActivityLadderRaceBattle(ctx, pl, msg)
	case *proto_activity.C2SGoFish: // 钓鱼
		activity.ReqActivityGoFish(ctx, pl, msg)
	case *proto_activity.C2SFishSign: // 钓鱼签到
		activity.ReqActivityFishSign(ctx, pl, msg)
	case *proto_activity.C2SFishLevelAward: // 钓鱼等级
		activity.ReqActivityFishLevelAward(ctx, pl, msg)
	case *proto_activity.C2SPassportGetAward: // 通行证领奖
		activity.ReqPassportGetAward(ctx, pl, msg)

		//系统-大闹天宫
	case *proto_danaotiangong.C2SReqTiangongData:
		danaotiangong.ReqInitDanaotiangong(ctx, pl, msg)
	case *proto_danaotiangong.C2SChallengeDanaoTiangongBattle:
		danaotiangong.ReqDntgBattleChallenge(ctx, pl, msg)
	case *proto_danaotiangong.C2SReqTiangongRecord:
		danaotiangong.ReqDntgBattleRecord(ctx, pl, msg)

	//战斗
	case *proto_game.C2SChallengeBattleReport:
		battle.ReqChallengeBattleReport(ctx, pl, msg)
	case *proto_game.C2SChallengePlayerBattle:
		battle.ReqChallengePlayerBattle(ctx, pl, msg)

	// 排行榜
	case *proto_rank.C2SRankData: // 请求数据
		rank.ReqRankingData(ctx, pl, msg)

	//副本
	case *proto_mission.C2SInitMissionStageData:
		mission.ReqInitMission(ctx, pl, msg)
	case *proto_mission.C2SChallengeMissionBattle:
		mission.ReqMissionBattleChallenge(ctx, pl, msg)

	//时装
	case *proto_fashion.C2SInitFashion:
		fashion.ReqInitFashion(ctx, pl, msg)
	case *proto_fashion.C2SUseFashion:
		fashion.ReqUseFashion(ctx, pl, msg)
	case *proto_fashion.C2SGetFashionHandBookAward:
		fashion.FashionHandbookAward(ctx, pl, msg)

		//头饰
	case *proto_fashion.C2SInitHeadWear:
		fashion.ReqInitHeadWear(ctx, pl, msg)
	case *proto_fashion.C2SUseHeadWear:
		fashion.ReqUseHeadWear(ctx, pl, msg)
	case *proto_fashion.C2SGetHeadWearHandBookAward:
		fashion.HeadWearHandbookAward(ctx, pl, msg)

	// 交易所
	case *proto_transaction.C2SGetTransaction:
		transaction.ReqGetTransaction(ctx, pl, msg)
	case *proto_transaction.C2SGetTransactionList:
		transaction.ReqGetTransactionList(ctx, pl, msg)
	case *proto_transaction.C2SSendTransaction:
		transaction.ReqSendTransaction(ctx, pl, msg)
	case *proto_transaction.C2SLogicTransaction:
		transaction.ReqLogicTransaction(ctx, pl, msg)
	case *proto_transaction.C2STransactionRecord:
		transaction.ReqTransactionRecord(ctx, pl, msg)

	// 花果山
	case *proto_huaguoshan.C2SInitHuaguoshan:
		paradise.ReqInitHuaguoshan(ctx, pl, msg)
	case *proto_huaguoshan.C2SInitPartner:
		paradise.ReqInitPartner(ctx, pl, msg)
	case *proto_huaguoshan.C2SGetPartnerInviteList:
		paradise.ReqGetPartnerInviteList(ctx, pl, msg)
	case *proto_huaguoshan.C2SPartnerInvite:
		paradise.ReqPartnerInvite(ctx, pl, msg)
	case *proto_huaguoshan.C2SLogicPartnerInvite:
		paradise.ReqLogicPartnerInvite(ctx, pl, msg)
	case *proto_huaguoshan.C2SRelieveParterner:
		paradise.ReqRelievePartner(ctx, pl, msg)
	case *proto_huaguoshan.C2SParternerGive:
		paradise.ReqPartnerGive(ctx, pl, msg)
	case *proto_huaguoshan.C2SStartMakeWine:
		paradise.ReqStartMakeWine(ctx, pl, msg)
	case *proto_huaguoshan.C2SCutWineRack:
		paradise.ReqCutMakeWine(ctx, pl, msg)
	case *proto_huaguoshan.C2SCollectWine:
		paradise.ReqCollectMakeWine(ctx, pl, msg)
	case *proto_huaguoshan.C2SStartPlantPeach:
		paradise.ReqStartPlantPeach(ctx, pl, msg)
	case *proto_huaguoshan.C2SLogicPlantPeach:
		paradise.ReqLogicPlantPeach(ctx, pl, msg)

	case *messages.SysMessage: // 系统消息
		dispatchSysMessage(ctx, pl, msg.Content)
	case *messages.GetPlayerDataMessage: //获取玩家数据
		return fromProps(pl)
	default:
		//return game.Process(ctx, pl, msg)
	}
	return nil
}

// dispatchSysMessage 处理模块下发的系统指令
func dispatchSysMessage(ctx global.IPlayer, pl *model.Player, content any) {
	switch cmd := content.(type) {
	case *messages.SysKick:
		log.Debug("player sys kick player_id=%d reason=%s", pl.Id, cmd.Reason)
		ctx.Send(&proto_player.S2CKick{})
		ctx.OnSave(true)
		ctx.Stop()
	case *messages.SysRefreshActivity:
		log.Debug("player sys refresh activity player_id=%d", pl.Id)
		// TODO: 刷新活动数据
	case *messages.SysGrantItems:
		awards := make([]conf2.ItemE, 0, len(cmd.Items))
		for _, e := range cmd.Items {
			awards = append(awards, conf2.ItemE{ItemId: e.ItemId, ItemType: e.ItemType, ItemNum: e.ItemNum})
		}
		bag.AddAward(ctx, pl, awards, true)
	default:
		log.Debug("player unknown sys message: %T", content)
	}
}
