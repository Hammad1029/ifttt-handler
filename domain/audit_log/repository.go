package audit_log

import "ifttt/handler/domain/request_data"

type AuditLog interface {
	Initialize(key string, requestData *request_data.RequestData)
	InitExecOrder(flowId uint)
	AddExecState(exState ExecState, flowId uint)
	AddExecLog(logUser string, logType string, logData any)
	EndLog()
	GetSystemErrorLogs() []string
	GetUserErrorLogs() []string
	SetFinalResponse(res map[string]any)
}

type ApiAuditLogRepository interface {
	InsertLog(log *APIAuditLog) error
}

type CronAuditLogRepository interface {
	InsertLog(log *CronAuditLog) error
}
