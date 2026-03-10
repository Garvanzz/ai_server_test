package define

const (
	AccountTable        = "account"          // 玩家账号表
	AdminTable          = "admin"            // GM 后台管理员表
	AdminMailTable      = "admin_mail"       // 后台邮件表
	PlayerMailInfoTable = "player_mail_info" // 个人邮件表
	SysMailInfoTable    = "sys_mail_info"    // 系统邮件表
	ServerGroupTable    = "server_group"     // 区服组表（仅分组：id, name, sort_order）
	GameServerTable     = "game_server"      // 区服表（区服 group_id>0；游戏服进程 group_id=0）
	PayCacheOrderTable  = "pay_cache_order"  // 支付缓存订单信息
	PayOrderTable       = "pay_order"        // 支付缓存订单信息
	HotUpdateTable      = "hot_update"       // 热更
	NoticeTable         = "notice"           // 公告
	GuildTable          = "guild"            // 公会
	GuildApplyTable     = "guild_apply"      // 公会申请
	GuildLogTable       = "guild_log"        // 公会日志
	FriendApplyTable    = "friend_apply"     // 好友申请
	FriendBlockTable    = "friend_block"     // 好友黑名单
)
