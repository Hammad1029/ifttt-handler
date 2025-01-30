package common

import "regexp"

const (
	DateTimeFormat        = "2006-01-02 15:04:05"
	DateTimeFormatGeneric = "YYYY-MM-DD HH:mm:ss"
)

const (
	DependencyRawQueryRepo = iota
	DependencyAppCacheRepo
	DependencyLogger
	DependencyOrmCacheRepo
	DependencyOrmQueryRepo
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
	ContextIter
	ContextResponseProfiles
)

const (
	RedisApis            = "api"
	RedisCrons           = "cron"
	RedisSchemas         = "schema"
	RedisAssociatons     = "association"
	RedisResponseProfile = "response_profile"
	RedisInternalTags    = "internal_tags"
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
	LogStageExecution  = "execution"
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
	EventSuccess IntIota = iota
	EventExhaust
	EventSystemMalfunction
	EventNotFound
	EventBadRequest
)

var EventCodes = map[IntIota]uint{
	EventSuccess:           0,
	EventExhaust:           10,
	EventBadRequest:        400,
	EventNotFound:          404,
	EventSystemMalfunction: 500,
}

const (
	ResponseHeaderTracer      = "tracer"
	ResponseHeaderContentType = "Content-Type"
)

const (
	DataTypeText    = "text"
	DataTypeNumber  = "number"
	DataTypeBoolean = "boolean"
	DataTypeArray   = "array"
	DataTypeMap     = "map"
)

const (
	ExternalTripQuery = "query"
	ExternalTripApi   = "api"
)

const (
	CastToString  = "string"
	CastToNumber  = "number"
	CastToBoolean = "boolean"
)

const (
	EnvConfig      = "configStore"
	EnvData        = "dataStore"
	EnvCache       = "cacheStore"
	EnvAppCache    = "appCache"
	EnvConfigStore = "cacheStore"
	EnvDBName      = "db"
)

const (
	OrmSelect = "select"
	OrmUpdate = "update"
	OrmInsert = "insert"
	OrmDelete = "delete"
)

const (
	DatabaseTypeString  = "string"
	DatabaseTypeNumber  = "number"
	DatabaseTypeBoolean = "boolean"
)

const (
	AssociationsHasOne        = "hasOne"
	AssociationsHasMany       = "hasMany"
	AssociationsBelongsTo     = "belongsTo"
	AssociationsBelongsToMany = "belongsToMany"
)

const (
	DateOperatorAdd      = "+"
	DateOperatorSubtract = "-"
)

var ResponseDefaultMalfunction = map[string]any{
	"responseCode":        EventCodes[EventSystemMalfunction],
	"responseDescription": "Could not map response to profile",
}

var SQLRegex = map[string]*regexp.Regexp{
	"SELECT":   regexp.MustCompile(`FROM\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	"INSERT":   regexp.MustCompile(`INTO\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	"UPDATE":   regexp.MustCompile(`UPDATE\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	"DELETE":   regexp.MustCompile(`FROM\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	"TRUNCATE": regexp.MustCompile(`TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	"ALTER":    regexp.MustCompile(`TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	"DROP":     regexp.MustCompile(`TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	"CREATE":   regexp.MustCompile(`TABLE\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
}

const (
	InternalTagErrorValidation = "validation"
	InternalTagErrorSystem     = "system"
	InternalTagErrorUser       = "user"
)
