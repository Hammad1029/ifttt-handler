package models

import (
	"handler/config"
	"handler/scylla"
	"handler/utils"
	"log"
	"strconv"
	"time"

	"github.com/scylladb/gocqlx/v2/table"
)

type LogModel struct {
	ApiGroup       string       `cql:"api_group"`
	ApiName        string       `cql:"api_name"`
	ExecutionOrder []int        `cql:"execution_order"`
	ExecutionLogs  []ExecLog    `cql:"execution_logs"`
	RequestData    *RequestData `cql:"request_data"`
	Start          time.Time    `cql:"start"`
	StartPartition time.Time    `cql:"start_partition"`
	End            time.Time    `cql:"end"`
	TimeTaken      int          `cql:"time_taken"`
}

var LogsMetadata = table.Metadata{
	Name:    "Logs",
	Columns: []string{"api_group", "api_name", "execution_order", "request_data", "start", "start_partition", "end", "time_taken"},
	PartKey: []string{"api_group", "start_partition"},
	SortKey: []string{"api_name", "start"},
}

type ExecLog struct {
	LogUser string `cql:"log_user"`
	LogType string `cql:"log_type"`
	LogData string `cql:"log_data"`
}

func (l *LogModel) Initialize(r *RequestData, api *ApiModel) {
	now := time.Now()
	log.Printf("Request recieved at %s\n", now.String())

	l.ApiGroup = api.ApiGroup
	l.ApiName = api.ApiName
	l.ExecutionOrder = []int{}
	l.RequestData = r
	l.Start = now

	timeSlot, err := strconv.Atoi(config.GetConfigProp("app.logPartitionSeconds"))
	if err != nil {
		l.AddExecLog("system", "error", err.Error())
	}
	l.StartPartition = utils.GetTimeSlot(now, timeSlot)
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

	l.RequestData.ReqBody = reqBodySerialized
	l.RequestData.Store = storeSerialized
	l.RequestData.Response = responseSerialized

	l.End = time.Now()
	LogsTable := table.New(LogsMetadata)
	l.TimeTaken = int(l.End.Sub(l.Start).Milliseconds())

	q := scylla.GetScylla().Query(LogsTable.Insert()).BindStruct(&l)
	if err := q.ExecRelease(); err != nil {
		log.Printf("error in saving log: %s", err.Error())
	}

	log.Printf("execution time: %+v\n", l.TimeTaken)
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
