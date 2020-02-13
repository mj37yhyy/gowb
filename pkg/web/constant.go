package web

const (
	// http-header请求参数名

	X_REQUEST_ID      = "X-REQUEST-ID"
	X_B3_TRACEID      = "X-B3-TRACEID"
	X_B3_SPANID       = "X-B3-SPANID"
	X_B3_PARENTSPANID = "X-B3-PARENTSPANID"
	X_B3_SAMPLED      = "X-B3-SAMPLED"
	X_B3_FLAGS        = "X-B3-FLAGS"
	X_OT_SPAN_CONTEXT = "X-OT-SPAN-CONTEXT"

	ConfigKey  = "config"
	RoutersKey = "routers"
	ContextKey = "context"
	LoggerKey  = "logger"

	BodyKey   = "body"
	HeaderKey = "header"
	ParamsKey = "params"
)
