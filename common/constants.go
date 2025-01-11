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
	DependencyResponseProfileCacheRepo
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
)

const (
	RedisApis            = "api"
	RedisCrons           = "cron"
	RedisSchemas         = "schema"
	RedisAssociatons     = "association"
	RedisResponseProfile = "response_profile"
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
	ResponseCodeSuccess IntIota = iota
	ResponseCodeExhaust
	ResponseCodeSystemMalfunction
	ResponseCodeBadRequest
)

var ResponseCodes = map[IntIota]string{
	ResponseCodeSuccess:           "000",
	ResponseCodeExhaust:           "010",
	ResponseCodeBadRequest:        "400",
	ResponseCodeSystemMalfunction: "500",
}

const (
	InternalResponseCode        = "responseCode"
	InternalResponseDescription = "responseDescription"
	InternalResponseData        = "data"
	InternalResponseErrors      = "errors"
)

const (
	ResponseHeaderTracer = "tracer"
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
