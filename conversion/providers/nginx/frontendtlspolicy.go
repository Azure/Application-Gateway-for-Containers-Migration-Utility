package nginx

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	crds_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// ========== TLS Policy Handlers ==========

func (p Provider) handleSSLCiphers(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC uses predefined TLS policies, not custom cipher suites
	// Map to an appropriate predefined policy
	listenerNames := routeCtx.GetParentHTTPSListenerNames(*gwCtx)
	if len(listenerNames) == 0 {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNoHTTPSListenersForSSLProfile, nil))
		return nil
	}

	gwNN := types.NamespacedName{Name: gwCtx.Gateway.Name, Namespace: gwCtx.Gateway.Namespace}

	for _, listener := range listenerNames {
		policy := output.GetOrCreateFrontendTLSPolicy(gwNN, listener)
		if policy.Spec.Default.FrontendTLSPolicyType == nil {
			policy.Spec.Default.FrontendTLSPolicyType = &crds_v1.PolicyType{
				FrontendTLSPolicyType: crds_v1.PredefinedFrontendTLSPolicyType,
				Name:                  crds_v1.PredefinedPolicy202306,
			}
		}

		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
	}

	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXSSLPolicyConversion, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleSSLPreferServerCiphers(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC predefined policies handle cipher preference internally
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXSSLPolicyConversion, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleSSLProtocols(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Map SSL protocols to appropriate AGC predefined policy
	listenerNames := routeCtx.GetParentHTTPSListenerNames(*gwCtx)
	if len(listenerNames) == 0 {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNoHTTPSListenersForSSLProfile, nil))
		return nil
	}

	// Determine policy based on protocols
	protocols := strings.ToLower(annotationCtx.Value)

	var agcPolicy crds_v1.FrontendTLSPolicyTypeName

	if strings.Contains(protocols, "tlsv1.3") && !strings.Contains(protocols, "tlsv1.1") {
		agcPolicy = crds_v1.PredefinedPolicy202306Strict
	} else {
		agcPolicy = crds_v1.PredefinedPolicy202306
	}

	gwNN := types.NamespacedName{Name: gwCtx.Gateway.Name, Namespace: gwCtx.Gateway.Namespace}
	for _, listener := range listenerNames {
		policy := output.GetOrCreateFrontendTLSPolicy(gwNN, listener)
		if policy.Spec.Default.FrontendTLSPolicyType == nil {
			policy.Spec.Default.FrontendTLSPolicyType = &crds_v1.PolicyType{
				FrontendTLSPolicyType: crds_v1.PredefinedFrontendTLSPolicyType,
				Name:                  agcPolicy,
			}
		}

		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
	}

	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXSSLPolicyConversion, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}
