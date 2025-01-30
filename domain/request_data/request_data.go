package request_data

import (
	"sync"
)

type RequestData struct {
	Mtx                sync.Mutex
	Errors             []error           `json:"errors" mapstructure:"errors"`
	Headers            map[string]string `json:"headers" mapstructure:"headers"`
	AggregatedResponse map[string]any    `json:"aggregatedResponse" mapstructure:"aggregatedResponse"`
	Store              map[string]any    `json:"store" mapstructure:"store"`
	ExternalTrips      []ExternalTrip    `json:"externalTrips" mapstructure:"externalTrips"`
}

type ExternalTrip struct {
	Key        string          `json:"key" mapstructure:"key"`
	Identifier string          `json:"identifier" mapstructure:"identifier"`
	TimeTaken  uint64          `json:"timeTaken" mapstructure:"timeTaken"`
	Data       *map[string]any `json:"data" mapstructure:"data"`
}
