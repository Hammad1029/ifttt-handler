package audit_log

type ExecState struct {
	State uint `json:"state" mapstructure:"state"`
	Rule  uint `json:"rule" mapstructure:"rule"`
}

type execLog struct {
	LogUser string `json:"logUser" mapstructure:"logUser"`
	LogType string `json:"logType" mapstructure:"logType"`
	LogData string `json:"logData" mapstructure:"logData"`
}
