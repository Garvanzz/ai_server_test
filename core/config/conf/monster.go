package conf

type Monster struct {
	Id         int32   `json:"Id"`
	Job        int32   `json:"Job"`
	Star       int32   `json:"Star"`
	HeroId     int32   `json:"HeroId"`
	Cards      []int32 `json:"Cards"`
	Hp         int32   `json:"Hp"`
	Atk        int32   `json:"Atk"`
	Def        int32   `json:"Def"`
	MagAtk     int32   `json:"MagAtk"`
	MagDef     int32   `json:"MagDef"`
	Equips     []int32 `json:"Equips"`
	CanCatch   bool    `json:"CanCatch"`
	CatchPetId int32   `json:"CatchPetId"`
}

type StageMonsterGroup struct {
	Id         int32           `json:"Id"`
	Type       int32           `json:"Type"`
	Num        int32           `json:"num"`
	MonsterId  []int32         `json:"MonsterId"`
	MonsterNum []int32         `json:"MonsterNum"`
	KillExp    map[int32]int32 `json:"KillExp"`
	KillAward  [][]int32       `json:"KillAward"`
}

type MonsterGroup struct {
	Id         int32   `json:"Id"`
	Type       int32   `json:"Type"`
	MonsterId  []int32 `json:"MonsterId"`
	MonsterNum []int32 `json:"MonsterNum"`
	IsSkip     bool    `json:"IsSkip"`
}
