package model

import (
	"xfx/proto/proto_equip"
)

type Equip struct {
	Equips   []*EquipOption           //装备列表
	Mount    *MountOption             //坐骑
	Weaponry *WeaponryOption          //神兵
	Enchant  map[int32]*EnchantOption //附魔
	Succinct *SuccinctOption          //洗练
	Brace    *BraceOption             //背饰
}

// EnchantOption 附魔
type EnchantOption struct {
	Id    int32 // 符咒ID
	Level int32 // 等级
	Exp   int32 // 经验
}

// 背饰
type BraceOption struct {
	BraceAuraItems    map[int32]*BraceAuraItem
	BraceItems        map[int32]*BraceItem
	GetAuraStageAward []int32                     //领取共鸣等级
	BraceTalentIndexs map[int32]*BraceTalentIndex //天赋方案
	BraceTalentIndex  int                         //正在使用的方案
	HandbookExp       int32
	HandbookIds       []int32
}

// 背饰-灵韵
type BraceAuraItem struct {
	Type  int32
	Level int32
	Exp   int32
}

// 背饰-背饰
type BraceItem struct {
	Id    int32
	Num   int32
	Level int32
	IsUse bool
}

// 背饰-天赋
type BraceTalentIndex struct {
	Index           int32                     //方案序列
	Name            string                    //方案名字
	BraceTalentJobs map[int32]*BraceTalentJob //每个职业对应的天赋组
}

type BraceTalentJob struct {
	Job               int32                       //职业
	BraceTalentGroups map[int32]*BraceTalentGroup //每个组
}

type BraceTalentGroup struct {
	Group            int32                      //组
	BraceTalentItems map[int32]*BraceTalentItem //每个组对应的天赋点
}

// 背饰-天赋点
type BraceTalentItem struct {
	Id    int32
	Level int32
	Exp   int32
}

// 洗练结构体
type SuccinctOption struct {
	UseIndex            int                         //正在使用的方案
	Level               int                         //等级
	Exp                 int32                       //经验
	SuccinctIndexs      map[int]*SuccinctIndex      //方案集合
	CacheSuccinctIndexs map[int]*CacheSuccinctIndex //缓存的方案
	LevelAward          []int32                     //等级奖励
}

// 方案
type SuccinctIndex struct {
	Index   int             //序列
	SkillId map[int32]int32 //技能列表
	Name    string          //名字
}

// 缓存方案
type CacheSuccinctIndex struct {
	Index      int   //序列
	SkillId    int32 //技能ID
	EquipIndex int   //装备部位
}

type MountOption struct {
	Stage       int32
	Star        int32
	Exp         int32
	UseId       int32
	Mount       map[int32]*MountItemOption //坐骑
	MountEnergy map[int32]int32            //坐骑赋能
	HandbookExp int32
	HandbookIds []int32
}

type MountItemOption struct {
	Name  string
	Id    int32
	Num   int32
	Level int32 //等级
}

type WeaponryOption struct {
	Star          int32
	Exp           int32
	UseId         int32
	WeaponryItems map[int32]*WeaponryItem
	HandbookExp   int32
	HandbookIds   []int32
}

type WeaponryItem struct {
	Id    int32
	Level int32 //等级
	Num   int32
}

type EquipOption struct {
	CId   int32 //配置ID
	Id    int32 //唯一ID
	Level int32
	Num   int32
	Index int32 //1：主武器 2:头盔 3：项链 4：外衣 5：腰带 6：鞋子
	IsUse bool  //是否使用
}

func ToEquipProto(maps []*EquipOption) map[int32]*proto_equip.EquipOption {
	m := make(map[int32]*proto_equip.EquipOption, 0)
	for _, v := range maps {
		m[v.Id] = &proto_equip.EquipOption{
			Id:    v.Id,
			Num:   v.Num,
			Index: v.Index,
			Level: v.Level,
			CId:   v.CId,
			IsUse: v.IsUse,
		}
	}

	return m
}

func ToMountProto(opt *MountOption) *proto_equip.MountOption {
	mounts := make([]*proto_equip.MountItemOption, 0)
	for _, v := range opt.Mount {
		mounts = append(mounts, &proto_equip.MountItemOption{
			Id:    v.Id,
			Name:  v.Name,
			Num:   v.Num,
			Level: v.Level,
		})
	}
	return &proto_equip.MountOption{
		Star:  opt.Star,
		Stage: opt.Stage,
		Exp:   opt.Exp,
		UseId: opt.UseId,
		Mount: mounts,
		EnergyOption: &proto_equip.EnergyOption{
			EnergyLevel: opt.MountEnergy,
		},
	}
}

func ToWeaponryProto(opt *WeaponryOption) *proto_equip.WeaponryOption {
	weaponrys := make([]*proto_equip.WeaponryItemOption, 0)
	for _, v := range opt.WeaponryItems {
		weaponrys = append(weaponrys, &proto_equip.WeaponryItemOption{
			Id:    v.Id,
			Level: v.Level,
			Num:   v.Num,
		})
	}

	return &proto_equip.WeaponryOption{
		Star:      opt.Star,
		Exp:       opt.Exp,
		UseId:     opt.UseId,
		Weaponrys: weaponrys,
	}
}

func ToEnchantProto(opt map[int32]*EnchantOption) map[int32]*proto_equip.EquipEnchantOption {
	enchants := make(map[int32]*proto_equip.EquipEnchantOption)
	for k, v := range opt {
		enchants[int32(k)] = &proto_equip.EquipEnchantOption{
			Id:    v.Id,
			Level: v.Level,
			Exp:   v.Exp,
		}
	}

	return enchants
}

func ToSuccinctIndexProto(opt map[int]*SuccinctIndex) map[int32]*proto_equip.EquipSussinctOption {
	enchants := make(map[int32]*proto_equip.EquipSussinctOption)
	for k, v := range opt {
		enchants[int32(k)] = &proto_equip.EquipSussinctOption{
			Index:   int32(v.Index),
			SkillId: v.SkillId,
			Name:    v.Name,
		}
	}

	return enchants
}

func ToSuccinctCacheProto(opt map[int]*CacheSuccinctIndex) map[int32]*proto_equip.CacheSuccinctOpt {
	enchants := make(map[int32]*proto_equip.CacheSuccinctOpt)
	for k, v := range opt {
		enchants[int32(k)] = &proto_equip.CacheSuccinctOpt{
			Index:      int32(v.Index),
			SkillId:    v.SkillId,
			EquipIndex: int32(v.EquipIndex),
		}
	}

	return enchants
}

// 背饰-灵韵
func ToBraceAuraProto(opt map[int32]*BraceAuraItem) map[int32]*proto_equip.AuraItem {
	auras := make(map[int32]*proto_equip.AuraItem)
	for k, v := range opt {
		auras[k] = &proto_equip.AuraItem{
			Type:  v.Type,
			Level: v.Level,
		}
	}

	return auras
}

// 背饰-背饰
func ToBraceItemProto(opt map[int32]*BraceItem) map[int32]*proto_equip.BraceItem {
	auras := make(map[int32]*proto_equip.BraceItem)
	for k, v := range opt {
		auras[k] = &proto_equip.BraceItem{
			Id:    v.Id,
			Num:   v.Num,
			Use:   v.IsUse,
			Level: v.Level,
		}
	}

	return auras
}

// 背饰-天赋
func ToBraceTalentIndexProto(opt map[int32]*BraceTalentIndex) map[int32]*proto_equip.BraceTalentIndex {
	auras := make(map[int32]*proto_equip.BraceTalentIndex)
	for k, v := range opt {
		auras[k] = &proto_equip.BraceTalentIndex{
			Index:    v.Index,
			Name:     v.Name,
			BraceJob: ToBraceTalentJobProto(v.BraceTalentJobs),
		}
	}

	return auras
}

// 背饰-天赋【职业】
func ToBraceTalentJobProto(opt map[int32]*BraceTalentJob) map[int32]*proto_equip.BraceTalentJob {
	auras := make(map[int32]*proto_equip.BraceTalentJob)
	for k, v := range opt {
		auras[k] = &proto_equip.BraceTalentJob{
			Job:        v.Job,
			BraceGroup: ToBraceTalentGroupProto(v.BraceTalentGroups),
		}
	}
	return auras
}

// 背饰-天赋
func ToBraceTalentGroupProto(opt map[int32]*BraceTalentGroup) map[int32]*proto_equip.BraceTalentGroup {
	auras := make(map[int32]*proto_equip.BraceTalentGroup)
	for k, v := range opt {
		auras[k] = &proto_equip.BraceTalentGroup{
			Group:     v.Group,
			BraceItem: ToBraceTalentItemProto(v.BraceTalentItems),
		}
	}
	return auras
}

// 背饰-天赋
func ToBraceTalentItemProto(opt map[int32]*BraceTalentItem) map[int32]*proto_equip.BraceTalentItem {
	auras := make(map[int32]*proto_equip.BraceTalentItem)
	for k, v := range opt {
		auras[k] = &proto_equip.BraceTalentItem{
			Id:    v.Id,
			Level: v.Level,
			Exp:   v.Exp,
		}
	}
	return auras
}

// 背饰-天赋
func ToBraceTalentSingleIndexProto(opts map[int32]*BraceTalentIndex, index, job, group, id int32) map[int32]*proto_equip.BraceTalentIndex {
	auras := make(map[int32]*proto_equip.BraceTalentIndex)
	opt := opts[index]

	//天赋点
	auraItem := make(map[int32]*proto_equip.BraceTalentItem)
	optItem := opts[index].BraceTalentJobs[job].BraceTalentGroups[group].BraceTalentItems[id]
	auraItem[id] = &proto_equip.BraceTalentItem{
		Id:    optItem.Id,
		Level: optItem.Level,
		Exp:   optItem.Exp,
	}

	//组
	auraGroup := make(map[int32]*proto_equip.BraceTalentGroup)
	auraGroup[group] = &proto_equip.BraceTalentGroup{
		Group:     group,
		BraceItem: auraItem,
	}

	//职业
	auraJob := make(map[int32]*proto_equip.BraceTalentJob)
	auraJob[job] = &proto_equip.BraceTalentJob{
		Job:        job,
		BraceGroup: auraGroup,
	}

	//方案
	auras[opt.Index] = &proto_equip.BraceTalentIndex{
		Index:    opt.Index,
		Name:     opt.Name,
		BraceJob: auraJob,
	}

	return auras
}

// 背饰-天赋
func ToBraceTalentSingleGroupProto(opts map[int32]*BraceTalentGroup, job, id int32) map[int32]*proto_equip.BraceTalentGroup {
	auras := make(map[int32]*proto_equip.BraceTalentGroup)
	opt := opts[job]
	auras[opt.Group] = &proto_equip.BraceTalentGroup{
		Group:     opt.Group,
		BraceItem: ToBraceTalentSingleItemProto(opt.BraceTalentItems, id),
	}
	return auras
}

// 背饰-天赋
func ToBraceTalentSingleItemProto(opts map[int32]*BraceTalentItem, id int32) map[int32]*proto_equip.BraceTalentItem {
	auras := make(map[int32]*proto_equip.BraceTalentItem)
	opt := opts[id]
	auras[opt.Id] = &proto_equip.BraceTalentItem{
		Id:    opt.Id,
		Level: opt.Level,
		Exp:   opt.Exp,
	}

	return auras
}
