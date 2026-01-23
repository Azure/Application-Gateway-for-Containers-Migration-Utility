package reporting

import (
	"testing"
	"time"

	agicmigration "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility"
	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	networking_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestNewReport(t *testing.T) {
	agicResources := resources.NewAGICResources(
		[]networking_v1.Ingress{testutil.MakeIngress("namespace", "ingress-1", testutil.IngressOptions{})},
		[]appgwrewrite.AzureApplicationGatewayRewrite{testutil.MakeRewriteRuleSet("namespace", "rewrite-1", testutil.RewriteRuleSetOptions{})},
	)
	agcResources := resources.NewAGCResourceGraph()
	backendTLS := testutil.MakeBackendTLSPolicy("namespace", "backend-tls", "namespace", "service-1", 443)
	httpRoute := testutil.MakeHTTPRoute("namespace", "ingress-1-all-hosts", testutil.HTTPRouteOptions{})
	gateway := testutil.MakeGateway("namespace", "gateway-1")
	healthCheckPolicy := testutil.MakeHealthCheckPolicy("namespace", "health-check", "service-namespace", "service-name", testutil.HealthCheckPolicyOptions{})
	referenceGrant := testutil.MakeReferenceGrant("namespace", "reference-grant")
	routePolicy := testutil.MakeRoutePolicy("namespace", "route-policy", "route-namespace", "route-name", testutil.RoutePolicyOptions{})
	wafPolicy := testutil.MakeWAFPolicy("namespace", "waf-policy", "route-namespace", "route-name", "waf-policy-id")

	agcResources.HTTPRoutes[k8snames.NamespacedName(&httpRoute)] = &httpRoute
	agcResources.BackendTLSPolicies[k8snames.NamespacedName(&backendTLS)] = &backendTLS
	agcResources.Gateway = &gateway
	agcResources.HealthCheckPolicies[k8snames.NamespacedName(&healthCheckPolicy)] = &healthCheckPolicy
	agcResources.ReferenceGrants[k8snames.NamespacedName(&referenceGrant)] = &referenceGrant
	agcResources.RoutePolicies[k8snames.NamespacedName(&routePolicy)] = &routePolicy
	agcResources.WAFPolicies[k8snames.NamespacedName(&wafPolicy)] = &wafPolicy

	report, err := NewReport(agicResources, agcResources, AGICResourcesSourceFiles, conversion.Options{BYOResourceID: "my-byo-id"})
	if err != nil {
		t.Fatalf("unexpected error creating report: %v", err)
	}

	if len(report.AGICResources.Ingresses) != 1 {
		t.Errorf("expected 1 ingress in report, got %d", len(report.AGICResources.Ingresses))
	}

	if len(report.AGCResources.HTTPRoutes) != 1 {
		t.Errorf("expected 1 HTTPRoute in report, got %d", len(report.AGCResources.HTTPRoutes))
	}

	if len(report.AGCResources.BackendTLSPolicies) != 1 {
		t.Errorf("expected 1 BackendTLSPolicy in report, got %d", len(report.AGCResources.BackendTLSPolicies))
	}

	if len(report.AGCResources.HealthCheckPolicies) != 1 {
		t.Errorf("expected 1 HealthCheckPolicy in report, got %d", len(report.AGCResources.HealthCheckPolicies))
	}

	if len(report.AGCResources.RoutePolicies) != 1 {
		t.Errorf("expected 1 RoutePolicy in report, got %d", len(report.AGCResources.RoutePolicies))
	}

	if len(report.AGCResources.WebApplicationFirewalPolicies) != 1 {
		t.Errorf("expected 1 WAFPolicy in report, got %d", len(report.AGCResources.WebApplicationFirewalPolicies))
	}

	if report.Metadata.AGICResourcesSource != AGICResourcesSourceFiles {
		t.Errorf("expected AGICResourcesSource to be %s, got %s", AGICResourcesSourceFiles, report.Metadata.AGICResourcesSource)
	}

	if report.Metadata.AGCBYOResourceID != "my-byo-id" {
		t.Errorf("expected AGCBYOResourceID to be 'my-byo-id', got %s", report.Metadata.AGCBYOResourceID)
	}

	if report.Metadata.ToolVersion != agicmigration.GetVersion() {
		t.Errorf("expected ToolVersion to be %s, got %s", agicmigration.GetVersion(), report.Metadata.ToolVersion)
	}

	if time.Now().Before(report.Metadata.GeneratedAt.Add(-time.Hour)) {
		t.Errorf("expected GeneratedAt to be within the last hour, got %s", report.Metadata.GeneratedAt)
	}
}

func TestAGICResources_sortContents(t *testing.T) {
	agicResources := AGICResources{
		Ingresses: []Ingress{
			{
				NamespacedName: types.NamespacedName{Namespace: "namespace", Name: "ingress-b"},
			},
			{
				NamespacedName: types.NamespacedName{Namespace: "namespace", Name: "ingress-a"},
			},
		},
	}

	agicResources.sortContents()

	if agicResources.Ingresses[0].Name != "ingress-a" {
		t.Errorf("expected first ingress to be 'ingress-a', got '%s'", agicResources.Ingresses[0].Name)
	}
}
