package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
	"ifttt/handler/domain/request_data"
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
		common.LogWithTracer(common.LogUser, "error in dumping to db", err, true, ctx)
		return nil, err
	}
	end := time.Now()
	timeTaken := uint64(end.Sub(start).Milliseconds())

	request_data.AddExternalTrip(common.ExternalTripDump, nil, timeTaken, ctx)

	return nil, nil
}
