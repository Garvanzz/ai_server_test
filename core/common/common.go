// Package common 保留对 core 与 main_server 的兼容入口。
// 通用工具已集中到 xfx/pkg/utils，新代码请直接使用 pkg/utils。
package common

import (
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

// SelectByOdds 按概率 upNum/downNum 判定是否命中（委托 pkg/utils）
func SelectByOdds(upNum, downNum int32) bool {
	return utils.SelectByOdds(upNum, downNum)
}

// StringToInt64 字符串转 int64，解析失败时打日志并返回 0（业务侧常用语义）
func StringToInt64(str string) int64 {
	num, err := utils.ParseInt64(str)
	if err != nil {
		log.Error("字符串转换int64出错：%v", str)
		return 0
	}
	return num
}

// RemoveFirstByValueInt32 删除切片中第一个等于 value 的元素（委托 pkg/utils）
func RemoveFirstByValueInt32(slice []int32, value int32) []int32 {
	return utils.RemoveFirstInt32(slice, value)
}

// IsHaveValueIntArray 判断切片是否包含指定值（委托 pkg/utils）
func IsHaveValueIntArray(arr []int32, value int32) bool {
	return utils.ContainsInt32(arr, value)
}

// IsHaveValueStringArray 判断切片是否包含指定字符串（委托 pkg/utils）
func IsHaveValueStringArray(arr []string, value string) bool {
	return utils.ContainsString(arr, value)
}

// IsHaveArrayValueIntArray 判断 arr 是否包含 value 中全部元素（委托 pkg/utils）
func IsHaveArrayValueIntArray(arr, value []int32) bool {
	return utils.ContainsAllInt32(arr, value)
}

// HasDuplicate 检查切片是否有重复值（委托 pkg/utils）
func HasDuplicate(nums []int32) (bool, int32) {
	return utils.HasDuplicateInt32(nums)
}
