package configuration

import (
	"context"
	"handler/config"
	"time"
)

type Configuration struct {
	IsActive          bool      `json:"isActive" mapstructure:"isActive"`
	ConfigurationJSON string    `json:"configurationJSON" mapstructure:"configurationJSON"`
	CreatedAt         time.Time `json:"createdAt" mapstructure:"createdAt"`
}

func (u *Configuration) Resolve(ctx context.Context) (any, error) {
	return config.GetUserConfig().AllSettings(), nil
}
