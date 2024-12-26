package infrastructure

import (
	"ifttt/handler/domain/api"

	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type crons struct {
	gorm.Model
	Name           string          `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
	Description    string          `gorm:"type:text;default:''" mapstructure:"description"`
	Cron           string          `gorm:"type:varchar(30);default:''" mapstructure:"description"`
	PreConfig      pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"preConfig"`
	TriggerFlowRef []trigger_flows `gorm:"many2many:cron_trigger_flows;joinForeignKey:CronId;joinReferences:FlowId;" mapstructure:"triggerFlows"`
	TriggerFlows   pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"triggerConditions"`
}

type apis struct {
	gorm.Model
	Name         string          `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
	Path         string          `gorm:"type:varchar(50);not null;unique" mapstructure:"path"`
	Method       string          `gorm:"type:varchar(10);not null" mapstructure:"method"`
	Description  string          `gorm:"type:text;default:''" mapstructure:"description"`
	Request      pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"request"`
	PreConfig    pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"preConfig"`
	Triggers     []trigger_flows `gorm:"many2many:api_trigger_flows_main;joinForeignKey:ApiId;joinReferences:FlowId;" mapstructure:"triggerFlows"`
	TriggerFlows pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"triggerConditions"`
}

type api_trigger_flow_json struct {
	If      api.Condition `json:"if" mapstructure:"if"`
	Trigger uint          `json:"trigger" mapstructure:"trigger"`
}

type trigger_flows struct {
	gorm.Model
	Name        string       `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
	Description string       `gorm:"type:text;default:''" mapstructure:"description"`
	StartState  uint         `gorm:"type:int;not null" mapstructure:"startState"`
	Rules       []rules      `gorm:"many2many:trigger_rules;joinForeignKey:FlowId;joinReferences:RuleId;" mapstructure:"rules"`
	BranchFlows pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"branchFlows"`
}

type rules struct {
	gorm.Model
	Name        string       `gorm:"type:varchar(50);not null;unique" mapstructure:"name"`
	Description string       `gorm:"type:text;default:''" mapstructure:"description"`
	Pre         pgtype.JSONB `gorm:"type:jsonb;default:'[]';not null" mapstructure:"pre"`
	Switch      pgtype.JSONB `gorm:"type:jsonb;default:'{\"cases\":[],\"default\":{\"do\":[],\"return\":{\"resolveType\":\"const\",\"resolveData\":\"\"}}}';not null" mapstructure:"switch"`
	Finally     pgtype.JSONB `gorm:"type:jsonb;default:'[]';not null" mapstructure:"finally"`
}

type data_scehma struct {
	gorm.Model
	Table   string                `gorm:"type:varchar(50);not null;unique" mapstructure:"table"`
	Columns []data_schema_columns `gorm:"many2many:data_schema_column_bindings;joinForeignKey:SchemaId;joinReferences:ColumnId;" mapstructure:"columns"`
}

type data_schema_columns struct {
	Get      bool        `mapstructure:"get"`
	As       string      `gorm:"type:bool;default:true" mapstructure:"as"`
	DataType string      `mapstructure:"data_type"`
	SubModel data_scehma `mapstructure:"sub_model"`
	Populate bool        `gorm:"type:bool;default:false" mapstructure:"populate"`
}
