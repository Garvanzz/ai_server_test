package conf

type Succinct struct {
	Id       int32   `json:"Id"`
	Level    int32   `json:"Level"`
	Exp      int32   `json:"Exp"`
	Weight   []int32 `json:"Weight"`
	Cost     []ItemE `json:"Cost"`
	AddAtk   int32   `json:"AddAtk"`
	AddDef   int32   `json:"AddDef"`
	AddHp    int32   `json:"AddHp"`
	AddForce int32   `json:"AddForce"`
}
