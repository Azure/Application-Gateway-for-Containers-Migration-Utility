package agic

import (
	"testing"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	networking_v1 "k8s.io/api/networking/v1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestHandleBackendProtocol(t *testing.T) {
	t.Run("backend protocol is http", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationBackendProtocol, "http")
		_ = provider.handleBackendProtocol(graph, gwCtx, routeCtx, ingressCtx, annoCtx)

		if annoCtx.Status() != resources.MigrationStatusCompleted {
			t.Fatalf("Expected annotation status to be completed, got: %q", resources.MigrationStatusCompleted)
		}
	})

	t.Run("backend protocol is https", func(t *testing.T) {
		ingress := testutil.MakeIngress(
			"default",
			"ingress-1",
			testutil.IngressOptions{Service: "service-1", ServicePortNumber: 443},
			AnnotationBackendProtocol, "https")

		tc := testCase{
			testCaseInput: testCaseInput{
				ingresses: []networking_v1.Ingress{ingress},
			},
			testCaseOutput: testCaseOutput{
				backendTLSPolicy: []albcontrollerapi_v1.BackendTLSPolicy{
					testutil.MakeBackendTLSPolicy("default", "service-1-policy", "default", "service-1", 443),
				},
				httpRoute: []gatewayapi_v1.HTTPRoute{
					testutil.MakeHTTPRoute("default", "ingress-1-all-hosts", testutil.HTTPRouteOptions{}),
				},
			},
		}
		tc.Run(t)
	})

	t.Run("backend protocol is invalid", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationBackendProtocol, "invalid")
		_ = provider.handleBackendProtocol(graph, gwCtx, routeCtx, ingressCtx, annoCtx)

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("Expected annotation status to be %q, got: %q", resources.MigrationStatusError, annoCtx.Status())
		}
	})
}

func TestHandleAppGWTrustedRootCertificates(t *testing.T) {
	provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationAppGwSSLCertificate, "my-cert")
	_ = provider.handleAppGWTrustedRootCertificates(graph, gwCtx, routeCtx, ingressCtx, annoCtx)

	if annoCtx.Status() != resources.MigrationStatusNotSupported {
		t.Fatalf("expected annotation status to be %s, got %s", resources.MigrationStatusNotSupported, annoCtx.Status())
	}
}
