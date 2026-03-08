package serverdb

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
)

// Manager 服务器与 DB 管理：连接仅来自 Config，服列表仅元数据。
// 方案二：一进程一 Redis + 共享 MySQL；GetEngine 仅返回本进程唯一引擎。
type Manager struct {
	cfg     Config
	engine  *Engine
	servers []ServerMeta
}

// NewManager 创建管理器（未连接，需调用 Start）
func NewManager(cfg Config) *Manager {
	return &Manager{cfg: cfg}
}

// Start 根据 Config 建立本服 Redis 与共享 MySQL，并可选从 MySQL 加载服列表元数据。
// 不依赖 server 表里的 Redis/MySQL 字段；表仅用于服列表展示。
func (m *Manager) Start() error {
	redisPool := m.newRedisPool()
	conn := redisPool.Get()
	_, err := conn.Do("PING")
	conn.Close()
	if err != nil {
		return fmt.Errorf("serverdb redis ping: %w", err)
	}

	mysqlEngine, err := m.newMysqlEngine()
	if err != nil {
		redisPool.Close()
		return fmt.Errorf("serverdb mysql: %w", err)
	}

	m.engine = &Engine{Redis: redisPool, Mysql: mysqlEngine}
	return nil
}

func (m *Manager) newRedisPool() *redis.Pool {
	cfg := m.cfg
	idleTimeout := 60 * 60 * time.Second
	if cfg.RedisIdleTimeout > 0 {
		idleTimeout = cfg.RedisIdleTimeout
	}
	maxIdle, maxActive := 200, 2000
	if cfg.RedisMaxIdle > 0 {
		maxIdle = cfg.RedisMaxIdle
	}
	if cfg.RedisMaxActive > 0 {
		maxActive = cfg.RedisMaxActive
	}
	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisAddr,
				redis.DialPassword(cfg.RedisPassword),
				redis.DialDatabase(cfg.RedisDB),
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < 5*time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func (m *Manager) newMysqlEngine() (*xorm.Engine, error) {
	cfg := m.cfg
	eng, err := xorm.NewEngine("mysql", cfg.MysqlAddr)
	if err != nil {
		return nil, err
	}
	eng.Logger().SetLevel(xlog.LOG_OFF)
	eng.ShowSQL(false)
	maxIdle, maxOpen, maxLife := 240, 1200, 4*time.Hour
	if cfg.MysqlMaxIdle > 0 {
		maxIdle = cfg.MysqlMaxIdle
	}
	if cfg.MysqlMaxOpen > 0 {
		maxOpen = cfg.MysqlMaxOpen
	}
	if cfg.MysqlMaxLifetime > 0 {
		maxLife = cfg.MysqlMaxLifetime
	}
	eng.SetMaxIdleConns(maxIdle)
	eng.SetMaxOpenConns(maxOpen)
	eng.SetConnMaxLifetime(maxLife)
	if err := eng.Ping(); err != nil {
		_ = eng.Close()
		return nil, err
	}
	return eng, nil
}

// Close 关闭 Redis 与 MySQL
func (m *Manager) Close() {
	if m.engine != nil {
		m.engine.Close(true)
		m.engine = nil
	}
}

// Engine 返回本进程唯一引擎（方案二：单服单 Redis）
func (m *Manager) Engine() (*Engine, error) {
	if m.engine == nil {
		return nil, ErrNotStarted
	}
	return m.engine, nil
}

// GetEngine 根据服 ID 取引擎。方案二下单进程只服务一个服，仅当 serverId == Config.ServerId 返回引擎。
func (m *Manager) GetEngine(serverId int) (*Engine, error) {
	if m.engine == nil {
		return nil, ErrNotStarted
	}
	if serverId != m.cfg.ServerId {
		return nil, ErrWrongServer
	}
	return m.engine, nil
}

// GetEngineByPlayerId 根据玩家 ID 取引擎。playerId 高段为 serverId，仅本服玩家返回引擎。
func (m *Manager) GetEngineByPlayerId(playerId int64) (*Engine, error) {
	if m.engine == nil {
		return nil, ErrNotStarted
	}
	serverId := int(playerId / PlayerIdBase)
	if serverId != m.cfg.ServerId {
		return nil, errors.New("serverdb: player not on this server, serverId=" + strconv.Itoa(serverId))
	}
	return m.engine, nil
}

// ServerId 返回本进程服 ID
func (m *Manager) ServerId() int {
	return m.cfg.ServerId
}

// LoadServerList 从 MySQL 加载服列表元数据到内存（仅元数据，无连接信息）。
// 表名见 ServerMeta.TableName；若表为 servergroup，可先查入再映射到 ServerMeta。
func (m *Manager) LoadServerList(tableName string) error {
	if m.engine == nil || m.engine.Mysql == nil {
		return ErrNotStarted
	}
	var list []ServerMeta
	err := m.engine.Mysql.Table(tableName).Find(&list)
	if err != nil {
		return err
	}
	m.servers = list
	return nil
}

// Servers 返回已加载的服列表元数据（只读）
func (m *Manager) Servers() []ServerMeta {
	return m.servers
}

// GetServerMeta 按服 ID 查元数据
func (m *Manager) GetServerMeta(serverId int) (ServerMeta, bool) {
	for _, s := range m.servers {
		if int(s.Id) == serverId {
			return s, true
		}
	}
	return ServerMeta{}, false
}
