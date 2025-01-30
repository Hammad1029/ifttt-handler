package configuration

type InternalTag struct {
	ID   uint   `json:"id" mapstructure:"id"`
	Name string `json:"name" mapstructure:"name"`
}

type InternalTagInMap struct {
	InternalTag string `json:"internalTag" mapstructure:"internalTag"`
}
