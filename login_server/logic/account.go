package logic

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
	"xfx/login_server/define"
	"xfx/login_server/model"
	"xfx/pkg/log"
)

// Key 加密使用
const Key = "12348578902223367877723456789012"

const TokenExpire = 60 * 60 * 2  // token失效(秒) 2小时
const VerifyCodeExpire = 60 * 10 // 验证码失效(秒)

// 注册账号
func Accountregister(c *gin.Context) {
	var registerUser model.RegisterUser
	if err := c.ShouldBindJSON(&registerUser); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}

	if registerUser.Platform < 1 || registerUser.Platform > 3 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "platform type error")
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
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "register account error")
		return
	}

	p := new(model.Account)

	if has, _ := AccountEngine.IsTableExist(define.AccountTable); !has {
		// 同步结构体与数据库表
		err := AccountEngine.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	p.Account = registerUser.Account
	p.Platform = registerUser.Platform
	uid := ""

	has, err := AccountEngine.Table(define.AccountTable).Where("account = ?", p.Account).Exist()
	if err != nil {
		log.Error("register find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if has {
		httpRetGame(c, ERR_ACCOUNT_EXISTS, "account exists")
		return
	}

	//检测uid重复否 重复则重新拿
	uid = RandomNumeric(10)
	for {
		has, err := AccountEngine.Table(define.AccountTable).Where("uid = ?", uid).Exist()
		if err != nil {
			log.Error("register find uid err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
		if !has {
			break
		}
		uid = RandomNumeric(10)
	}

	now := time.Now()

	p.Uid = uid
	p.Password = registerUser.Password
	p.CreateTime = now
	p.ServerId = registerUser.ServerId

	_, err = AccountEngine.Table(define.AccountTable).Insert(p)
	if err != nil {
		log.Error("register player account db err : %v", err)
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	httpRetGame(c, SUCCESS, "success", map[string]interface{}{"account": p.Account})
}

// 登录
func Login(c *gin.Context) {
	var loginUser model.LoginUser
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" account %v, version %v", loginUser.Account, loginUser.Version)

	// 检查版本号
	//if loginUser.Version != conf.Server.ClientVersion {
	//	httpRetGame(c, ErrAccountClientVersionUnmatched, "client version unmatched")
	//	return
	//}

	//判断平台id是否正确
	if loginUser.Platform < 1 || loginUser.Platform > 100 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "platform type error")
		return
	}

	//账号不能为空
	if len(loginUser.Account) <= 0 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "account not found")
		return
	}

	player := new(model.Account)
	player.Account = loginUser.Account
	has, err := AccountEngine.Table(define.AccountTable).Get(player)
	if err != nil {
		log.Error("account failed, err : %v", err)
		httpRetGame(c, ERR_DB, "db err")
		return
	}
	if !has {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	if player.LoginBan != 0 && player.LoginBan > time.Now().Unix() {
		httpRetGame(c, ERR_ACCOUNT_BANNED, "account banned")
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
		httpRetGame(c, ERR_ACCOUNT_PASSWORD_FAILED, "password failed")
		return
	}

	var serverItem model.ServerItem
	has, err = AccountEngine.Table(define.ServerGroup).Where("id = ?", loginUser.ServerId).Get(&serverItem)
	if err != nil {
		log.Error("get server list find db err :%v", err.Error())
		httpRetGame(c, ERR_DB, "get server list find db err")
		return
	}

	if !has {
		log.Error("get server list id error")
		httpRetGame(c, ERR_DB, "get server list id error")
		return
	}

	//删除老的token
	if len(player.LastToken) > 0 {
		_, err = RedisExec("del", fmt.Sprintf("%s:%v", define.LoginToken, player.LastToken))
		if err != nil {
			log.Error("del last token failed, err : %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	//生成登录token
	loginToken := fmt.Sprintf("%x", RandomNumeric(15))
	player.LastToken = loginToken
	player.OnlineTime = time.Now()

	//把token存入db 还有最近一次登录时间
	n, err := AccountEngine.Table(define.AccountTable).Where("uid = ?", player.Uid).Cols("last_token", "online_time").Update(player)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		log.Error("login save last token failed, err :%v", err)
		return
	}
	if n == 0 {
		log.Error("login save last token failed")
		httpRetGame(c, ERR_DB, "login failed")
		return
	}

	//token存入redis
	_, err = RedisExec("set", fmt.Sprintf("%s:%v", define.LoginToken, loginToken), player.Uid, "ex", fmt.Sprintf("%v", TokenExpire))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	//登录服务器存入redis
	_, err = RedisExec("set", fmt.Sprintf("%s:%v", define.PlayerLastLoginServer, player.RedisId), serverItem.Id)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	log.Debug("login success")

	httpRetGame(c, SUCCESS, "success",
		map[string]any{
			"serverId": serverItem.Id,
			"token":    loginToken,
			"uid":      player.Uid,
		})
}
