package logic

import (
	"fmt"
	"time"
	dto2 "xfx/login_server/dto"

	"github.com/gin-gonic/gin"
	"xorm.io/xorm"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/core/multiserver"
	"xfx/login_server/internal/middleware"
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

const TokenExpire = 60 * 60 * 2  // token 失效(秒) 2 小时
const VerifyCodeExpire = 60 * 10 // 验证码失效(秒)

func getLogicServerID(serverItem *model.ServerItem) int {
	logicServerId := int(serverItem.LogicServerId)
	if logicServerId <= 0 {
		logicServerId = int(serverItem.Id)
	}
	return logicServerId
}

func ensureAccountRole(session *xorm.Session, account *model.Account, entryServerId int, logicServerId int, now time.Time) (*model.AccountRole, error) {
	role := new(model.AccountRole)
	has, err := session.Table(define.AccountRoleTable).Where("uid = ? AND entry_server_id = ?", account.Uid, entryServerId).Get(role)
	if err != nil {
		return nil, err
	}
	if has {
		return role, nil
	}

	role.AccountId = account.Id
	role.Uid = account.Uid
	role.EntryServerId = entryServerId
	role.LogicServerId = logicServerId
	role.OriginServerId = entryServerId
	role.CreateTime = now
	role.LastLoginTime = now
	if _, err = session.Table(define.AccountRoleTable).Insert(role); err != nil {
		return nil, err
	}
	return role, nil
}

func Register(c *gin.Context) {
	var registerUser dto2.RegisterRequest
	if err := c.ShouldBindJSON(&registerUser); err != nil {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}

	if registerUser.Platform < 1 || registerUser.Platform > 3 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "platform type error")
		return
	}
	if len(registerUser.Account) == 0 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "register account error")
		return
	}

	serverItem := new(model.ServerItem)
	has, err := AccountEngine.Table(define.GameServerTable).Cols("logic_server_id").Where("id = ?", registerUser.ServerID).Get(serverItem)
	if err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}
	if !has {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "server not found")
		return
	}
	logicServerId := getLogicServerID(serverItem)
	now := time.Now()

	session := AccountEngine.NewSession()
	defer session.Close()
	if err = session.Begin(); err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	has, err = session.Table(define.AccountTable).Where("account = ?", registerUser.Account).Exist()
	if err != nil {
		_ = session.Rollback()
		log.Error("register find err :%v", err.Error())
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}
	if has {
		_ = session.Rollback()
		middleware.RetGame(c, dto2.ERR_ACCOUNT_EXISTS, "account exists")
		return
	}

	uid := utils.RandomNumeric(10)
	has, err = session.Table(define.AccountTable).Where("uid = ?", uid).Exist()
	if err != nil {
		_ = session.Rollback()
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}
	for has {
		uid = utils.RandomNumeric(10)
		has, err = session.Table(define.AccountTable).Where("uid = ?", uid).Exist()
		if err != nil {
			_ = session.Rollback()
			middleware.RetGame(c, dto2.ERR_DB, err.Error())
			return
		}
	}

	account := &model.Account{
		Uid:        uid,
		Account:    registerUser.Account,
		Password:   registerUser.Password,
		Platform:   registerUser.Platform,
		CreateTime: now,
	}
	if _, err = session.Table(define.AccountTable).Insert(account); err != nil {
		_ = session.Rollback()
		log.Error("register player account db err : %v", err)
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	if _, err = ensureAccountRole(session, account, registerUser.ServerID, logicServerId, now); err != nil {
		_ = session.Rollback()
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	if err = session.Commit(); err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	middleware.RetGameData(c, dto2.SUCCESS, "success", dto2.RegisterResponse{Account: account.Account}, map[string]interface{}{"account": account.Account})
}

// Login 登录
func Login(c *gin.Context) {
	var loginUser dto2.LoginRequest
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" account %v, version %v", loginUser.Account, loginUser.Version)

	if loginUser.Platform < 1 || loginUser.Platform > 100 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "platform type error")
		return
	}
	if len(loginUser.Account) <= 0 {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "account not found")
		return
	}

	account := new(model.Account)
	has, err := AccountEngine.Table(define.AccountTable).Where("account = ?", loginUser.Account).Get(account)
	if err != nil {
		log.Error("account failed, err : %v", err)
		middleware.RetGame(c, dto2.ERR_DB, "db err")
		return
	}
	if !has {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}
	if account.LoginBan != 0 && account.LoginBan > time.Now().Unix() {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_BANNED, "account banned")
		return
	}
	if account.Password != loginUser.Password {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PASSWORD_FAILED, "password failed")
		return
	}

	serverItem := new(model.ServerItem)
	has, err = AccountEngine.Table(define.GameServerTable).Cols("logic_server_id").Where("id = ?", loginUser.ServerID).Get(serverItem)
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

	logicServerId := getLogicServerID(serverItem)
	now := time.Now()
	lastLoginServerId := int64(account.LastLoginEntryServerId)

	session := AccountEngine.NewSession()
	defer session.Close()
	if err = session.Begin(); err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	role, err := ensureAccountRole(session, account, loginUser.ServerID, logicServerId, now)
	if err != nil {
		_ = session.Rollback()
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	if len(role.LastToken) > 0 {
		if _, err = RedisExec("del", fmt.Sprintf("%s:%v", define.LoginToken, role.LastToken)); err != nil {
			_ = session.Rollback()
			log.Error("del last token failed, err : %v", err)
			middleware.RetGame(c, dto2.ERR_DB, err.Error())
			return
		}
	}

	loginToken := fmt.Sprintf("%x", []byte(utils.RandomNumeric(15)))
	role.LastToken = loginToken
	role.OnlineTime = now
	role.LastLoginTime = now
	role.LogicServerId = logicServerId
	if role.OriginServerId == 0 {
		role.OriginServerId = loginUser.ServerID
	}
	if _, err = session.Table(define.AccountRoleTable).Where("id = ?", role.Id).
		Cols("last_token", "online_time", "last_login_time", "logic_server_id", "origin_server_id").Update(role); err != nil {
		_ = session.Rollback()
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	account.OnlineTime = now
	account.LastLoginEntryServerId = loginUser.ServerID
	account.LastLoginLogicServerId = logicServerId
	if _, err = session.Table(define.AccountTable).Where("id = ?", account.Id).
		Cols("online_time", "last_login_entry_server_id", "last_login_logic_server_id").Update(account); err != nil {
		_ = session.Rollback()
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		log.Error("login save account state failed, err :%v", err)
		return
	}

	if err = session.Commit(); err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	payload := multiserver.LoginTokenPayload{
		UID:           account.Uid,
		AccountID:     account.Id,
		RoleID:        role.Id,
		EntryServerID: loginUser.ServerID,
		LogicServerID: logicServerId,
		PlayerID:      role.RedisId,
		IssuedAt:      now.Unix(),
	}
	_, err = RedisExec("set", fmt.Sprintf("%s:%v", define.LoginToken, loginToken), multiserver.EncodeLoginTokenPayload(payload), "ex", fmt.Sprintf("%v", TokenExpire))
	if err != nil {
		middleware.RetGame(c, dto2.ERR_DB, err.Error())
		return
	}

	_, err = RedisExec("set", fmt.Sprintf("%s:uid:%s", define.PlayerLastLoginServer, account.Uid), loginUser.ServerID, "ex", fmt.Sprintf("%v", TokenExpire*2))
	if err != nil {
		log.Error("login save last server by uid failed, err :%v", err)
	}

	log.Debug("login success")

	resp := dto2.LoginResponse{
		ServerID:          logicServerId,
		EntryServerID:     serverItem.Id,
		RoleID:            role.Id,
		Token:             loginToken,
		UID:               account.Uid,
		LastLoginServerID: lastLoginServerId,
	}
	middleware.RetGameData(c, dto2.SUCCESS, "success", resp, map[string]any{
		"serverId":          logicServerId,
		"entryServerId":     serverItem.Id,
		"roleId":            role.Id,
		"token":             loginToken,
		"uid":               account.Uid,
		"lastLoginServerId": lastLoginServerId,
	})
}
