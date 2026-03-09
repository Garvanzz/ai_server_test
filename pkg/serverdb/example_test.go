package serverdb_test

import (
	"xfx/pkg/serverdb"
)

// 使用示例（逻辑自洽：连接仅来自 Config，服列表仅元数据）：
//
//	cfg := serverdb.Config{
//	    ServerId: 1,
//	    RedisAddr: "127.0.0.1:6379",
//	    RedisPassword: "",
//	    RedisDB: 0,
//	    MysqlAddr: "user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4",
//	}
//	m := serverdb.NewManager(cfg)
//	if err := m.Start(); err != nil { ... }
//	defer m.Close()
//	_ = m.LoadServerList("server_list")
//	serverdb.SetGlobal(m)
//
//	eng, _ := serverdb.GetEngine(cfg.ServerId)
//	eng, _ = serverdb.DefaultEngine()
func Example_usage() {
	_ = serverdb.PlayerIdBase
	_ = serverdb.ServerStateNormal
}
