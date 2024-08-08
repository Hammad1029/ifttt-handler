package configuration

type Repository interface {
	GetUserConfigFromDb() (*UserConfiguration, error)
}
