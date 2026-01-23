// Package nginx provides annotation handlers for migrating NGINX Ingress Controller
// ingress resources to Gateway API resources for Azure Application Gateway for Containers.
package nginx

import (
	"log/slog"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/logging"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// Provider implements the Provider interface for NGINX Ingress Controller annotations.
type Provider struct {
	log   *slog.Logger
	input resources.AGICResources
}

// NewProvider creates a new NGINX provider.
func NewProvider(input resources.AGICResources) Provider {
	return Provider{
		log:   logging.New("nginx-converter"),
		input: input,
	}
}

// GetAnnotationHandlers returns the annotation handlers for NGINX.
func (p Provider) GetAnnotationHandlers() map[string]conversion.AnnotationHandler {
	return map[string]conversion.AnnotationHandler{
		// Client certificate authentication (mTLS)
		AnnotationAuthTLSSecret:             p.handleAuthTLSSecret,
		AnnotationAuthTLSVerifyClient:       p.handleAuthTLSVerifyClient,
		AnnotationAuthTLSVerifyDepth:        p.handleAuthTLSVerifyDepth,
		AnnotationAuthTLSErrorPage:          p.handleAuthTLSErrorPage,
		AnnotationAuthTLSPassCertToUpstream: p.handleAuthTLSPassCertToUpstream,

		// Session affinity
		AnnotationAffinity:               p.handleAffinity,
		AnnotationAffinityMode:           p.handleAffinityMode,
		AnnotationAffinityCanaryBehavior: p.handleAffinityCanaryBehavior,
		AnnotationSessionCookieName:      p.handleSessionCookieName,
		AnnotationSessionCookiePath:      p.handleSessionCookiePath,
		AnnotationSessionCookieExpires:   p.handleSessionCookieExpires,
		AnnotationSessionCookieMaxAge:    p.handleSessionCookieMaxAge,
		AnnotationSessionCookieSameSite:  p.handleSessionCookieSameSite,

		// URL rewriting
		AnnotationRewriteTarget:    p.handleRewriteTarget,
		AnnotationUseRegex:         p.handleUseRegex,
		AnnotationAppRoot:          p.handleAppRoot,
		AnnotationXForwardedPrefix: p.handleXForwardedPrefix,

		// SSL/TLS redirect
		AnnotationSSLRedirect:      p.handleSSLRedirect,
		AnnotationForceSSLRedirect: p.handleForceSSLRedirect,

		// Custom headers and configuration
		AnnotationCustomHeaders:         p.handleConfigurationSnippet,
		AnnotationConnectionProxyHeader: p.handleConnectionProxyHeader,
		AnnotationConfigurationSnippet:  p.handleConfigurationSnippet,
		AnnotationServerSnippet:         p.handleServerSnippet,

		// Backend protocol
		AnnotationBackendProtocol: p.handleBackendProtocol,

		// Default backend
		AnnotationDefaultBackend: p.handleDefaultBackend,

		// ModSecurity/WAF
		AnnotationEnableModSecurity:        p.handleEnableModSecurity,
		AnnotationEnableOWASPCoreRules:     p.handleEnableOWASPCoreRules,
		AnnotationModSecuritySnippet:       p.handleModSecuritySnippet,
		AnnotationModSecurityTransactionID: p.handleModSecurityTransactionID,

		// TLS Policy
		AnnotationSSLCiphers:             p.handleSSLCiphers,
		AnnotationSSLPreferServerCiphers: p.handleSSLPreferServerCiphers,

		// Proxy settings
		AnnotationProxyConnectTimeout: p.handleProxyConnectTimeout,
		AnnotationProxySendTimeout:    p.handleProxySendTimeout,
		AnnotationProxyReadTimeout:    p.handleProxyReadTimeout,
		AnnotationProxyBodySize:       p.handleProxyBodySize,
		AnnotationProxyBuffering:      p.handleProxyBuffering,

		// Load balancing
		AnnotationLoadBalance:   p.handleLoadBalance,
		AnnotationUpstreamVhost: p.handleUpstreamVhost,

		// Redirects
		AnnotationPermanentRedirect:     p.handlePermanentRedirect,
		AnnotationPermanentRedirectCode: p.handlePermanentRedirectCode,
		AnnotationTemporalRedirect:      p.handleTemporalRedirect,
		AnnotationTemporalRedirectCode:  p.handleTemporalRedirectCode,
		AnnotationFromToWWWRedirect:     p.handleFromToWWWRedirect,

		// Server aliases
		AnnotationServerAlias: p.handleServerAlias,

		// Canary deployments
		AnnotationCanary:                p.handleCanary,
		AnnotationCanaryWeight:          p.handleCanaryWeight,
		AnnotationCanaryWeightTotal:     p.handleCanaryWeightTotal,
		AnnotationCanaryByHeader:        p.handleCanaryByHeader,
		AnnotationCanaryByHeaderValue:   p.handleCanaryByHeaderValue,
		AnnotationCanaryByHeaderPattern: p.handleCanaryByHeaderPattern,
		AnnotationCanaryByCookie:        p.handleCanaryByCookie,

		// Standard annotations to ignore
		resources.LastAppliedConfiguration: p.handleIgnoredAnnotation,
	}
}
