package reporting

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestNewIngress(t *testing.T) {
	newID := func(name, kind string) resources.K8sResourceID {
		return resources.K8sResourceID{
			GroupVersionKind: schema.GroupVersionKind{
				Group:   "group",
				Version: "version",
				Kind:    kind,
			},
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: "default",
			},
		}
	}
	k8sIngress := testutil.MakeIngress("namespace", "ingress-1", testutil.IngressOptions{})
	ingressCtx := resources.NewIngressContext(k8sIngress)
	ingressCtx.Status = resources.MigrationStatusCompleted
	ingressCtx.HTTPRoutes = []types.NamespacedName{
		{
			Name:      "route-1",
			Namespace: "default",
		},
	}

	addAnnotation := func(name string, destName string, kind string, status resources.MigrationStatus, issues ...resources.Issue) {
		annoCtx := resources.NewIngressAnnotationContext(name, "some-value")
		annoCtx.DestinationResources = sets.New(newID(destName, kind))
		annoCtx.Issues = issues
		annoCtx.SetStatus(status)
		ingressCtx.Annotations[name] = annoCtx
	}

	addAnnotation("backend tls", "backend-tls-1", k8snames.KindBackendTLSPolicy, resources.MigrationStatusCompleted)
	addAnnotation("route policy", "route-3", k8snames.KindRoutePolicy, resources.MigrationStatusCompleted)
	addAnnotation("frontend tls", "frontend-tls-1", k8snames.KindFrontendTLSPolicy, resources.MigrationStatusNotSupported)
	addAnnotation("health check", "health-check-1", k8snames.KindHealthCheckPolicy, resources.MigrationStatusError, resources.NewIssue(resources.IssueCouldNotFindRoute, errors.New("error")))
	addAnnotation("waf policy", "waf-policy-1", k8snames.KindWebApplicationFirewallPolicy, resources.MigrationStatusCompleted)
	addAnnotation("http route", "http-route-1", k8snames.KindHTTPRoute, resources.MigrationStatusCompleted)
	addAnnotation("warning", "", "", resources.MigrationStatusWarning, resources.NewIssue(resources.IssueFrontendTLSPolicyProfileCipherWarning, nil))

	route1NN := types.NamespacedName{Name: "route-1", Namespace: "namespace"}
	route2NN := types.NamespacedName{Name: "route-2", Namespace: "namespace"}
	ingressCtx.HTTPRoutes = []types.NamespacedName{
		route1NN,
		route2NN,
	}

	reportIngress := newIngress(ingressCtx)
	if reportIngress.Namespace != "namespace" {
		t.Errorf("expected Namespace to be 'namespace', got '%s'", reportIngress.Namespace)
	}

	if reportIngress.Name != "ingress-1" {
		t.Errorf("expected Name to be 'ingress-1', got '%s'", reportIngress.Name)
	}

	if reportIngress.Result != resources.MigrationStatusCompleted {
		t.Errorf("expected Result to be 'Completed', got '%s'", reportIngress.Result)
	}

	if len(reportIngress.MigratedIngressResources.HTTPRoutes) != 3 {
		t.Errorf("expected 2 HTTPRoutes, got %d", len(reportIngress.MigratedIngressResources.HTTPRoutes))
	}

	if len(reportIngress.Annotations.Completed) != 4 {
		t.Errorf("expected 6 completed annotations, got %d", len(reportIngress.Annotations.Completed))
	}

	if len(reportIngress.Annotations.Failed) != 1 {
		t.Errorf("expected 1 failed annotation, got %d", len(reportIngress.Annotations.Failed))
	}

	if len(reportIngress.Annotations.Warnings) != 1 {
		t.Errorf("expected 1 warning annotation, got %d", len(reportIngress.Annotations.Warnings))
	}

	if len(reportIngress.Annotations.Unsupported) != 1 {
		t.Errorf("expected 1 unsupported annotation, got %d", len(reportIngress.Annotations.Unsupported))
	}

	if len(reportIngress.MigratedIngressResources.BackendTLSPolicies) != 1 {
		t.Errorf("expected 1 BackendTLSPolicy, got %d", len(reportIngress.MigratedIngressResources.BackendTLSPolicies))
	}

	if len(reportIngress.MigratedIngressResources.FrontendTLSPolicies) != 1 {
		t.Errorf("expected 1 FrontendTLSPolicy, got %d", len(reportIngress.MigratedIngressResources.FrontendTLSPolicies))
	}

	if len(reportIngress.MigratedIngressResources.HealthCheckPolicies) != 1 {
		t.Errorf("expected 1 HealthCheckPolicy, got %d", len(reportIngress.MigratedIngressResources.HealthCheckPolicies))
	}

	if len(reportIngress.MigratedIngressResources.RoutePolicies) != 1 {
		t.Errorf("expected 1 RoutePolicy, got %d", len(reportIngress.MigratedIngressResources.RoutePolicies))
	}

	if len(reportIngress.MigratedIngressResources.WebApplicationFirewalPolicies) != 1 {
		t.Errorf("expected 1 WebApplicationFirewallPolicy, got %d", len(reportIngress.MigratedIngressResources.WebApplicationFirewalPolicies))
	}
}
