package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/audit_log"
	"time"
)

type dbDumpResolvable struct {
	Columns map[string]Resolvable `json:"columns" mapstructure:"columns"`
	Table   string                `json:"table" mapstructure:"table"`
}

type DbDumpRepository interface {
	InsertDump(dump map[string]any, table string) error
}

func (d *dbDumpResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	dumpResolved, err := resolveIfNested(d.Columns, ctx, dependencies)
	if err != nil {
		return nil, err
	}

	dumpMap, ok := dumpResolved.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("could not cast dump map")
	}

	dumpRepo, ok := dependencies[common.DependencyDbDumpRepo].(DbDumpRepository)
	if !ok {
		return nil, fmt.Errorf("could not cast dump repo")
	}

	start := time.Now()
	err = dumpRepo.InsertDump(dumpMap, d.Table)
	if err != nil {
		audit_log.AddExecLog(common.LogUser, common.LogError, err, ctx)
	}
	end := time.Now()
	timeTaken := uint64(end.Sub(start).Milliseconds())

	if log := audit_log.GetAuditLogFromContext(ctx); log != nil {
		(*log).AddExternalTime(timeTaken)
	}

	return nil, nil
}
