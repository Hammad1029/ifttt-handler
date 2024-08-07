package audit_log

import "context"

type Repository interface {
	InsertLog(log PostableAuditLog, ctx context.Context) error
}
