package infrastructure

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"strings"
)

type MySqlRawQueryRepository struct {
	*MySqlBaseRepository
}

func NewMySqlRawQueryRepository(base *MySqlBaseRepository) *MySqlRawQueryRepository {
	return &MySqlRawQueryRepository{MySqlBaseRepository: base}
}

func (p *MySqlRawQueryRepository) Positional(queryString string, parameters []any, ctx context.Context) (*[]map[string]any, error) {
	var results []map[string]any
	lowerQuery := strings.TrimSpace(strings.ToLower(queryString))

	tx := p.client.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if strings.HasPrefix(lowerQuery, common.OrmSelect) {
		if err := p.client.WithContext(ctx).Raw(queryString, parameters...).Scan(&results).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if strings.HasPrefix(lowerQuery, common.OrmInsert) {
		if err := tx.Exec(queryString, parameters...).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		tableName := p.getTableName(queryString)
		if tableName == "" {
			tx.Rollback()
			return nil, fmt.Errorf("table not found in query")
		}
		returnQuery := fmt.Sprintf("SELECT * FROM %s WHERE id = LAST_INSERT_ID()", tableName)
		if err := tx.Raw(returnQuery).Scan(&results).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if strings.HasPrefix(lowerQuery, common.OrmUpdate) {
		if err := tx.Exec(queryString, parameters...).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		// tableName := p.getTableName(queryString)
		// if tableName == "" {
		// 	tx.Rollback()
		// 	return nil, fmt.Errorf("table not found in query")
		// }
		// whereClause := p.getWhereClause(queryString)
		// returnQuery := fmt.Sprintf("SELECT * FROM %s %s", tableName, whereClause)
		// if err := tx.Raw(returnQuery, parameters...).Scan(&results).Error; err != nil {
		// 	tx.Rollback()
		// 	return nil, err
		// }
	} else if strings.HasPrefix(lowerQuery, common.OrmDelete) {
		tableName := p.getTableName(queryString)
		if tableName == "" {
			tx.Rollback()
			return nil, fmt.Errorf("table not found in query")
		}
		whereClause := p.getWhereClause(queryString)
		returnQuery := fmt.Sprintf("SELECT * FROM %s %s", tableName, whereClause)
		if err := tx.Raw(returnQuery, parameters...).Scan(&results).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Exec(queryString, parameters...).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return &results, nil
}

func (p *MySqlRawQueryRepository) Named(queryString string, parameters map[string]any, ctx context.Context) (*[]map[string]any, error) {
	var results []map[string]any
	lowerQuery := strings.TrimSpace(strings.ToLower(queryString))

	tx := p.client.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if strings.HasPrefix(lowerQuery, common.OrmSelect) {
		if err := p.client.WithContext(ctx).Raw(queryString, parameters).Scan(&results).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if strings.HasPrefix(lowerQuery, common.OrmInsert) {
		if err := tx.Exec(queryString, parameters).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		tableName := p.getTableName(queryString)
		if tableName == "" {
			tx.Rollback()
			return nil, fmt.Errorf("table not found in query")
		}
		returnQuery := fmt.Sprintf("SELECT * FROM %s WHERE id = LAST_INSERT_ID() LIMIT 1", tableName)
		if err := tx.Raw(returnQuery).Scan(&results).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if strings.HasPrefix(lowerQuery, common.OrmUpdate) {
		if err := tx.Exec(queryString, parameters).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		// tableName := p.getTableName(queryString)
		// if tableName == "" {
		// 	tx.Rollback()
		// 	return nil, fmt.Errorf("table not found in query")
		// }
		// whereClause := p.getWhereClause(queryString)
		// returnQuery := fmt.Sprintf("SELECT * FROM %s %s", tableName, whereClause)
		// if err := tx.Raw(returnQuery, parameters).Scan(&results).Error; err != nil {
		// 	tx.Rollback()
		// 	return nil, err
		// }
	} else if strings.HasPrefix(lowerQuery, common.OrmDelete) {
		tableName := p.getTableName(queryString)
		if tableName == "" {
			tx.Rollback()
			return nil, fmt.Errorf("table not found in query")
		}
		whereClause := p.getWhereClause(queryString)
		returnQuery := fmt.Sprintf("SELECT * FROM %s %s", tableName, whereClause)
		if err := tx.Raw(returnQuery, parameters).Scan(&results).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Exec(queryString, parameters).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return &results, nil
}

func (p *MySqlRawQueryRepository) getTableName(query string) string {
	parts := strings.Fields(query)
	for i, part := range parts {
		if strings.ToUpper(part) == "INTO" ||
			strings.ToUpper(part) == "UPDATE" || strings.ToUpper(part) == "FROM" {
			if i+1 < len(parts) {
				return strings.TrimSpace(parts[i+1])
			}
		}
	}
	return ""
}

func (p *MySqlRawQueryRepository) getWhereClause(query string) string {
	parts := strings.SplitN(strings.ToUpper(query), "WHERE", 2)
	if len(parts) > 1 {
		return "WHERE " + parts[1]
	}
	return ""
}
