package utils

import (
	"math/rand"
	"strconv"
	"time"
)

var rnd *rand.Rand

func init() {
	UpdateRand(time.Now().UnixNano())
}

func UpdateRand(t int64) {
	rnd = rand.New(rand.NewSource(t))
}

func Random(length int) (nid string) {
	nid = ""

	for i := 0; i < length; i++ {
		var (
			r1 int
		)

		r1 = rnd.Intn(9)

		if i == 0 {
			for r1 == 0 {
				r1 = rnd.Intn(9)
			}
		}

		nid += strconv.Itoa(r1)
	}
	return
}

// MicsSlice 随机列表
func MicsSlice[T comparable](arr []T, count int) []T {
	tmp := make([]T, len(arr))
	copy(tmp, arr)

	rnd.Shuffle(len(tmp), func(i int, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	ret := make([]T, 0, count)
	for i := 0; i < count; i++ {
		ret = append(ret, tmp[i])
	}
	return ret
}

// WeightedRandom 根据权重取固定数量的值
func WeightedRandom[T ~int | ~int32 | ~int64, E any](weights []T, values []E, num int) []E {
	cv := make([]E, len(values))
	copy(cv, values)

	wh := make([]T, len(weights))
	copy(wh, weights)

	if num <= 0 {
		return nil
	}

	if len(wh) != len(values) {
		return nil
	}

	if num > len(wh) {
		return nil
	}

	l := make([]E, 0, num)
	for i := 0; i < num; i++ {
		sum := T(0)
		for _, w := range wh {
			sum += w
		}

		r := T(rnd.Float64() * float64(sum))
		t := T(0)
		for i, w := range wh {
			t += w
			if t > r {
				l = append(l, cv[i])
				wh = append(wh[:i], wh[i+1:]...)
				cv = append(cv[:i], cv[i+1:]...)
				break
			}
		}
	}
	return l
}

func WeightIndex[T ~int | ~int32 | ~int64](arr []T) int {
	tmp := make([]T, len(arr))
	copy(tmp, arr)

	sum := T(0)
	for _, w := range tmp {
		sum += w
	}

	v := T(rnd.Intn(int(sum)))
	t := T(0)
	for i, w := range tmp {
		t += w
		if t > v {
			return i
		}
	}

	return 0
}

// SelectByOdds 按概率 upNum/downNum 判定是否命中（几分之几的几率）
func SelectByOdds(upNum, downNum int32) bool {
	if downNum < 1 || upNum < 1 {
		return false
	}
	if upNum >= downNum {
		return true
	}
	return (1 + int32(rnd.Float64()*float64(downNum))) <= upNum
}

// RandInt 生成指定区间随机数
func RandInt[T ~int | ~int32 | ~int64](min, max T) T {
	if min >= max {
		return max
	}

	var ret T
	if min >= 0 && max >= 0 { // 正数
		v := rnd.Intn(int(max) - int(min) + 1)
		ret = T(v) + min
	} else {
		offset := -min

		_min := min + offset
		_max := max + offset
		v := rnd.Intn(int(_max) - int(_min) + 1)
		ret = T(v) + _min - offset
	}

	return ret
}
