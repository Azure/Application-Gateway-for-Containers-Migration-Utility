package conversion

import "testing"

func TestHandleWAFPolicyForGateway(t *testing.T) {
	policyID := "waf-id"
	graph, gatewayCtx, converter := ingressTestSetup()
	converter.handleWAFPolicyForGateway(graph, gatewayCtx, policyID)

	if len(graph.WAFPolicies) != 1 {
		t.Fatalf("expected 1 waf policy, got %d", len(graph.WAFPolicies))
	}

	for _, policy := range graph.WAFPolicies {
		if policy.Spec.WebApplicationFirewallPolicy.ID != policyID {
			t.Fatalf("Expected policy ID %q, got %q", policyID, policy.Spec.WebApplicationFirewallPolicy.ID)
		}
	}
}
