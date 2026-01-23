package agic

import (
	"fmt"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

const (
	appGWSSLPolicy20150501  = "AppGwSslPolicy20150501"
	appGWSSLPolicy20170401  = "AppGwSslPolicy20170401"
	appGWSSLPolicy20170401S = "AppGwSslPolicy20170401S"
	appGWSSLPolicy20220101  = "AppGwSslPolicy20220101"
	appGWSSLPolicy20220101S = "AppGwSslPolicy20220101S"
)

func (p Provider) handleAppGWSSLProfile(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	var agcPolicy albcontrollerapi_v1.FrontendTLSPolicyTypeName

	switch annotationCtx.Value {
	case appGWSSLPolicy20170401S, appGWSSLPolicy20220101S:
		agcPolicy = albcontrollerapi_v1.PredefinedPolicy202306Strict
	default:
		agcPolicy = albcontrollerapi_v1.PredefinedPolicy202306
	}

	if len(routeCtx.GetParentHTTPSListenerNames(*gwCtx)) == 0 {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNoHTTPSListenersForSSLProfile, nil))
		return nil
	}

	for _, listener := range routeCtx.GetParentHTTPSListenerNames(*gwCtx) {
		policy := output.GetOrCreateFrontendTLSPolicy(k8snames.NamespacedName(gwCtx), listener)
		if policy.Spec.Default.FrontendTLSPolicyType == nil {
			policy.Spec.Default.FrontendTLSPolicyType = &albcontrollerapi_v1.PolicyType{
				FrontendTLSPolicyType: albcontrollerapi_v1.PredefinedFrontendTLSPolicyType,
				Name:                  agcPolicy,
			}

			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueFrontendTLSPolicyProfileCipherWarning, nil))
			annotationCtx.DestinationResources.Insert(resources.NewK8sResourceID(policy))
		} else if policy.Spec.Default.FrontendTLSPolicyType.Name != agcPolicy {
			err := fmt.Errorf("FrontendTLSPolicy %q already has conflicting SSL profile %q defined from a different Ingress", k8snames.NamespacedName(policy), policy.Spec.Default.FrontendTLSPolicyType.Name)
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueFrontendTLSPolicyProfileConflict, err))

			return err
		}
	}

	return nil
}
