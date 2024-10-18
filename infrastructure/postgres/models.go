package infrastructure

import (
	"ifttt/handler/domain/api"
	"time"

	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type apis struct {
	gorm.Model
	Name           string          `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
	Path           string          `gorm:"type:varchar(50);not null;unique" mapstructure:"path"`
	Method         string          `gorm:"type:varchar(10);not null" mapstructure:"method"`
	Description    string          `gorm:"type:text;default:''" mapstructure:"description"`
	Request        pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"request"`
	PreConfig      pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"preConfig"`
	TriggerFlowRef []trigger_flows `gorm:"many2many:api_trigger_flows;joinForeignKey:ApiId;joinReferences:FlowId;" mapstructure:"triggerFlows"`
	TriggerFlows   pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"triggerConditions"`
}

type api_trigger_flow_json struct {
	If      api.Condition `json:"if" mapstructure:"if"`
	Trigger uint          `json:"trigger" mapstructure:"trigger"`
}

type classes struct {
	gorm.Model
	Name string `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
}

type trigger_flows struct {
	gorm.Model
	Name        string       `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
	Description string       `gorm:"type:text;default:''" mapstructure:"description"`
	ClassId     uint         `gorm:"type:int;not null" mapstructure:"classId"`
	Class       classes      `mapstructure:"class"`
	StartRules  []rules      `gorm:"many2many:trigger_start_rules;joinForeignKey:FlowId;joinReferences:RuleId;" mapstructure:"startRules"`
	AllRules    []rules      `gorm:"many2many:trigger_all_rules;joinForeignKey:FlowId;joinReferences:RuleId;" mapstructure:"allRules"`
	BranchFlow  pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"branchFlows"`
}

type rules struct {
	gorm.Model
	Name        string       `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
	Description string       `gorm:"type:text;default:''" mapstructure:"description"`
	Pre         pgtype.JSONB `gorm:"type:jsonb;default:'[]';not null" mapstructure:"pre"`
	Switch      pgtype.JSONB `gorm:"type:jsonb;default:'{\"cases\":[],\"default\":{\"do\":[],\"return\":{\"resolveType\":\"const\",\"resolveData\":\"\"}}}';not null" mapstructure:"switch"`
}

type audit_log struct {
	gorm.Model
	ApiID          uint         `gorm:"not null"`
	Api            apis         `gorm:"foreignKey:ApiID" mapstructure:"apiID"`
	ApiName        string       `gorm:"type: varchar(50);not null" mapstructure:"apiName"`
	ApiPath        string       `gorm:"type: varchar(50);not null" mapstructure:"apiPath"`
	ExecutionOrder pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"executionOrder"`
	ExecutionLogs  pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"executionLogs"`
	RequestData    pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"requestData"`
	Start          time.Time    `gorm:"type:timestamp;not null" mapstructure:"start"`
	End            time.Time    `gorm:"type:timestamp;not null" mapstructure:"end"`
	TimeTaken      uint64       `gorm:"type:int;not null" mapstructure:"timeTaken"`
}
