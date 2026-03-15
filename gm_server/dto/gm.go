package dto

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

// GMGetServerList 区服列表项（与 login 区服设计一致，数据来自 game_server）
type GMGetServerList struct {
	Name      string `json:"name"`
	Id        int64  `json:"id"`
	Time      string `json:"time"`
	GroupId   int    `json:"groupId,omitempty"`   // 区服组 ID
	GroupName string `json:"groupName,omitempty"` // 区服组名称
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

// GMRespServerItem 区服详情（含进程状态），数据来自 game_server 区服行
type GMRespServerItem struct {
	Id                int64  `json:"id"`
	LogicServerId     int64  `json:"logicServerId"`
	MergeState        int    `json:"mergeState"`
	MergeStateText    string `json:"mergeStateText"`
	MergeTime         int64  `json:"mergeTime"`
	ServerName        string `json:"serverName"`
	GroupId           int    `json:"groupId"`   // 区服组 ID
	GroupName         string `json:"groupName"` // 区服组名称
	Channel           int    `json:"channel"`
	Ip                string `json:"ip"`
	Port              int    `json:"port"`
	MainServerHttpUrl string `json:"mainServerHttpUrl"` // 大厅服 HTTP 地址，GM 转发用
	ServerState       string `json:"serverState"`       // 正常/拥挤/爆满/维护/未开服/停服
	OpenServerTime    string `json:"openServerTime"`
	StopServerTime    string `json:"stopServerTime"`
	RunState          string `json:"runState"` // 运行中/离线（大厅服进程）
}

// GMGameRespServerItem 游戏服进程简要信息（game_server 表 group_id=0）
type GMGameRespServerItem struct {
	Id         int64  `json:"id"`
	ServerName string `json:"serverName"`
	ExeName    string `json:"exeName"`
	ExePath    string `json:"exePath"`
	RunState   string `json:"runState"`
}

// 订单信息
type RechargeOrder struct {
	Amount        float32 `json:"amount"`
	ProductId     string  `json:"product_id"`
	ProductName   string  `json:"product_name"`
	UserId        string  `json:"user_id"`
	OrderId       string  `json:"order_id"`
	GameUserId    string  `json:"game_user_id"`
	ServerId      int     `json:"server_id"`
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
	ServerId      int
	PaymentTime   string
	ChannelNumber string
	Award         string
}
