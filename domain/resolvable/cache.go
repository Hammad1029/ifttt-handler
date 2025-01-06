package resolvable

import (
	"context"
	"fmt"
	"ifttt/handler/common"
)

type setCache struct {
	Key   Resolvable `json:"key" mapstructure:"key"`
	Value Resolvable `json:"value" mapstructure:"value"`
	TTL   uint       `json:"ttl" mapstructure:"ttl"`
}

type getCache struct {
	Key Resolvable `json:"key" mapstructure:"key"`
}

type AppCacheRepository interface {
	SetKey(key string, val any, ttl uint, ctx context.Context) error
	GetKey(key string, ctx context.Context) (any, error)
}

func (s *setCache) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	keyResolved, err := s.Key.Resolve(ctx, dependencies)
	if err != nil {
		return nil, err
	}

	valResolved, err := s.Value.Resolve(ctx, dependencies)
	if err != nil {
		return nil, err
	}

	appCacheRepo := dependencies[common.DependencyAppCacheRepo].(AppCacheRepository)
	return nil, appCacheRepo.SetKey(fmt.Sprint(keyResolved), valResolved, s.TTL, ctx)
}

func (g *getCache) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	keyResolved, err := g.Key.Resolve(ctx, dependencies)
	if err != nil {
		return nil, err
	}

	appCacheRepo := dependencies[common.DependencyAppCacheRepo].(AppCacheRepository)
	return appCacheRepo.GetKey(fmt.Sprint(keyResolved), ctx)
}
