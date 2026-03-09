package logic

import (
	"fmt"
	"time"
	"xfx/login_server/model/entity"
	"xfx/pkg/log"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
	xlog "xorm.io/xorm/log"
)

var AccountEngine *xorm.Engine // 账户库（account、server_group、game_server、hot_update、notice）

// EnsureServerTables 确保区组/区服表存在
func EnsureServerTables() {
	if AccountEngine == nil {
		return
	}
	if err := AccountEngine.Sync2(new(entity.ServerGroupMeta)); err != nil {
		log.Error("Sync2 server_group err: %v", err)
	}
	if err := AccountEngine.Sync2(new(entity.ServerItem)); err != nil {
		log.Error("Sync2 game_server err: %v", err)
	}
	if err := AccountEngine.Sync2(new(entity.HotUpdateItem)); err != nil {
		log.Error("Sync2 hot_update err: %v", err)
	}
	if err := AccountEngine.Sync2(new(entity.NoticeItem)); err != nil {
		log.Error("Sync2 notice err: %v", err)
	}
}

func NewMysqlEngine(addr string) *xorm.Engine {
	_engine, err := xorm.NewEngine("mysql", addr)
	if err != nil {
		panic(err)
	}
	_engine.Logger().SetLevel(xlog.LOG_OFF)         //不要日志
	_engine.ShowSQL(false)                          //不显示命令
	_engine.SetMaxIdleConns(240)                    //设置pool里可留存的空闲conn
	_engine.SetMaxOpenConns(1200)                   //设置最大打开连接数 mysql里设置的1752
	_engine.SetConnMaxLifetime(time.Second * 14400) //超时时间 mysql那边默认是8小时28800 外网设置的是24小时 这里这个值必须小于mysql的时间

	err = _engine.Ping()
	if err != nil {
		fmt.Println("数据库地址:", addr)
		panic("mysql数据库连接失败")
	}
	return _engine
}
