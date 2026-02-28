package env

import (
	"github.com/spf13/viper"
	"time"
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
	v  *viper.Viper
	ID int // 机器id

	Name string // 当前节点名
	Host string // 当前节点地址
	Port int    // 当前节点端口

	Debug    bool
	ConfPath string // 配置路径
	HttpUrl  string

	Mysql   *Mysql
	Redis   *Redis
	Gate    *Gate    // ws
	TcpGate *TcpGate // tcp
	Log     *Log
}

type Gate struct {
	Host              string
	WriteWait         time.Duration
	PongWait          time.Duration
	PingPeriod        time.Duration
	MaxMessageSize    int
	MessageBufferSize int
}

type TcpGate struct {
	Port              int // 只有这个配置在用
	Ip                string
	WriteWait         time.Duration
	PongWait          time.Duration
	PingPeriod        time.Duration
	MaxMessageSize    int
	MessageBufferSize int
}

type Mysql struct {
	CommonAddr string
}

type Redis struct {
	Host     string
	Password string
	DbNum    int
}

type Log struct {
	WriteFile bool
	FilePath  string
	Level     string
	Prefix    string
	Format    string // json fmtlog text--default text

	// file write
	MaxSize    int  // 文件最大大小(MB)
	MaxBackups int  // 最大备份文件数
	MaxAge     int  // 文件最大保存天数
	Compress   bool // 是否压缩备份文件
}

func (e *Env) Get(key string) interface{}            { return e.v.Get(key) }
func (e *Env) GetBool(key string) bool               { return e.v.GetBool(key) }
func (e *Env) GetDuration(key string) time.Duration  { return e.v.GetDuration(key) }
func (e *Env) GetInt(key string) interface{}         { return e.v.GetInt(key) }
func (e *Env) GetInt32(key string) interface{}       { return e.v.GetInt32(key) }
func (e *Env) GetInt64(key string) interface{}       { return e.v.GetInt64(key) }
func (e *Env) GetIntSlice(key string) interface{}    { return e.v.GetIntSlice(key) }
func (e *Env) GetString(key string) interface{}      { return e.v.GetString(key) }
func (e *Env) GetStringMap(key string) interface{}   { return e.v.GetStringMap(key) }
func (e *Env) GetStringSlice(key string) interface{} { return e.v.GetStringSlice(key) }
func (e *Env) GetTime(key string) interface{}        { return e.v.GetTime(key) }
