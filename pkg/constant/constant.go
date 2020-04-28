package constant

type BindingType string
type Operate string

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

	BodyKey           = "body"
	HeaderKey         = "header"
	ParamsKey         = "params"
	RequestKey        = "request"
	BindKey           = "bind"
	BindWithKey       = "bindWith"
	ShouldBindKey     = "shouldBind"
	ShouldBindWithKey = "shouldBindWith"
	TransactionKey    = "tx"

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

	AuditOperateKey    = "Operate"
	AuditClusterKey    = "Cluster"
	AuditNamespaceKey  = "Namespace"
	AuditObjectTypeKey = "ObjectType"
	AuditObjectKey     = "Object"
	AuditUserKey       = "User"
	AuditAccountKey    = "Account"

	Create Operate = "create"
	Delete Operate = "delete"
	Modify Operate = "modify"
	Query  Operate = "query"
)
