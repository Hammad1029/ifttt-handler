package configuration

type ResponseProfile struct {
	ID                 uint           `json:"id" mapstructure:"id"`
	Name               string         `json:"name" mapstructure:"name"`
	BodyFormat         map[string]any `json:"bodyFormat" mapstructure:"bodyFormat"`
	ResponseHTTPStatus int            `json:"responseHTTPStatus" mapstructure:"responseHTTPStatus"`
}

type ResponseProfilePersistentRepository interface {
	GetAllProfiles() (*[]ResponseProfile, error)
}
