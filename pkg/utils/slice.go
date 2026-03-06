// Package utils 提供与业务无关的通用工具，供全项目使用。
// 切片、集合类辅助函数。

package utils

// ContainsInt32 判断切片是否包含指定值
func ContainsInt32(arr []int32, value int32) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// ContainsString 判断切片是否包含指定字符串
func ContainsString(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// ContainsAllInt32 判断 arr 是否包含 value 中全部元素
func ContainsAllInt32(arr, value []int32) bool {
	for _, v := range value {
		if !ContainsInt32(arr, v) {
			return false
		}
	}
	return true
}

// RemoveFirstInt32 删除切片中第一个等于 value 的元素
func RemoveFirstInt32(slice []int32, value int32) []int32 {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// HasDuplicateInt32 检查切片是否有重复值，若有则返回 true 和重复的那个值
func HasDuplicateInt32(nums []int32) (bool, int32) {
	seen := make(map[int32]bool)
	for _, num := range nums {
		if seen[num] {
			return true, num
		}
		seen[num] = true
	}
	return false, 0
}
