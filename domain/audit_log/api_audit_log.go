package audit_log

import (
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"sync"
	"time"

	"github.com/google/uuid"
)

type APIAuditLog struct {
	ApiID            uint                      `json:"apiID" mapstructure:"apiID"`
	ApiName          string                    `json:"apiName" mapstructure:"apiName"`
	ApiPath          string                    `json:"apiPath" mapstructure:"apiPath"`
	RequestToken     string                    `json:"requestToken" mapstructure:"requestToken"`
	ExecutionOrder   *sync.Map                 `json:"executionOrder" mapstructure:"executionOrder"`
	ExecutionLogs    *ExecLogGrouped           `json:"executionLogs" mapstructure:"executionLogs"`
	RequestData      *request_data.RequestData `json:"requestData" mapstructure:"requestData"`
	Start            time.Time                 `json:"start" mapstructure:"start"`
	End              time.Time                 `json:"end" mapstructure:"end"`
	ExecTime         uint64                    `json:"execTime" mapstructure:"execTime"`
	InternalExecTime uint64                    `json:"internalExecTime" mapstructure:"internalExecTime"`
	ExternalExecTime uint64                    `json:"externalExecTime" mapstructure:"externalExecTime"`
	FinalResponse    map[string]any            `json:"finalResponse" mapstructure:"finalResponse"`
	ResponseSent     bool                      `json:"responseSent" mapstructure:"responseSent"`
}

func (l *APIAuditLog) Initialize(apiPath string, requestData *request_data.RequestData) {
	l.ApiPath = apiPath
	l.ExecutionOrder = &sync.Map{}
	l.RequestData = requestData
	l.ExecutionLogs = &ExecLogGrouped{}
	if token, err := uuid.NewRandom(); err != nil {
		l.RequestToken = common.RequestTokenDefault
	} else {
		l.RequestToken = token.String()
	}

	now := time.Now()
	l.Start = now
	fmt.Printf("token: %s | path %s | timestamp %s | request received \n", l.RequestToken, apiPath, l.Start.Format(common.DateTimeFormat))
}

func (l *APIAuditLog) InitExecOrder(flowId uint) {
	l.ExecutionOrder.Store(flowId, &[]ExecState{})
}

func (l *APIAuditLog) AddExecState(exState ExecState, flowId uint) {
	if execOrder, ok := l.ExecutionOrder.Load(flowId); ok {
		execOrderCasted := execOrder.(*[]ExecState)
		*execOrderCasted = append(*execOrderCasted, exState)
		l.ExecutionOrder.Store(flowId, execOrderCasted)
	}
}

func (l *APIAuditLog) AddExecLog(logUser string, logType string, logData any) {
	log := execLog{
		LogUser: logUser,
		LogType: logType,
		LogData: fmt.Sprint(logData),
	}

	switch {
	case log.LogUser == common.LogUser && log.LogType == common.LogInfo:
		l.ExecutionLogs.UserInfo = append(l.ExecutionLogs.UserInfo, log)
	case log.LogUser == common.LogUser && log.LogType == common.LogError:
		l.ExecutionLogs.UserError = append(l.ExecutionLogs.UserError, log)
	case log.LogUser == common.LogSystem && log.LogType == common.LogError:
		l.ExecutionLogs.SystemError = append(l.ExecutionLogs.SystemError, log)
	case log.LogUser == common.LogSystem && log.LogType == common.LogInfo:
		l.ExecutionLogs.SystemInfo = append(l.ExecutionLogs.SystemInfo, log)
	}
}

func (l *APIAuditLog) EndLog() {
	l.End = time.Now()
	l.ExecTime = uint64(l.End.Sub(l.Start).Milliseconds())
	l.InternalExecTime = l.ExecTime - l.ExternalExecTime
	fmt.Printf("token: %s | path: %s | timestamp: %s | execution Time: %d (internal: %d, external: %d) | request ending\n",
		l.RequestToken, l.ApiPath, l.End.Format(common.DateTimeFormat), l.ExecTime, l.InternalExecTime, l.ExternalExecTime)
}

func (l *APIAuditLog) GetLogs() *ExecLogGrouped {
	return l.ExecutionLogs
}

func (l *APIAuditLog) SetFinalResponse(res map[string]any) {
	l.FinalResponse = res
}

func (l *APIAuditLog) SetResponseSent() bool {
	if l.ResponseSent {
		return false
	}
	l.ResponseSent = true
	return true
}

func (l *APIAuditLog) AddExternalTime(t uint64) {
	l.ExternalExecTime += t
}

func (l *APIAuditLog) GetRequestToken() string {
	return l.RequestToken
}
