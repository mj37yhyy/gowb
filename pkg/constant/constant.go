package constant

type BindingType string

const (
	// http-header请求参数名

	X_REQUEST_ID      = "X-REQUEST-ID"
	X_B3_TRACEID      = "X-B3-TRACEID"
	X_B3_SPANID       = "X-B3-SPANID"
	X_B3_PARENTSPANID = "X-B3-PARENTSPANID"
	X_B3_SAMPLED      = "X-B3-SAMPLED"
	X_B3_FLAGS        = "X-B3-FLAGS"
	X_OT_SPAN_CONTEXT = "X-OT-SPAN-CONTEXT"

	ConfigKey      = "config"
	RoutersKey     = "routers"
	ContextKey     = "context"
	LoggerKey      = "logger"
	AuditLoggerKey = "auditLogger"
	TraceKey       = "trace"
	MiddlewareKey  = "middleware"

	RouterConfigsKey = "router_configs"

	BodyKey           = "body"
	HeaderKey         = "header"
	ParamsKey         = "params"
	RequestKey        = "request"
	BindKey           = "bind"
	BindWithKey       = "bindWith"
	ShouldBindKey     = "shouldBind"
	ShouldBindWithKey = "shouldBindWith"
	TransactionKey    = "tx"
	ResponseKey       = "response"

	BindingUri           BindingType = "uri"
	BindingForm          BindingType = "form"
	BindingFormPost      BindingType = "formPost"
	BindingFormMultipart BindingType = "multipart"
	BindingQuery         BindingType = "query"
	BindingHeader        BindingType = "header"
	BindingJson          BindingType = "json"
	BindingYaml          BindingType = "yaml"
	BindingXml           BindingType = "xml"
	BindingValidator     BindingType = "validator"
	BindingMsgPack       BindingType = "msgPack"
	BindingProtoBuf      BindingType = "protoBuf"

	AuditModuleKey      = "Module"
	AuditOperateKey     = "Operate"
	AuditClusterKey     = "Cluster"
	AuditNamespaceKey   = "Namespace"
	AuditDateKey        = "Date"
	AuditObjectTypeKey  = "ObjectType"
	AuditObjectKey      = "Object"
	AuditClientIPKey    = "ClusterIP"
	AuditAccountTypeKey = "AccountType"
	AuditLogLevelKey    = "LogLevel"
	AuditUserKey        = "User"
	AuditAccountKey     = "Account"
)
