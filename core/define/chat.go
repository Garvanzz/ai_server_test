package define

const (
	ChatTypeKuafu    = 1 //跨服
	ChatTypeWorld    = 2 //世界
	ChatTypeZudui    = 3 //组队
	ChatTypePrivate  = 4 //私聊
	ChatTypeGuild    = 5 //帮会
	ChatTypeChuanwen = 6 //传闻
)

const ChatMsgLen = 30

const (
	PrivateChatDataKey    = "private_chat_data:"
	PrivateChatExpiration = 86400 * 3
	MsgMaxLen             = 20
)

const (
	ChuanwenMsg_DrawCard = "恭喜%s获得%d品质的%s"
)
