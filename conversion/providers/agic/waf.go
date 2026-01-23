package agic

import (
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

func (p Provider) handleWAFPolicyForPath(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	policy := output.GetOrCreateWAFPolicyForRoute(k8snames.NamespacedName(routeCtx))
	if policy.Spec.WebApplicationFirewallPolicy.ID != "" &&
		policy.Spec.WebApplicationFirewallPolicy.ID != annotationCtx.Value {
		// I don't think this should ever happen
		p.log.Warn(
			"overwriting conflicting WAF policy ID",
			"old-id",
			policy.Spec.WebApplicationFirewallPolicy.ID,
			"new-id",
			annotationCtx.Value,
			"name",
			k8snames.NamespacedName(policy),
		)
	}

	policy.Spec.WebApplicationFirewallPolicy.ID = annotationCtx.Value
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueWAFPotentialIncompatibility, nil))
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}
