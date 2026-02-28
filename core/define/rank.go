package define

const (
	RankTypePerfect        = iota + 1 //帮会完美榜
	RankTypeGrow                      //帮会成长榜
	RankTypeGuildBattle               //帮会战斗榜
	RankTypeDrawHero                  //招募排行榜
	RankTypeRecharge                  //充值排行榜
	RankTypeClimbTower                //爬塔排行榜
	RankTypeTheCompetition            //巅峰决斗排行榜
	RankTypeArena                     //竞技场排行榜
	RankTypeTianti                    //天梯排行榜
	RankTypeGoFish                    //钓鱼争霸赛
	RankTypePower                     //战力排行榜
)

const (
	RankPerfectKey            = "rank_perfect"
	RankGrowKey               = "rank_grow"
	RankGuildBattleKey        = "rank_guild_battle"
	RankDrawHeroKey           = "rank_draw_hero"
	RankRechargeKey           = "rank_recharge"
	RankClimbTowerKey         = "rank_climb_tower"
	RankTypeTheCompetitionKey = "rank_the_competition"
	RankTypeArenaKey          = "rank_arena"
	RankTypeArenaRecordKey    = "rank_arena_record"
	RankTypeTiantiKey         = "rank_tianti"
	RankTypeTiantiRecordKey   = "rank_tianti_record"
	RankTypeGoFishKey         = "rank_gofish"
	RankTypePowerKey          = "rank_power"
)

var RankTypeToKey = map[int]string{
	RankTypePerfect:        RankPerfectKey,
	RankTypeGrow:           RankGrowKey,
	RankTypeGuildBattle:    RankGuildBattleKey,
	RankTypeDrawHero:       RankDrawHeroKey,
	RankTypeRecharge:       RankRechargeKey,
	RankTypeClimbTower:     RankClimbTowerKey,
	RankTypeTheCompetition: RankTypeTheCompetitionKey,
	RankTypeArena:          RankTypeArenaKey,
	RankTypeTianti:         RankTypeTiantiKey,
	RankTypeGoFish:         RankTypeGoFishKey,
}

const (
	RankTop = 100
)
