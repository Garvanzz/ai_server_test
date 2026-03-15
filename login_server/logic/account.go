package logic

import (
	"fmt"
	"time"
	dto2 "xfx/login_server/dto"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/login_server/internal/middleware"
	"xfx/pkg/log"
	"xfx/pkg/utils"

	"github.com/gin-gonic/gin"
)

const TokenExpire = 60 * 60 * 2  // token 失效(秒) 2 小时
const VerifyCodeExpire = 60 * 10 // 验证码失效(秒)

func Register(c *gin.Context) {
	var registerUser dto2.RegisterUser
	if err := c.ShouldBindJSON(&registerUser); err != nil {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}

	if registerUser.Platform < 1 || registerUser.Platform > 3 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "platform type error")
		return
	}

	//密码加密hash
	//passwordHash, err := bcrypt.GenerateFromPassword([]byte(registerUser.Password), bcrypt.DefaultCost)
	//if err != nil {
	//	httpRetGame(c, ErrDb, "db err")
	//	return
	//}

	//玩家注册的话 账号不能为空
	if len(registerUser.Account) == 0 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "register account error")
		return
	}

	p := new(model.Account)
	p.Account = registerUser.Account
	p.Platform = registerUser.Platform
	uid := ""

	has, err := AccountEngine.Table(define.AccountTable).Where("account = ?", p.Account).Exist()
	if err != nil {
		log.Error("register find err :%v", err.Error())
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}
	if has {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_EXISTS, "account exists")
		return
	}

	now := time.Now()
	serverItem := new(model.ServerItem)
	has, err = AccountEngine.Table(define.GameServerTable).Where("id = ?", registerUser.ServerId).Get(serverItem)
	if err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}
	if !has {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "server not found")
		return
	}
	logicServerId := int(serverItem.LogicServerId)
	if logicServerId <= 0 {
		logicServerId = int(serverItem.Id)
	}

	uid = utils.RandomNumeric(10)
	has, err = AccountEngine.Table(define.AccountTable).Where("uid = ? AND server_id = ?", uid, logicServerId).Exist()
	if err != nil {
		log.Error("register find uid err :%v", err.Error())
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}
	for has {
		uid = utils.RandomNumeric(10)
		has, err = AccountEngine.Table(define.AccountTable).Where("uid = ? AND server_id = ?", uid, logicServerId).Exist()
		if err != nil {
			log.Error("register find uid err :%v", err.Error())
			middleware.RetGame(c, dto2.ERR_DB, err.Error())
			return
		}
	}

	p.Uid = uid
	p.Password = registerUser.Password
	p.CreateTime = now
	p.ServerId = logicServerId
	p.OriginServerId = registerUser.ServerId

	_, err = AccountEngine.Table(define.AccountTable).Insert(p)
	if err != nil {
		log.Error("register player account db err : %v", err)
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	middleware.RetGame(c, dto2.SUCCESS, "success", map[string]interface{}{"account": p.Account})
}

// Login 登录
func Login(c *gin.Context) {
	var loginUser dto2.LoginUser
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" account %v, version %v", loginUser.Account, loginUser.Version)

	// 检查版本号
	//if loginUser.Version != conf.Server.ClientVersion {
	//	httpRetGame(c, ErrAccountClientVersionUnmatched, "client version unmatched")
	//	return
	//}

	// 判断平台 id 是否正确
	if loginUser.Platform < 1 || loginUser.Platform > 100 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "platform type error")
		return
	}

	//账号不能为空
	if len(loginUser.Account) <= 0 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "account not found")
		return
	}

	player := new(model.Account)
	has, err := AccountEngine.Table(define.AccountTable).Where("account = ?", loginUser.Account).Get(player)
	if err != nil {
		log.Error("account failed, err : %v", err)
		middleware.RetGame(c, dto2.ERR_DB, "db err")
		return
	}
	if !has {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	if player.LoginBan != 0 && player.LoginBan > time.Now().Unix() {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_BANNED, "account banned")
		return
	}

	//-------开始密码效验--------
	//拿密码 先dehex一手
	//passByte, err := hex.DecodeString(loginUser.Password)
	//if err != nil {
	//	log.Error("login decrypt password hex err : %v", err)
	//	httpRetGame(c, ErrAccountParamsError, "params error2")
	//	return
	//}
	//
	////密码需要再decrypt一次
	//passwordRaw, err := crypto.AesPkcs7Decrypt(passByte, []byte(Key))
	//if err != nil {
	//	log.Error("login decrypt password failed, err : %v", err)
	//	httpRetGame(c, ErrAccountParamsError, "params error3")
	//	return
	//}

	// TODO:跳过密码校验

	//需要对比密码是否正确
	//err = bcrypt.CompareHashAndPassword([]byte(player.Password), passwordRaw)
	//if err != nil {
	//	httpRetGame(c, ErrAccountPasswordFailed, "password failed")
	//	return
	//}

	//验证密码
	if player.Password != loginUser.Password {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PASSWORD_FAILED, "password failed")
		return
	}

	var serverItem model.ServerItem
	has, err = AccountEngine.Table(define.GameServerTable).Where("id = ?", loginUser.ServerId).Get(&serverItem)
	if err != nil {
		log.Error("get server list find db err :%v", err.Error())
		middleware.RetGame(c, dto2.ERR_DB, "get server list find db err")
		return
	}

	if !has {
		log.Error("get server list id error")
		middleware.RetGame(c, dto2.ERR_DB, "get server list id error")
		return
	}

	//删除老的token
	if len(player.LastToken) > 0 {
		_, err = RedisExec("del", fmt.Sprintf("%s:%v", define.LoginToken, player.LastToken))
		if err != nil {
			log.Error("del last token failed, err : %v", err)
			middleware.RetGame(c, dto2.ERR_DB, err.Error())
			return
		}
	}

	logicServerId := int(serverItem.LogicServerId)
	if logicServerId <= 0 {
		logicServerId = int(serverItem.Id)
	}

	// 上次登录服（更新前）供客户端选服展示
	lastLoginServerId := int64(player.ServerId)

	// 生成登录 token，并更新账号 last_token、online_time、server_id（本次所选服）
	loginToken := fmt.Sprintf("%x", []byte(utils.RandomNumeric(15)))
	player.LastToken = loginToken
	player.OnlineTime = time.Now()
	player.ServerId = logicServerId
	if player.OriginServerId == 0 {
		player.OriginServerId = loginUser.ServerId
	}

	n, err := AccountEngine.Table(define.AccountTable).Where("id = ?", player.Id).
		Cols("last_token", "online_time", "server_id", "origin_server_id").Update(player)
	if err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		log.Error("login save last token failed, err :%v", err)
		return
	}
	if n == 0 {
		log.Error("login save last token failed")
		middleware.RetGame(c, dto2.ERR_DB, "login failed")
		return
	}

	// token 存入 redis
	_, err = RedisExec("set", fmt.Sprintf("%s:%v", define.LoginToken, loginToken), player.Uid, "ex", fmt.Sprintf("%v", TokenExpire))
	if err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	// 按 uid 记录上次登录服（渠道与服解耦：不依赖 RedisId，login_server 即可维护）
	_, err = RedisExec("set", fmt.Sprintf("%s:uid:%s", define.PlayerLastLoginServer, player.Uid), logicServerId, "ex", fmt.Sprintf("%v", TokenExpire*2))
	if err != nil {
		log.Error("login save last server by uid failed, err :%v", err)
		// 不阻断登录
	}

	log.Debug("login success")

	middleware.RetGame(c, dto2.SUCCESS, "success",
		map[string]any{
			"serverId":          logicServerId,
			"entryServerId":     serverItem.Id,
			"token":             loginToken,
			"uid":               player.Uid,
			"lastLoginServerId": lastLoginServerId,
		})
}
