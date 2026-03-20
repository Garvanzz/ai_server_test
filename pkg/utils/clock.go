// Package utils 提供游戏逻辑时间源，支持时间偏移以便测试/调试时"调时间"。
//
// 使用方式：
//
//	// 1. 在 main_server 初始化时注入存储层（如 Redis）
//	utils.SetClockStorage(&redisStorage{client: redisClient})
//
//	// 2. 初始化时钟系统
//	utils.InitClock(utils.ClockConfig{
//	    AllowOffset: config.IsDebugMode(), // 调试服true，正式服false
//	})
//
//	// 3. 游戏中使用 utils.Now() 获取带偏移的时间
//	now := utils.Now()
//
package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

// ClockStorage 定义时间偏移的存储接口。
// 由 main_server 注入具体的实现（如 Redis、文件等）。
type ClockStorage interface {
	// SaveOffset 保存时间偏移（单位：纳秒）。
	// 该函数在后台 goroutine 中调用，应保证线程安全。
	SaveOffset(nanos int64) error

	// LoadOffset 加载时间偏移。
	// 返回存储的纳秒值，如果不存在返回 0 和 nil。
	LoadOffset() (int64, error)
}

// ClockConfig 时钟配置。
type ClockConfig struct {
	// AllowOffset 是否允许时间偏移。
	// 正式服应设为 false，此时偏移恒为 0，GM 无法修改。
	// 调试服设为 true，允许 GM 调整时间。
	AllowOffset bool

	// OffsetKey 存储偏移的键名，默认使用 "game:time_offset_nanos"。
	OffsetKey string
}

var (
	// 当前时间偏移（纳秒）
	offsetNanos atomic.Int64

	// 是否允许时间偏移
	allowOffset atomic.Bool

	// 存储层（由外部注入）
	storage     ClockStorage
	storageMu   sync.RWMutex

	// 存储键名
	offsetKey   = "game:time_offset_nanos"
)

// ==================== 初始化 ====================

// SetClockStorage 设置时间偏移的存储层。
// 应在 InitClock 之前调用，通常由 main_server 在启动时注入 Redis 实现。
func SetClockStorage(s ClockStorage) {
	storageMu.Lock()
	storage = s
	storageMu.Unlock()
}

// InitClock 初始化时钟系统。
// 应在应用启动时调用一次，根据配置决定是否启用时间偏移。
func InitClock(cfg ClockConfig) {
	// 设置存储键名
	if cfg.OffsetKey != "" {
		offsetKey = cfg.OffsetKey
	}

	// 设置是否允许偏移
	allowOffset.Store(cfg.AllowOffset)

	// 如果不允许偏移，直接返回（偏移保持为 0）
	if !cfg.AllowOffset {
		offsetNanos.Store(0)
		return
	}

	// 从存储加载之前的偏移（如果有）
	loadOffsetFromStorage()
}

// ==================== 时间偏移操作 ====================

// Now 返回当前游戏逻辑时间（真实时间 + 偏移）。
// 正式服下恒为真实时间。
func Now() time.Time {
	return time.Now().Add(GetOffset())
}

// GetOffset 返回当前的时间偏移量。
// 正式服下恒为 0。
func GetOffset() time.Duration {
	if !allowOffset.Load() {
		return 0
	}
	return time.Duration(offsetNanos.Load())
}

// SetOffset 设置时间偏移。
// 仅当 AllowOffset=true 时生效；正式服下为 no-op。
// 偏移量会被持久化到存储层。
func SetOffset(offset time.Duration) {
	if !allowOffset.Load() {
		return
	}

	nanos := offset.Nanoseconds()
	offsetNanos.Store(nanos)

	// 异步持久化
	saveOffsetToStorageAsync(nanos)
}

// AddOffset 增加时间偏移（相对当前偏移）。
// 例如 AddOffset(time.Hour) 会让游戏时间快进 1 小时。
func AddOffset(delta time.Duration) {
	if !allowOffset.Load() {
		return
	}

	current := GetOffset()
	newOffset := current + delta
	SetOffset(newOffset)
}

// ResetOffset 重置时间偏移为 0。
func ResetOffset() {
	SetOffset(0)
}

// IsOffsetEnabled 返回是否允许时间偏移。
func IsOffsetEnabled() bool {
	return allowOffset.Load()
}

// OffsetInfo 返回当前偏移的详细信息（用于 GM 查询等）。
type OffsetInfo struct {
	Enabled   bool          // 是否启用偏移
	Offset    time.Duration // 当前偏移量
	OffsetKey string        // 存储键名
}

// GetOffsetInfo 返回偏移的详细信息。
func GetOffsetInfo() OffsetInfo {
	return OffsetInfo{
		Enabled:   allowOffset.Load(),
		Offset:    GetOffset(),
		OffsetKey: offsetKey,
	}
}

// ==================== 存储层操作 ====================

// loadOffsetFromStorage 从存储层加载偏移。
func loadOffsetFromStorage() {
	storageMu.RLock()
	s := storage
	storageMu.RUnlock()

	if s == nil {
		return // 未设置存储层，保持默认值 0
	}

	nanos, err := s.LoadOffset()
	if err != nil {
		// 加载失败，保持默认值 0
		return
	}

	offsetNanos.Store(nanos)
}

// saveOffsetToStorageAsync 异步保存偏移到存储层。
func saveOffsetToStorageAsync(nanos int64) {
	storageMu.RLock()
	s := storage
	storageMu.RUnlock()

	if s == nil {
		return // 未设置存储层，不持久化
	}

	// 后台异步保存，不阻塞主流程
	go func() {
		_ = s.SaveOffset(nanos) // 忽略错误，可根据需要添加日志
	}()
}

// ReloadOffset 从存储层重新加载偏移。
// 用于 GM 后台强制刷新偏移。
func ReloadOffset() {
	if !allowOffset.Load() {
		return
	}
	loadOffsetFromStorage()
}

// ==================== 便捷函数 ====================

// TimeAfter 返回当前时间之后 duration 的时间点。
func TimeAfter(duration time.Duration) time.Time {
	return Now().Add(duration)
}

// UnixNow 返回当前游戏时间的 Unix 时间戳（秒）。
func UnixNow() int64 {
	return Now().Unix()
}

// UnixMilliNow 返回当前游戏时间的 Unix 时间戳（毫秒）。
func UnixMilliNow() int64 {
	return Now().UnixMilli()
}

// UnixNanoNow 返回当前游戏时间的 Unix 时间戳（纳秒）。
func UnixNanoNow() int64 {
	return Now().UnixNano()
}

// Since 返回从 t 到现在经过的时间（考虑偏移）。
func Since(t time.Time) time.Duration {
	return Now().Sub(t)
}

// Until 返回从现在到 t 还有多长时间（考虑偏移）。
func Until(t time.Time) time.Duration {
	return t.Sub(Now())
}
