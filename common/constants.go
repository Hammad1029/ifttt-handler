package common

const (
	DateTimeFormat = "2006-01-02 15:04:05.000"
)

const (
	DependencyRawQueryRepo = iota
	DependencyAppCacheRepo
	DependencyLogger
	DependencyOrmCacheRepo
	DependencyOrmQueryRepo
	DependencyEventProfileCacheRepo
)

var ReservedPaths = []string{"^/test/.*", "^/auth/.*"}

const (
	ContextState IntIota = iota
	ContextLogger
	ContextRequestData
	ContextEventChannel
	ContextTracer
	ContextExternalExecTime
	ContextResponseSent
	ContextLogStage
	ContextIter
)

const (
	RedisApis         = "api"
	RedisCrons        = "cron"
	RedisSchemas      = "schema"
	RedisAssociatons  = "association"
	RedisEventProfile = "event_profile"
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

var EventCodes = map[IntIota]string{
	EventSuccess:           "000",
	EventExhaust:           "010",
	EventBadRequest:        "400",
	EventNotFound:          "404",
	EventSystemMalfunction: "500",
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

var ResponseDefaultMalfunction = map[string]string{
	"responseCode":        EventCodes[EventSystemMalfunction],
	"responseDescription": "Could not map response to profile",
}
