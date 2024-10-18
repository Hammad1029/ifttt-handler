package common

const (
	DateTimeFormat = "2006-01-02T15:04:05.000"
)

const (
	DependencyRawQueryRepo = "rawQueryRepo"
)

var ReservedPaths = []string{"^/test/.*", "^/auth/.*"}

const (
	ContextState contextStateKey = iota
	ContextLog
	ContextRequestData
	ContextResponseChannel
	ContextApiData
)
