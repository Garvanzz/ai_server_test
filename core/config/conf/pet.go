package conf

type Pet struct {
	Id            int32     `json:"Id"`
	Rate          int32     `json:"Rate"`
	Type          int32     `json:"Type"`
	Fragment      int32     `json:"Fragment"`
	AttributeType int32     `json:"AttributeType"`
	SkillUnlock   [][]int32 `json:"SkillUnlock"`
	GiftUnlock    [][]int32 `json:"GiftUnlock"`
	NormalAtk     []int32   `json:"NormalAtk"`
	UtilSkill     []int32   `json:"UtilSkill"`
	PassiveSkill  [][]int32 `json:"PassiveSkill"`
}

type PetDrawPool struct {
	Id       int32 `json:"Id"`
	PoolType int32 `json:"PoolType"`
	Type     int32 `json:"Type"`
	Value    int32 `json:"Value"`
	Rate     int32 `json:"Rate"`
	Num      int32 `json:"Num"`
	Weight   int32 `json:"Weight"`
}

type PetEquip struct {
	Id        int32   `json:"Id"`
	Rate      int32   `json:"Rate"`
	Type      int32   `json:"Type"`
	BreakDown []ItemE `json:"BreakDown"`
}

type PetEquipLevel struct {
	Id       int32   `json:"Id"`
	Level    int32   `json:"Level"`
	Exp      int32   `json:"Exp"`
	Rate     int32   `json:"Rate"`
	NeedCost []ItemE `json:"NeedCost"`
	NeedNum  int32   `json:"NeedNum"`
}

type PetEquipHandbook struct {
	Id       int32 `json:"Id"`
	TargetId int32 `json:"TargetId"`
	Exp      int32 `json:"Exp"`
}

type PetEquipHandbookAward struct {
	Id           int32 `json:"Id"`
	Level        int32 `json:"Level"`
	Exp          int32 `json:"Exp"`
	BasicDef     int32 `json:"BasicDef"`
	BasicHp      int32 `json:"BasicHp"`
	BasicAtk     int32 `json:"BasicAtk"`
	BasicForce   int32 `json:"BasicForce"`
	AddHp        int32 `json:"AddHp"`
	AddAtk       int32 `json:"AddAtk"`
	AddDef       int32 `json:"AddDef"`
	AddHeroAtk   int32 `json:"AddHeroAtk"`
	AddForce     int32 `json:"AddForce"`
	AddHeroHp    int32 `json:"AddHeroHp"`
	AddHeroDef   int32 `json:"AddHeroDef"`
	AddHeroForce int32 `json:"AddHeroForce"`
}

type PetUpLevel struct {
	Id               int32   `json:"Id"`
	AtkRatio         int32   `json:"AtkRatio"`
	ForceRatio       int32   `json:"ForceRatio"`
	UpLevelCondition []int32 `json:"UpLevelCondition"`
}

type PetUpStage struct {
	Id               int32   `json:"Id"`
	AtkRatio         int32   `json:"AtkRatio"`
	ForceRatio       int32   `json:"ForceRatio"`
	UpStageCondition []int32 `json:"UpStageCondition"`
}

type PetUpStar struct {
	Id          int32   `json:"Id"`
	AtkRatio    int32   `json:"AtkRatio"`
	ForceRatio  int32   `json:"ForceRatio"`
	NeedId      int32   `json:"NeedId"`
	NeedCostNum []int32 `json:"NeedCostNum"`
}

type PetGiftCost struct {
	Id         int32     `json:"Id"`
	Rate       int32     `json:"Rate"`
	NormalCost [][]int32 `json:"NormalCost"`
	PointCost  [][]int32 `json:"PointCost"`
	Cost       []ItemE   `json:"Cost"`
}

type PetGift struct {
	Id        int32   `json:"Id"`
	Rate      int32   `json:"Rate"`
	MatchPet  []int32 `json:"MatchPet"`
	Rare      int32   `json:"Rare"`
	BindSkill int32   `json:"BindSkill"`
}

type PetSkill struct {
	Id        int32 `json:"Id"`
	Rate      int32 `json:"Rate"`
	PetId     int32 `json:"PetId"`
	PetBookId int32 `json:"PetBookId"`
	BindSkill int32 `json:"BindSkill"`
	Type      int32 `json:"Type"`
}
