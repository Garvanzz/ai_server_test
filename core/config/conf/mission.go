package conf

type Mission struct {
	Id            int32   `json:"Id"`
	OpenType      int32   `json:"OpenType"`
	OpenValue     []int32 `json:"OpenValue"`
	ChallengeCost []ItemE `json:"challengeCost"`
	ChallengeNum  int32   `json:"challengeNum"`
	MonsetrGroup  int32   `json:"monsetrGroup"`
	Award         []ItemE `json:"Award"`
	AddRate       int32   `json:"addRate"`
	Type          int32   `json:"Type"`
}
type ClimbTower struct {
	Id           int32   `json:"Id"`
	Flower       int32   `json:"flower"`
	AddRate      int32   `json:"addRate"`
	AwardAddRate int32   `json:"awardAddRate"`
	MonsetrGroup int32   `json:"monsetrGroup"`
	Award        []ItemE `json:"Award"`
}
