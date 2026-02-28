package model

// PlayerCdkey 玩家兑换码数据
type PlayerCdkey struct {
	UsedKeys map[string]int32 // key是兑换码，value是已使用次数
}
