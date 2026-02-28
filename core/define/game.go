package define

const (
	GameStage_Card            = 1 //操作卡阶段
	GameStage_PreBattle       = 2 //战斗准备阶段
	GameStage_Battle          = 3 //战斗进行阶段
	GameStage_GameTime        = 4 //游戏时长
	GameStage_StageChangeTime = 5 //阶段变换倒计时
)

const (
	MaxRound = 30
)

// 卡类型
const (
	CardTypeSkill = 1
	CardTypeEquip = 2
)

// 阵营
const (
	BattleGroupLeft  = 1
	BattleGroupRight = 2
)

// 技能分类
const (
	SkillTypeUtil      = 1
	SkillTypePassive   = 2
	SkillTypeNormalAtk = 3
)

// 技能效果分类
const (
	SkillEffectDamage               = 101 //技能伤害
	SkillEffectLowerHpKill          = 102 //低血斩杀
	SkillEffectRealDamage           = 103 //真实伤害
	SkillEffectCombo                = 104 //连击
	SkillEffectAddAttribute         = 201 //增益属性
	SkillEffectHeal                 = 202 //治疗
	SkillEffectRevive               = 203 //复活
	SkillEffectRemoveAllDebuff      = 204 //去除所有的debuff
	SkillEffectRemoveAllControl     = 205 //解除身上的控制
	SkillEffectReduceAllDebuffRound = 206 //减少身上所有debuff的持续回合
	SkillEffectReduceAllSkillRound  = 207 //减少所有技能cd
	SkillEffectReduceDef            = 301 //减防
	SkillEffectTriggerBuff          = 401 //触发BUFF
)

// BUFF效果分类
const (
	BuffEffectResDamageShield      = 1  //减伤护盾
	BuffEffectKillDamageShield     = 2  //挡伤护盾
	BuffEffectInvincible           = 3  //无敌
	BuffEffectVertigo              = 4  //眩晕
	BuffEffectHeal                 = 5  //生命恢复
	BuffEffectReduceAtt            = 6  //减少属性
	BuffEffectAddAtt               = 7  //增加属性
	BuffEffectBackfire             = 8  //反伤
	BuffEffectNoUseSkill           = 9  //不能释放技能
	BuffEffectReduceHeal           = 10 //减少生命恢复
	BuffEffectRemoveAllDebuff      = 11 //解除身上的所有debuff
	BuffEffectBurn                 = 12 //灼烧
	BuffEffectImmunityControl      = 13 //免疫控制
	BuffEffectImmunityAtkDamage    = 14 //免疫物理伤害
	BuffEffectImmunityMagAtkDamage = 15 //免疫法术伤害
	BuffEffectReduceDamage         = 16 //减伤
	BuffEffectLowBloodHighAtk      = 17 //低血高攻
)

// 属性Id
const (
	AttributeIdHp                     = 1001 // 气血
	AttributeIdAttack                 = 1002 // 攻击
	AttributeIdDef                    = 1003 // 防御
	AttributeIdForce                  = 1004 // 灵气
	AttributeIdMoveSpeed              = 1005 // 移速
	AttributeIdAddDamage              = 1006 // 伤害加成
	AttributeIdDamageRes              = 1007 // 伤害抗性
	AttributeIdAnger                  = 1008 // 怒气效率
	AttributeIdIgnoreDodge            = 1009 // 命中率
	AttributeIdAddDodge               = 1010 // 闪避率
	AttributeIdCrit                   = 1011 // 暴击率
	AttributeIdIgnoreCrit             = 1012 // 抗暴率
	AttributeIdCritDamage             = 1013 // 暴伤
	AttributeIdIgnoreCritDamage       = 1014 // 韧性
	AttributeIdAtkSpeed               = 1015 // 攻速
	AttributeIdContDamageRes          = 1016 // 持续伤害抗性
	AttributeIdNormalAtkDamage        = 1017 // 普攻伤害
	AttributeIdNormalAtkDamageRes     = 1018 // 普攻伤害抗性
	AttributeIdSkillDamage            = 1019 // 技能伤害
	AttributeIdSkillDamageRes         = 1020 // 技能伤害抗性
	AttributeIdContDamage             = 1021 // 持续伤害
	AttributeIdFinalDef               = 1022 // 最终防御
	AttributeIdFinalAtk               = 1023 // 最终攻击
	AttributeIdFinalHp                = 1024 // 最终气血
	AttributeIdFinalForce             = 1025 // 最终内力
	AttributeIdBossRestrain           = 1026 // 首领克制
	AttributeIdFinalAddDamage         = 1027 // 最终伤害加成
	AttributeIdFinalDamageRes         = 1028 // 最终伤害抗性
	AttributeIdBasicAttackDamage      = 1029 // 基础伤害加成
	AttributeIdBasicSkillDamage       = 1030 // 基础技能伤害加成
	AttributeIdZengyiBuffContinueTime = 1031 // 增益类BUFF持续时间
	AttributeIdJianyiBuffContinueTime = 1032 // 减益类BUFF持续时间
)

// Buff目标
const (
	BuffTargetFollow  = 1 //携带者
	BuffTargetCreate  = 2 //创建者
	BuffTargetTrigger = 3 //触发者
)

// Buff触发类型
const (
	BuffTriggerDie             = 1 //死亡者
	BuffTriggerAttecked        = 2 //受到攻击
	BuffTriggerRoundStart      = 3 //回合开始
	BuffTriggerReplyHp         = 4 //生命回复
	BuffTriggerImmediately     = 5 //立即触发
	BuffTriggerRoundEnd        = 6 //回合结束
	BuffTriggerImAndRoundStart = 7 //立即触发&回合开始
	BuffTriggerImAndRoundEnd   = 8 //立即触发&回合结束
)

// Buff添加类型
const (
	BuffAdd         = 1 //添加
	BuffRemove      = 2 //移除
	BuffValueChange = 3 //变值
)

// Buff移除类型
const (
	BuffRemoveTypeUse      = 1 //使用完
	BuffRemoveTypeRoundEnd = 2 //回合数完
	BuffRemoveTypeSystem   = 3 //系统
)

// 技能范围
const (
	SkillTargetTypeEnemy     = 1
	SkillTargetTypeAllEnemy  = 2
	SkillTargetTypeFriend    = 3
	SkillTargetTypeAllFriend = 4
)

// 技能限制范围
const (
	SkillTargetLimitMySelf = 1
	SkillTargetLimitLeft   = 2
	SkillTargetLimitRight  = 3
	SkillTargetLimitIndex  = 4
)

const (
	SETTLE_RARE_Null = 0
	SETTLE_RARE_S    = 1
	SETTLE_RARE_SS   = 2
	SETTLE_RARE_SSS  = 3
	SETTLE_RARE_SSSS = 4
)

// 同步
const (
	SyncActionRefreshPool = 1
	SyncActionMoney       = 2
	SyncActionUseCard     = 3
	SyncActionAttribute   = 4
	SyncActionCard        = 5
	SyncActionReport      = 6
	SyncActionRefCount    = 7
	SyncActionStageTime   = 8
)
