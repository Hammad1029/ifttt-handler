package audit_log

import "context"

type Repository interface {
	InsertLog(log AuditLog, ctx context.Context) error
}
