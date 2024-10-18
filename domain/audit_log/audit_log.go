package audit_log

import (
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
	"sync"
	"time"
)

type AuditLog struct {
	ApiID          uint                      `json:"apiID" mapstructure:"apiID"`
	ApiName        string                    `json:"apiName" mapstructure:"apiName"`
	ApiPath        string                    `json:"apiPath" mapstructure:"apiPath"`
	ExecutionOrder *sync.Map                 `json:"executionOrder" mapstructure:"executionOrder"`
	ExecutionLogs  *[]execLog                `json:"executionLogs" mapstructure:"executionLogs"`
	RequestData    *request_data.RequestData `json:"requestData" mapstructure:"requestData"`
	Start          time.Time                 `json:"start" mapstructure:"start"`
	End            time.Time                 `json:"end" mapstructure:"end"`
	TimeTaken      uint64                    `json:"timeTaken" mapstructure:"timeTaken"`
}

type execLog struct {
	LogUser string `json:"logUser" mapstructure:"logUser"`
	LogType string `json:"logType" mapstructure:"logType"`
	LogData string `json:"logData" mapstructure:"logData"`
}

func (l *AuditLog) Initialize(apiPath string, requestData *request_data.RequestData) {
	l.ApiPath = apiPath
	l.ExecutionOrder = &sync.Map{}
	l.RequestData = requestData
	l.ExecutionLogs = &[]execLog{}

	now := time.Now()
	l.Start = now
	fmt.Printf("Request recieved: Path %s DateTime %s\n", apiPath, l.Start.Format(common.DateTimeFormat))
}

func (l *AuditLog) AddExecLog(logUser string, logType string, AuditLog string) {
	execLog := execLog{
		LogUser: logUser,
		LogType: logType,
		LogData: AuditLog,
	}

	if execLog.LogUser != "user" && execLog.LogUser != "system" {
		execLog.LogUser = "system"
		execLog.LogType = "error"
		execLog.LogData = "invalid log attempt: illegal user"
	}

	if execLog.LogType != "info" && execLog.LogType != "error" {
		execLog.LogUser = "system"
		execLog.LogType = "error"
		execLog.LogData = "invalid log attempt: illegal type"
	}

	*l.ExecutionLogs = append(*l.ExecutionLogs, execLog)
}

func (l *AuditLog) EndLog() {
	l.End = time.Now()
	l.TimeTaken = uint64(l.End.Sub(l.Start).Milliseconds())
	fmt.Printf("Request ended: Path %s DateTime %s Time taken %s\n",
		l.ApiPath, l.End.Format(common.DateTimeFormat), fmt.Sprint(l.TimeTaken))
}

func (l *AuditLog) GetSystemErrorLogs() []string {
	errLogs := []string{}
	for _, log := range *l.ExecutionLogs {
		if log.LogUser == "system" && log.LogType == "error" {
			errLogs = append(errLogs, log.LogData)
		}
	}
	return errLogs
}

func (l *AuditLog) GetUserErrorLogs() []string {
	errLogs := []string{}
	for _, log := range *l.ExecutionLogs {
		if log.LogUser == "user" && log.LogType == "error" {
			errLogs = append(errLogs, log.LogData)
		}
	}
	return errLogs
}
