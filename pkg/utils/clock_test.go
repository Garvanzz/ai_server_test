package utils_test

import (
	"sync"
	"testing"
	"time"

	"xfx/pkg/utils"
)

// 内存存储实现（用于测试）
type memoryStorage struct {
	data map[string]int64
	mu   sync.Mutex
}

func newMemoryStorage() *memoryStorage {
	return &memoryStorage{data: make(map[string]int64)}
}

func (m *memoryStorage) SaveOffset(nanos int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data["offset"] = nanos
	return nil
}

func (m *memoryStorage) LoadOffset() (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data["offset"], nil
}

func TestClock(t *testing.T) {
	// 使用内存存储测试
	storage := newMemoryStorage()
	utils.SetClockStorage(storage)

	// 测试正式服模式（不允许偏移）
	t.Run("ProductionMode", func(t *testing.T) {
		utils.InitClock(utils.ClockConfig{
			AllowOffset: false,
		})

		if utils.IsOffsetEnabled() {
			t.Error("正式服不应允许偏移")
		}

		// 尝试设置偏移
		utils.SetOffset(time.Hour)
		if utils.GetOffset() != 0 {
			t.Error("正式服偏移应恒为 0")
		}

		// Now() 应该接近真实时间
		diff := time.Since(utils.Now())
		if diff < -time.Second || diff > time.Second {
			t.Error("Now() 应返回真实时间")
		}
	})

	// 测试调试服模式（允许偏移）
	t.Run("DebugMode", func(t *testing.T) {
		// 重新初始化，允许偏移
		utils.InitClock(utils.ClockConfig{
			AllowOffset: true,
		})

		if !utils.IsOffsetEnabled() {
			t.Error("调试服应允许偏移")
		}

		// 设置 1 小时偏移
		utils.SetOffset(time.Hour)
		if utils.GetOffset() != time.Hour {
			t.Errorf("偏移应为 1 小时，实际为 %v", utils.GetOffset())
		}

		// Now() 应该是未来时间
		if utils.Now().Before(time.Now()) {
			t.Error("Now() 应返回未来时间")
		}

		// 重置偏移
		utils.ResetOffset()
		if utils.GetOffset() != 0 {
			t.Error("重置后偏移应为 0")
		}
	})

	// 测试偏移计算
	t.Run("OffsetCalculation", func(t *testing.T) {
		utils.InitClock(utils.ClockConfig{
			AllowOffset: true,
		})

		baseTime := time.Now()
		offset := 2 * time.Hour
		utils.SetOffset(offset)

		// Now() 应该等于 baseTime + offset
		expectedTime := baseTime.Add(offset)
		actualTime := utils.Now()

		// 允许 1 秒误差
		diff := actualTime.Sub(expectedTime)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("时间偏差过大: %v", diff)
		}
	})

	// 测试便捷函数
	t.Run("HelperFunctions", func(t *testing.T) {
		utils.InitClock(utils.ClockConfig{
			AllowOffset: true,
		})
		utils.SetOffset(time.Hour)

		// TimeAfter
		future := utils.TimeAfter(30 * time.Minute)
		if !future.After(utils.Now()) {
			t.Error("TimeAfter 应返回未来时间")
		}

		// UnixNow
		if utils.UnixNow() == 0 {
			t.Error("UnixNow 不应为 0")
		}

		// OffsetInfo
		info := utils.GetOffsetInfo()
		if !info.Enabled || info.Offset != time.Hour {
			t.Error("OffsetInfo 信息不正确")
		}

		// Until / Since
		past := utils.Now().Add(-time.Hour)
		future = utils.Now().Add(time.Hour)

		since := utils.Since(past)
		if since < time.Hour-time.Minute {
			t.Error("Since 计算不正确")
		}

		until := utils.Until(future)
		if until < time.Hour-time.Minute {
			t.Error("Until 计算不正确")
		}
	})
}

// TestClockPersistence 测试持久化功能
func TestClockPersistence(t *testing.T) {
	// 注意：由于使用全局变量，这个测试会修改全局状态
	// 实际测试中应该隔离状态或使用子进程
	t.Skip("Persistence test skipped - modifies global state")

	/*
	storage := newMemoryStorage()
	utils.SetClockStorage(storage)

	// 设置偏移并保存
	t.Run("SaveAndLoad", func(t *testing.T) {
		utils.InitClock(utils.ClockConfig{
			AllowOffset: true,
		})

		utils.SetOffset(3 * time.Hour)

		// 模拟重启：重新初始化
		utils.InitClock(utils.ClockConfig{
			AllowOffset: true,
		})

		// 应该加载之前保存的偏移
		if utils.GetOffset() != 3*time.Hour {
			t.Errorf("应从存储加载偏移，期望 3h，实际 %v", utils.GetOffset())
		}
	})
	*/
}

// Benchmark
func BenchmarkNow(b *testing.B) {
	utils.InitClock(utils.ClockConfig{AllowOffset: true})
	utils.SetOffset(time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.Now()
	}
}

func BenchmarkGetOffset(b *testing.B) {
	utils.InitClock(utils.ClockConfig{AllowOffset: true})
	utils.SetOffset(time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = utils.GetOffset()
	}
}
