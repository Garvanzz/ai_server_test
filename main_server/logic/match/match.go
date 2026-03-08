package match

import (
	"xfx/core/define"
	"xfx/pkg/utils"
	"xfx/core/model"
	"xfx/pkg/log"
)

var (
	matchLenWeights = []int32{8, 13, 18, 20, 19, 15, 7}
	matchSorWeights = []int32{22, 28, 30, 12, 8}
)

type Match struct {
	waitIds    []int32 // 需要等待的房间id
	sortIds    []int32 // 排队Id
	curMatchId int32   // 当前正在处理的Id
	teamNum    int32   // 队伍数

	matchPool map[int32]map[int32]struct{} // 匹配池
	teamInfo  map[int32]*model.MatchTeam   // 队伍信息

	matchlock bool
}

// 初始化
func (m *Match) init() {
	m.waitIds = make([]int32, 0)
	m.sortIds = make([]int32, 0)

	m.teamInfo = make(map[int32]*model.MatchTeam, 0)

	m.matchPool = make(map[int32]map[int32]struct{})
	for i := define.PlayerRankNull; i <= define.PlayerRankZuiQiangWangzhe; i++ {
		m.matchPool[int32(i)] = make(map[int32]struct{})
	}
}

func (m *Match) update() {
	if m.matchlock == false {
		m.matchlock = true

		m.curMatchId = m.getCurMatchId()
		if m.curMatchId == 0 {
			//没获取到队列id
		} else {

			//匹配对手
			team := m.getTeamById(m.curMatchId)
			enemy := m.matchEnemy(team)

			//没有匹配到对手，进入等地队列
			if enemy == nil {
				m.waitIds = append(m.waitIds, m.curMatchId)
				m.matchlock = false
			} else {
				// 通知匹配到的队伍
				notifyTeam(team, enemy)

				//删除匹配到的队伍
				m.deleteTeam(enemy.Id)
				m.deleteTeam(team.Id)

				m.teamNum -= 2
				m.matchlock = false
			}
		}
	}
}

// 写进匹配池
func (m *Match) putInMatchPool(team *model.MatchTeam) bool {
	pool, ok := m.matchPool[team.AverageRank]
	if !ok {
		log.Error("put in match pool error:%v", team.AverageRank)
		return false
	}

	m.teamInfo[team.Id] = team
	pool[team.Id] = struct{}{}

	m.teamNum++
	m.sortIds = append(m.sortIds, team.Id)
	return true
}

func (m *Match) startMatch(team *model.MatchTeam) bool {
	//进队列
	m.matchlock = false

	return m.putInMatchPool(team)

}

func (m *Match) cancelMatch(roomId int32) bool {
	m.deleteTeam(roomId)
	m.teamNum -= 1
	arr := utils.RemoveFirstInt32(m.waitIds, roomId)
	m.waitIds = arr
	arr = utils.RemoveFirstInt32(m.sortIds, roomId)
	m.sortIds = arr
	log.Debug("取消匹配:%v", roomId)
	return true
}

// 通过段位获取池子
func (m *Match) getPoolByRank(rank int32) map[int32]struct{} {
	return m.matchPool[rank]
}

// 通过队列id获取队伍信息
func (m *Match) getTeamById(id int32) *model.MatchTeam {
	return m.teamInfo[id]
}

// 通过Id删除匹配队伍
func (m *Match) deleteTeam(id int32) {
	team, ok := m.teamInfo[id]
	if !ok {
		log.Error("deleteTeam id error:%v", id)
		return
	}

	pool, ok := m.matchPool[team.AverageRank]
	if !ok {
		log.Error("delete team from pool error:%v", id)
		return
	}

	delete(pool, id)
	delete(m.teamInfo, id)
}

// 获取池子里随机team
func (m *Match) randomTeamFromPool(pool map[int32]struct{}, exclude int32, isGroup bool) *model.MatchTeam {
	for k := range pool {
		if k != exclude {
			team := m.getTeamById(k)
			if team == nil {
				log.Error("randomTeamFromPool team id error:%v", k)
				return nil
			}

			if team.IsGroup == isGroup {
				return team
			}
		}
	}
	return nil
}

// 获取匹配队列的Id
func (m *Match) getCurMatchId() int32 {
	curId := int32(0)

	if len(m.sortIds) <= 0 {
		if len(m.waitIds) > 0 {
			curId = m.waitIds[0]
			m.waitIds = m.waitIds[1:]
		}
	} else {
		curId = m.sortIds[0]
		m.sortIds = m.sortIds[1:]
	}

	return curId
}

// 根据段位通过权重获取匹配值
func getRankIndex(rank int32) int32 {
	matchRank := rank
	//小于铂金，上下取3， 大于上下取2
	if rank <= define.PlayerRankBojin {
		index := utils.WeightIndex(matchLenWeights)
		if index < 3 {
			matchRank = rank - int32(index)
		} else if index == 3 {
			matchRank = rank
		} else {
			matchRank = rank + int32(index)
		}
	} else {
		index := utils.WeightIndex(matchSorWeights)
		if index < 2 {
			matchRank = rank - int32(index)
		} else if index == 2 {
			matchRank = rank
		} else {
			matchRank = rank + int32(index)
		}
	}

	//临界值判断
	if matchRank <= 0 {
		//默认青铜
		matchRank = define.PlayerRankBaiyin
	}

	if matchRank > define.PlayerRankZuiQiangWangzhe {
		matchRank = define.PlayerRankZuiQiangWangzhe
	}

	return matchRank
}

// 匹配对手
func (m *Match) matchEnemy(curTeam *model.MatchTeam) *model.MatchTeam {
	if m.teamNum <= 1 {
		return nil
	}

	var matPlayer *model.MatchTeam
	//判断人数,测试用先写5个,小于这个数 直接随机获取，大于才判断段位
	if m.teamNum <= 5 {
		mat := true
		for mat {
			val := utils.RandInt(int32(define.PlayerRankBaiyin), define.PlayerRankZuiQiangWangzhe)
			pool := m.getPoolByRank(val)
			if len(pool) > 0 {
				matPlayer = m.randomTeamFromPool(pool, curTeam.Id, curTeam.IsGroup)
				mat = false
			}
		}
	} else {
		mat := true
		for mat {
			val := getRankIndex(curTeam.AverageRank)
			pool := m.getPoolByRank(val)
			if len(pool) > 0 {
				matPlayer = m.randomTeamFromPool(pool, curTeam.Id, curTeam.IsGroup)
				mat = false
			}
		}
	}

	if matPlayer != nil && matPlayer.Id > 0 {
		return matPlayer
	}

	return nil
}
