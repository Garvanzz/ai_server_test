package config

import "fmt"

var registry = map[string]func(raw any, name string) any{}

type Table[T any] struct {
	name string
}

// NewTable 注册一张配置表并返回其类型安全访问器。
func NewTable[T any](jsonName string) *Table[T] {
	t := &Table[T]{name: jsonName}
	registry[jsonName] = func(raw any, name string) any {
		return ParseToStruct[T](raw, name)
	}
	return t
}

// Get 按 ID 获取配置，不存在则
func (t *Table[T]) Get(id int64) T {
	m := t.All()
	v, ok := m[id]
	if !ok {
		panic(fmt.Sprintf("config %s: id %d not found", t.name, id))
	}
	return v
}

// Get32 是 int32 ID 的便捷方法。
func (t *Table[T]) Get32(id int32) T { return t.Get(int64(id)) }

// Find 按 ID 查找配置，返回值和是否存在。
func (t *Table[T]) Find(id int64) (T, bool) {
	m := t.All()
	v, ok := m[id]
	return v, ok
}

// Find32 是 int32 ID 的便捷方法。
func (t *Table[T]) Find32(id int32) (T, bool) { return t.Find(int64(id)) }

// All 返回整张配置表 map
func (t *Table[T]) All() map[int64]T {
	allConfigs := CfgMgr.getAll()
	return allConfigs[t.name].(map[int64]T)
}

// Range 遍历所有条目，回调返回 false 停止。
func (t *Table[T]) Range(fn func(int64, T) bool) {
	for id, v := range t.All() {
		if !fn(id, v) {
			return
		}
	}
}

// Len 返回条目数。
func (t *Table[T]) Len() int { return len(t.All()) }

// ---------------------------------------------------------------------------
// Single — 单对象配置 (JSON key-value → struct T)
// ---------------------------------------------------------------------------

// Single 提供对单对象配置（如 Global）的类型安全访问。
type Single[T any] struct {
	name string
}

// NewSingle 注册一个单对象配置并返回其类型安全访问器。
func NewSingle[T any](jsonName string) *Single[T] {
	s := &Single[T]{name: jsonName}
	registry[jsonName] = func(raw any, name string) any {
		return AttachToStruct[T](raw, name)
	}
	return s
}

// Get 返回配置对象
func (s *Single[T]) Get() T {
	allConfigs := CfgMgr.getAll()
	return allConfigs[s.name].(T)
}
