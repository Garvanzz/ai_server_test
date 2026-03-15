// slice.go 提供切片、集合类辅助函数。
package utils

import (
	"cmp"
	"slices"
)

// Contains 判断切片是否包含指定值
func Contains[T comparable](arr []T, value T) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// ContainsInt32 兼容旧调用：判断 int32 切片包含关系
func ContainsInt32(arr []int32, value int32) bool {
	return Contains(arr, value)
}

// ContainsAll 判断 arr 是否包含 values 中的所有元素
func ContainsAll[T comparable](arr, values []T) bool {
	for _, v := range values {
		if !Contains(arr, v) {
			return false
		}
	}
	return true
}

// ContainsAny 判断 arr 是否包含 values 中的任意一个元素
func ContainsAny[T comparable](arr, values []T) bool {
	for _, v := range values {
		if Contains(arr, v) {
			return true
		}
	}
	return false
}

// IndexOf 返回元素在切片中的索引，不存在返回 -1
func IndexOf[T comparable](arr []T, value T) int {
	for i, v := range arr {
		if v == value {
			return i
		}
	}
	return -1
}

// LastIndexOf 返回元素在切片中最后一次出现的索引，不存在返回 -1
func LastIndexOf[T comparable](arr []T, value T) int {
	for i := len(arr) - 1; i >= 0; i-- {
		if arr[i] == value {
			return i
		}
	}
	return -1
}

// RemoveFirst 删除切片中第一个等于 value 的元素
func RemoveFirst[T comparable](slice []T, value T) []T {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// RemoveAll 删除切片中所有等于 value 的元素
func RemoveAll[T comparable](slice []T, value T) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != value {
			result = append(result, v)
		}
	}
	return result
}

// RemoveAt 删除指定索引位置的元素
func RemoveAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// InsertAt 在指定索引位置插入元素
func InsertAt[T any](slice []T, index int, value T) []T {
	if index < 0 || index > len(slice) {
		index = len(slice)
	}
	result := make([]T, len(slice)+1)
	copy(result[:index], slice[:index])
	result[index] = value
	copy(result[index+1:], slice[index:])
	return result
}

// Unique 去除切片中的重复元素，保持原有顺序
func Unique[T comparable](arr []T) []T {
	seen := make(map[T]bool, len(arr))
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// HasDuplicate 检查切片是否有重复值
func HasDuplicate[T comparable](arr []T) bool {
	seen := make(map[T]bool, len(arr))
	for _, v := range arr {
		if seen[v] {
			return true
		}
		seen[v] = true
	}
	return false
}

// Intersect 返回两个切片的交集
func Intersect[T comparable](a, b []T) []T {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	set := make(map[T]bool, len(b))
	for _, v := range b {
		set[v] = true
	}
	result := make([]T, 0)
	seen := make(map[T]bool)
	for _, v := range a {
		if set[v] && !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Union 返回两个切片的并集（去重）
func Union[T comparable](a, b []T) []T {
	result := make([]T, 0, len(a)+len(b))
	seen := make(map[T]bool)
	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	for _, v := range b {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Difference 返回在 a 中但不在 b 中的元素（a - b）
func Difference[T comparable](a, b []T) []T {
	if len(a) == 0 {
		return nil
	}
	if len(b) == 0 {
		return append([]T(nil), a...)
	}
	set := make(map[T]bool, len(b))
	for _, v := range b {
		set[v] = true
	}
	result := make([]T, 0)
	for _, v := range a {
		if !set[v] {
			result = append(result, v)
		}
	}
	return result
}

// Filter 按条件过滤切片元素
func Filter[T any](arr []T, fn func(T) bool) []T {
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map 对切片每个元素进行转换
func Map[T any, R any](arr []T, fn func(T) R) []R {
	result := make([]R, len(arr))
	for i, v := range arr {
		result[i] = fn(v)
	}
	return result
}

// Reverse 反转切片顺序
func Reverse[T any](arr []T) []T {
	result := make([]T, len(arr))
	for i, j := 0, len(arr)-1; i <= j; i, j = i+1, j-1 {
		result[i], result[j] = arr[j], arr[i]
	}
	return result
}

// Chunk 将切片分割为指定大小的子切片
func Chunk[T any](arr []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(arr)+size-1)/size)
	for i := 0; i < len(arr); i += size {
		end := i + size
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}
	return chunks
}

// Flatten 将二维切片展平为一维切片
func Flatten[T any](arr [][]T) []T {
	totalLen := 0
	for _, sub := range arr {
		totalLen += len(sub)
	}
	result := make([]T, 0, totalLen)
	for _, sub := range arr {
		result = append(result, sub...)
	}
	return result
}

// Sum 计算切片元素的总和（数值类型）
func Sum[T cmp.Ordered](arr []T) T {
	var sum T
	for _, v := range arr {
		sum += v
	}
	return sum
}

// MinValue 返回切片中的最小值，空切片返回零值
func MinValue[T cmp.Ordered](arr []T) T {
	if len(arr) == 0 {
		var zero T
		return zero
	}
	min := arr[0]
	for _, v := range arr[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// MaxValue 返回切片中的最大值，空切片返回零值
func MaxValue[T cmp.Ordered](arr []T) T {
	if len(arr) == 0 {
		var zero T
		return zero
	}
	max := arr[0]
	for _, v := range arr[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// SortAsc 升序排序（会修改原切片，返回排序后的切片）
func SortAsc[T cmp.Ordered](arr []T) []T {
	slices.Sort(arr)
	return arr
}

// SortDesc 降序排序（会修改原切片，返回排序后的切片）
func SortDesc[T cmp.Ordered](arr []T) []T {
	slices.SortFunc(arr, func(a, b T) int {
		if a > b {
			return -1
		}
		if a < b {
			return 1
		}
		return 0
	})
	return arr
}

// Clone 深拷贝切片
func Clone[T any](arr []T) []T {
	result := make([]T, len(arr))
	copy(result, arr)
	return result
}

// IsEmpty 判断切片是否为空
func IsEmpty[T any](arr []T) bool {
	return len(arr) == 0
}

// IsNotEmpty 判断切片是否非空
func IsNotEmpty[T any](arr []T) bool {
	return len(arr) > 0
}

// First 返回切片的第一个元素，空切片返回零值
func First[T any](arr []T) T {
	if len(arr) == 0 {
		var zero T
		return zero
	}
	return arr[0]
}

// FirstOrDefault 返回切片的第一个元素，空切片返回默认值
func FirstOrDefault[T any](arr []T, defaultValue T) T {
	if len(arr) == 0 {
		return defaultValue
	}
	return arr[0]
}

// Last 返回切片的最后一个元素，空切片返回零值
func Last[T any](arr []T) T {
	if len(arr) == 0 {
		var zero T
		return zero
	}
	return arr[len(arr)-1]
}

// LastOrDefault 返回切片的最后一个元素，空切片返回默认值
func LastOrDefault[T any](arr []T, defaultValue T) T {
	if len(arr) == 0 {
		return defaultValue
	}
	return arr[len(arr)-1]
}

// Take 取切片前 n 个元素
func Take[T any](arr []T, n int) []T {
	if n <= 0 {
		return nil
	}
	if n > len(arr) {
		n = len(arr)
	}
	return append([]T(nil), arr[:n]...)
}

// Skip 跳过切片前 n 个元素，返回剩余部分
func Skip[T any](arr []T, n int) []T {
	if n <= 0 {
		return append([]T(nil), arr...)
	}
	if n >= len(arr) {
		return nil
	}
	return append([]T(nil), arr[n:]...)
}
