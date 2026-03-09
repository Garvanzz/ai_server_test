package skill

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_skill"
)

func Init(pl *model.Player) {
	pl.Skill = new(model.Skill)
	pl.Skill.Ids = make(map[int32]int32)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Skill)
	if err != nil {
		log.Error("player[%v],save skill marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerSkill, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerSkill, pl.Id))
	if err != nil {
		log.Error("player[%v],load stage error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Skill)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load skill unmarshal error:%v", pl.Id, err)
	}

	pl.Skill = m
}

// ReqSkillList 请求技能等级
func ReqSkillList(ctx global.IPlayer, pl *model.Player, req *proto_skill.C2SInitSkill) {
	ctx.Send(&proto_skill.S2CInitSkill{
		SkillIds: pl.Skill.Ids,
	})
}
