package initializer

import (
	"context"
	"gorm.io/gorm"
)

// Initializer 初始化器接口
type Initializer interface {
	// Name 返回初始化器名称
	Name() string

	// Order 返回初始化顺序（数字越小越先执行）
	Order() int

	// MigrateTable 创建表结构
	MigrateTable(ctx context.Context, db *gorm.DB) error

	// InitializeData 初始化数据
	InitializeData(ctx context.Context, db *gorm.DB) error

	// IsTableCreated 检查表是否已创建
	IsTableCreated(ctx context.Context, db *gorm.DB) bool

	// IsDataInitialized 检查数据是否已初始化
	IsDataInitialized(ctx context.Context, db *gorm.DB) bool
}

// orderedInitializer 带顺序的初始化器
type orderedInitializer struct {
	order int
	Initializer
}

// initSlice 初始化器切片，用于排序
type initSlice []*orderedInitializer

// Len 实现 sort.Interface
func (s initSlice) Len() int {
	return len(s)
}

// Less 实现 sort.Interface
func (s initSlice) Less(i, j int) bool {
	return s[i].order < s[j].order
}

// Swap 实现 sort.Interface
func (s initSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
