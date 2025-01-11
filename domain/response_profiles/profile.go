package responseprofiles

type Profile struct {
	MappedCode    string   `json:"mappedCode" mapstructure:"mappedCode"`
	HttpStatus    int      `json:"httpStatus" mapstructure:"httpStatus"`
	Code          Field    `json:"code" mapstructure:"code"`
	Description   Field    `json:"description" mapstructure:"description"`
	Data          Field    `json:"data" mapstructure:"data"`
	Errors        Field    `json:"errors" mapstructure:"errors"`
	MappedProfile *Profile `json:"mappedProfile" mapstructure:"mappedProfile"`
}

type Field struct {
	Key      string `json:"key" mapstructure:"key"`
	Default  any    `json:"default" mapstructure:"default"`
	Disabled bool   `json:"disabled" mapstructure:"disabled"`
}
