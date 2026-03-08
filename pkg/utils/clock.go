// 游戏逻辑时间源：支持时间偏移，便于测试/调试时“调时间”。
// 仅当 SetTimeOffsetEnabled(true)（一般 Debug 模式）时偏移生效，可由 GM 后台设置；正式服应调用 SetTimeOffsetEnabled(false)，此时偏移恒为 0，GM 也无法修改。
package utils

import (
	"sync/atomic"
	"time"
)

var (
	timeOffsetNanos atomic.Int64
	// allowTimeOffset 为 true 时允许设置/使用时间偏移（Debug 模式）；为 false 时正式服，偏移恒为 0。
	allowTimeOffset atomic.Bool
)

// SetTimeOffsetEnabled 设置是否允许时间偏移。应在启动时根据配置 Debug 调用一次：正式服传 false，调试服传 true。
// 正式服下即使 GM 调 SetTimeOffset 也不会生效，Now() 始终为真实时间。
func SetTimeOffsetEnabled(allow bool) {
	allowTimeOffset.Store(allow)
	if !allow {
		timeOffsetNanos.Store(0)
	}
}

// TimeOffsetEnabled 返回当前是否允许时间偏移（一般与 Debug 一致）。供 GM 接口判断是否可设置偏移。
func TimeOffsetEnabled() bool {
	return allowTimeOffset.Load()
}

// SetTimeOffset 设置游戏逻辑时间相对真实时间的偏移量。
// 仅当 SetTimeOffsetEnabled(true) 时生效；正式服下为 no-op。
func SetTimeOffset(offset time.Duration) {
	if !allowTimeOffset.Load() {
		return
	}
	timeOffsetNanos.Store(int64(offset))
}

// GetTimeOffset 返回当前配置的时间偏移量，供 GM 查询等使用。正式服下恒为 0。
func GetTimeOffset() time.Duration {
	if !allowTimeOffset.Load() {
		return 0
	}
	return time.Duration(timeOffsetNanos.Load())
}

// Now 返回当前游戏逻辑时间（真实时间 + 偏移）。正式服下恒为真实时间。
func Now() time.Time {
	return time.Now().Add(GetTimeOffset())
}
