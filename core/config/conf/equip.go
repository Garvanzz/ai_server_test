package conf

type Equip struct {
	Id               int32 `json:"Id"`
	HeroAttId        int32 `json:"HeroAttId"`
	Index            int32 `json:"Index"`
	Rate             int32 `json:"Rate"`
	LinkSkillId      int32 `json:"LinkSkillId"`
	LinkBuffId       int32 `json:"LinkBuffId"`
	BasicAtk         int32 `json:"BasicAtk"`
	BasicDef         int32 `json:"BasicDef"`
	BasicHp          int32 `json:"BasicHp"`
	BasicForce       int32 `json:"BasicForce"`
	RAtk             int32 `json:"RAtk"`
	Rdef             int32 `json:"Rdef"`
	RHp              int32 `json:"RHp"`
	RForce           int32 `json:"RForce"`
	AddDamage        int32 `json:"AddDamage"`
	AddAniDamage     int32 `json:"AddAniDamage"`
	AddMp            int32 `json:"AddMp"`
	AddIgnoreDodge   int32 `json:"AddIgnoreDodge"`
	AddDodge         int32 `json:"AddDodge"`
	AddCrit          int32 `json:"AddCrit"`
	AddIgnoreCrit    int32 `json:"AddIgnoreCrit"`
	AddAtkSpeed      int32 `json:"AddAtkSpeed"`
	AddTimeAniDamage int32 `json:"AddTimeAniDamage"`
	AddNormalAtk     int32 `json:"AddNormalAtk"`
	AddAniNormalAtk  int32 `json:"AddAniNormalAtk"`
	AddSkilDamage    int32 `json:"AddSkilDamage"`
	AddAniSkilDamage int32 `json:"AddAniSkilDamage"`
	AddTimeDmage     int32 `json:"AddTimeDmage"`
}

type EquipSell struct {
	Id    int32   `json:"Id"`
	Award []ItemE `json:"Award"`
	Rate  int32   `json:"Rate"`
	Index int32   `json:"Index"`
}
