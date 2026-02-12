package initializer

import (
	"fmt"
	"sort"
)

// InitializerRegistry 初始化器注册表
type InitializerRegistry struct {
	initializers initSlice
	cache        map[string]*orderedInitializer
}

var registry = &InitializerRegistry{
	initializers: make(initSlice, 0),
	cache:        make(map[string]*orderedInitializer),
}

// Register 注册初始化器
func Register(init Initializer) {
	name := init.Name()
	if _, existed := registry.cache[name]; existed {
		panic(fmt.Sprintf("初始化器名称冲突: %s", name))
	}

	ordered := &orderedInitializer{
		order:       init.Order(),
		Initializer: init,
	}

	registry.initializers = append(registry.initializers, ordered)
	registry.cache[name] = ordered
}

// GetAll 获取所有初始化器（已排序）
func GetAll() []Initializer {
	// 按 Order 排序
	sort.Sort(&registry.initializers)

	result := make([]Initializer, len(registry.initializers))
	for i, ordered := range registry.initializers {
		result[i] = ordered.Initializer
	}
	return result
}

// GetByName 根据名称获取初始化器
func GetByName(name string) (Initializer, bool) {
	ordered, ok := registry.cache[name]
	if !ok {
		return nil, false
	}
	return ordered.Initializer, true
}

// Clear 清空注册表（用于测试）
func Clear() {
	registry.initializers = make(initSlice, 0)
	registry.cache = make(map[string]*orderedInitializer)
}
