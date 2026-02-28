package conf

type Hero struct {
	Id           int32     `json:"Id"`
	Rate         int32     `json:"Rate"`
	Sex          int32     `json:"Sex"`
	Job          int32     `json:"Job"`
	Type         int32     `json:"Type"`
	Fragment     int32     `json:"Fragment"`
	State        []int32   `json:"State"`
	NormalAtk    int32     `json:"NormalAtk"`
	UtilSkill    int32     `json:"UtilSkill"`
	PassiveSkill []int32   `json:"PassiveSkill"`
	SkillUnlock  [][]int32 `json:"SkillUnlock"`
	SkillWakeUp  [][]int32 `json:"SkillWakeUp"`
}

type HeroBasicAttribute struct {
	Id             int32 `json:"Id"`
	BasicHp        int32 `json:"BasicHp"`
	BasicAtk       int32 `json:"BasicAtk"`
	BasicForce     int32 `json:"BasicForce"`
	BasicDef       int32 `json:"BasicDef"`
	BasicMoveSpeed int32 `json:"BasicMoveSpeed"`
	BasicAtkSpeed  int32 `json:"BasicAtkSpeed"`
	Range          int32 `json:"Range"`
}

type HeroCultivation struct {
	Id            int32           `json:"Id"`
	Job           int32           `json:"Job"`
	Stage         int32           `json:"Stage"`
	Cultivation   int32           `json:"Cultivation"`
	CostCult      int32           `json:"CostCult"`
	BasicHp       int32           `json:"BasicHp"`
	BasicAtk      int32           `json:"BasicAtk"`
	BasicForce    int32           `json:"BasicForce"`
	BasicDef      int32           `json:"BasicDef"`
	AttibuteRatio float32         `json:"AttibuteRatio"`
	CostRatio     float32         `json:"CostRatio"`
	SkillLevel    map[int32]int32 `json:"SkillLevel"`
}

type HeroMagic struct {
	Id   int32 `json:"Id"`
	Type int32 `json:"Type"`
	Rate int32 `json:"Rare"`
}

type HeroMagicLevel struct {
	Id           int32 `json:"Id"`
	MagicId      int32 `json:"MagicId"`
	Level        int32 `json:"Level"`
	UplevelCost  int32 `json:"UplevelCost"`
	LinkSkill    int32 `json:"LinkSkill"`
	LinkBuff     int32 `json:"LinkBuff"`
	HeroAtk      int32 `json:"HeroAtk"`
	HeroDef      int32 `json:"HeroDef"`
	HeroForce    int32 `json:"HeroForce"`
	HeroHP       int32 `json:"HeroHP"`
	PerHeroAtk   int32 `json:"PerHeroAtk"`
	PerHeroDef   int32 `json:"PerHeroDef"`
	PerHeroForce int32 `json:"PerHeroForce"`
	PerHeroHP    int32 `json:"PerHeroHP"`
}

type HeroPool struct {
	Id       int32 `json:"Id"`
	Type     int32 `json:"Type"`
	PoolType int32 `json:"PoolType"`
	Value    int32 `json:"Value"`
	Rate     int32 `json:"Rate"`
	Weight   int32 `json:"Weight"`
	Level    int32 `json:"Level"`
}

type HeroUpLevel struct {
	Id               int32   `json:"Id"`
	HpRatio          int32   `json:"HpRatio"`
	AtkRatio         int32   `json:"AtkRatio"`
	DefRatio         int32   `json:"DefRatio"`
	ForceRatio       int32   `json:"ForceRatio"`
	MoveSpeedRatio   int32   `json:"MoveSpeedRatio"`
	AtkSpeedRatio    int32   `json:"AtkSpeedRatio"`
	UpLevelCondition []int32 `json:"UpLevelCondition"`
}

type HeroUpStage struct {
	Id               int32   `json:"Id"`
	HpRatio          int32   `json:"HpRatio"`
	AtkRatio         int32   `json:"AtkRatio"`
	DefRatio         int32   `json:"DefRatio"`
	ForceRatio       int32   `json:"ForceRatio"`
	MoveSpeedRatio   int32   `json:"MoveSpeedRatio"`
	AtkSpeedRatio    int32   `json:"AtkSpeedRatio"`
	UpStageCondition []int32 `json:"UpStageCondition"`
}

type HeroUpStar struct {
	Id             int32   `json:"Id"`
	HpRatio        int32   `json:"HpRatio"`
	AtkRatio       int32   `json:"AtkRatio"`
	DefRatio       int32   `json:"DefRatio"`
	ForceRatio     int32   `json:"ForceRatio"`
	MoveSpeedRatio int32   `json:"MoveSpeedRatio"`
	AtkSpeedRatio  int32   `json:"AtkSpeedRatio"`
	NeedId         int32   `json:"NeedId"`
	NeedCostNum    []int32 `json:"NeedCostNum"`
}
