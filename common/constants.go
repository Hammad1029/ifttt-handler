package common

const (
	DateTimeFormat = "2006-01-02T15:04:05.000"
)

const (
	DependencyRawQueryRepo = iota
	DependencyAppCacheRepo
)

var ReservedPaths = []string{"^/test/.*", "^/auth/.*"}

const (
	ContextState IntIota = iota
	ContextLog
	ContextRequestData
	ContextResponseChannel
)

const (
	RedisApis  = "api"
	RedisCrons = "cron"
)

const (
	LogUser   = "user"
	LogSystem = "system"
	LogError  = "error"
	LogInfo   = "info"
)
