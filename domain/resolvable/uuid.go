package resolvable

import (
	"ifttt/handler/common"

	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type generateUUID struct{}

func (u *generateUUID) Resolve(ctx context.Context, dependencies map[common.IntIota]any) (any, error) {
	newUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return newUUID.String(), nil
}
