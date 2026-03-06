package gm_model

type GMLogin struct {
	UserName string
	Password string
}

type GMUserInfo struct {
	Xiaoxiaoxiyou string
	UserName      string
}

type GMRespUserInfo struct {
	Permissions []string
	Username    string
	Avatar      []string
}

type GMGetServerList struct {
	Name string
	Id   int64
	Time string
}

type GmAccount struct {
	Id         int64  `json:"id"`
	UserName   string `json:"username"` //账号
	Password   string `json:"password"` //密码
	Token      string
	Permission int
	Name       string
}

type GmReqPlayerInfo struct {
	Uid      string
	ServerId int
}

type GMPlayerInfo struct {
	Id          int64  `redis:"id"`
	Uid         string `redis:"uid"`
	Name        string `redis:"name"`         // 名字
	Level       int32  `redis:"level"`        // 等级
	Exp         int32  `redis:"exp"`          // 经验
	FaceId      int32  `redis:"face_id"`      // 头像ID
	FaceSlotId  int32  `redis:"face_slot_id"` // 头像框id
	OfflineTime int64  `redis:"offline_time"` // 上次登录时间
	Rank        int32  `redis:"rank"`         // 段位0：无
	Title       int32  `redis:"title"`        // 称号
	Job         int32  `redis:"job"`          // 职业
	Sex         int32  `redis:"sex"`          // 性别
	Clan        string `redis:"clan"`         // 帮会
	ClanId      int32  `redis:"clanid"`       // 帮会ID
	HeroId      int32  `redis:"heroId"`       // 主角ID
}

// 背包
type GmReqPlayerBag struct {
	Uid      string
	ServerId int
	ItemId   int
	ItemNum  int
	Ids      []int
}

type GmRespPlayerBag struct {
	ItemId   int32
	ItemName string
	ItemNum  int32
}

type Bag struct {
	Items map[int32]int32
}

// 装备
type GmReqPlayerEquip struct {
	Uid      string
	ServerId int
	EquipId  int32
	Ids      []int
}

type GmRespPlayerEquip struct {
	EquipId    int32
	EquipCId   int32
	EquipNum   int32
	EquipName  string
	EquipLevel int32
	EquipIndex string
	EquipIsUse bool
}

type Equip struct {
	Equips   []*EquipOption           //装备列表
	Mount    *MountOption             //坐骑
	Weaponry *WeaponryOption          //神兵
	Enchant  map[int32]*EnchantOption //附魔
}

type EnchantOption struct {
	Id    int32 //符咒ID
	Level int32 //等级
	Exp   int32 //经验
}

type MountOption struct {
	Stage       int32
	Star        int32
	Exp         int32
	UseId       int32
	Mount       []*MountItemOption //坐骑
	MountEnergy map[int32]int32    //坐骑赋能
}

type MountItemOption struct {
	Name string
	Id   int32
	Num  int32
}

type WeaponryOption struct {
	Star          int32
	Exp           int32
	UseId         int32
	WeaponryItems []*WeaponryItem
}

type WeaponryItem struct {
	Id    int32
	Level int32
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

// 角色
type GmReqPlayerHero struct {
	Uid      string
	ServerId int
	HeroId   int32
	Data     *GMEditHeroOption
}

type GMEditHeroOption struct {
	HeroId    int32
	HeroLevel int32
	HeroStage int32
	HeroStar  int32
	HeroExp   int32
}

type GMModHero struct {
	Hero map[int32]*GMModHeroOption
	Skin map[int32]*GMModSkinOption
}

type GMModHeroOption struct {
	Id          int32
	Star        int32
	Level       int32
	Stage       int32
	Exp         int32
	Skin        string
	Cultivation map[int32]int32 //修为
}

type GMModSkinOption struct {
	Id    int32
	SrcId int32
}

type GMRespHero struct {
	HeroId         int32
	HeroName       string
	HeroLevel      int32
	HeroStage      int32
	HeroStar       int32
	HeroExp        int32
	HeroIsUse      string
	HeroIsMainHero string
}

// 布阵
type GMLineUp struct {
	LineUps map[int32]*GMLineUpOption
}

type GMLineUpOption struct {
	Type   int32
	HeroId []int32
}

// 服务器中心
type GMResServerCenter struct {
	ServerType  string
	ServerName  string
	ServerState string
}

// 游戏服务器列表
type GMRespServerItem struct {
	Channel        int
	Ip             string
	Port           int
	RedisPort      int
	MysqlAddr      string
	ServerState    string
	OpenServerTime string
	StopServerTime string
	ServerName     string
	Id             int64
	LoginServerUrl string
	Group          int
	RunState       string
	GameServer     string
	GameServerId   int
}

// 游戏游戏服务器列表
type GMGameRespServerItem struct {
	Ip         string
	Port       int
	RedisPort  int
	MysqlAddr  string
	ServerName string
	Id         int64
	RunState   string
}

// 订单信息
type RechargeOrder struct {
	Amount        float32 `json:"amount"`
	ProductId     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	UserId        string  `json:"user_id"`
	OrderId       string  `json:"order_id"`
	GameUserId    string  `json:"game_user_id"`
	ServerId      string  `json:"server_id"`
	PaymentTime   string  `json:"payment_time"`
	ChannelNumber string  `json:"channel_number"`
}

type GMRespRechargeOrder struct {
	Amount        string
	ProductId     string
	ProductName   string
	UserId        string
	OrderId       string
	GameUserId    string
	ServerId      string
	PaymentTime   string
	ChannelNumber string
	Award         string
}
