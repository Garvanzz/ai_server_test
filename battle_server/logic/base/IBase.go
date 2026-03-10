package base

type IBase interface {
	OnGameClose()
	OnGameStart() //游戏开始
	OnGameOver()  //游戏结束
}
