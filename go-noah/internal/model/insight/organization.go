package insight

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Organization 组织架构
// Key 全局唯一
// ParentID+Name 组成唯一索引，表示同一级别下节点名不能重复
type Organization struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(32);not null;uniqueIndex:uniq_name;comment:节点名" json:"name"`
	ParentID  uint64         `gorm:"not null;default:0;uniqueIndex:uniq_name;comment:父节点ID,0值表示父节点" json:"parent_id"`
	Key       string         `gorm:"type:varchar(256);default:null;uniqueIndex:uniq_key;comment:搜索路径" json:"key"`
	Level     uint64         `gorm:"not null;default:1;comment:当前节点到根节点的距离或者层级,父节点起始值为1" json:"level"`
	Path      datatypes.JSON `gorm:"type:json;null;default:null;comment:绝对路径" json:"path"`
	Creator   string         `gorm:"type:varchar(64);default:null;comment:创建人" json:"creator"`
	Updater   string         `gorm:"type:varchar(64);default:null;comment:更新人" json:"updater"`
	CreatedAt time.Time      `gorm:"index:idx_created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"index:idx_updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (Organization) TableName() string {
	return "organizations"
}

// OrganizationUser 用户组织关联
type OrganizationUser struct {
	gorm.Model
	UID             uint64 `gorm:"type:bigint;not null;uniqueIndex:uniq_uid;comment:用户ID" json:"uid"`
	OrganizationKey string `gorm:"type:varchar(256);not null;index:organization_key;comment:搜索路径" json:"organization_key"`
}

func (OrganizationUser) TableName() string {
	return "organization_users"
}

