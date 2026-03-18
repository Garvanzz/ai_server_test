package login

import (
	"fmt"
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/core/multiserver"
	"xfx/main_server/messages"
	"xfx/main_server/player"
	"xfx/pkg/agent"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/proto/proto_public"

	"github.com/gomodule/redigo/redis"
)

var Module = func() module.Module {
	login := new(Login)
	login.players = make(map[int64]agent.PID)
	return login
}

type Login struct {
	modules.BaseModule
	players map[int64]agent.PID
}

func loadAccountRoleByPayload(payload multiserver.LoginTokenPayload, logicServerId int) (*model.AccountRole, error) {
	role := new(model.AccountRole)
	var has bool
	var err error
	switch {
	case payload.RoleID > 0:
		has, err = db.Engine.Mysql.Table(define.AccountRoleTable).Where("id = ?", payload.RoleID).Get(role)
	case payload.UID != "" && payload.EntryServerID > 0:
		has, err = db.Engine.Mysql.Table(define.AccountRoleTable).Where("uid = ? AND entry_server_id = ?", payload.UID, payload.EntryServerID).Get(role)
	default:
		return nil, nil
	}
	if err != nil || !has {
		return nil, err
	}
	if logicServerId > 0 {
		role.LogicServerId = logicServerId
	}
	return role, nil
}

func (l *Login) OnInit(app module.App) {
	l.BaseModule.OnInit(app)
	l.Register("login", l.login)
	l.Register("logout", l.logout)
	l.Register("disconnect", l.disconnect)
	l.Register("castAgent", l.castAgent)
	l.Register("castAgents", l.castAgents)
	l.Register("boardCast", l.boardCast)
	l.Register("isOnline", l.isOnline)
	l.Register("getPlayerPid", l.getPlayerPid)
}

func (l *Login) GetType() string { return define.ModuleLogin }

func (l *Login) OnTick(delta time.Duration) {}

func (l *Login) OnMessage(msg any) any {
	log.Debug("* login message %v", msg)
	return nil
}

// 获取玩家是否在线
func (l *Login) isOnline(id int64) bool {
	_, exist := l.players[id]
	return exist
}

// 玩家登录
func (l *Login) login(msg *messages.Login) (*messages.LoginResult, error) {
	resp := new(messages.LoginResult)
	resp.Result = int(proto_public.CommonState_Faild)

	// 从登录服 Redis 取 uid（与 login_server 共用，token 由登录服写入）
	rawTokenValue, err := redis.String(db.RedisLoginExec("get", fmt.Sprintf("%s:%s", define.LoginToken, msg.Request.Token)))
	if err != nil {
		log.Error("login token lookup error:%v", err)
		return resp, err
	}
	payload := multiserver.DecodeLoginTokenPayload(rawTokenValue)
	uid := payload.UID
	if uid == "" {
		log.Error("login token payload missing uid")
		return resp, nil
	}
	role, err := loadAccountRoleByPayload(payload, l.App.GetEnv().ID)
	if err != nil {
		log.Error("load account role error:%v", err)
		return resp, err
	}

	var pid agent.PID
	var dbId int64
	entryServerId := l.App.GetEnv().ID
	if payload.EntryServerID > 0 {
		entryServerId = payload.EntryServerID
	}
	roleRedisKey := multiserver.AccountRoleRedisKey(uid, entryServerId)
	reply, err := db.RedisExec("get", roleRedisKey)
	if err != nil {
		log.Error("login db error:%v", err)
		return resp, err
	}
	if reply == nil && payload.EntryServerID == 0 {
		reply, err = db.RedisExec("get", fmt.Sprintf("%s:%s", define.Account, uid))
		if err != nil {
			log.Error("login legacy db error:%v", err)
			return resp, err
		}
	}
	if role != nil && role.RedisId > 0 {
		dbId = role.RedisId
		if reply == nil {
			_, _ = db.RedisExec("set", roleRedisKey, dbId)
		}
	} else if reply != nil {
		dbId, err = redis.Int64(reply, nil)
		if err != nil {
			log.Error("login convert player id error:%v", err)
			return resp, err
		}
	}

	if dbId == 0 {
		log.Debug("无账号玩家=====")
		playerData, err := player.Born(uid, l.App.GetEnv().ID, entryServerId)
		if err != nil {
			log.Error("born db error:%v", err)
			return resp, err
		}

		_, err = db.RedisExec("set", roleRedisKey, playerData.Id)
		if err != nil {
			return resp, err
		}

		if role != nil {
			role.NickName = playerData.Base.Name
			role.RedisId = playerData.Id
			role.LogicServerId = l.App.GetEnv().ID
			role.LastLoginTime = time.Now()
			if _, err = db.Engine.Mysql.Table(define.AccountRoleTable).Where("id = ?", role.Id).
				Cols("nick_name", "redis_id", "logic_server_id", "last_login_time").Update(role); err != nil {
				log.Error("player born update mysql error:%v,uid:%v", err, uid)
				return resp, err
			}
		}

		pl := player.New(playerData, msg.Session, l.App)
		// 落库一次
		pl.OnSave(false)

		pid, err = l.Context.Create(
			fmt.Sprintf("player#%d", playerData.Id),
			pl,
		)

		if err != nil {
			log.Error("new player actor error:%v", err)
			return resp, err
		}

		dbId = playerData.Id
		pl.OnInit(l.App)
		l.players[playerData.Id] = pid
	} else { // 加载玩家数据
		log.Info("dbId:%v, %v", dbId, uid)
		log.Debug("有账号玩家:%v", dbId)

		var ok = false
		pid, ok = l.players[dbId]
		if ok { // 玩家在线顶号
			_, err = l.Context.Call(pid, &messages.LoginReplace{
				Session: msg.Session,
			})
			if err != nil {
				log.Error("login replace error:%v", err)
				return resp, err
			}
		} else {
			playerData, err := player.LoadPlayerData(dbId)
			if err != nil {
				log.Error("login load player data error:%v, playerId:%d, uid:%s (possible merge data not synced or redis key missing)", err, dbId, uid)
				return resp, err
			}
			playerData.SetProp(define.PlayerPropServerId, int64(l.App.GetEnv().ID), false)
			playerData.SetProp(define.PlayerPropEntryServerId, int64(entryServerId), false)
			if role != nil {
				role.LogicServerId = l.App.GetEnv().ID
				role.LastLoginTime = time.Now()
				if playerData.Base != nil {
					role.NickName = playerData.Base.Name
				}
				_, _ = db.Engine.Mysql.Table(define.AccountRoleTable).Where("id = ?", role.Id).
					Cols("logic_server_id", "last_login_time", "nick_name").Update(role)
			}

			pl := player.New(playerData, msg.Session, l.App)
			pid, err = l.Context.Create(
				fmt.Sprintf("player#%d", playerData.Id),
				pl,
			)

			if err != nil {
				log.Error("new player actor error:%v", err)
				return resp, err
			}

			pl.OnInit(l.App)
			l.players[playerData.Id] = pid
		}
	}

	resp.PlayerId = dbId
	resp.PlayerPid = pid
	resp.Result = int(proto_public.CommonState_Success)
	return resp, err
}

// 获取玩家pid
func (l *Login) getPlayerPid(dbId int64) agent.PID {
	if pid, ok := l.players[dbId]; ok {
		return pid
	}

	return nil
}

// 玩家登出 — Cast 异步通知 Player，避免阻塞 Login Actor
func (l *Login) logout(playerId int64) error {
	pid, ok := l.players[playerId]
	if ok {
		l.Context.Cast(pid, &messages.Logout{})
		delete(l.players, playerId)
	}
	return nil
}

// 断开连接 — Cast 异步通知 Player，避免阻塞 Login Actor
func (l *Login) disconnect(playerId int64) error {
	pid, ok := l.players[playerId]
	if ok {
		l.Context.Cast(pid, &messages.Disconnect{})
		delete(l.players, playerId)
	}
	return nil
}

// 消息广播
func (l *Login) boardCast(msg any) {
	for _, pid := range l.players {
		l.Context.Cast(pid, msg)
	}
}

// 消息转发
func (l *Login) castAgent(playerId int64, msg any) {
	pid, ok := l.players[playerId]
	if ok {
		l.Context.Cast(pid, msg)
	}
}

// 消息转发
func (l *Login) castAgents(playerIds []int64, msg any) {
	for i := 0; i < len(playerIds); i++ {
		playerId := playerIds[i]
		pid, ok := l.players[playerId]
		if ok {
			l.Context.Cast(pid, msg)
		}
	}
}
