package define

var EnchantType = map[int]int32{
	1: ItemIdTallyAtk,
	2: ItemIdTallyDef,
	3: ItemIdTallyHp,
	4: ItemIdTallyForce,
}

var EnchantTypeById = map[int32]int{
	ItemIdTallyAtk:   1,
	ItemIdTallyDef:   2,
	ItemIdTallyHp:    3,
	ItemIdTallyForce: 4,
}

const (
	BraceTalentLevelLimit = 5
)
