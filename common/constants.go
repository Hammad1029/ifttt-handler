package common

const (
	TimeFormat = "2006-01-02 15:04:05"
)

const (
	DependencyRawQueryRepo = "rawQueryRepo"
)

const (
	RestMethodGet    = "GET"
	RestMethodPost   = "POST"
	RestMethodPut    = "PUT"
	RestMethodDelete = "DELETE"
)

var ReservedPaths = []string{"^/test/.*", "^/auth/.*"}
