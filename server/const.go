package server

import "time"

const (
	applicationName = "docker-deploy-server" // Application name.
	version         = "1.0.0"                // Application and server version.
	maxRandom       = 16                     // Max Random chars to use for unique directory names.

	// Config file defaults (yml)
	DefaultConfigPrefix       = applicationName           // Server configuration file name (YML)
	DefaultConfigPath         = "/etc/" + applicationName // Server configuration file location.
	DefaultServerName         = applicationName
	DefaultDomain             = applicationName
	DefaultHostname           = "localhost"
	DefaultPort               = 8080
	DefaultDSN                = "root:root@tcp(mysql:3306)/" + applicationName
	DefaultRedisHost          = "redis"
	DefaultRedisPort          = "6379"
	DefaultRedisKeyQueue      = applicationName + ":queue"
	DefaultRedisKeyLastDeploy = applicationName + ":lastdeploy"
	DefaultRedisPollInt       = 5 // sec.
	DefaultProject            = "docker"
	DefaultTempPath           = "/tmp/" + applicationName
	DefaultImageTag           = "latest"
	DefaultNumCont            = 2

	// http: routes.
	httpRouteV1Health  = "/v1.0/health"
	httpRouteV1Info    = "/v1.0/info"
	httpRouteV1Metrics = "/v1.0/metrics"
	httpRouteV1Deploy  = "/v1.0/deploy"
	httpRouteV1Status  = "/v1.0/status/"

	// Connections.
	TCPReadTimeout  = 10 * time.Second
	TCPWriteTimeout = 10 * time.Second

	// http commands
	httpGet    = "GET"
	httpPost   = "POST"
	httpPut    = "PUT"
	httpDelete = "DELETE"
	httpHead   = "HEAD"
	httpTrace  = "TRACE"
	httpPatch  = "PATCH"

	// Error messages.
	InvalidMediaType         = "Invalid Content-Type or Accept header value."
	InvalidMethod            = "Invalid Method for this route."
	InvalidBody              = "Invalid body of text in request."
	InvalidJSONText          = "Invalid JSON format in text of body in request."
	InvalidJSONAttribute     = "Invalid - 'text' attribute in JSON not found."
	InvalidAuthorization     = "Invalid authorization."
	InvalidEnvAuthorization  = "Invalid authorization for environment."
	InvalidDeployEnv         = "Invalid 'deployEnvironment'."
	InvalidDeployImage       = "Invalid 'image'."
	InvalidDeployCannotQueue = "Cannot queue deploy request at this time."
)
