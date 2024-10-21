package infrastructure

import (
	domain_audit_log "ifttt/handler/domain/audit_log"
)

type PostgresAPIAuditLogRepository struct {
	*PostgresBaseRepository
}

func NewPostgresAPIAuditLogRepository(base *PostgresBaseRepository) *PostgresAPIAuditLogRepository {
	return &PostgresAPIAuditLogRepository{PostgresBaseRepository: base}
}

func (p *PostgresAPIAuditLogRepository) InsertLog(dLog *domain_audit_log.APIAuditLog) error {
	var audit_log api_audit_log
	if err := audit_log.fromDomain(dLog); err != nil {
		return err
	}
	if err := p.client.Create(&audit_log).Error; err != nil {
		return err
	}
	return nil
}
