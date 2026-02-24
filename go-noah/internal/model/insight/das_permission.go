package insight

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
)

// ============ 权限模板 ============

// PermissionObject 权限对象（用于JSON存储）
type PermissionObject struct {
	InstanceID string `json:"instance_id"`
	Schema     string `json:"schema"`
	Table      string `json:"table,omitempty"` // 可选，表权限
}

// PermissionObjects 权限对象数组（用于JSON存储）
type PermissionObjects []PermissionObject

// Value 实现 driver.Valuer 接口
func (p PermissionObjects) Value() (driver.Value, error) {
	if len(p) == 0 {
		return "[]", nil
	}
	return json.Marshal(p)
}

// Scan 实现 sql.Scanner 接口
func (p *PermissionObjects) Scan(value interface{}) error {
	if value == nil {
		*p = PermissionObjects{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("cannot scan non-string value into PermissionObjects")
	}

	return json.Unmarshal(bytes, p)
}

// DASPermissionTemplate 权限模板
type DASPermissionTemplate struct {
	gorm.Model
	Name        string            `gorm:"type:varchar(128);not null;comment:模板名称;uniqueIndex:uk_name" json:"name"`
	Description string            `gorm:"type:varchar(512);default:'';comment:模板描述" json:"description"`
	Permissions PermissionObjects `gorm:"type:json;not null;comment:权限配置（JSON数组）" json:"permissions"`
}

func (DASPermissionTemplate) TableName() string {
	return "das_permission_templates"
}

// ============ 角色权限 ============

// PermissionType 权限类型
type PermissionType string

const (
	PermissionTypeObject  PermissionType = "object"  // 直接权限对象
	PermissionTypeTemplate PermissionType = "template" // 权限模板
)

// DASRolePermission 角色权限
type DASRolePermission struct {
	gorm.Model
	Role           string        `gorm:"type:varchar(128);not null;comment:角色标识（Casbin role）;index:idx_role" json:"role"`
	PermissionType PermissionType `gorm:"type:varchar(50);not null;comment:权限类型：object, template" json:"permission_type"`
	PermissionID   uint          `gorm:"type:bigint unsigned;not null;comment:权限ID（权限对象/模板/组的ID）;index:idx_permission" json:"permission_id"`
	// 直接权限对象时，存储具体权限信息（用于快速查询，避免关联查询）
	InstanceID     string        `gorm:"type:varchar(128);default:'';comment:实例ID（permission_type=object时使用）" json:"instance_id,omitempty"`
	Schema         string        `gorm:"type:varchar(128);default:'';comment:库名（permission_type=object时使用）" json:"schema,omitempty"`
	Table          string        `gorm:"type:varchar(128);default:'';comment:表名（permission_type=object时使用）" json:"table,omitempty"`
}

func (DASRolePermission) TableName() string {
	return "das_role_permissions"
}

// ============ 用户权限（与角色权限配置方式一致：object/template，无 rule）============

// DASUserPermission 用户权限，与角色权限同构：权限类型 object/template，无规则
type DASUserPermission struct {
	gorm.Model
	Username       string        `gorm:"type:varchar(128);not null;comment:用户名;index:idx_username" json:"username"`
	PermissionType PermissionType `gorm:"type:varchar(50);not null;comment:权限类型：object, template" json:"permission_type"`
	PermissionID   uint          `gorm:"type:bigint unsigned;not null;comment:权限ID（模板ID或对象占位）;index:idx_permission" json:"permission_id"`
	InstanceID     string        `gorm:"type:varchar(128);default:'';comment:实例ID（permission_type=object时使用）" json:"instance_id,omitempty"`
	Schema         string        `gorm:"type:varchar(128);default:'';comment:库名（permission_type=object时使用）" json:"schema,omitempty"`
	Table          string        `gorm:"type:varchar(128);default:'';comment:表名（permission_type=object时使用）" json:"table,omitempty"`
}

func (DASUserPermission) TableName() string {
	return "das_user_permissions"
}

