package infrastructure

import (
	domain_audit_log "ifttt/handler/domain/audit_log"
)

type PostgresCronAuditLogRepository struct {
	*PostgresBaseRepository
}

func NewPostgresCronAuditLogRepository(base *PostgresBaseRepository) *PostgresCronAuditLogRepository {
	return &PostgresCronAuditLogRepository{PostgresBaseRepository: base}
}

func (p *PostgresCronAuditLogRepository) InsertLog(dLog *domain_audit_log.CronAuditLog) error {
	var audit_log cron_audit_log
	if err := audit_log.fromDomain(dLog); err != nil {
		return err
	}
	if err := p.client.Create(&audit_log).Error; err != nil {
		return err
	}
	return nil
}
