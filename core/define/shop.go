package define

// 商品类型
const (
	SHOPTYPE_DAYTEHUI               = 1  //每日特惠
	SHOPTYPE_DAYLIMIT               = 2  //每日限购
	SHOPTYPE_WEEKLIMIT              = 3  //每周限购     = 3
	SHOPTYPE_XIANYURECHARGE         = 4  //仙玉充值
	SHOPTYPE_LINGYURECHARGE         = 5  //灵玉充值
	SHOPTYPE_YONGJIURECHARGE        = 6  //永久充值
	SHOPTYPE_FASHIONSHOP            = 7  //时装商城
	SHOPTYPE_LINGYUSHOP_LIMITDAY    = 8  //灵玉商城【每日】
	SHOPTYPE_QUICKBUY               = 9  //快捷购买
	SHOPTYPE_LINGYUSHOP             = 10 //灵玉商城【无】
	SHOPTYPE_GEMAPPRAISALGIFT       = 11 //鉴宝礼包
	SHOPTYPE_GEMAPPRAISALLMONTHCATD = 12 //鉴宝月卡
	SHOPTYPE_PETEQUIPWEEK           = 14 //宠物装备-每周限购
	SHOPTYPE_PETEQUIPZHONGSHEN      = 15 //宠物装备-终身限购
	SHOPTYPE_PETGIFTDAY             = 16 //宠物礼包-每日
	SHOPTYPE_PETGIFTWEEK            = 17 //宠物礼包-每周
	SHOPTYPE_NORMALMONTHCARD        = 18 //常规月卡
	SHOPTYPE_MAINLINEFUND           = 19 //主线基金
	SHOPTYPE_LEVELFUND              = 20 //成长基金
	SHOPTYPE_BOXFUND                = 21 //宝箱基金
	SHOPTYPE_PASSPORT               = 31 //通行证
	SHOPTYPE_PASSPORT_ADVANCE       = 32 //通行证高级
	SHOPTYPE_PASSPORT_SCOREGIFT     = 33 //通行证积分礼包
)

const (
	ShopLimitTypeNull    = 0 //无
	ShopLimitTypeDay     = 1 //每天
	ShopLimitTypeWeek    = 2 //每周
	ShopLimitTypeYongJiu = 3 //永久
	ShopLimitTypeMonth   = 4 //每月
	ShopLimitTypeSeason  = 5 //赛季
)

var ShopRefreshType = map[int32]int{
	// 每日刷新
	SHOPTYPE_DAYTEHUI:            ShopLimitTypeDay,
	SHOPTYPE_DAYLIMIT:            ShopLimitTypeDay,
	SHOPTYPE_PETGIFTDAY:          ShopLimitTypeDay,
	SHOPTYPE_LINGYUSHOP_LIMITDAY: ShopLimitTypeDay,

	// 周刷新
	SHOPTYPE_WEEKLIMIT:    ShopLimitTypeWeek,
	SHOPTYPE_PETGIFTWEEK:  ShopLimitTypeWeek,
	SHOPTYPE_PETEQUIPWEEK: ShopLimitTypeWeek,

	// 月刷新
	SHOPTYPE_GEMAPPRAISALLMONTHCATD: ShopLimitTypeMonth,
	SHOPTYPE_NORMALMONTHCARD:        ShopLimitTypeMonth,

	//赛季刷新
	SHOPTYPE_PASSPORT:           ShopLimitTypeSeason,
	SHOPTYPE_PASSPORT_ADVANCE:   ShopLimitTypeSeason,
	SHOPTYPE_PASSPORT_SCOREGIFT: ShopLimitTypeSeason,

	// 不刷新
	SHOPTYPE_XIANYURECHARGE:    ShopLimitTypeNull,
	SHOPTYPE_LINGYURECHARGE:    ShopLimitTypeNull,
	SHOPTYPE_YONGJIURECHARGE:   ShopLimitTypeNull,
	SHOPTYPE_FASHIONSHOP:       ShopLimitTypeNull,
	SHOPTYPE_QUICKBUY:          ShopLimitTypeNull,
	SHOPTYPE_LINGYUSHOP:        ShopLimitTypeNull,
	SHOPTYPE_GEMAPPRAISALGIFT:  ShopLimitTypeNull,
	SHOPTYPE_PETEQUIPZHONGSHEN: ShopLimitTypeNull,
	SHOPTYPE_MAINLINEFUND:      ShopLimitTypeNull,
	SHOPTYPE_LEVELFUND:         ShopLimitTypeNull,
	SHOPTYPE_BOXFUND:           ShopLimitTypeNull,
}
