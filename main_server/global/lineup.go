package global

import (
	"encoding/json"
	"fmt"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_public"
)

// GetPlayerInfo 获取玩家布阵
func GetPlayerLineUpInfo(dbId int64) map[int32]*proto_public.CommonPlayerLineUpInfo {
	lineup := make(map[int32]*proto_public.CommonPlayerLineUpInfo)
	rdb, err := db.GetEngineByPlayerId(dbId)
	if err != nil {
		log.Error("GetPlayerInfo error, no this server:%v", err)
		return lineup
	}

	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerLineUp, dbId))
	if err != nil {
		log.Error("player[%v],load Collection error:%v", dbId, err)
		return lineup
	}

	if reply == nil {
		return lineup
	}

	dst := new(model.LineUp)
	err = json.Unmarshal(reply.([]byte), &dst)
	if err != nil {
		log.Error("player[%v],load Equip unmarshal error:%v", dbId, err)
		return lineup
	}

	for _, v := range dst.LineUps {
		heros := make([]*proto_public.CommonPlayerLineUpItemInfo, 0)
		for k := 0; k < len(v.HeroId); k++ {
			_lineup := new(proto_public.CommonPlayerLineUpItemInfo)
			_lineup.Id = v.HeroId[k]
			_lineup.Level = 0
			_lineup.Star = 0
			heros = append(heros, _lineup)
		}
		lineup[v.Type] = &proto_public.CommonPlayerLineUpInfo{
			Type:   v.Type,
			HeroId: heros,
		}
	}

	return lineup
}

// 获取阵容
func GetLineupLiupai(lineupSelf, lineupOther []int32) (bool, int32, int32, int32) {
	//流派阵容
	basicAtkDamage := int32(0)
	basicSkillDamage := int32(0)
	zengyiBuffConTime := int32(0)
	isLiupai, job, num := getLineUpOtion(lineupSelf)
	if !isLiupai {
		return false, basicAtkDamage, basicSkillDamage, zengyiBuffConTime
	}

	//判断对方有没有阵容
	isLiupaiOther, jobOther, _ := getLineUpOtion(lineupOther)
	if !isLiupaiOther {
		return false, basicAtkDamage, basicSkillDamage, zengyiBuffConTime
	}

	//判断是不是克制关系
	LiupaiRestrainConfs := config.CfgMgr.AllJson["LiupaiRestrain"].(map[int64]conf2.LiupaiRestrain)
	restrain := false
	for _, v := range LiupaiRestrainConfs {
		if v.Job == job && v.Restrain == jobOther {
			if job == define.PlayerJobYao {
				if num == 4 {
					basicAtkDamage = v.Restrain_Attribute[0]
				} else if num == 5 {
					basicAtkDamage = v.Restrain_Attribute[1]
				} else if num == 6 {
					basicAtkDamage = v.Restrain_Attribute[2]
				}
			} else if job == define.PlayerJobShen {
				if num == 4 {
					basicSkillDamage = v.Restrain_Attribute[0]
				} else if num == 5 {
					basicSkillDamage = v.Restrain_Attribute[1]
				} else if num == 6 {
					basicSkillDamage = v.Restrain_Attribute[2]
				}
			} else if job == define.PlayerJobFo {
				if num == 4 {
					zengyiBuffConTime = v.Restrain_Attribute[0]
				} else if num == 5 {
					zengyiBuffConTime = v.Restrain_Attribute[1]
				} else if num == 6 {
					zengyiBuffConTime = v.Restrain_Attribute[2]
				}
			}
			restrain = true
			break
		}
	}

	return restrain, (-1) * basicAtkDamage, (-1) * basicSkillDamage, zengyiBuffConTime
}

func getLineUpOtion(lineup []int32) (bool, int32, int32) {
	liupai := make(map[int32]int32)
	isLiupai := false
	conf := config.CfgMgr.AllJson["Hero"].(map[int64]conf2.Hero)
	for _, v := range lineup {
		if v <= 0 {
			continue
		}

		_conf := conf[int64(v)]
		if _, ok := liupai[_conf.Job]; !ok {
			liupai[_conf.Job] = 0
		}
		liupai[_conf.Job] += 1
	}

	liupaiConfs := config.CfgMgr.AllJson["Liupai"].(map[int64]conf2.Liupai)
	Job := int32(0)
	Num := int32(0)
	for job, num := range liupai {
		if num < 4 {
			continue
		}

		for _, c := range liupaiConfs {
			if c.Number == num {
				Job = job
				isLiupai = true
				Num = num
				break
			}
		}
	}

	return isLiupai, Job, Num
}
