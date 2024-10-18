package audit_log

type Repository interface {
	InsertLog(log *AuditLog) error
}
