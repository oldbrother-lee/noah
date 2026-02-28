package insight

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DASUserSchemaPermission 用户的库权限
type DASUserSchemaPermission struct {
	gorm.Model
	Username   string    `gorm:"type:varchar(128);not null;comment:用户;uniqueIndex:uniq_schema" json:"username"`
	Schema     string    `gorm:"type:varchar(128);not null;default:'';comment:库名;uniqueIndex:uniq_schema" json:"schema"`
	InstanceID uuid.UUID `gorm:"type:char(36);comment:关联db_configs的instance_id;uniqueIndex:uniq_schema;index:idx_instance_id" json:"instance_id"`
}

func (DASUserSchemaPermission) TableName() string {
	return "das_user_schema_permissions"
}

// RuleType 规则类型
type RuleType string

const (
	RuleAllow RuleType = "allow"
	RuleDeny  RuleType = "deny"
)

// DASUserTablePermission 用户的表权限
type DASUserTablePermission struct {
	gorm.Model
	Username   string    `gorm:"type:varchar(128);not null;comment:用户;uniqueIndex:uniq_table" json:"username"`
	Schema     string    `gorm:"type:varchar(128);not null;default:'';comment:库名;uniqueIndex:uniq_table" json:"schema"`
	Table      string    `gorm:"type:varchar(128);not null;default:'';comment:表名;uniqueIndex:uniq_table" json:"table"`
	InstanceID uuid.UUID `gorm:"type:char(36);comment:关联db_configs的instance_id;uniqueIndex:uniq_table;index:idx_instance_id" json:"instance_id"`
	Rule       RuleType  `gorm:"type:varchar(10);default:'allow';comment:规则" json:"rule"`
}

func (DASUserTablePermission) TableName() string {
	return "das_user_table_permissions"
}

// DASAllowedOperation 允许用户执行的操作
type DASAllowedOperation struct {
	gorm.Model
	Name     string `gorm:"type:varchar(128);not null;comment:语句类型;uniqueIndex:uniq_name" json:"name"`
	IsEnable bool   `gorm:"type:boolean;null;default:false;comment:是否启用,0未启用,1启用" json:"is_enable"`
	Remark   string `gorm:"type:varchar(1024);not null;default:'';comment:备注" json:"remark"`
}

func (DASAllowedOperation) TableName() string {
	return "das_allowed_operations"
}

// DASRecord SQL执行记录
type DASRecord struct {
	gorm.Model
	Username   string    `gorm:"type:varchar(128);not null;index:idx_username;comment:用户" json:"username"`
	InstanceID uuid.UUID `gorm:"type:char(36);index:idx_instance_id;comment:实例ID" json:"instance_id"`
	Schema     string    `gorm:"type:varchar(128);not null;comment:库名" json:"schema"`
	SQL        string    `gorm:"type:text;comment:SQL语句" json:"sql"`
	Duration   int64     `gorm:"type:bigint;comment:执行时长(ms)" json:"duration"`
	RowCount   int64     `gorm:"type:bigint;comment:返回行数" json:"row_count"`
	Error      string    `gorm:"type:text;comment:错误信息" json:"error"`
	IsFinish   bool      `gorm:"type:boolean;default:false;comment:是否已完成,0未完成,1已完成" json:"is_finish"`
}

func (DASRecord) TableName() string {
	return "das_records"
}

// DASFavorite 收藏夹
type DASFavorite struct {
	gorm.Model
	Username string `gorm:"type:varchar(128);not null;index:idx_username;comment:用户" json:"username"`
	Title    string `gorm:"type:varchar(256);not null;comment:标题" json:"title"`
	SQL      string `gorm:"type:text;comment:SQL语句" json:"sql"`
}

func (DASFavorite) TableName() string {
	return "das_favorites"
}

// UserAuthorizedSchema 用户授权的 schema 返回结构（用于关联查询）
type UserAuthorizedSchema struct {
	InstanceID string `json:"instance_id"`
	Schema     string `json:"schema"`
	DbType     string `json:"db_type"`
	Hostname   string `json:"hostname"`
	Port       int    `json:"port"`
	IsDeleted  bool   `json:"is_deleted"`
	Remark     string `json:"remark"`
}

