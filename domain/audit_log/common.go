package audit_log

import (
	"context"
	"fmt"
	"ifttt/handler/common"
)

type ExecState struct {
	State uint `json:"state" mapstructure:"state"`
	Rule  uint `json:"rule" mapstructure:"rule"`
}

type ExecLogGrouped struct {
	UserInfo    []execLog `json:"userInfo" mapstructure:"userInfo"`
	UserError   []execLog `json:"userError" mapstructure:"userError"`
	SystemInfo  []execLog `json:"systemInfo" mapstructure:"systemInfo"`
	SystemError []execLog `json:"systemErrors" mapstructure:"systemError"`
}

type execLog struct {
	LogUser string `json:"logUser" mapstructure:"logUser"`
	LogType string `json:"logType" mapstructure:"logType"`
	LogData string `json:"logData" mapstructure:"logData"`
}

func GetAuditLogFromContext(ctx context.Context) *AuditLog {
	if log, ok := common.GetRequestState(ctx).Load(common.ContextLog); ok {
		return log.(*AuditLog)
	}
	return nil
}

func AddExecLog(logUser string, logType string, logData any, ctx context.Context) {
	log := GetAuditLogFromContext(ctx)
	if log != nil {
		fmt.Printf("token: %s | user: %s | type: %s | log: %s\n",
			(*log).GetRequestToken(), logUser, logType, logData)
		(*log).AddExecLog(logUser, logType, logData)
	} else {
		fmt.Printf("token: %s | user: %s | type: %s | log: %s\n",
			common.RequestTokenDefault, logUser, logType, logData)
	}
}
