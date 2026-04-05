package env

import (
	"time"

	"github.com/spf13/viper"
)

var (
	v *viper.Viper
)

func LoadEnv() (*Env, error) {
	v = viper.New()

	v.SetConfigType("toml")
	v.AutomaticEnv()

	v.SetConfigName("env") // 文件名（不含扩展名）
	v.AddConfigPath("./")  // 搜索当前目录
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	e := new(Env)
	if err := v.Unmarshal(&e); err != nil {
		return nil, err
	}
	e.v = v
	return e, nil
}

type Env struct {
	v          *viper.Viper
	ID         int      // 机器 id
	RemoteName string   // Remote 节点名
	RemoteHost string   // Remote 监听地址
	RemotePort int      // Remote 监听端口
	Debug      bool     // 是否是DEBUG模式(Debug模式不可设置时间偏移)
	ConfPath   string   // json 文件路径
	HttpUrl    string   // http 服务地址
	GmToken    string   // GM HTTP 接口鉴权 Token（X-GM-Token 请求头）；为空则拒绝所有请求
	GmAllowIPs []string // 允许访问 GM 接口的 IP 白名单；GmToken 非空时此字段不生效
	TcpGate    *TcpGate // TCP 网关配置
	Gate       *Gate    // WebSocket 网关配置
	Mysql      *Mysql   // 通用 Mysql
	Redis      *Redis   // 本服 Redis
	LoginRedis *Redis   // 登录服 Redis
	Log        *Log     // 日志
}

type TcpGate struct {
	Ip                string
	Port              int
	WriteWait         time.Duration
	PongWait          time.Duration
	PingPeriod        time.Duration
	MaxMessageSize    int
	MessageBufferSize int
}

type Gate struct {
	Host              string
	Port              int
	WriteWait         time.Duration
	PongWait          time.Duration
	PingPeriod        time.Duration
	MaxMessageSize    int
	MessageBufferSize int
}

type Mysql struct {
	CommonAddr      string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

type Redis struct {
	Host           string
	Password       string
	DbNum          int
	MaxIdle        int
	MaxActive      int
	IdleTimeout    time.Duration
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type Log struct {
	WriteFile bool
	FilePath  string
	Level     string
	Prefix    string
	Format    string

	// file write
	MaxSize    int  // 文件最大大小(MB)
	MaxBackups int  // 最大备份文件数
	MaxAge     int  // 文件最大保存天数
	Compress   bool // 是否压缩备份文件
}
