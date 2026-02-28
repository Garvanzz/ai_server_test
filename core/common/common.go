package common

import (
	"math/rand"
	"strconv"
	"xfx/pkg/log"
)

// 获取几分之的几率
func SelectByOdds(upNum, downNum int32) bool {
	if downNum < 1 {
		return false
	}
	if upNum < 1 {
		return false
	}
	if upNum > downNum-1 {
		return true
	}
	return (1 + int32((float64(rand.Int63())/(1<<63))*float64(downNum))) <= upNum
}

func StringToInt64(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log.Error("字符串转换int64出错：%v", str)
		return 0
	}
	return num
}

// 删除切片中的第一个指定值的元素
func RemoveFirstByValueInt32(slice []int32, value int32) []int32 {
	for i, v := range slice {
		if v == value {
			// 删除找到的元素，拼接前后部分
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice // 如果没有找到指定值，返回原切片
}

// 查找数组包含某个值
func IsHaveValueIntArray(arr []int32, value int32) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// 查找数组包含某个值
func IsHaveValueStringArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// 查找数组包含某个值
func IsHaveArrayValueIntArray(arr []int32, value []int32) bool {
	have := true
	for _, v := range value {
		if !IsHaveValueIntArray(arr, v) {
			have = false
			break
		}
	}
	return have
}

// 数组是否有相同的值
func HasDuplicate(nums []int32) (bool, int32) {
	seen := make(map[int32]bool) // 使用 map 记录是否已经出现过
	for _, num := range nums {
		if seen[num] { // 如果已经存在，说明有重复
			return true, num
		}
		seen[num] = true // 标记为已出现
	}
	return false, 0 // 遍历完没有重复
}
