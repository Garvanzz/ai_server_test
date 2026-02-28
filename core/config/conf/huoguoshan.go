package conf

// 伴侣亲密度
type ParternerIntimacy struct {
	Id                   int32 `json:"Id"`
	UnLockBuffValue      int32 `json:"UnLockBuffValue"`
	UnLockHeadFrameValue int32 `json:"UnLockHeadFrameValue"`
	UnLockHeadWearValue  int32 `json:"UnLockHeadWearValue"`
	UnLockBraceValue     int32 `json:"UnLockBraceValue"`
	UnLockMountValue     int32 `json:"UnLockMountValue"`
	UnLockSkillValue     int32 `json:"UnLockSkillValue"`
	Stage                int32 `json:"Stage"`
	Exp                  int32 `json:"Exp"`
}

// 伴侣副本
type ParternerMission struct {
	Id           int32   `json:"Id"`
	Flower       int32   `json:"flower"`
	MonsterGroup int32   `json:"monsterGroup"`
	Awards       []ItemE `json:"Awards"`
	AddRate      int32   `json:"addRate"`
	awardAddRate int32   `json:"awardAddRate"`
}

// 酒架
type WineRack struct {
	Id                int32   `json:"Id"`
	Type              []int32 `json:"type"`
	MakeTime          int32   `json:"makeTime"`
	XiuweineedCost    []ItemE `json:"xiuweineedCost"`
	zhanlineedCost    []ItemE `json:"zhanlineedCost"`
	yuanfenneedCost   []ItemE `json:"yuanfenneedCost"`
	zhenniangneedCost []ItemE `json:"zhenniangneedCost"`
}

// GetWineCostByType 根据酒类型获取消耗
func (w *WineRack) GetWineCostByType(wineType int32) []ItemE {
	switch wineType {
	case 1: // 修为酒
		return w.XiuweineedCost
	case 2: // 战力酒
		return w.zhanlineedCost
	case 3: // 缘分酒
		return w.yuanfenneedCost
	case 4: // 真酿酒
		return w.zhenniangneedCost
	default:
		return nil
	}
}

// CanMakeWineType 检查是否可以酿造指定类型的酒
func (w *WineRack) CanMakeWineType(wineType int32) bool {
	for _, t := range w.Type {
		if t == wineType {
			return true
		}
	}
	return false
}

// 桃树
type PeachTree struct {
	Id               int32   `json:"Id"`
	Type             []int32 `json:"type"`
	MakeTime         int32   `json:"makeTime"`
	Stage            []int32 `json:"Stage"`
	Award            []ItemE `json:"award"`
	CoolDownTime     int32   `json:"coolDownTime"`      // 浇水缩短的时间(秒)
	AddNum           int32   `json:"addNum"`            // 施肥增加的产出数量
	AddNumneedCost   []ItemE `json:"addNumneedCost"`   // 施肥消耗
	CoolDownneedCost []ItemE `json:"coolDownneedCost"` // 浇水消耗
}

// GetStageTime 获取指定阶段的时长
func (p *PeachTree) GetStageTime(stage int32) int32 {
	if stage < 1 || stage > int32(len(p.Stage)) {
		return 0
	}
	return p.Stage[stage-1]
}
