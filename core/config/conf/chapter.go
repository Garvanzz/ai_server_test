package conf

type Chapter struct {
	Id               int32  `json:"Id"`
	Monster          int32  `json:"monster"`
	Front            int32  `json:"front"`
	Next             int32  `json:"next"`
	StartStage       int32  `json:"startStage"`
	UnlockStoryType  int32  `json:"UnlockStoryType"`
	UnlockStoryValue string `json:"UnlockStoryValue"`
}
