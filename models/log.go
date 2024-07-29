package models

import (
	"fmt"
	"handler/common"
	"handler/config"
	"handler/scylla"
	"strconv"
	"time"

	"github.com/scylladb/gocqlx/v2/table"
)

type LogModel struct {
	ApiGroup       string           `cql:"api_group"`
	ApiName        string           `cql:"api_name"`
	ExecutionOrder []string         `cql:"execution_order"`
	ExecutionLogs  []ExecLog        `cql:"execution_logs"`
	RequestData    RequestDataModel `cql:"request_data"`
	Start          time.Time        `cql:"start"`
	StartPartition time.Time        `cql:"start_partition"`
	End            time.Time        `cql:"end"`
	TimeTaken      int              `cql:"time_taken"`
}

type LogData struct {
	ApiGroup       string       `json:"apiGroup" mapstructure:"apiGroup"`
	ApiName        string       `json:"apiName" mapstructure:"apiName"`
	ExecutionOrder []string     `json:"executionOrder" mapstructure:"executionOrder"`
	ExecutionLogs  []ExecLog    `json:"executionLogs" mapstructure:"executionLogs"`
	RequestData    *RequestData `json:"requestData" mapstructure:"requestData"`
	Start          time.Time    `json:"start" mapstructure:"start"`
	StartPartition time.Time    `json:"startPartition" mapstructure:"startPartition"`
	End            time.Time    `json:"end" mapstructure:"end"`
	TimeTaken      int          `json:"timeTaken" mapstructure:"timeTaken"`
}

var LogsMetadata = table.Metadata{
	Name:    "Logs",
	Columns: []string{"api_group", "api_name", "execution_order", "execution_logs", "request_data", "start", "start_partition", "end", "time_taken"},
	PartKey: []string{"api_group", "start_partition"},
	SortKey: []string{"api_name", "start"},
}

type ExecLog struct {
	LogUser string `cql:"log_user"`
	LogType string `cql:"log_type"`
	LogData string `cql:"log_data"`
}

func (l *LogData) StartLog() {
	now := time.Now()
	fmt.Printf("Request recieved at %s\n", now.String())
	l.Start = now
}

func (l *LogData) Initialize(r *RequestData, api *ApiModel) {
	l.ApiGroup = api.ApiGroup
	l.ApiName = api.ApiName
	l.ExecutionOrder = []string{}
	l.RequestData = r

	timeSlot, err := strconv.Atoi(config.GetConfigProp("app.logPartitionSeconds"))
	if err != nil {
		l.AddExecLog("system", "error", err.Error())
	}
	l.StartPartition = common.GetTimeSlot(l.Start, timeSlot)
}

func (l *LogData) Post() error {
	newLog := LogModel{
		ApiGroup:       l.ApiGroup,
		ApiName:        l.ApiName,
		ExecutionOrder: l.ExecutionOrder,
		ExecutionLogs:  l.ExecutionLogs,
		Start:          l.Start,
		StartPartition: l.StartPartition,
	}

	if serializedRequestData, err := l.RequestData.serialize(); err == nil {
		newLog.RequestData = serializedRequestData
	} else {
		return fmt.Errorf("method Post: error in serializing request data, %s", err)
	}

	newLog.End = time.Now()
	LogsTable := table.New(LogsMetadata)
	timeSubtracted := newLog.End.Sub(newLog.Start)
	newLog.TimeTaken = int(timeSubtracted.Milliseconds())

	q := scylla.GetScylla().Query(LogsTable.Insert()).BindStruct(&newLog)
	if err := q.ExecRelease(); err != nil {
		return fmt.Errorf("method Post: error in saving log: %s", err)
	}

	fmt.Printf("execution time: %+vs %+vms %+vµs %+vns \n", timeSubtracted.Seconds(), timeSubtracted.Milliseconds(), timeSubtracted.Microseconds(), timeSubtracted.Nanoseconds())
	return nil
}

func (l *LogData) AddExecLog(logUser string, logType string, logData string) {
	execLog := ExecLog{
		LogUser: logUser,
		LogType: logType,
		LogData: logData,
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

func (l *LogData) GetUserErrorLogs() []ExecLog {
	errLogs := []ExecLog{}
	for _, log := range l.ExecutionLogs {
		if log.LogUser == "user" && log.LogType == "error" {
			errLogs = append(errLogs, log)
		}
	}
	return errLogs
}
