package conversion

import (
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

func (c converter) handleWAFPolicyForGateway(
	output resources.AGCResourceGraph,
	gatewayCtx *GatewayContext,
	policyID string,
) {
	policy := output.GetOrCreateWAFPolicyForGateway(k8snames.NamespacedName(gatewayCtx))
	gatewayCtx.RegisterIssue(resources.NewIssue(resources.IssueWAFPotentialIncompatibility, nil))

	policy.Spec.WebApplicationFirewallPolicy.ID = policyID
}
