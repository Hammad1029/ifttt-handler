package infrastructure

import (
	"ifttt/handler/domain/api"
	"time"

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
	PreWare      []trigger_flows `gorm:"many2many:api_trigger_flows_pre;joinForeignKey:ApiId;joinReferences:FlowId;" mapstructure:"triggerFlows"`
	MainWare     []trigger_flows `gorm:"many2many:api_trigger_flows_main;joinForeignKey:ApiId;joinReferences:FlowId;" mapstructure:"triggerFlows"`
	PostWare     []trigger_flows `gorm:"many2many:api_trigger_flows_post;joinForeignKey:ApiId;joinReferences:FlowId;" mapstructure:"triggerFlows"`
	TriggerFlows pgtype.JSONB    `gorm:"type:jsonb;default:'{}';not null" mapstructure:"triggerConditions"`
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
}

type api_audit_log struct {
	gorm.Model
	ApiID            uint         `gorm:"default:null"`
	Api              apis         `gorm:"foreignKey:ApiID" mapstructure:"apiID"`
	ApiName          string       `gorm:"type: varchar(50);not null" mapstructure:"apiName"`
	ApiPath          string       `gorm:"type: varchar(50);not null" mapstructure:"apiPath"`
	RequestToken     string       `gorm:"varchar(50)" mapstructure:"requestToken"`
	ExecutionOrder   pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"executionOrder"`
	ExecutionLogs    pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"executionLogs"`
	RequestData      pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"requestData"`
	Start            time.Time    `gorm:"type:timestamp;not null" mapstructure:"start"`
	End              time.Time    `gorm:"type:timestamp;not null" mapstructure:"end"`
	ExecTime         uint64       `gorm:"type:int;not null" mapstructure:"execTime"`
	InternalExecTime uint64       `gorm:"type:int;not null" mapstructure:"internalExecTime"`
	ExternalExecTime uint64       `gorm:"type:int;not null" mapstructure:"externalExecTime"`
	FinalResponse    pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"finalResponse"`
	ResponseSent     bool         `gorm:"type:boolean;default:false;not null" mapstructure:"responseSent"`
}

type cron_audit_log struct {
	gorm.Model
	CronID           uint         `gorm:"default:null"`
	Cron             crons        `gorm:"foreignKey:CronID" mapstructure:"cronID"`
	CronName         string       `gorm:"type:varchar(50);not null" mapstructure:"cronName"`
	RequestToken     string       `gorm:"varchar(50)" mapstructure:"requestToken"`
	ExecutionOrder   pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"executionOrder"`
	ExecutionLogs    pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"executionLogs"`
	RequestData      pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"requestData"`
	Start            time.Time    `gorm:"type:timestamp;not null" mapstructure:"start"`
	End              time.Time    `gorm:"type:timestamp;not null" mapstructure:"end"`
	ExecTime         uint64       `gorm:"type:int;not null" mapstructure:"execTime"`
	InternalExecTime uint64       `gorm:"type:int;not null" mapstructure:"internalExecTime"`
	ExternalExecTime uint64       `gorm:"type:int;not null" mapstructure:"externalExecTime"`
	FinalResponse    pgtype.JSONB `gorm:"type:jsonb;default:'{}';not null" mapstructure:"finalResponse"`
	ResponseSent     bool         `gorm:"type:boolean;default:false;not null" mapstructure:"responseSent"`
}
