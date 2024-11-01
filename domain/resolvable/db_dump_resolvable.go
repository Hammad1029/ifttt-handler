package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
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

	return nil, dumpRepo.InsertDump(dumpMap, d.Table)
}
