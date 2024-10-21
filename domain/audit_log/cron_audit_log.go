package audit_log

import (
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"sync"
	"time"
)

type CronAuditLog struct {
	CronID         uint                      `json:"cronID" mapstructure:"cronID"`
	Name           string                    `json:"name" mapstructure:"name"`
	ExecutionOrder *sync.Map                 `json:"executionOrder" mapstructure:"executionOrder"`
	ExecutionLogs  *[]execLog                `json:"executionLogs" mapstructure:"executionLogs"`
	RequestData    *request_data.RequestData `json:"requestData" mapstructure:"requestData"`
	Start          time.Time                 `json:"start" mapstructure:"start"`
	End            time.Time                 `json:"end" mapstructure:"end"`
	TimeTaken      uint64                    `json:"timeTaken" mapstructure:"timeTaken"`
	FinalResponse  map[string]any            `json:"finalResponse" mapstructure:"finalResponse"`
}

func (l *CronAuditLog) Initialize(name string, requestData *request_data.RequestData) {
	l.Name = name
	l.ExecutionOrder = &sync.Map{}
	l.RequestData = requestData
	l.ExecutionLogs = &[]execLog{}

	now := time.Now()
	l.Start = now
	fmt.Printf("Cron job start: Name %s DateTime %s\n", name, l.Start.Format(common.DateTimeFormat))
}

func (l *CronAuditLog) InitExecOrder(flowId uint) {
	l.ExecutionOrder.Store(flowId, &[]ExecState{})
}

func (l *CronAuditLog) AddExecState(exState ExecState, flowId uint) {
	if execOrder, ok := l.ExecutionOrder.Load(flowId); ok {
		execOrderCasted := execOrder.(*[]ExecState)
		*execOrderCasted = append(*execOrderCasted, exState)
		l.ExecutionOrder.Store(flowId, execOrderCasted)
	}
}

func (l *CronAuditLog) AddExecLog(logUser string, logType string, logData any) {
	execLog := execLog{
		LogUser: logUser,
		LogType: logType,
		LogData: fmt.Sprint(logData),
	}

	if execLog.LogUser != common.LogUser && execLog.LogUser != common.LogSystem {
		execLog.LogUser = common.LogSystem
		execLog.LogType = common.LogError
		execLog.LogData = "invalid log attempt: illegal user"
	}

	if execLog.LogType != common.LogInfo && execLog.LogType != common.LogError {
		execLog.LogUser = common.LogSystem
		execLog.LogType = common.LogError
		execLog.LogData = "invalid log attempt: illegal type"
	}

	*l.ExecutionLogs = append(*l.ExecutionLogs, execLog)
}

func (l *CronAuditLog) EndLog() {
	l.End = time.Now()
	l.TimeTaken = uint64(l.End.Sub(l.Start).Milliseconds())
	fmt.Printf("Cron job end: Name %s DateTime %s Time taken %s\n",
		l.Name, l.End.Format(common.DateTimeFormat), fmt.Sprint(l.TimeTaken))
}

func (l *CronAuditLog) GetSystemErrorLogs() []string {
	errLogs := []string{}
	for _, log := range *l.ExecutionLogs {
		if log.LogUser == "system" && log.LogType == "error" {
			errLogs = append(errLogs, log.LogData)
		}
	}
	return errLogs
}

func (l *CronAuditLog) GetUserErrorLogs() []string {
	errLogs := []string{}
	for _, log := range *l.ExecutionLogs {
		if log.LogUser == "user" && log.LogType == "error" {
			errLogs = append(errLogs, log.LogData)
		}
	}
	return errLogs
}

func (l *CronAuditLog) SetFinalResponse(res map[string]any) {
	l.FinalResponse = res
}
