package interceptors

//DefInterceptor default interceptor Id
type DefInterceptor int

const (
	DefRecovery         DefInterceptor                                        = 1 << iota //nolint
	DefLogLevelOverride                                                                   //nolint
	DefLogServerAPI                                                                       //nolint
	DefAll              = DefLogLevelOverride | DefRecovery | DefLogServerAPI             //nolint
)
