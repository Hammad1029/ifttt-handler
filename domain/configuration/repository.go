package configuration

type Repository interface {
	GetConfigFromDb() (*Configuration, error)
}
