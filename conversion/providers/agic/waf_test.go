package agic

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestHandleWAFPolicyForPath(t *testing.T) {
	policyID := "waf-id"
	// converter, graph, gatewayCtx, routeCtx, ingressCtx, annoCtx := ingressAnnotationTestSetup(AnnotationWAFPolicyForPath, policyID)
	provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationWAFPolicyForPath, policyID)
	err := provider.handleWAFPolicyForPath(graph, gwCtx, routeCtx, ingressCtx, annoCtx)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if annoCtx.Status() != resources.MigrationStatusWarning {
		t.Fatalf("expected annotation status: %q, got %q", resources.MigrationStatusWarning, annoCtx.Status())
	}

	if len(annoCtx.DestinationResources) != 1 {
		t.Fatalf("expected 1 destination resource, got %d", len(annoCtx.DestinationResources))
	}

	wafNN := annoCtx.DestinationResources.UnsortedList()[0].NamespacedName
	if got, ok := graph.WAFPolicies[wafNN]; !ok {
		t.Fatalf("WAF policy %q not present in output graph", wafNN)
	} else if got.Spec.WebApplicationFirewallPolicy.ID != policyID {
		t.Fatalf("Expected policy ID %q, got %q", policyID, got.Spec.WebApplicationFirewallPolicy.ID)
	}
}
