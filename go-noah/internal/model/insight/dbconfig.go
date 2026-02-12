package insight

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UseType 数据库用途类型
type UseType string

const (
	UseTypeQuery UseType = "查询"
	UseTypeOrder UseType = "工单"
)

// DbType 数据库类型
type DbType string

const (
	DbTypeMySQL      DbType = "MySQL"
	DbTypeTiDB       DbType = "TiDB"
	DbTypeClickHouse DbType = "ClickHouse"
)

// DBConfig 数据库实例配置
type DBConfig struct {
	gorm.Model
	InstanceID       uuid.UUID      `gorm:"type:char(36);uniqueIndex:uniq_instance_id" json:"instance_id"`
	Hostname         string         `gorm:"type:varchar(128);not null;default:'';uniqueIndex:uniq_hostname;comment:主机名" json:"hostname"`
	Port             int            `gorm:"type:int;not null;default:3306;uniqueIndex:uniq_hostname;comment:端口" json:"port"`
	UserName         string         `gorm:"type:varchar(128);not null;default:'';comment:账号" json:"user_name"`
	Password         string         `gorm:"type:varchar(256);not null;default:'';comment:密码" json:"password"`
	UseType          UseType        `gorm:"type:varchar(20);default:'工单';uniqueIndex:uniq_hostname;comment:用途" json:"use_type"`
	DbType           DbType         `gorm:"type:varchar(20);default:'MySQL';comment:数据库类型" json:"db_type"`
	Environment      int            `gorm:"type:int;null;default:null;comment:环境;index" json:"environment"`
	InspectParams    datatypes.JSON `gorm:"type:json;null;default:null;comment:语法审核参数" json:"inspect_params"`
	OrganizationKey  string         `gorm:"type:varchar(256);not null;index:organization_key;comment:搜索路径" json:"organization_key"`
	OrganizationPath datatypes.JSON `gorm:"type:json;null;default:null;comment:绝对路径" json:"organization_path"`
	Remark           string         `gorm:"type:varchar(256);not null;default:'';comment:备注" json:"remark"`
}

func (DBConfig) TableName() string {
	return "db_configs"
}

func (u *DBConfig) BeforeCreate(tx *gorm.DB) (err error) {
	u.InstanceID, _ = uuid.NewUUID()
	return
}

// DBSchema 自动采集的库信息
type DBSchema struct {
	gorm.Model
	InstanceID uuid.UUID `gorm:"type:char(36);comment:关联db_configs的instance_id;uniqueIndex:uniq_schema" json:"instance_id"`
	Schema     string    `gorm:"type:varchar(128);not null;default:'';comment:库名;uniqueIndex:uniq_schema" json:"schema"`
	IsDeleted  bool      `gorm:"type:boolean;not null;default:false;comment:是否删除;uniqueIndex:uniq_schema" json:"is_deleted"`
}

func (DBSchema) TableName() string {
	return "db_schemas"
}

