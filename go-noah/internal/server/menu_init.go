package server

import (
	"context"
	"encoding/json"
	"go-noah/api"
	"go-noah/internal/model"
	"go-noah/pkg/log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitializeMenuDataIfNeeded 检查并初始化菜单数据（如果不存在则初始化）
func InitializeMenuDataIfNeeded(ctx context.Context, db *gorm.DB, logger *log.Logger) error {
	// 检查菜单数据是否已存在
	var count int64
	if err := db.Model(&model.Menu{}).Count(&count).Error; err != nil {
		logger.Error("检查菜单数据失败", zap.Error(err))
		return err
	}

	// 如果已有菜单数据，跳过初始化
	if count > 0 {
		logger.Debug("菜单数据已存在，跳过初始化", zap.Int64("count", count))
		return nil
	}

	// 从 context 获取 menuData，如果没有则从 GetMenuData() 获取
	menuDataStr, ok := ctx.Value("menuData").(string)
	if !ok || menuDataStr == "" {
		menuDataStr = GetMenuData()
	}

	menuList := make([]api.MenuDataItem, 0)
	if err := json.Unmarshal([]byte(menuDataStr), &menuList); err != nil {
		logger.Error("解析菜单数据失败", zap.Error(err))
		return err
	}

	// 只创建不存在的菜单
	createdCount := 0
	for _, item := range menuList {
		var existingMenu model.Menu
		if err := db.Where("id = ?", item.ID).First(&existingMenu).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 菜单不存在，创建新菜单
				menu := model.Menu{
					Model: gorm.Model{
						ID: item.ID,
					},
					ParentID:   item.ParentID,
					Path:       item.Path,
					Title:      item.Title,
					Name:       item.Name,
					Component:  item.Component,
					Locale:     item.Locale,
					Weight:     item.Weight,
					Icon:       item.Icon,
					Redirect:   item.Redirect,
					URL:        item.URL,
					KeepAlive:  item.KeepAlive,
					HideInMenu: item.HideInMenu,
				}
				if err := db.Create(&menu).Error; err != nil {
					logger.Warn("创建菜单失败", zap.Uint("id", item.ID), zap.Error(err))
				} else {
					createdCount++
				}
			} else {
				logger.Warn("检查菜单失败", zap.Uint("id", item.ID), zap.Error(err))
			}
		}
		// 如果菜单已存在，跳过创建
	}

	if createdCount > 0 {
		logger.Info("菜单初始化完成", zap.Int("created_count", createdCount))
	} else {
		logger.Debug("菜单已存在，跳过初始化")
	}

	return nil
}
