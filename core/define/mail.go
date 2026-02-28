package define

const (
	SystemMail = 1 // 全员系统邮件
	PlayerMail = 2 // 个人邮件
)

const (
	MailTypeNormal int32 = 1 // 普通邮件
	MailTypeGuild  int32 = 2 // 联盟邮件
)

const (
	MailExpiration   int64 = 15 // 邮件过期时间
	MailStorageLimit       = 50 // 玩家邮件上限
)

const (
	LanguageEnglish            = "EN"
	LanguageChinese            = "CHS"
	LanguageChineseTraditional = "CHT"
	LANGUAGE_JAPANESE          = "JP"
	LANGUAGE_FRENCH            = "FR"
	LANGUAGE_GERMAN            = "GE"
	LANGUAGE_ITALY             = "IT"
	LANGUAGE_KOREA             = "KR"
	LANGUAGE_RUSSIA            = "RU"
	LANGUAGE_SPANISH           = "SP"
)

const (
	MailAction_Delete           = 1 //删除邮件
	MailAction_ReturnAttachment = 2 //退还附件
	MailAction_GetAward         = 3 //领取邮件
	MailAction_BuyAttachment    = 4 //购买附件
)
