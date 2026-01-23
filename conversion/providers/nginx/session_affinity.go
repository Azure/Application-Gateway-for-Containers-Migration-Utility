package nginx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	crds_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

// ========== Client Certificate Authentication Handlers ==========

func (p Provider) handleAuthTLSSecret(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// NGINX auth-tls-secret specifies the secret containing the CA certificate
	// AGC uses FrontendTLSPolicy for client certificate verification
	listenerNames := routeCtx.GetParentHTTPSListenerNames(*gwCtx)
	if len(listenerNames) == 0 {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNoHTTPSListenersForSSLProfile, nil))
		return nil
	}

	// Parse the secret reference (format: namespace/name or just name)
	secretRef := annotationCtx.Value
	secretNamespace := ingressCtx.Ingress.Namespace
	secretName := secretRef

	if strings.Contains(secretRef, "/") {
		parts := strings.SplitN(secretRef, "/", 2)
		secretNamespace = parts[0]
		secretName = parts[1]
	}

	gwNN := types.NamespacedName{Name: gwCtx.Gateway.Name, Namespace: gwCtx.Gateway.Namespace}
	for _, listener := range listenerNames {
		policy := output.GetOrCreateFrontendTLSPolicy(gwNN, listener)
		if policy.Spec.Default.Verify == nil {
			policy.Spec.Default.Verify = &crds_v1.MTLSPolicyVerify{}
		}

		policy.Spec.Default.Verify.CaCertificateRef = &gatewayapi_v1.SecretObjectReference{
			Name:      gatewayapi_v1.ObjectName(secretName),
			Namespace: ptr.To(gatewayapi_v1.Namespace(secretNamespace)),
		}
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
	}

	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAuthTLSNotFullySupported, nil))

	return nil
}

func (p Provider) handleAuthTLSVerifyClient(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// NGINX values: "on", "off", "optional", "optional_no_ca"
	// AGC FrontendTLSPolicy supports client validation via Verify field
	listenerNames := routeCtx.GetParentHTTPSListenerNames(*gwCtx)
	if len(listenerNames) == 0 {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNoHTTPSListenersForSSLProfile, nil))
		return nil
	}

	value := strings.ToLower(annotationCtx.Value)
	switch value {
	case "on", "optional":
		// AGC requires the Verify field to be set for mTLS
		gwNN := types.NamespacedName{Name: gwCtx.Gateway.Name, Namespace: gwCtx.Gateway.Namespace}
		for _, listener := range listenerNames {
			policy := output.GetOrCreateFrontendTLSPolicy(gwNN, listener)
			if policy.Spec.Default.Verify == nil {
				policy.Spec.Default.Verify = &crds_v1.MTLSPolicyVerify{}
			}

			annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
		}

		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAuthTLSNotFullySupported, nil))
	case "off":
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	case "optional_no_ca":
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAuthTLSNotFullySupported, nil))
		annotationCtx.SetStatus(resources.MigrationStatusNotSupported)
	default:
		err := fmt.Errorf("unknown auth-tls-verify-client value: %s", annotationCtx.Value)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))

		return err
	}

	return nil
}

func (p Provider) handleAuthTLSVerifyDepth(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC doesn't have direct support for verify depth
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAuthTLSNotFullySupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleAuthTLSErrorPage(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC doesn't have direct support for custom error pages
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAuthTLSNotFullySupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

func (p Provider) handleAuthTLSPassCertToUpstream(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC doesn't have direct support for passing client cert to upstream
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAuthTLSNotFullySupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

// ========== Session Affinity Handlers ==========

func (p Provider) handleAffinity(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// NGINX affinity annotation typically has value "cookie"
	if strings.ToLower(annotationCtx.Value) != "cookie" {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAffinityTypeNotSupported, nil))
		return nil
	}

	policy, err := p.GetOrCreateRoutePolicy(output, routeCtx.HTTPRoute, annotationCtx)
	if err != nil {
		return err
	}

	if policy.Spec.Default.SessionAffinity != nil {
		p.log.Warn("overwriting SessionAffinity with managed cookie on", "RoutePolicy", k8snames.NamespacedName(policy))
	}

	policy.Spec.Default.SessionAffinity = &crds_v1.SessionAffinity{
		AffinityType: crds_v1.AffinityTypeManagedCookie,
	}
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

	return nil
}

func (p Provider) handleAffinityMode(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// NGINX affinity-mode can be "balanced" or "persistent"
	// AGC doesn't have a direct equivalent
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAffinityModeNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleAffinityCanaryBehavior(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAffinityCanaryBehaviorNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

func (p Provider) handleSessionCookieName(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC managed cookies don't support custom names
	// Log warning but continue with managed cookie
	p.log.Warn("NGINX session-cookie-name annotation value cannot be used with AGC managed cookies", "value", annotationCtx.Value)
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAffinityModeNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleSessionCookiePath(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC managed cookies don't support custom paths
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAffinityModeNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleSessionCookieExpires(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Parse cookie expiration and set on RoutePolicy if possible
	seconds, err := strconv.Atoi(annotationCtx.Value)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))
		return err
	}

	policy, err := p.GetOrCreateRoutePolicy(output, routeCtx.HTTPRoute, annotationCtx)
	if err != nil {
		return err
	}

	if policy.Spec.Default.SessionAffinity == nil {
		policy.Spec.Default.SessionAffinity = &crds_v1.SessionAffinity{
			AffinityType: crds_v1.AffinityTypeManagedCookie,
		}
	}

	policy.Spec.Default.SessionAffinity.CookieDuration = meta_v1.Duration{Duration: time.Duration(seconds) * time.Second}
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

	return nil
}

func (p Provider) handleSessionCookieMaxAge(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Same as expires - use TTL
	return p.handleSessionCookieExpires(output, gwCtx, routeCtx, ingressCtx, annotationCtx)
}

func (p Provider) handleSessionCookieSameSite(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC managed cookies don't support custom SameSite attribute
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXAffinityModeNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

// GetOrCreateRoutePolicy gets or creates a RoutePolicy for the given HTTPRoute.
func (p Provider) GetOrCreateRoutePolicy(
	output resources.AGCResourceGraph,
	httpRoute *gatewayapi_v1.HTTPRoute,
	_ *resources.IngressAnnotationContext,
) (*crds_v1.RoutePolicy, error) {
	return output.GetOrCreateRoutePolicy(k8snames.NamespacedName(httpRoute)), nil
}
