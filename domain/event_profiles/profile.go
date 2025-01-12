package eventprofiles

type Profile struct {
	ID                 uint           `json:"id" mapstructure:"id"`
	Trigger            string         `json:"trigger" mapstructure:"trigger"`
	UseBody            bool           `json:"useBody" mapstructure:"useBody"`
	Internal           bool           `json:"internal" mapstructure:"internal"`
	ResponseBody       map[string]any `json:"responseBody" mapstructure:"responseBody"`
	ResponseHTTPStatus int            `json:"responseHTTPStatus" mapstructure:"responseHTTPStatus"`
	MappedProfiles     *[]Profile     `json:"mappedProfiles" mapstructure:"mappedProfiles"`
}
