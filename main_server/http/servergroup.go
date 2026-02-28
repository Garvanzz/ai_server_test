package http

const (
	SUCCESS                              = 0
	ERR_SERVER_INTERNAL                  = 1    // 服务器内部错误
	ERR_PAY_ORDER_NOT_FOUND              = 2    // 订单不存在
	ERR_PAY_SIGN                         = 3    // 签名不正确
	ERR_INVITE_CODE_FAIL                 = 29   // 邀请码验证不通过
	ERR_DB                               = 1801 // 数据库错误
	ERR_ACCOUNT_EXISTS                   = 1802 // 账号已存在
	ERR_ACCOUNT_PASSWORD_FAILED          = 1803 // 账号密码错误
	ERR_ACCOUNT_TYPE_UNKNOWN             = 1804 // 账号类型错误
	ERR_ACCOUNT_NOT_FOUND                = 1805 // 账号不存在
	ERR_ACCOUNT_VERIFY_CODE_INCORRECT    = 1806 // 验证码不正确
	ERR_ACCOUNT_GET_VERIFY_CODE_FAILED   = 1807 // 获取验证码失败
	ERR_ACCOUNT_REGISTER_CLOSED          = 1808 // 注册服务关闭
	ERR_ACCOUNT_LOGIN_SERVER_MAINTAIN    = 1809 // 服务器维护 其实只有白名单账号可以进
	ERR_ACCOUNT_BANNED                   = 1810 // 账号被ban中
	ERR_ACCOUNT_PARAMS_ERROR             = 1811 // 参数错误
	ERR_ACCOUNT_CLIENT_VERSION_UNMATCHED = 1812 // 客户端版本不匹配
	ERR_ACCOUNT_SDK_TOKEN_AUTH_FAILED    = 1813 // 登录SDK Token效验失败
	ERR_ACCOUNT_SDK_TOKEN_EXPIRED        = 1814 // 登录SDK Token过期
	ERR_ACCOUNT_HAS_NO_NFT_HERO          = 1815 // 帐号没有nft英雄
	ERR_ACCOUNT_FORCED_OFFLINE           = 1816 // 帐号强制下线中
)
