package resolvable

import (
	"context"
	"handler/config"
	"handler/store"
	"time"
)

type UserConfigurationResolvable struct {
	IsActive          bool      `json:"isActive" mapstructure:"isActive"`
	ConfigurationJSON string    `json:"configurationJSON" mapstructure:"configurationJSON"`
	CreatedAt         time.Time `json:"createdAt" mapstructure:"createdAt"`
}

func (u *UserConfigurationResolvable) Resolve(ctx context.Context) (interface{}, error) {
	return config.GetUserConfig().AllSettings(), nil
}

func (u *UserConfigurationResolvable) ReadUserConfig() error {
	(*store.GetConfigStore()).GetUserConfiguration(u)
	return config.SetUserConfig(u.ConfigurationJSON)
}
