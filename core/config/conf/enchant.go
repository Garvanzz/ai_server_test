package conf

type Enchant struct {
	Id         int32 `json:"Id"`
	Level      int32 `json:"Level"`
	CostNum    int32 `json:"CostNum"`
	BasicAtk   int32 `json:"BasicAtk"`
	BasicDef   int32 `json:"BasicDef"`
	BasicHp    int32 `json:"BasicHp"`
	BasicForce int32 `json:"BasicForce"`
}

type EnchantStage struct {
	Id       int32 `json:"Id"`
	Level    int32 `json:"Level"`
	AddAtk   int32 `json:"AddAtk"`
	AddDef   int32 `json:"AddDef"`
	AddHp    int32 `json:"AddHp"`
	AddForce int32 `json:"AddForce"`
}
