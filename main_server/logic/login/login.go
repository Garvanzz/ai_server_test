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
	Proto_Public "xfx/proto/proto_public"

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

// TODO:玩家登录
func (l *Login) login(msg *messages.Login) (*messages.LoginResult, error) {
	resp := new(messages.LoginResult)
	resp.Result = int(Proto_Public.CommonState_Faild)

	rdb, _ := db.GetEngine(l.App.GetEnv().ID)

	uid, err := redis.String(rdb.RedisExec("get", fmt.Sprintf("%s:%s", define.LoginToken, msg.Request.Token)))
	if err != nil {
		log.Error("login db error:%v", err)
		return resp, err
	}

	// key->uid，value->id
	reply, err := rdb.RedisExec("get", fmt.Sprintf("%s:%s", define.Account, uid))
	if err != nil {
		log.Error("login db error:%v", err)
		return resp, err
	}

	var pid agent.PID
	var dbId int64
	if reply == nil { // 无账号创建及玩家
		playerData, err := player.Born(uid, l.App.GetEnv().ID)
		if err != nil {
			log.Error("born db error:%v", err)
			return resp, err
		}

		// 修改对应account表对应数据
		account := new(model.Account)
		account.NickName = playerData.Base.Name
		account.RedisId = playerData.Id

		// 同步到account表
		conn := db.CommonEngine.Mysql
		n, err := conn.Table(define.AccountTable).Where("uid = ?", uid).Cols("nick_name", "redis_id").Update(account)
		if err != nil || n == 0 {
			log.Error("player born update mysql error:%v,n:%v", err, n)
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
				log.Error("login load player data error:%v", err)
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
	resp.Result = int(Proto_Public.CommonState_Success)
	return resp, err
}

// 获取玩家pid
func (l *Login) getPlayerPid(dbId int64) agent.PID {
	if pid, ok := l.players[dbId]; ok {
		return pid
	}

	return nil
}

// 玩家登出
func (l *Login) logout(playerId int64) error {
	pid, ok := l.players[playerId]
	if ok {
		l.Context.Call(pid, &messages.Logout{}) // 转发给玩家进程
		delete(l.players, playerId)
	}
	return nil
}

// 断开连接
func (l *Login) disconnect(playerId int64) error {
	pid, ok := l.players[playerId]
	if ok {
		l.Context.Call(pid, &messages.Disconnect{}) // 转发给玩家进程
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
