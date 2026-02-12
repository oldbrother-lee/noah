package insight

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// InspectParams SQL审核参数配置
type InspectParams struct {
	gorm.Model
	Params datatypes.JSON `gorm:"type:json;null;default:null;comment:语法审核参数" json:"params"`
	Remark string         `gorm:"type:varchar(256);null;default:null;uniqueIndex:uniq_remark;comment:备注" json:"remark"`
}

func (InspectParams) TableName() string {
	return "inspect_params"
}

