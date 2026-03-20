// Package id 提供分布式唯一 ID 生成器（Twitter Snowflake 算法变体）。
//
// ID 格式（64位）：
//   - 40位时间戳（毫秒，从2020-01-01开始）
//   - 10位机器ID（支持 0-1023）
//   - 14位序列号（每毫秒最多16384个ID）
//
// 使用方法：
//
//	// 初始化（在应用启动时调用一次）
//	id.Init(1) // machineId: 0-1000为保留段，100-200用于机器人
//
//	// 生成ID
//	idVal, err := id.Generate()
//
//	// ID 压缩（转为62进制字符串）
//	shortId := id.Itoa(idVal)
//	originalId := id.Atoi(shortId)
//
package id

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	timestampBits = 40
	machineBits   = 10
	sequenceBits  = 14

	maxMachineID = (1 << machineBits) - 1 // 1023
	maxSequence  = (1 << sequenceBits) - 1 // 16383

	// 保留段定义
	reservedStart   = 0
	reservedEnd     = 1000
	robotRangeStart = 100
	robotRangeEnd   = 200
)

var (
	// epoch 是 ID 时间戳的起始时间（2020-01-01 11:11:11）
	epoch time.Time
	// generator 是全局单例
	generator *Snowflake
	// initOnce 确保只初始化一次
	initOnce sync.Once
)

func init() {
	var err error
	epoch, err = time.Parse("2006-01-01 15:04:05", "2020-01-01 11:11:11")
	if err != nil {
		panic(fmt.Sprintf("failed to parse epoch: %v", err))
	}
}

// Snowflake ID 生成器
type Snowflake struct {
	machineID     uint32
	sequence      uint32
	lastTimestamp uint64
	mu            sync.Mutex
}

// Init 初始化全局 ID 生成器。
// machineID 范围 0-1023，其中 0-1000 为保留段，100-200 用于机器人。
// 该函数只应在应用启动时调用一次。
func Init(machineID uint32) error {
	var initErr error
	initOnce.Do(func() {
		if machineID > maxMachineID {
			initErr = fmt.Errorf("machineID must be between 0 and %d", maxMachineID)
			return
		}
		generator = &Snowflake{
			machineID: machineID,
			sequence:  0,
		}
	})
	return initErr
}

// Generate 生成唯一 ID。
// 如果未初始化或时钟回拨，返回错误。
func Generate() (int64, error) {
	if generator == nil {
		return 0, errors.New("id generator not initialized")
	}
	return generator.generate()
}

// MustGenerate 生成唯一 ID，出错时 panic。
func MustGenerate() int64 {
	id, err := Generate()
	if err != nil {
		panic(err)
	}
	return id
}

// MachineID 返回当前配置的机器 ID。
func MachineID() uint32 {
	if generator == nil {
		return 0
	}
	return generator.machineID
}

// IsInitialized 返回是否已初始化。
func IsInitialized() bool {
	return generator != nil
}

// generate 内部生成逻辑。
func (sf *Snowflake) generate() (int64, error) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := sf.currentTime()

	if now < sf.lastTimestamp {
		return 0, fmt.Errorf("clock moved backwards: %d < %d", now, sf.lastTimestamp)
	}

	if now == sf.lastTimestamp {
		// 同一毫秒内，递增序列号
		sf.sequence = (sf.sequence + 1) & maxSequence
		if sf.sequence == 0 {
			// 序列号溢出，等待下一毫秒
			now = sf.waitNextMilli(now)
		}
	} else {
		// 新毫秒，重置序列号
		sf.sequence = 0
	}

	sf.lastTimestamp = now
	return sf.pack(now), nil
}

// pack 将时间戳、机器ID和序列号打包成64位整数。
func (sf *Snowflake) pack(timestamp uint64) int64 {
	return int64((timestamp << (machineBits + sequenceBits)) |
		(uint64(sf.machineID) << sequenceBits) |
		uint64(sf.sequence))
}

// currentTime 返回当前时间相对于 epoch 的毫秒数。
func (sf *Snowflake) currentTime() uint64 {
	return uint64(time.Since(epoch).Milliseconds())
}

// waitNextMilli 等待直到下一毫秒。
func (sf *Snowflake) waitNextMilli(current uint64) uint64 {
	for current <= sf.lastTimestamp {
		time.Sleep(time.Millisecond)
		current = sf.currentTime()
	}
	return current
}

// ExtractTimestamp 从 ID 中提取时间戳（毫秒，相对于 epoch）。
func ExtractTimestamp(id int64) uint64 {
	return uint64(id) >> (machineBits + sequenceBits)
}

// ExtractMachineID 从 ID 中提取机器 ID。
func ExtractMachineID(id int64) uint32 {
	return uint32((uint64(id) >> sequenceBits) & maxMachineID)
}

// ExtractSequence 从 ID 中提取序列号。
func ExtractSequence(id int64) uint32 {
	return uint32(uint64(id) & maxSequence)
}

// TimeFromID 从 ID 中提取生成时间。
func TimeFromID(id int64) time.Time {
	return epoch.Add(time.Duration(ExtractTimestamp(id)) * time.Millisecond)
}
