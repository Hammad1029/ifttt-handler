package common

const (
	DateTimeFormat = "2006-01-02T15:04:05.000"
)

const (
	DependencyRawQueryRepo = iota
	DependencyAppCacheRepo
	DependencyDbDumpRepo
	DependencyLogger
)

var ReservedPaths = []string{"^/test/.*", "^/auth/.*"}

const (
	ContextState IntIota = iota
	ContextLogger
	ContextRequestData
	ContextResponseChannel
	ContextTracer
	ContextExternalExecTime
	ContextResponseSent
	ContextLogStage
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

const (
	LogStageInitation  = "initiation"
	LogStageMemload    = "memload"
	LogStageParsing    = "parsing"
	LogStageValidation = "validation"
	LogStagePreConfig  = "preconfig"
	LogStagePreWare    = "preware"
	LogStageMainWare   = "mainware"
	LogStagePostWare   = "postware"
	LogStageEnding     = "end"
)

const (
	EncodeMD5          = "md5"
	EncodeSHA1         = "sha1"
	EncodeSHA2         = "sha2"
	EncodeBcrypt       = "bcrypt"
	EncodeBase64Decode = "base64-de"
	EncodeBase64Encode = "base64-en"
)

const (
	ResponseCodeSuccess            = "00"
	ResponseDescriptionSuccess     = "SUCCESS"
	ResponseCodeUserError          = "100"
	ResponseDescriptionUserError   = "USER ERROR"
	ResponseCodeSystemError        = "200"
	ResponseDescriptionSystemError = "SYSTEM ERROR"
)

const (
	DataTypeText    = "text"
	DataTypeNumber  = "number"
	DataTypeBoolean = "boolean"
	DataTypeArray   = "array"
	DataTypeMap     = "map"
)

const (
	ExternalTripDump  = "dump"
	ExternalTripQuery = "query"
	ExternalTripApi   = "api"
)
