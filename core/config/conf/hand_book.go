package conf

type HandBook struct {
	Id       int32 `json:"Id"`
	TargetId int32 `json:"TargetId"`
	Exp      int32 `json:"Exp"`
}

type HandBookAward struct {
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
	Type         int32 `json:"Type"`
}
