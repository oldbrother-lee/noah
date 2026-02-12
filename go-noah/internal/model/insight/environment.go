package insight

import "gorm.io/gorm"

// DBEnvironment 工单环境
type DBEnvironment struct {
	gorm.Model
	Name string `gorm:"type:varchar(32);not null;default:'';comment:环境名;uniqueIndex:uniq_name" json:"name"`
}

func (DBEnvironment) TableName() string {
	return "db_environments"
}

