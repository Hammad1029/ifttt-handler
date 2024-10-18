package infrastructure

import (
	domain_audit_log "ifttt/handler/domain/audit_log"
)

type PostgresAuditLogRepository struct {
	*PostgresBaseRepository
}

func NewPostgresAuditLogRepository(base *PostgresBaseRepository) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{PostgresBaseRepository: base}
}

func (p *PostgresAuditLogRepository) InsertLog(dLog *domain_audit_log.AuditLog) error {
	var audit_log audit_log
	if err := audit_log.fromDomain(dLog); err != nil {
		return err
	}
	if err := p.client.Create(&audit_log).Error; err != nil {
		return err
	}
	return nil
}
