package login

import (
	"fmt"
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
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
	uid, err := redis.String(db.RedisLoginExec("get", fmt.Sprintf("%s:%s", define.LoginToken, msg.Request.Token)))
	if err != nil {
		log.Error("login token lookup error:%v", err)
		return resp, err
	}

	// 本服 Redis：uid -> 本服 playerId（首次进服时创建并写入）
	reply, err := db.RedisExec("get", fmt.Sprintf("%s:%s", define.Account, uid))
	if err != nil {
		log.Error("login db error:%v", err)
		return resp, err
	}

	var pid agent.PID
	var dbId int64
	if reply == nil { // 本服 Redis 无 uid -> playerId 映射
		// 合服校验：若 MySQL 中该 uid 在本服已有角色（redis_id>0）但 Redis 无映射，说明数据未同步或异常，不允许建新号
		acc := new(model.Account)
		has, err := db.Engine.Mysql.Table(define.AccountTable).Where("uid = ? AND server_id = ?", uid, l.App.GetEnv().ID).Get(acc)
		if err != nil {
			log.Error("login check account error:%v, uid:%s", err, uid)
			return resp, err
		}
		if has && acc.RedisId > 0 {
			log.Warn("login reject: uid %s has player on this server (redis_id=%d) but no redis mapping, merge data not synced or abnormal", uid, acc.RedisId)
			return resp, nil
		}

		log.Debug("无账号玩家=====")
		playerData, err := player.Born(uid, l.App.GetEnv().ID)
		if err != nil {
			log.Error("born db error:%v", err)
			return resp, err
		}

		// 修改对应account表对应数据
		account := new(model.Account)
		account.NickName = playerData.Base.Name
		account.RedisId = playerData.Id

		_, err = db.RedisExec("set", fmt.Sprintf("%s:%s", define.Account, uid), playerData.Id)
		if err != nil {
			return resp, err
		}

		// 同步到account表（按本服更新，避免共享 MySQL 串服）
		conn := db.Engine.Mysql
		currentServerId := l.App.GetEnv().ID
		n, err := conn.Table(define.AccountTable).Where("uid = ? AND server_id = ?", uid, currentServerId).Cols("nick_name", "redis_id").Update(account)
		if err != nil || n == 0 {
			log.Error(" player born update mysql error:%v,n:%v,uid:%v", err, n, uid)
			return resp, err
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
		dbId, err = redis.Int64(reply, nil)
		log.Info("dbId:%v, %v", dbId, uid)
		if err != nil {
			log.Error("login convert player id error:%v", err)
			return resp, err
		}

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
