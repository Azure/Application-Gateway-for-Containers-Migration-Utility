package nginx

const (
	// =====================================================================
	// NGINX Ingress Controller Annotations
	// Reference: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/
	// =====================================================================

	// --- Canary Deployment ---
	// Canary annotations allow traffic splitting between services
	AnnotationCanary                = "nginx.ingress.kubernetes.io/canary"                   // IMPLEMENTED: Enables canary behavior
	AnnotationCanaryByHeader        = "nginx.ingress.kubernetes.io/canary-by-header"         // IMPLEMENTED: Route by header presence
	AnnotationCanaryByHeaderValue   = "nginx.ingress.kubernetes.io/canary-by-header-value"   // IMPLEMENTED: Route by header value
	AnnotationCanaryByHeaderPattern = "nginx.ingress.kubernetes.io/canary-by-header-pattern" // IMPLEMENTED: Route by header regex
	AnnotationCanaryByCookie        = "nginx.ingress.kubernetes.io/canary-by-cookie"         // NOT SUPPORTED: Gateway API doesn't support cookie routing
	AnnotationCanaryWeight          = "nginx.ingress.kubernetes.io/canary-weight"            // IMPLEMENTED: Weight set on HTTPBackendRef (user merges canary+main routes)
	AnnotationCanaryWeightTotal     = "nginx.ingress.kubernetes.io/canary-weight-total"      // IMPLEMENTED: Total weight for canary calculations

	// --- Rewrite ---
	AnnotationRewriteTarget = "nginx.ingress.kubernetes.io/rewrite-target" // IMPLEMENTED: URL rewrite target
	AnnotationAppRoot       = "nginx.ingress.kubernetes.io/app-root"       // IMPLEMENTED: Redirect from / to app root
	AnnotationUseRegex      = "nginx.ingress.kubernetes.io/use-regex"      // IMPLEMENTED: Treat paths as regex

	// --- Session Affinity ---
	AnnotationAffinity                             = "nginx.ingress.kubernetes.io/affinity"                                 // IMPLEMENTED: Enables cookie affinity
	AnnotationAffinityMode                         = "nginx.ingress.kubernetes.io/affinity-mode"                            // NOT IMPLEMENTED: balanced/persistent mode
	AnnotationAffinityCanaryBehavior               = "nginx.ingress.kubernetes.io/affinity-canary-behavior"                 // NOT IMPLEMENTED: sticky/legacy for canary
	AnnotationSessionCookieName                    = "nginx.ingress.kubernetes.io/session-cookie-name"                      // PARTIAL: AGC uses managed cookies
	AnnotationSessionCookiePath                    = "nginx.ingress.kubernetes.io/session-cookie-path"                      // PARTIAL: Cookie path setting
	AnnotationSessionCookieDomain                  = "nginx.ingress.kubernetes.io/session-cookie-domain"                    // NOT IMPLEMENTED
	AnnotationSessionCookieExpires                 = "nginx.ingress.kubernetes.io/session-cookie-expires"                   // IMPLEMENTED: Cookie expiration
	AnnotationSessionCookieMaxAge                  = "nginx.ingress.kubernetes.io/session-cookie-max-age"                   // IMPLEMENTED: Cookie max age
	AnnotationSessionCookieSameSite                = "nginx.ingress.kubernetes.io/session-cookie-samesite"                  // PARTIAL: SameSite attribute
	AnnotationSessionCookieSecure                  = "nginx.ingress.kubernetes.io/session-cookie-secure"                    // NOT IMPLEMENTED
	AnnotationSessionCookieConditionalSameSiteNone = "nginx.ingress.kubernetes.io/session-cookie-conditional-samesite-none" // NOT IMPLEMENTED
	AnnotationSessionCookieChangeOnFailure         = "nginx.ingress.kubernetes.io/session-cookie-change-on-failure"         // NOT IMPLEMENTED

	// --- Authentication (Basic/Digest) ---
	AnnotationAuthType       = "nginx.ingress.kubernetes.io/auth-type"        // NOT IMPLEMENTED: basic or digest auth
	AnnotationAuthSecret     = "nginx.ingress.kubernetes.io/auth-secret"      //nolint:gosec // annotation key, not a credential
	AnnotationAuthSecretType = "nginx.ingress.kubernetes.io/auth-secret-type" //nolint:gosec // annotation key, not a credential
	AnnotationAuthRealm      = "nginx.ingress.kubernetes.io/auth-realm"       // NOT IMPLEMENTED: Auth realm string

	// --- Client Certificate Authentication (mTLS) ---
	AnnotationAuthTLSSecret             = "nginx.ingress.kubernetes.io/auth-tls-secret"                       //nolint:gosec // annotation key, not a credential
	AnnotationAuthTLSVerifyClient       = "nginx.ingress.kubernetes.io/auth-tls-verify-client"                // IMPLEMENTED: on/off/optional
	AnnotationAuthTLSVerifyDepth        = "nginx.ingress.kubernetes.io/auth-tls-verify-depth"                 // PARTIAL: Cert chain depth
	AnnotationAuthTLSErrorPage          = "nginx.ingress.kubernetes.io/auth-tls-error-page"                   // NOT IMPLEMENTED
	AnnotationAuthTLSPassCertToUpstream = "nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream" //nolint:gosec // annotation key, not a credential // NOT IMPLEMENTED
	AnnotationAuthTLSMatchCN            = "nginx.ingress.kubernetes.io/auth-tls-match-cn"                     // NOT IMPLEMENTED: CN validation

	// --- External Authentication ---
	AnnotationAuthURL                 = "nginx.ingress.kubernetes.io/auth-url"                   // NOT IMPLEMENTED
	AnnotationAuthMethod              = "nginx.ingress.kubernetes.io/auth-method"                // NOT IMPLEMENTED
	AnnotationAuthSignin              = "nginx.ingress.kubernetes.io/auth-signin"                // NOT IMPLEMENTED
	AnnotationAuthSigninRedirectParam = "nginx.ingress.kubernetes.io/auth-signin-redirect-param" // NOT IMPLEMENTED
	AnnotationAuthResponseHeaders     = "nginx.ingress.kubernetes.io/auth-response-headers"      // NOT IMPLEMENTED
	AnnotationAuthProxySetHeaders     = "nginx.ingress.kubernetes.io/auth-proxy-set-headers"     // NOT IMPLEMENTED
	AnnotationAuthRequestRedirect     = "nginx.ingress.kubernetes.io/auth-request-redirect"      // NOT IMPLEMENTED
	AnnotationAuthCacheKey            = "nginx.ingress.kubernetes.io/auth-cache-key"             // NOT IMPLEMENTED
	AnnotationAuthCacheDuration       = "nginx.ingress.kubernetes.io/auth-cache-duration"        // NOT IMPLEMENTED
	AnnotationAuthKeepalive           = "nginx.ingress.kubernetes.io/auth-keepalive"             // NOT IMPLEMENTED
	AnnotationAuthKeepaliveShareVars  = "nginx.ingress.kubernetes.io/auth-keepalive-share-vars"  // NOT IMPLEMENTED
	AnnotationAuthKeepaliveRequests   = "nginx.ingress.kubernetes.io/auth-keepalive-requests"    // NOT IMPLEMENTED
	AnnotationAuthKeepaliveTimeout    = "nginx.ingress.kubernetes.io/auth-keepalive-timeout"     // NOT IMPLEMENTED
	AnnotationAuthAlwaysSetCookie     = "nginx.ingress.kubernetes.io/auth-always-set-cookie"     // NOT IMPLEMENTED
	AnnotationAuthSnippet             = "nginx.ingress.kubernetes.io/auth-snippet"               // NOT IMPLEMENTED
	AnnotationEnableGlobalAuth        = "nginx.ingress.kubernetes.io/enable-global-auth"         // NOT IMPLEMENTED

	// --- Backend Protocol ---
	AnnotationBackendProtocol = "nginx.ingress.kubernetes.io/backend-protocol" // IMPLEMENTED: HTTP/HTTPS/GRPC/GRPCS

	// --- Backend Certificate Authentication ---
	AnnotationProxySSLSecret      = "nginx.ingress.kubernetes.io/proxy-ssl-secret"       //nolint:gosec // annotation key, not a credential // NOT IMPLEMENTED
	AnnotationProxySSLVerify      = "nginx.ingress.kubernetes.io/proxy-ssl-verify"       // NOT IMPLEMENTED
	AnnotationProxySSLVerifyDepth = "nginx.ingress.kubernetes.io/proxy-ssl-verify-depth" // NOT IMPLEMENTED
	AnnotationProxySSLCiphers     = "nginx.ingress.kubernetes.io/proxy-ssl-ciphers"      // NOT IMPLEMENTED
	AnnotationProxySSLName        = "nginx.ingress.kubernetes.io/proxy-ssl-name"         // NOT IMPLEMENTED
	AnnotationProxySSLProtocols   = "nginx.ingress.kubernetes.io/proxy-ssl-protocols"    // NOT IMPLEMENTED
	AnnotationProxySSLServerName  = "nginx.ingress.kubernetes.io/proxy-ssl-server-name"  // NOT IMPLEMENTED

	// --- Custom Load Balancing ---
	AnnotationLoadBalance              = "nginx.ingress.kubernetes.io/load-balance"                 // PARTIAL: round_robin is default, others not supported
	AnnotationUpstreamHashBy           = "nginx.ingress.kubernetes.io/upstream-hash-by"             // NOT IMPLEMENTED: Consistent hashing key
	AnnotationUpstreamHashBySubset     = "nginx.ingress.kubernetes.io/upstream-hash-by-subset"      // NOT IMPLEMENTED
	AnnotationUpstreamHashBySubsetSize = "nginx.ingress.kubernetes.io/upstream-hash-by-subset-size" // NOT IMPLEMENTED
	AnnotationUpstreamVhost            = "nginx.ingress.kubernetes.io/upstream-vhost"               // IMPLEMENTED: Custom Host header to backend

	// --- Configuration Snippets ---
	AnnotationConfigurationSnippet = "nginx.ingress.kubernetes.io/configuration-snippet" // NOT IMPLEMENTED
	AnnotationServerSnippet        = "nginx.ingress.kubernetes.io/server-snippet"        // NOT IMPLEMENTED
	AnnotationStreamSnippet        = "nginx.ingress.kubernetes.io/stream-snippet"        // NOT IMPLEMENTED

	// --- Custom HTTP Errors ---
	AnnotationCustomHTTPErrors = "nginx.ingress.kubernetes.io/custom-http-errors" // NOT IMPLEMENTED
	AnnotationDefaultBackend   = "nginx.ingress.kubernetes.io/default-backend"    // NOT IMPLEMENTED

	// --- Custom Headers ---
	AnnotationCustomHeaders         = "nginx.ingress.kubernetes.io/custom-headers"          // NOT IMPLEMENTED
	AnnotationConnectionProxyHeader = "nginx.ingress.kubernetes.io/connection-proxy-header" // NOT IMPLEMENTED

	// --- CORS ---
	AnnotationEnableCORS           = "nginx.ingress.kubernetes.io/enable-cors"            // NOT IMPLEMENTED
	AnnotationCORSAllowOrigin      = "nginx.ingress.kubernetes.io/cors-allow-origin"      // NOT IMPLEMENTED
	AnnotationCORSAllowMethods     = "nginx.ingress.kubernetes.io/cors-allow-methods"     // NOT IMPLEMENTED
	AnnotationCORSAllowHeaders     = "nginx.ingress.kubernetes.io/cors-allow-headers"     // NOT IMPLEMENTED
	AnnotationCORSExposeHeaders    = "nginx.ingress.kubernetes.io/cors-expose-headers"    // NOT IMPLEMENTED
	AnnotationCORSAllowCredentials = "nginx.ingress.kubernetes.io/cors-allow-credentials" //nolint:gosec // annotation key, not a credential // NOT IMPLEMENTED
	AnnotationCORSMaxAge           = "nginx.ingress.kubernetes.io/cors-max-age"           // NOT IMPLEMENTED

	// --- SSL/TLS ---
	AnnotationSSLRedirect            = "nginx.ingress.kubernetes.io/ssl-redirect"              // IMPLEMENTED: Redirect HTTP to HTTPS
	AnnotationForceSSLRedirect       = "nginx.ingress.kubernetes.io/force-ssl-redirect"        // IMPLEMENTED: Force HTTPS redirect
	AnnotationSSLPassthrough         = "nginx.ingress.kubernetes.io/ssl-passthrough"           //nolint:gosec // boolean, not a secret // NOT IMPLEMENTED
	AnnotationSSLCiphers             = "nginx.ingress.kubernetes.io/ssl-ciphers"               // NOT IMPLEMENTED
	AnnotationSSLPreferServerCiphers = "nginx.ingress.kubernetes.io/ssl-prefer-server-ciphers" // NOT IMPLEMENTED
	AnnotationPreserveTrailingSlash  = "nginx.ingress.kubernetes.io/preserve-trailing-slash"   // NOT IMPLEMENTED
	AnnotationSSLProtocols           = "nginx.ingress.kubernetes.io/ssl-protocols"             // NOT IMPLEMENTED

	// --- Redirects ---
	AnnotationPermanentRedirect     = "nginx.ingress.kubernetes.io/permanent-redirect"      // IMPLEMENTED: 301 redirect to URL
	AnnotationPermanentRedirectCode = "nginx.ingress.kubernetes.io/permanent-redirect-code" // IMPLEMENTED: Custom redirect code (301, 308)
	AnnotationTemporalRedirect      = "nginx.ingress.kubernetes.io/temporal-redirect"       // IMPLEMENTED: 302 redirect to URL
	AnnotationTemporalRedirectCode  = "nginx.ingress.kubernetes.io/temporal-redirect-code"  // IMPLEMENTED: Custom redirect code (302, 303, 307)
	AnnotationFromToWWWRedirect     = "nginx.ingress.kubernetes.io/from-to-www-redirect"    // IMPLEMENTED: Redirect to www subdomain

	// --- Rate Limiting ---
	AnnotationLimitConnections     = "nginx.ingress.kubernetes.io/limit-connections"      // NOT IMPLEMENTED
	AnnotationLimitRPS             = "nginx.ingress.kubernetes.io/limit-rps"              // NOT IMPLEMENTED
	AnnotationLimitRPM             = "nginx.ingress.kubernetes.io/limit-rpm"              // NOT IMPLEMENTED
	AnnotationLimitBurstMultiplier = "nginx.ingress.kubernetes.io/limit-burst-multiplier" // NOT IMPLEMENTED
	AnnotationLimitRateAfter       = "nginx.ingress.kubernetes.io/limit-rate-after"       // NOT IMPLEMENTED
	AnnotationLimitRate            = "nginx.ingress.kubernetes.io/limit-rate"             // NOT IMPLEMENTED
	AnnotationLimitWhitelist       = "nginx.ingress.kubernetes.io/limit-whitelist"        // NOT IMPLEMENTED

	// --- Access Control ---
	AnnotationDenylistSourceRange  = "nginx.ingress.kubernetes.io/denylist-source-range"  // NOT IMPLEMENTED
	AnnotationWhitelistSourceRange = "nginx.ingress.kubernetes.io/whitelist-source-range" // NOT IMPLEMENTED

	// --- Proxy Settings ---
	AnnotationProxyConnectTimeout      = "nginx.ingress.kubernetes.io/proxy-connect-timeout"       // NOT IMPLEMENTED
	AnnotationProxySendTimeout         = "nginx.ingress.kubernetes.io/proxy-send-timeout"          // NOT IMPLEMENTED
	AnnotationProxyReadTimeout         = "nginx.ingress.kubernetes.io/proxy-read-timeout"          // NOT IMPLEMENTED
	AnnotationProxyNextUpstream        = "nginx.ingress.kubernetes.io/proxy-next-upstream"         // NOT IMPLEMENTED
	AnnotationProxyNextUpstreamTimeout = "nginx.ingress.kubernetes.io/proxy-next-upstream-timeout" // NOT IMPLEMENTED
	AnnotationProxyNextUpstreamTries   = "nginx.ingress.kubernetes.io/proxy-next-upstream-tries"   // NOT IMPLEMENTED
	AnnotationProxyRequestBuffering    = "nginx.ingress.kubernetes.io/proxy-request-buffering"     // NOT IMPLEMENTED
	AnnotationProxyRedirectFrom        = "nginx.ingress.kubernetes.io/proxy-redirect-from"         // NOT IMPLEMENTED
	AnnotationProxyRedirectTo          = "nginx.ingress.kubernetes.io/proxy-redirect-to"           // NOT IMPLEMENTED
	AnnotationProxyHTTPVersion         = "nginx.ingress.kubernetes.io/proxy-http-version"          // NOT IMPLEMENTED

	// --- Proxy Body/Buffer Settings ---
	AnnotationProxyBodySize        = "nginx.ingress.kubernetes.io/proxy-body-size"          // NOT IMPLEMENTED
	AnnotationClientBodyBufferSize = "nginx.ingress.kubernetes.io/client-body-buffer-size"  // NOT IMPLEMENTED
	AnnotationProxyBuffering       = "nginx.ingress.kubernetes.io/proxy-buffering"          // NOT IMPLEMENTED
	AnnotationProxyBuffersNumber   = "nginx.ingress.kubernetes.io/proxy-buffers-number"     // NOT IMPLEMENTED
	AnnotationProxyBufferSize      = "nginx.ingress.kubernetes.io/proxy-buffer-size"        // NOT IMPLEMENTED
	AnnotationProxyBusyBuffersSize = "nginx.ingress.kubernetes.io/proxy-busy-buffers-size"  // NOT IMPLEMENTED
	AnnotationProxyMaxTempFileSize = "nginx.ingress.kubernetes.io/proxy-max-temp-file-size" // NOT IMPLEMENTED

	// --- Proxy Cookie Settings ---
	AnnotationProxyCookieDomain = "nginx.ingress.kubernetes.io/proxy-cookie-domain" // NOT IMPLEMENTED
	AnnotationProxyCookiePath   = "nginx.ingress.kubernetes.io/proxy-cookie-path"   // NOT IMPLEMENTED

	// --- ModSecurity/WAF ---
	AnnotationEnableModSecurity        = "nginx.ingress.kubernetes.io/enable-modsecurity"         // IMPLEMENTED: Enable WAF
	AnnotationEnableOWASPCoreRules     = "nginx.ingress.kubernetes.io/enable-owasp-core-rules"    // IMPLEMENTED: Enable WAF
	AnnotationModSecuritySnippet       = "nginx.ingress.kubernetes.io/modsecurity-snippet"        // NOT IMPLEMENTED
	AnnotationModSecurityTransactionID = "nginx.ingress.kubernetes.io/modsecurity-transaction-id" // IMPLEMENTED: AGC uses trackingId automatically

	// --- Logging ---
	// Enable access logs on AGC : https://learn.microsoft.com/azure/application-gateway/for-containers/diagnostics
	AnnotationEnableAccessLog  = "nginx.ingress.kubernetes.io/enable-access-log"  // IMPLEMENTED: Enable access logging
	AnnotationEnableRewriteLog = "nginx.ingress.kubernetes.io/enable-rewrite-log" // IMPLEMENTED: Enable access logging

	// --- OpenTelemetry ---
	AnnotationEnableOpentelemetry            = "nginx.ingress.kubernetes.io/enable-opentelemetry"              // NOT IMPLEMENTED
	AnnotationOpentelemetryTrustIncomingSpan = "nginx.ingress.kubernetes.io/opentelemetry-trust-incoming-span" // NOT IMPLEMENTED

	// --- Miscellaneous ---
	AnnotationServerAlias      = "nginx.ingress.kubernetes.io/server-alias"       // IMPLEMENTED: Adds aliases to HTTPRoute hostnames
	AnnotationServiceUpstream  = "nginx.ingress.kubernetes.io/service-upstream"   // NOT IMPLEMENTED
	AnnotationHTTP2PushPreload = "nginx.ingress.kubernetes.io/http2-push-preload" // NOT IMPLEMENTED
	AnnotationXForwardedPrefix = "nginx.ingress.kubernetes.io/x-forwarded-prefix" // IMPLEMENTED: Adds X-Forwarded-Prefix header
	AnnotationSatisfy          = "nginx.ingress.kubernetes.io/satisfy"            // NOT IMPLEMENTED

	// --- Mirror ---
	AnnotationMirrorTarget      = "nginx.ingress.kubernetes.io/mirror-target"       // NOT IMPLEMENTED
	AnnotationMirrorRequestBody = "nginx.ingress.kubernetes.io/mirror-request-body" // NOT IMPLEMENTED
	AnnotationMirrorHost        = "nginx.ingress.kubernetes.io/mirror-host"         // NOT IMPLEMENTED
)
