package serverdb

import "sync"

var (
	global   *Manager
	globalMu sync.RWMutex
)

// SetGlobal 设置全局 Manager（建议在进程启动时 Start 后调用一次）
func SetGlobal(m *Manager) {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = m
}

// GetGlobal 获取全局 Manager
func GetGlobal() *Manager {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return global
}

// DefaultEngine 便捷：取全局唯一引擎，未设置或未 Start 返回 nil, ErrNotStarted
func DefaultEngine() (*Engine, error) {
	m := GetGlobal()
	if m == nil {
		return nil, ErrNotStarted
	}
	return m.Engine()
}

// GetEngine 便捷：按服 ID 取引擎（仅本服有数据时返回）
func GetEngine(serverId int) (*Engine, error) {
	m := GetGlobal()
	if m == nil {
		return nil, ErrNotStarted
	}
	return m.GetEngine(serverId)
}
