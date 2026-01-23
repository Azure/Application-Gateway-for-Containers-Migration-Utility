package conversion

import (
	"testing"

	networking_v1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestIngressHTTPS(t *testing.T) {
	graph, gwCtx, converter := ingressTestSetup()
	ingress := testutil.MakeIngress("default", "https-ingrss", testutil.IngressOptions{})
	ingress.Spec.TLS = []networking_v1.IngressTLS{
		{
			Hosts:      []string{"example-1.com"},
			SecretName: "secret-1",
		},
		{
			Hosts:      []string{"example-2.com"},
			SecretName: "secret-2",
		},
	}

	ingress.Spec.Rules[0].Host = "example-1.com"
	ingress.Spec.Rules = append(ingress.Spec.Rules, ingress.Spec.Rules[0])
	ingress.Spec.Rules[1].Host = "example-2.com"

	ingressCtx := resources.NewIngressContext(ingress)
	if err := converter.handleIngress(&graph, gwCtx, ingressCtx); err != nil {
		t.Fatalf("handleIngress returned error: %v", err)
	}

	if len(ingressCtx.HTTPRoutes) != 2 {
		t.Fatalf("expected 2 HTTPRoutes, got %d", len(ingressCtx.HTTPRoutes))
	}

	route1NN := ingressCtx.HTTPRoutes[0]
	route, ok := graph.HTTPRoutes[route1NN]

	if !ok {
		t.Fatalf("expected HTTPRoute %q to exist", route1NN)
	}

	if len(route.Spec.ParentRefs) != 1 {
		t.Fatalf("expected 1 ParentRef, got %d", len(route.Spec.ParentRefs))
	}

	parentRef := route.Spec.ParentRefs[0]
	if parentRef.SectionName == nil {
		t.Fatalf("expected SectionName to be set")
	}

	if *parentRef.SectionName != "https-example-1-com" {
		t.Errorf("expected SectionName to be 'https-example-1-com', got %s", *parentRef.SectionName)
	}

	if gwCtx.HTTPSListeners == nil {
		t.Fatalf("expected httpsListeners to be initialized")
	}

	listener1Key := ListenerKey{
		Hostname: "example-1.com",
		Port:     443,
	}

	if _, ok := gwCtx.HTTPSListeners[listener1Key]; !ok {
		t.Fatalf("expected httpsListener %q to exist", listener1Key)
	}

	listener2Key := ListenerKey{
		Hostname: "example-2.com",
		Port:     443,
	}

	if _, ok := gwCtx.HTTPSListeners[listener2Key]; !ok {
		t.Fatalf("expected httpsListener %q to exist", listener2Key)
	}
}

func TestHandleIngressAnnotationUnsupported(t *testing.T) {
	graph, gwCtx, converter := ingressTestSetup()
	ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "https-ingrss", testutil.IngressOptions{}, "unsupported-annotation", "value"))
	routeCtx := HTTPRouteContext{
		HTTPRoute: &gatewayapi_v1.HTTPRoute{},
	}

	if err := converter.handleIngressAnnotations(&graph, gwCtx, ingressCtx, &routeCtx); err != nil {
		t.Fatalf("handleIngressAnnotations returned error: %v", err)
	}

	annoCtx, exists := ingressCtx.Annotations["unsupported-annotation"]
	if !exists {
		t.Fatalf("expected annotation context for unsupported-annotation to exist")
	}

	if len(annoCtx.Issues) != 1 {
		t.Fatalf("expected 1 issue for unsupported-annotation, got %d", len(annoCtx.Issues))
	}

	issue := annoCtx.Issues[0]
	if issue.Code != resources.IssueUnsupportedAnnotationGeneric {
		t.Errorf("expected issue code to be IssueUnsupportedAnnotationGeneric, got %v", issue.Code)
	}
}

func TestHandleIngressAnnotationIgnored(t *testing.T) {
	graph, gwCtx, converter := ingressTestSetup()
	annoCtx := resources.NewIngressAnnotationContext("key", "val")
	_ = converter.handleIgnoredAnnotation(graph, gwCtx, nil, nil, annoCtx)

	if annoCtx.Status() != resources.MigrationStatusIgnored {
		t.Fatalf("Expected annotation status to be MigrationStatusIgnored, got: %v", annoCtx.Status())
	}
}

func ingressTestSetup() (resources.AGCResourceGraph, *GatewayContext, converter) {
	graph := resources.NewAGCResourceGraph()
	gwCtx := NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1")))
	converter := newConverter(resources.AGICResources{}, DummyProvider{})

	return graph, gwCtx, converter
}
