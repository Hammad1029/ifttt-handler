package models

import (
	"fmt"
	"handler/config"
	"handler/scylla"
	"handler/utils"
	"strconv"
	"time"

	"github.com/scylladb/gocqlx/v2/table"
)

type LogModel struct {
	ApiGroup       string       `cql:"api_group"`
	ApiName        string       `cql:"api_name"`
	ExecutionOrder []string     `cql:"execution_order"`
	ExecutionLogs  []ExecLog    `cql:"execution_logs"`
	RequestData    *RequestData `cql:"request_data"`
	Start          time.Time    `cql:"start"`
	StartPartition time.Time    `cql:"start_partition"`
	End            time.Time    `cql:"end"`
	TimeTaken      int          `cql:"time_taken"`
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

func (l *LogModel) StartLog() {
	now := time.Now()
	fmt.Printf("Request recieved at %s\n", now.String())
	l.Start = now
}

func (l *LogModel) Initialize(r *RequestData, api *ApiModel) {
	l.ApiGroup = api.ApiGroup
	l.ApiName = api.ApiName
	l.ExecutionOrder = []string{}
	l.RequestData = r

	timeSlot, err := strconv.Atoi(config.GetConfigProp("app.logPartitionSeconds"))
	if err != nil {
		l.AddExecLog("system", "error", err.Error())
	}
	l.StartPartition = utils.GetTimeSlot(l.Start, timeSlot)
}

func (l *LogModel) Post() {
	reqBodySerialized, err := utils.SerializeMap(l.RequestData.ReqBody)
	if err != nil {
		l.AddExecLog("system", "error", "could not serialize request body")
	}
	storeSerialized, err := utils.SerializeMap(l.RequestData.Store)
	if err != nil {
		l.AddExecLog("system", "error", "could not serialize store")
	}
	responseSerialized, err := utils.SerializeMap(l.RequestData.Response)
	if err != nil {
		l.AddExecLog("system", "error", "could not serialize response")
	}
	queryResSerialized, err := utils.SerializeMap(l.RequestData.QueryRes)
	if err != nil {
		l.AddExecLog("system", "error", "could not serialize query results")
	}

	l.RequestData.ReqBody = reqBodySerialized
	l.RequestData.Store = storeSerialized
	l.RequestData.Response = responseSerialized
	l.RequestData.QueryRes = queryResSerialized

	l.End = time.Now()
	LogsTable := table.New(LogsMetadata)
	timeSubtracted := l.End.Sub(l.Start)
	l.TimeTaken = int(timeSubtracted.Milliseconds())

	q := scylla.GetScylla().Query(LogsTable.Insert()).BindStruct(&l)
	if err := q.ExecRelease(); err != nil {
		fmt.Printf("error in saving log: %s", err)
	}

	fmt.Printf("execution time: %+vs %+vms %+vµs %+vns \n", timeSubtracted.Seconds(), timeSubtracted.Milliseconds(), timeSubtracted.Microseconds(), timeSubtracted.Nanoseconds())
}

func (l *LogModel) AddExecLog(logUser string, logType string, logData string) {
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

func (l *LogModel) GetUserErrorLogs() []ExecLog {
	errLogs := []ExecLog{}
	for _, log := range l.ExecutionLogs {
		if log.LogUser == "user" && log.LogType == "error" {
			errLogs = append(errLogs, log)
		}
	}
	return errLogs
}
