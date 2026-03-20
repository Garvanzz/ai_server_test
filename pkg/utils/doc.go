// Package utils 提供游戏开发中常用的通用工具函数。
//
// 本包包含以下功能模块：
//
// # 时间工具 (clock.go, time.go)
//   - 游戏逻辑时间源（支持时间偏移，便于调试）
//   - 日期计算（同一天/同一周/同一月判断）
//   - 时间戳转换和格式化
//   - 常用时间边界计算（当天开始/结束、周计算等）
//
// # 随机工具 (random.go)
//   - 随机数生成
//   - 随机字符串生成（数字、字母、混合）
//   - 随机选择和加权随机
//   - 概率判定（命中率计算）
//
// # 切片工具 (slice.go)
//   - 查找、包含、去重
//   - 过滤、映射、排序
//   - 集合操作（交集、并集、差集）
//   - 分页、分块
//
// # 数学工具 (math.go)
//   - 最小值、最大值、钳制
//   - 绝对值、符号、范围判断
//   - 线性插值、取整
//   - 百分比计算
//
// # 加密工具 (crypto.go)
//   - MD5、SHA1、SHA256 哈希
//   - Base64 编码/解码
//   - AES-ECB 加密/解密（兼容旧系统）
//   - 十六进制编码
//   - XOR 操作
//
// # 类型转换 (conv.go, reply.go)
//   - 字符串转数字
//   - Redis reply 类型转换
//   - 布尔值转换
//
// # 子包
//
// ## id - 分布式唯一 ID 生成器
//
// 使用 Twitter Snowflake 算法生成全局唯一 ID：
//
//	import "xfx/pkg/utils/id"
//
//	// 初始化（应用启动时调用一次）
//	err := id.Init(1) // machineId: 0-1023
//
//	// 生成 ID
//	idVal, err := id.Generate()
//
//	// ID 压缩（62进制短字符串）
//	shortId := id.Itoa(idVal)
//	originalId := id.Atoi(shortId)
//
// ## sensitive - 敏感词过滤
//
// 使用 DFA 算法实现高效敏感词过滤：
//
//	import "xfx/pkg/utils/sensitive"
//
//	// 创建过滤器
//	filter := sensitive.New()
//
//	// 加载敏感词
//	filter.LoadFromFile("dict.txt")
//	filter.AddWords("词1", "词2")
//
//	// 检测和处理
//	if filter.IsSensitive("测试文本") { ... }
//	clean := filter.Replace("测试文本", '*')
//	clean = filter.Remove("测试文本")
//
// 使用示例：
//
//	import "xfx/pkg/utils"
//
//	// 时间工具
//	todayEnd := utils.TodayEndUnix()
//	isSameDay := utils.IsSameDayBySec(timestamp1, timestamp2)
//
//	// 随机工具
//	randomStr := utils.RandomAlphanumeric(10)
//	selected := utils.Sample(items, 3)
//	if utils.HitPercent(50) { /* 50% 概率 */ }
//
//	// 切片工具
//	unique := utils.Unique(items)
//	common := utils.Intersect(listA, listB)
//	isDuplicate, dup := utils.HasDuplicate(nums)
//
//	// 数学工具
//	min := utils.Min(a, b)
//	clamped := utils.Clamp(value, 0, 100)
//	pct := utils.Percent(value, total)
//
//	// 加密工具
//	hash := utils.MD5("data")
//	encoded := utils.Base64EncodeString("data")
//
// 注意事项：
//   - 所有函数都是线程安全的（除非特别说明）
//   - 泛型函数需要 Go 1.18+ 版本
//   - clock.go 中的时间偏移功能仅在 Debug 模式下生效
//   - crypto.go 中的 ECB 模式仅用于兼容旧系统，新项目请使用更安全的模式
//
package utils
