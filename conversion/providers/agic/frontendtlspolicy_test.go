package agic

import (
	"testing"

	networking_v1 "k8s.io/api/networking/v1"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestHandleAppGWSSLProfile(t *testing.T) {
	setup := func() (Provider, resources.AGCResourceGraph, *conversion.GatewayContext, *conversion.HTTPRouteContext, *resources.IngressContext) {
		gw := &gatewayapi_v1.Gateway{}
		gw.Name = "test-gateway"
		gw.Namespace = "test-ns"
		gwCtx := &conversion.GatewayContext{Gateway: gw, HTTPSListeners: map[conversion.ListenerKey]conversion.HTTPSListener{}}
		lName := gatewayapi_v1.SectionName("https-listener")
		gwCtx.HTTPSListeners[conversion.ListenerKey{Hostname: "example.com", Port: 443}] = conversion.HTTPSListener{Listener: &gatewayapi_v1.Listener{Name: lName}}

		route := &gatewayapi_v1.HTTPRoute{}
		route.Spec.ParentRefs = []gatewayapi_v1.ParentReference{{SectionName: &lName}}
		routeCtx := &conversion.HTTPRouteContext{HTTPRoute: route}
		ingressCtx := resources.NewIngressContext(networking_v1.Ingress{})

		return NewProvider(resources.AGICResources{}), resources.NewAGCResourceGraph(), gwCtx, routeCtx, ingressCtx
	}

	t.Run("no HTTPS listeners", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx := setup()
		gwCtx.HTTPSListeners = nil
		annoCtx := resources.NewIngressAnnotationContext(AnnotationAppGwSSLProfile, appGWSSLPolicy20150501)
		_ = provider.handleAppGWSSLProfile(graph, gwCtx, routeCtx, ingressCtx, annoCtx)

		if len(annoCtx.Issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(annoCtx.Issues))
		}

		if annoCtx.Issues[0].Code != resources.IssueNoHTTPSListenersForSSLProfile {
			t.Fatalf("expected issue code %q, got %q", resources.IssueNoHTTPSListenersForSSLProfile, annoCtx.Issues[0].Code)
		}
	})

	t.Run("valid strict profile", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx := setup()
		annoCtx := resources.NewIngressAnnotationContext(AnnotationAppGwSSLProfile, appGWSSLPolicy20170401S)
		_ = provider.handleAppGWSSLProfile(graph, gwCtx, routeCtx, ingressCtx, annoCtx)

		if len(annoCtx.Issues) != 1 {
			t.Fatalf("expected 1 issues, got %d", len(annoCtx.Issues))
		}

		if annoCtx.Issues[0].Code != resources.IssueFrontendTLSPolicyProfileCipherWarning {
			t.Fatalf("expected Issue Code %d got %d", resources.IssueFrontendTLSPolicyProfileCipherWarning, annoCtx.Issues[0].Code)
		}

		if len(graph.FrontendTLSPolicies) != 1 {
			t.Fatalf("expected 1 FrontendTLSPolicy, got %d", len(graph.FrontendTLSPolicies))
		}

		for _, p := range graph.FrontendTLSPolicies {
			if p.Spec.Default.FrontendTLSPolicyType.Name != albcontrollerapi_v1.PredefinedPolicy202306Strict {
				t.Fatalf("expected policy type %q, got %q", albcontrollerapi_v1.PredefinedPolicy202306Strict, p.Spec.Default.FrontendTLSPolicyType.Name)
			}
		}
	})

	t.Run("conflicting profile", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx := setup()
		annoCtx := resources.NewIngressAnnotationContext(AnnotationAppGwSSLProfile, appGWSSLPolicy20170401S)
		_ = provider.handleAppGWSSLProfile(graph, gwCtx, routeCtx, ingressCtx, annoCtx)
		annoCtx = resources.NewIngressAnnotationContext(AnnotationAppGwSSLProfile, appGWSSLPolicy20170401)
		_ = provider.handleAppGWSSLProfile(graph, gwCtx, routeCtx, ingressCtx, annoCtx)

		if len(annoCtx.Issues) != 1 {
			t.Fatalf("expected 1 issues, got %d", len(annoCtx.Issues))
		}

		if annoCtx.Issues[0].Code != resources.IssueFrontendTLSPolicyProfileConflict {
			t.Fatalf("expected Issue Code %d got %d", resources.IssueFrontendTLSPolicyProfileConflict, annoCtx.Issues[0].Code)
		}

		if len(graph.FrontendTLSPolicies) != 1 {
			t.Fatalf("expected 1 FrontendTLSPolicy, got %d", len(graph.FrontendTLSPolicies))
		}

		for _, p := range graph.FrontendTLSPolicies {
			if p.Spec.Default.FrontendTLSPolicyType.Name != albcontrollerapi_v1.PredefinedPolicy202306Strict {
				t.Fatalf("expected policy type %q, got %q", albcontrollerapi_v1.PredefinedPolicy202306Strict, p.Spec.Default.FrontendTLSPolicyType.Name)
			}
		}
	})
}
