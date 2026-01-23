package agic

import (
	"log/slog"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/logging"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

type Provider struct {
	log   *slog.Logger
	input resources.AGICResources
}

func NewProvider(input resources.AGICResources) Provider {
	return Provider{
		log:   logging.New("agic-converter"),
		input: input,
	}
}

func (p Provider) GetAnnotationHandlers() map[string]conversion.AnnotationHandler {
	return map[string]conversion.AnnotationHandler{
		AnnotationRequestTimeout:                p.handleRequestTimeout,
		AnnotationCookieBasedAffinity:           p.handleCookieBasedAffinity,
		AnnotationHealthProbeHostname:           p.handleHealthProbeHostname,
		AnnotationHealthProbePath:               p.handleHealthProbePath,
		AnnotationHealthProbeInterval:           p.handleHealthProbeInterval,
		AnnotationHealthProbePort:               p.handleHealthProbePort,
		AnnotationHealthProbeStatusCode:         p.handleHealthProbeStatusCode,
		AnnotationHealthProbeTimeout:            p.handleHealthProbeTimeout,
		AnnotationHealthProbeUnhealthyThreshold: p.handleHealthProbeUnhealthyThreshold,
		AnnotationHostnameExtension:             p.handleHostNameExtension,
		AnnotationBackendHostname:               p.handleBackendHostnameRewrite,
		AnnotationRewriteRuleSetCustomResource:  p.handleRewriteRuleSetCustomResource,
		AnnotationWAFPolicyForPath:              p.handleWAFPolicyForPath,
		AnnotationBackendPathPrefix:             p.handleBackendPathPrefix,
		AnnotationSSLRedirect:                   p.handleSSLRedirect,
		AnnotationBackendProtocol:               p.handleBackendProtocol,
		AnnotationAppGwTrustedRootCertificate:   p.handleAppGWTrustedRootCertificates,
		AnnotationAppGwSSLProfile:               p.handleAppGWSSLProfile,
	}
}
