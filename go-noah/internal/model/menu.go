package model

import "gorm.io/gorm"

type Menu struct {
	gorm.Model
	ParentID   uint   `json:"parentId,omitempty" gorm:"column:parent_id;index;comment:父级菜单的id，使用整数表示"`     // 父级菜单的id，使用整数表示
	Path       string `json:"path" gorm:"column:path;type:varchar(255);comment:地址"`                        // 地址
	Title      string `json:"title" gorm:"column:title;type:varchar(100);comment:标题，使用字符串表示"`              // 标题，使用字符串表示
	Name       string `json:"name,omitempty" gorm:"column:name;type:varchar(100);comment:同路由中的name，用于保活"`  // 同路由中的name，用于保活
	Component  string `json:"component,omitempty" gorm:"column:component;type:varchar(255);comment:绑定的组件"` // 绑定的组件，默认类型：Iframe、RouteView、ComponentError
	Locale     string `json:"locale,omitempty" gorm:"column:locale;type:varchar(100);comment:本地化标识"`       // 本地化标识
	Icon       string `json:"icon,omitempty" gorm:"column:icon;type:varchar(100);comment:图标，使用字符串表示"`      // 图标，使用字符串表示
	Redirect   string `json:"redirect,omitempty" gorm:"column:redirect;type:varchar(255);comment:重定向地址"`   // 重定向地址
	URL        string `json:"url,omitempty" gorm:"column:url;type:varchar(255);comment:iframe模式下的跳转url"`   // iframe模式下的跳转url，不能与path重复
	KeepAlive  bool   `json:"keepAlive,omitempty" gorm:"column:keep_alive;default:false;comment:是否保活"`     // 是否保活
	HideInMenu bool   `json:"hideInMenu,omitempty" gorm:"column:hide_in_menu;default:false;comment:是否保活"`  // 是否保活
	Target     string `json:"target,omitempty" gorm:"column:target;type:varchar(20);comment:全连接跳转模式"`      // 全连接跳转模式：'_blank'、'_self'、'_parent'
	Weight     int    `json:"weight" gorm:"column:weight;type:int;default:0;comment:排序权重"`
	
	// Soybean-admin 格式字段
	MenuType   string `json:"menuType,omitempty" gorm:"column:menu_type;type:varchar(10);default:'2';comment:菜单类型:1-目录,2-菜单"` // 菜单类型:1-目录,2-菜单
	MenuName   string `json:"menuName,omitempty" gorm:"column:menu_name;type:varchar(100);comment:菜单名称"`                          // 菜单名称
	RouteName  string `json:"routeName,omitempty" gorm:"column:route_name;type:varchar(100);comment:路由名称"`                      // 路由名称
	RoutePath  string `json:"routePath,omitempty" gorm:"column:route_path;type:varchar(255);comment:路由路径"`                     // 路由路径
	I18nKey    string `json:"i18nKey,omitempty" gorm:"column:i18n_key;type:varchar(100);comment:国际化key"`                        // 国际化key
	IconType   string `json:"iconType,omitempty" gorm:"column:icon_type;type:varchar(10);default:'1';comment:图标类型:1-iconify,2-local"` // 图标类型:1-iconify,2-local
	Order      int    `json:"order,omitempty" gorm:"column:order;type:int;default:0;comment:排序"`                                 // 排序
	Status     string `json:"status,omitempty" gorm:"column:status;type:varchar(10);default:'1';comment:状态:1-启用,2-禁用"`          // 状态:1-启用,2-禁用
	MultiTab   bool   `json:"multiTab,omitempty" gorm:"column:multi_tab;default:false;comment:是否多标签"`                          // 是否多标签
	ActiveMenu string `json:"activeMenu,omitempty" gorm:"column:active_menu;type:varchar(100);comment:激活菜单"`                    // 激活菜单
	Constant   bool   `json:"constant,omitempty" gorm:"column:constant;default:false;comment:是否常量"`                              // 是否常量
	Href       string `json:"href,omitempty" gorm:"column:href;type:varchar(255);comment:外部链接"`                              // 外部链接
}

func (m *Menu) TableName() string {
	return "menu"
}
