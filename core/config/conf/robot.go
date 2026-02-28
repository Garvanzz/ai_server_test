package conf

type Robot struct {
	Id        int32 `json:"Id"`
	HeroId    int32 `json:"HeroId"`
	Range     int32 `json:"Range"`
	Hp        int32 `json:"Hp"`
	Atk       int32 `json:"Atk"`
	Def       int32 `json:"Def"`
	MoveSpeed int32 `json:"MoveSpeed"`
	AtkSpeed  int32 `json:"AtkSpeed"`
	Power     int32 `json:"Power"`
	Level     int32 `json:"Level"`
	Star      int32 `json:"Star"`
}

type RobotGroup struct {
	Id      int32   `json:"Id"`
	Power   int64   `json:"Power"`
	RobotId []int32 `json:"RobotId"`
	Mode    int32   `json:"mode"`
}
