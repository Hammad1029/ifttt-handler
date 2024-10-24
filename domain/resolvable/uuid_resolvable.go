package resolvable

import (
	"ifttt/handler/common"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type uuidResolvable struct{}

func (u *uuidResolvable) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return newUUID.String(), nil
}
