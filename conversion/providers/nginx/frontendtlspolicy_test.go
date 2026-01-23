package nginx

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	crds_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func setupFrontendTLSPolicyTestInputs(key, val string) (
	Provider,
	resources.AGCResourceGraph,
	*conversion.GatewayContext,
	*conversion.HTTPRouteContext,
	*resources.IngressContext,
	*resources.IngressAnnotationContext,
) {
	provider := NewProvider(resources.NewAGICResources(nil, nil))
	graph := resources.NewAGCResourceGraph()
	gw := testutil.MakeGateway("default", "gw-1")
	// Add HTTPS listener
	gw.Spec.Listeners = append(gw.Spec.Listeners, gatewayapi_v1.Listener{
		Name:     "https",
		Protocol: gatewayapi_v1.HTTPSProtocolType,
		Port:     443,
	})
	gwCtx := conversion.NewGatewayContext(ptr.To(gw))
	gwCtx.HTTPSListeners = map[conversion.ListenerKey]conversion.HTTPSListener{
		{Hostname: "*", Port: 443}: {Listener: &gatewayapi_v1.Listener{Name: "https"}},
	}

	route := testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})
	route.Spec.ParentRefs = []gatewayapi_v1.ParentReference{
		{
			Name:        "gw-1",
			SectionName: ptr.To(gatewayapi_v1.SectionName("https")),
		},
	}
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(route))
	ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{}))
	annoCtx := resources.NewIngressAnnotationContext(key, val)

	return provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx
}

func TestNGINXHandleSSLCiphers(t *testing.T) {
	t.Run("with HTTPS listener creates FrontendTLSPolicy", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSPolicyTestInputs(
			AnnotationSSLCiphers, "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256",
		)

		err := provider.handleSSLCiphers(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleSSLCiphers returned error: %v", err)
		}

		if len(graph.FrontendTLSPolicies) != 1 {
			t.Fatalf("expected 1 FrontendTLSPolicy, got %d", len(graph.FrontendTLSPolicies))
		}

		// Check policy was created with predefined type
		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.FrontendTLSPolicyType == nil {
				t.Fatal("expected FrontendTLSPolicyType to be set")
			}

			if policy.Spec.Default.FrontendTLSPolicyType.FrontendTLSPolicyType != crds_v1.PredefinedFrontendTLSPolicyType {
				t.Errorf("expected PredefinedFrontendTLSPolicyType, got %s", policy.Spec.Default.FrontendTLSPolicyType.FrontendTLSPolicyType)
			}
		}

		// Check status and issues
		if annotationCtx.Status() != resources.MigrationStatusWarning {
			t.Errorf("expected status Warning, got %s", annotationCtx.Status())
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issues to be registered")
		}
	})

	t.Run("without HTTPS listener registers issue", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationSSLCiphers, "ECDHE-ECDSA-AES128-GCM-SHA256",
		)
		// Clear HTTPS listeners
		gwCtx.HTTPSListeners = nil

		err := provider.handleSSLCiphers(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleSSLCiphers returned error: %v", err)
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issue to be registered for no HTTPS listeners")
		}

		if annotationCtx.Issues[0].Code != resources.IssueNoHTTPSListenersForSSLProfile {
			t.Errorf("expected issue code %d, got %d", resources.IssueNoHTTPSListenersForSSLProfile, annotationCtx.Issues[0].Code)
		}
	})
}

func TestNGINXHandleSSLPreferServerCiphers(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSSLPreferServerCiphers, "true",
	)

	err := provider.handleSSLPreferServerCiphers(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleSSLPreferServerCiphers returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered")
	}

	if annotationCtx.Issues[0].Code != resources.IssueNGINXSSLPolicyConversion {
		t.Errorf("expected issue code %d, got %d", resources.IssueNGINXSSLPolicyConversion, annotationCtx.Issues[0].Code)
	}
}

func TestNGINXHandleSSLProtocols(t *testing.T) {
	t.Run("with TLSv1.2 and TLSv1.3 (no TLSv1.1) uses strict policy", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSPolicyTestInputs(
			AnnotationSSLProtocols, "TLSv1.2 TLSv1.3",
		)

		err := provider.handleSSLProtocols(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleSSLProtocols returned error: %v", err)
		}

		if len(graph.FrontendTLSPolicies) != 1 {
			t.Fatalf("expected 1 FrontendTLSPolicy, got %d", len(graph.FrontendTLSPolicies))
		}

		// Check the policy name - TLSv1.3 present and no TLSv1.1 means strict
		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.FrontendTLSPolicyType == nil {
				t.Fatal("expected FrontendTLSPolicyType to be set")
			}

			if policy.Spec.Default.FrontendTLSPolicyType.Name != crds_v1.PredefinedPolicy202306Strict {
				t.Errorf("expected policy name %s, got %s", crds_v1.PredefinedPolicy202306Strict, policy.Spec.Default.FrontendTLSPolicyType.Name)
			}
		}
	})

	t.Run("with TLSv1.3 only and no TLSv1.1 uses strict policy", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSPolicyTestInputs(
			AnnotationSSLProtocols, "TLSv1.3",
		)

		err := provider.handleSSLProtocols(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleSSLProtocols returned error: %v", err)
		}

		if len(graph.FrontendTLSPolicies) != 1 {
			t.Fatalf("expected 1 FrontendTLSPolicy, got %d", len(graph.FrontendTLSPolicies))
		}

		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.FrontendTLSPolicyType.Name != crds_v1.PredefinedPolicy202306Strict {
				t.Errorf("expected strict policy, got %s", policy.Spec.Default.FrontendTLSPolicyType.Name)
			}
		}
	})

	t.Run("with TLSv1.1 uses non-strict policy", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSPolicyTestInputs(
			AnnotationSSLProtocols, "TLSv1.1 TLSv1.2 TLSv1.3",
		)

		err := provider.handleSSLProtocols(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleSSLProtocols returned error: %v", err)
		}

		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.FrontendTLSPolicyType.Name != crds_v1.PredefinedPolicy202306 {
				t.Errorf("expected non-strict policy, got %s", policy.Spec.Default.FrontendTLSPolicyType.Name)
			}
		}
	})

	t.Run("without HTTPS listener registers issue", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationSSLProtocols, "TLSv1.2 TLSv1.3",
		)
		gwCtx.HTTPSListeners = nil

		err := provider.handleSSLProtocols(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleSSLProtocols returned error: %v", err)
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issue to be registered for no HTTPS listeners")
		}

		if annotationCtx.Issues[0].Code != resources.IssueNoHTTPSListenersForSSLProfile {
			t.Errorf("expected issue code %d, got %d", resources.IssueNoHTTPSListenersForSSLProfile, annotationCtx.Issues[0].Code)
		}
	})

	t.Run("adds destination to annotation context", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSPolicyTestInputs(
			AnnotationSSLProtocols, "TLSv1.2",
		)

		err := provider.handleSSLProtocols(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleSSLProtocols returned error: %v", err)
		}

		if annotationCtx.DestinationResources.Len() == 0 {
			t.Error("expected destination to be added to annotation context")
		}
	})
}

func TestNGINXHandleSSLCiphersIdempotent(t *testing.T) {
	// Test that calling the handler multiple times doesn't create duplicate policies
	provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSPolicyTestInputs(
		AnnotationSSLCiphers, "ECDHE-RSA-AES128-GCM-SHA256",
	)

	// Call twice
	_ = provider.handleSSLCiphers(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
	_ = provider.handleSSLCiphers(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)

	if len(graph.FrontendTLSPolicies) != 1 {
		t.Errorf("expected 1 FrontendTLSPolicy after multiple calls, got %d", len(graph.FrontendTLSPolicies))
	}
}

func TestNGINXFrontendTLSPolicyMultipleListeners(t *testing.T) {
	provider := NewProvider(resources.NewAGICResources(nil, nil))
	graph := resources.NewAGCResourceGraph()
	gw := testutil.MakeGateway("default", "gw-1")
	gw.Spec.Listeners = []gatewayapi_v1.Listener{
		{Name: "https-1", Protocol: gatewayapi_v1.HTTPSProtocolType, Port: 443, Hostname: ptr.To(gatewayapi_v1.Hostname("example.com"))},
		{Name: "https-2", Protocol: gatewayapi_v1.HTTPSProtocolType, Port: 443, Hostname: ptr.To(gatewayapi_v1.Hostname("other.com"))},
	}
	gwCtx := conversion.NewGatewayContext(ptr.To(gw))
	gwCtx.HTTPSListeners = map[conversion.ListenerKey]conversion.HTTPSListener{
		{Hostname: "example.com", Port: 443}: {Listener: &gatewayapi_v1.Listener{Name: "https-1"}},
		{Hostname: "other.com", Port: 443}:   {Listener: &gatewayapi_v1.Listener{Name: "https-2"}},
	}

	route := testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})
	route.Spec.ParentRefs = []gatewayapi_v1.ParentReference{
		{Name: "gw-1", SectionName: ptr.To(gatewayapi_v1.SectionName("https-1"))},
		{Name: "gw-1", SectionName: ptr.To(gatewayapi_v1.SectionName("https-2"))},
	}
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(route))
	ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{}))
	annotationCtx := resources.NewIngressAnnotationContext(AnnotationSSLCiphers, "ECDHE-RSA-AES128-GCM-SHA256")

	err := provider.handleSSLCiphers(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
	if err != nil {
		t.Errorf("handleSSLCiphers returned error: %v", err)
	}

	// Should create policies for both listeners
	if len(graph.FrontendTLSPolicies) < 1 {
		t.Errorf("expected at least 1 FrontendTLSPolicy for multiple listeners, got %d", len(graph.FrontendTLSPolicies))
	}

	// Check destinations were added
	if annotationCtx.DestinationResources.Len() == 0 {
		t.Error("expected destinations to be added for multiple listeners")
	}
}

func TestNGINXFrontendTLSPolicyGatewayNamespacing(t *testing.T) {
	provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSPolicyTestInputs(
		AnnotationSSLCiphers, "ECDHE-RSA-AES128-GCM-SHA256",
	)

	err := provider.handleSSLCiphers(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
	if err != nil {
		t.Errorf("handleSSLCiphers returned error: %v", err)
	}

	// Check that policy is in the right namespace
	gwNN := types.NamespacedName{Name: gwCtx.Gateway.Name, Namespace: gwCtx.Gateway.Namespace}
	found := false

	for nn := range graph.FrontendTLSPolicies {
		if nn.Namespace == gwNN.Namespace {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected FrontendTLSPolicy to be in the same namespace as the Gateway")
	}
}
