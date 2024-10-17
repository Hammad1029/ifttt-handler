package audit_log

import (
	"fmt"
	"time"
)

type AuditLog struct {
	ApiGroup       string    `json:"apiGroup" mapstructure:"apiGroup"`
	ApiName        string    `json:"apiName" mapstructure:"apiName"`
	ExecutionOrder []uint    `json:"executionOrder" mapstructure:"executionOrder"`
	ExecutionLogs  []ExecLog `json:"executionLogs" mapstructure:"executionLogs"`
	Start          time.Time `json:"start" mapstructure:"start"`
	StartPartition time.Time `json:"startPartition" mapstructure:"startPartition"`
	End            time.Time `json:"end" mapstructure:"end"`
	TimeTaken      int       `json:"timeTaken" mapstructure:"timeTaken"`
}

type PostableAuditLog struct {
	ApiGroup       string    `json:"apiGroup" mapstructure:"apiGroup"`
	ApiName        string    `json:"apiName" mapstructure:"apiName"`
	ExecutionOrder []string  `json:"executionOrder" mapstructure:"executionOrder"`
	ExecutionLogs  []ExecLog `json:"executionLogs" mapstructure:"executionLogs"`
	Start          time.Time `json:"start" mapstructure:"start"`
	StartPartition time.Time `json:"startPartition" mapstructure:"startPartition"`
	End            time.Time `json:"end" mapstructure:"end"`
	TimeTaken      int       `json:"timeTaken" mapstructure:"timeTaken"`
}

type ExecLog struct {
	LogUser string `json:"logUser" mapstructure:"logUser"`
	LogType string `json:"logType" mapstructure:"logType"`
	LogData string `json:"logData" mapstructure:"logData"`
}

func (l *AuditLog) StartLog() {
	now := time.Now()
	fmt.Printf("Request recieved at %s\n", now.String())
	l.Start = now
}

func (l *AuditLog) Initialize(apiGroup string, apiName string) {
	// l.ApiGroup = apiGroup
	// l.ApiName = apiName
	// l.ExecutionOrder = []string{}
	// l.RequestData = r

	// timeSlot, err := strconv.Atoi(config.GetConfigProp("app.logPartitionSeconds"))
	// if err != nil {
	// 	l.AddExecLog("system", "error", err.Error())
	// }
	// l.StartPartition = common.GetTimeSlot(l.Start, timeSlot)
}

func (l *AuditLog) Post() (*PostableAuditLog, error) {
	return nil, nil
	// postableLog := PostableAuditLog{
	// 	ApiGroup:       l.ApiGroup,
	// 	ApiName:        l.ApiName,
	// 	ExecutionOrder: l.ExecutionOrder,
	// 	ExecutionLogs:  l.ExecutionLogs,
	// 	Start:          l.Start,
	// 	StartPartition: l.StartPartition,
	// }

	// if serializedRequestData, err := l.RequestData.Serialize(); err == nil {
	// 	postableLog.RequestData = request_data.SerializedRequestData{
	// 		ReqBody:  serializedRequestData.ReqBody,
	// 		Store:    serializedRequestData.Store,
	// 		Response: serializedRequestData.Response,
	// 		QueryRes: serializedRequestData.QueryRes,
	// 		ApiRes:   serializedRequestData.ApiRes,
	// 	}
	// } else {
	// 	return postableLog, fmt.Errorf("method ScyllaAuditLogRepository.InsertLog: error in serializing request data, %s", err)
	// }

	// postableLog.End = time.Now()
	// timeSubtracted := postableLog.End.Sub(l.Start)
	// postableLog.TimeTaken = int(timeSubtracted.Microseconds())

	// return postableLog, nil
}

func (l *AuditLog) AddExecLog(logUser string, logType string, AuditLog string) {
	execLog := ExecLog{
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

	l.ExecutionLogs = append(l.ExecutionLogs, execLog)
}

func (l *AuditLog) GetSystemErrorLogs() []string {
	errLogs := []string{}
	for _, log := range l.ExecutionLogs {
		if log.LogUser == "system" && log.LogType == "error" {
			errLogs = append(errLogs, log.LogData)
		}
	}
	return errLogs
}

func (l *AuditLog) GetUserErrorLogs() []string {
	errLogs := []string{}
	for _, log := range l.ExecutionLogs {
		if log.LogUser == "user" && log.LogType == "error" {
			errLogs = append(errLogs, log.LogData)
		}
	}
	return errLogs
}
