// 游戏逻辑时间源：支持配置偏移，便于测试/调试时“调时间”。
// Now() 返回 真实时间 + 偏移。偏移仅在 Debug 模式下由 run 根据 TimeOffsetDays 设置；线上 Debug=false 时偏移为 0，即使用服务器真实时间。
package utils

import (
	"sync"
	"time"
)

var (
	timeOffset time.Duration
	offsetMu   sync.RWMutex
)

// SetTimeOffset 设置游戏逻辑时间相对真实时间的偏移量，启动时根据配置调用一次即可。
// 例如 offset = 7*24*time.Hour 表示“当前游戏时间比真实时间快 7 天”。
func SetTimeOffset(offset time.Duration) {
	offsetMu.Lock()
	timeOffset = offset
	offsetMu.Unlock()
}

// Now 返回当前游戏逻辑时间（真实时间 + 配置的偏移），时间会随系统时钟自然增加。
func Now() time.Time {
	offsetMu.RLock()
	d := timeOffset
	offsetMu.RUnlock()
	return time.Now().Add(d)
}
