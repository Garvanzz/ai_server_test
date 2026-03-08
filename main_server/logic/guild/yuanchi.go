package guild

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/model"
	"xfx/pkg/utils"
	"xfx/proto/proto_guild"
	"xfx/proto/proto_player"
)

// 元池
func (mgr *Manager) getYuanchiData(ctx *proto_player.Context) (*proto_guild.S2CInitYuanchi, error) {
	//获取帮派信息
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return nil, errors.New("no guild")
	}

	if info.GuildId == 0 {
		return nil, errors.New("no guild")
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return nil, errors.New("no entity")
	}

	res := &proto_guild.S2CInitYuanchi{}
	if ent.guild.Yuanchi == nil {
		ent.guild.Yuanchi = new(model.GuildYuanchi)
	}

	//原材料
	if ent.guild.Yuanchi.Materials == nil {
		ent.guild.Yuanchi.Materials = make(map[int32]int32)
	}

	res.RawMaterials = ent.guild.Yuanchi.Materials

	//炼制中
	if ent.guild.Yuanchi.Refinings == nil {
		ent.guild.Yuanchi.Refinings = make(map[int32]*model.YuanchiRefining)
	}

	res.Refinings = make(map[int32]*proto_guild.YuanchiRefining)
	for _, v := range ent.guild.Yuanchi.Refinings {
		ref := new(proto_guild.YuanchiRefining)
		ref.Id = v.Id
		ref.Alltime = v.AllTime
		ref.Time = v.Time
		ref.Rare = v.Rate

		if ref.YuanchiItems == nil {
			ref.YuanchiItems = make(map[int32]*proto_guild.YuanchiItem)
		}
		ref.YuanchiItems = make(map[int32]*proto_guild.YuanchiItem)
		for _, v := range ref.YuanchiItems {
			ref.YuanchiItems[v.Id] = &proto_guild.YuanchiItem{
				Id:  v.Id,
				Num: v.Num,
			}
		}
		res.Refinings[v.Id] = ref
	}

	//元素
	if ent.guild.Yuanchi.Elements == nil {
		ent.guild.Yuanchi.Elements = make(map[int32]int32)
	}
	res.Bags = ent.guild.Yuanchi.Elements

	return res, nil
}

// 元池-材料增加
func (mgr *Manager) addYuanchiMaterials(ctx *proto_player.Context, materials map[int32]int32) error {
	//获取帮派信息
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return errors.New("no guild")
	}

	if info.GuildId == 0 {
		return errors.New("no guild")
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return errors.New("no entity")
	}

	if ent.guild.Yuanchi == nil {
		ent.guild.Yuanchi = new(model.GuildYuanchi)
	}

	//原材料
	if ent.guild.Yuanchi.Materials == nil {
		ent.guild.Yuanchi.Materials = make(map[int32]int32)
	}

	for k, v := range materials {
		if _, ok := ent.guild.Yuanchi.Materials[k]; !ok {
			ent.guild.Yuanchi.Materials[k] = v
		} else {
			ent.guild.Yuanchi.Materials[k] += v
		}
	}

	//触发炼制判断
	ent.guildRefiningLogic()

	//同步变化
	ent.boardCast(&proto_guild.PushMaterialChange{
		RawMaterial: ent.guild.Yuanchi.Materials,
	})
	return nil
}

// tick
func (ent *entity) updateYuanchi() {
	//炼制中
	if ent.guild.Yuanchi == nil {
		return
	}

	if ent.guild.Yuanchi.Refinings == nil {
		return
	}

	// 先标记要删除的元素
	toRemove := make(map[int32]bool)
	for i, v := range ent.guild.Yuanchi.Refinings {
		if v.LastTime == 0 {
			v.LastTime = utils.Now().Unix()
		} else {
			if utils.Now().Unix()-v.LastTime >= 60 {
				v.LastTime = utils.Now().Unix()

				v.Time += 60
				// 计算当前成功率（随时间动态变化）
				currentRate := calculateCurrentRate(float64(v.AllTime), float64(v.Time), float64(v.Rate))

				if v.Time >= v.AllTime {
					ent.addGuildRefiningLog(v, true)
					toRemove[i] = true

					ent.addGuildElements(v.Id)
				} else if rand.Float64() > currentRate {
					ent.addGuildRefiningLog(v, false)
					toRemove[i] = true
					//推送
					ent.pushChangeGuildRefining()
				}
			}
		}
	}

	// 压缩切片
	if len(toRemove) > 0 {
		for k, _ := range toRemove {
			delete(ent.guild.Yuanchi.Refinings, k)
		}
	}
}

func calculateCurrentRate(TotalTime, ElapsedTime, BaseRate float64) float64 {
	remainingTime := TotalTime - ElapsedTime
	remainingTime = math.Max(0, remainingTime)

	// 使用剩余时间比例计算成功率
	// 这里使用平方关系使曲线更平滑
	timeRatio := remainingTime / TotalTime
	rate := BaseRate + (1-BaseRate)*(1-timeRatio*timeRatio)

	return math.Max(0.01, math.Min(0.99, rate))
}

// 添加炼制记录到redis
func (ent *entity) addGuildRefiningLog(data *model.YuanchiRefining, state bool) {
	guildLog := new(model.GuildRefiningLog)
	guildLog.Data = data
	guildLog.Time = utils.Now().Unix()
	guildLog.State = state

	rdb, _ := db.GetEngine(Mgr.App.GetEnv().ID)
	key := fmt.Sprintf("guild_refing_history:%d", ent.guild.Id)

	js, _ := json.Marshal(guildLog)
	//尾部添
	rdb.RedisExec("RPUSH", key, js)
}

// 炼制判断
func (ent *entity) guildRefiningLogic() {
	confs := config.GuildElement.All()

	// 将map转换为切片
	confSlice := make([]conf.GuildElement, 0, len(confs))
	for _, elem := range confs {
		confSlice = append(confSlice, elem)
	}

	// 按照某个字段排序，例如按ID升序
	slices.SortFunc(confSlice, func(a, b conf.GuildElement) int {
		if a.Id < b.Id {
			return -1
		} else if a.Id > b.Id {
			return 1
		}
		return 0
	})

	//加入炼制中
	for i := 0; i < len(confSlice); i++ {
		if !confSlice[i].IsElement {
			continue
		}

		//是否满足
		if ent.guild.Yuanchi.Refinings != nil {
			if _, ok := ent.guild.Yuanchi.Refinings[int32(confSlice[i].Id)]; ok {
				continue
			}
		}

		needMaterial := confSlice[i].Material
		if len(needMaterial) <= 0 {
			continue
		}
		for k, v := range needMaterial {
			if ent.guild.Yuanchi.Materials == nil {
				continue
			}

			num := ent.guild.Yuanchi.Materials[k]
			if num < v {
				continue
			}
		}

		//添加到炼制中
		ref := &model.YuanchiRefining{
			Id:          confSlice[i].Id,
			Time:        0,
			AllTime:     confSlice[i].BasicTime,
			Rate:        confSlice[i].BasicSuccessRare,
			YuanchiItem: make(map[int32]*model.YuanchiItem),
		}
		ent.addGuildRefining(ref)
	}
}

// 增加炼制中
func (ent *entity) addGuildRefining(ref *model.YuanchiRefining) {
	if ent.guild.Yuanchi.Refinings == nil {
		ent.guild.Yuanchi.Refinings = make(map[int32]*model.YuanchiRefining)
	}

	//添加到炼制中
	ent.guild.Yuanchi.Refinings[ref.Id] = ref

	ent.pushChangeGuildRefining()
}

// 同步变化
func (ent *entity) pushChangeGuildRefining() {
	refinings := make(map[int32]*proto_guild.YuanchiRefining)
	for _, v := range ent.guild.Yuanchi.Refinings {
		refs := new(proto_guild.YuanchiRefining)
		refs.Id = v.Id
		refs.Alltime = v.AllTime
		refs.Time = v.Time
		refs.Rare = v.Rate

		if refs.YuanchiItems == nil {
			refs.YuanchiItems = make(map[int32]*proto_guild.YuanchiItem)
		}
		refs.YuanchiItems = make(map[int32]*proto_guild.YuanchiItem)
		for _, v := range refs.YuanchiItems {
			refs.YuanchiItems[v.Id] = &proto_guild.YuanchiItem{
				Id:  v.Id,
				Num: v.Num,
			}
		}
		refinings[v.Id] = refs
	}
	ent.boardCast(&proto_guild.PushRefiningChange{
		Refining: refinings,
	})
}

// 增加元素
func (ent *entity) addGuildElements(id int32) {
	//增加元素
	if ent.guild.Yuanchi.Elements == nil {
		ent.guild.Yuanchi.Elements = make(map[int32]int32)
	}

	num := ent.guild.Yuanchi.Elements[id]
	num += 1
	ent.guild.Yuanchi.Elements[id] = num

	//同步变化
	ent.boardCast(&proto_guild.PushBagChange{
		Bag: ent.guild.Yuanchi.Elements,
	})
}
