package infrastructure

import (
	"context"
	"ifttt/handler/domain/audit_log"
	"time"

	"github.com/scylladb/gocqlx/v3/table"
)

type scyllaAuditLog struct {
	ApiGroup       string                       `cql:"api_group"`
	ApiName        string                       `cql:"api_name"`
	ExecutionOrder []string                     `cql:"execution_order"`
	ExecutionLogs  []scyllaExecAuditLog         `cql:"execution_logs"`
	RequestData    scyllaSerizalizedRequestData `cql:"request_data"`
	Start          time.Time                    `cql:"start"`
	StartPartition time.Time                    `cql:"start_partition"`
	End            time.Time                    `cql:"end"`
	TimeTaken      int                          `cql:"time_taken"`
}

type scyllaExecAuditLog struct {
	LogUser string `cql:"log_user"`
	LogType string `cql:"log_type"`
	LogData string `cql:"log_data"`
}

type scyllaSerizalizedRequestData struct {
	ReqBody  string `cql:"req_body"`
	Store    string `cql:"store"`
	Response string `cql:"response"`
	QueryRes string `cql:"query_res"`
	ApiRes   string `cql:"api_res"`
}

var scyllaAuditLogMetadata = table.Metadata{
	Name:    "Logs",
	Columns: []string{"api_group", "api_name", "execution_order", "execution_logs", "request_data", "start", "start_partition", "end", "time_taken"},
	PartKey: []string{"api_group", "start_partition"},
	SortKey: []string{"api_name", "start"},
}

var scyllaAuditLogTable *table.Table

type ScyllaAuditLogRepository struct {
	ScyllaBaseRepository
}

func NewScyllaAuditLogRepository(base ScyllaBaseRepository) *ScyllaAuditLogRepository {
	return &ScyllaAuditLogRepository{ScyllaBaseRepository: base}
}

func (s *ScyllaAuditLogRepository) getTable() *table.Table {
	if scyllaAuditLogTable == nil {
		scyllaAuditLogTable = table.New(scyllaAuditLogMetadata)
	}
	return scyllaAuditLogTable
}

func (s *ScyllaAuditLogRepository) InsertLog(log audit_log.AuditLog, ctx context.Context) error {
	// newLog := scyllaAuditLog{
	// 	ApiGroup:       log.ApiGroup,
	// 	ApiName:        log.ApiName,
	// 	ExecutionOrder: log.ExecutionOrder,
	// 	Start:          log.Start,
	// 	StartPartition: log.StartPartition,
	// }

	// var scyllaExecAuditLogs []scyllaExecAuditLog
	// for _, eL := range log.ExecutionLogs {
	// 	scyllaExecAuditLogs = append(scyllaExecAuditLogs, scyllaExecAuditLog{
	// 		LogUser: eL.LogUser,
	// 		LogType: eL.LogType,
	// 		LogData: eL.LogData,
	// 	})
	// }
	// newLog.ExecutionLogs = scyllaExecAuditLogs
	// // newLog.RequestData = scyllaSerizalizedRequestData{
	// // 	ReqBody:  log.RequestData.ReqBody,
	// // 	Store:    log.RequestData.Store,
	// // 	Response: log.RequestData.Response,
	// // 	QueryRes: log.RequestData.QueryRes,
	// // 	ApiRes:   log.RequestData.ApiRes,
	// // }

	// LogsTable := s.getTable()
	// query := LogsTable.InsertQuery(*s.session).BindStruct(&newLog)
	// if err := query.ExecRelease(); err != nil {
	// 	return fmt.Errorf("method ScyllaLogRepository.InsertLog: error in inserting log: %s", err)
	// }

	return nil
}
