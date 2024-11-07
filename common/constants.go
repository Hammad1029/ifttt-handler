package common

const (
	DateTimeFormat = "2006-01-02T15:04:05.000"
)

const (
	DependencyRawQueryRepo = iota
	DependencyAppCacheRepo
	DependencyDbDumpRepo
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

const (
	EncodeMD5          = "md5"
	EncodeSHA1         = "sha1"
	EncodeSHA2         = "sha2"
	EncodeBcrypt       = "bcrypt"
	EncodeBase64Decode = "base64-de"
	EncodeBase64Encode = "base64-en"
)

const (
	RequestTokenDefault = "uuid-error"
)

const (
	ResponseCodeSuccess            = "00"
	ResponseDescriptionSuccess     = "SUCCESS"
	ResponseCodeUserError          = "100"
	ResponseDescriptionUserError   = "USER ERROR"
	ResponseCodeSystemError        = "200"
	ResponseDescriptionSystemError = "SYSTEM ERROR"
)
