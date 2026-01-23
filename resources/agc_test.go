package resources

import (
	"reflect"
	"testing"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestNewAGCResourceGraph(t *testing.T) {
	instance := NewAGCResourceGraph()

	// Check that all of the maps have been initialized
	val := reflect.ValueOf(instance)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if field.Kind() == reflect.Map {
			if field.IsNil() {
				t.Errorf("Map field %s is nil", fieldType.Name)
			}
		}
	}
}

func TestGetOrCreateRoutePolicy(t *testing.T) {
	graph := NewAGCResourceGraph()
	ingress := testutil.MakeIngress("default", "test", testutil.IngressOptions{})
	routeNN := types.NamespacedName{Name: "route-1", Namespace: ingress.Namespace}

	var policy *albcontrollerapi_v1.RoutePolicy

	t.Run("doesn't exist", func(t *testing.T) {
		policy = graph.GetOrCreateRoutePolicy(routeNN)
		if _, ok := graph.RoutePolicies[k8snames.NamespacedName(policy)]; !ok {
			t.Fatal("RoutePolicy was not added to graph")
		}
	})

	t.Run("policy exists", func(t *testing.T) {
		// edit the WAF policy, then get it again
		policy.Spec.Default.RouteTimeouts = &albcontrollerapi_v1.RouteTimeouts{
			RouteTimeout: meta_v1.Duration{
				Duration: time.Hour,
			},
		}
		policy = graph.GetOrCreateRoutePolicy(routeNN)

		if policy.Spec.Default.RouteTimeouts.RouteTimeout.Duration != time.Hour {
			t.Fatal("Retrieved RoutePolicy did not have updated RouteTimeout setting")
		}
	})
}

func TestGetOrCreateHealthCheckPolicy(t *testing.T) {
	graph := NewAGCResourceGraph()
	ingress := testutil.MakeIngress("default", "test", testutil.IngressOptions{})
	routeNN := types.NamespacedName{Name: "route-1", Namespace: ingress.Namespace}

	var policy *albcontrollerapi_v1.HealthCheckPolicy

	t.Run("doesn't exist", func(t *testing.T) {
		policy = graph.GetOrCreateHealthCheckPolicy(routeNN)
		if _, ok := graph.HealthCheckPolicies[k8snames.NamespacedName(policy)]; !ok {
			t.Fatal("HealthCheckPolicy was not added to graph")
		}
	})

	t.Run("policy exists", func(t *testing.T) {
		// edit the WAF policy, then get it again
		policy.Spec.Default.Port = int32(1234)
		policy = graph.GetOrCreateHealthCheckPolicy(routeNN)

		if policy.Spec.Default.Port != int32(1234) {
			t.Fatal("Retrieved HealthCheckPolicy did not have updated RouteTimeout setting")
		}
	})
}

func TestGetOrCreateWAFPolicy(t *testing.T) {
	graph := NewAGCResourceGraph()
	ingress := testutil.MakeIngress("default", "test", testutil.IngressOptions{})
	ingress.Annotations["appgw.ingress.kubernetes.io/waf-policy-for-path"] = "waf-policy-1"
	routeNN := types.NamespacedName{Name: "route-1", Namespace: ingress.Namespace}

	var policy *albcontrollerapi_v1.WebApplicationFirewallPolicy

	t.Run("WAF Policy doesn't exist", func(t *testing.T) {
		policy = graph.GetOrCreateWAFPolicyForRoute(routeNN)
		if _, ok := graph.WAFPolicies[k8snames.NamespacedName(policy)]; !ok {
			t.Fatal("WAF Policy was not added to graph")
		}
	})

	t.Run("WAF policy exists", func(t *testing.T) {
		// edit the WAF policy, then get it again
		policy.Spec.WebApplicationFirewallPolicy.ID = "id"

		policy = graph.GetOrCreateWAFPolicyForRoute(routeNN)
		if policy.Spec.WebApplicationFirewallPolicy.ID != "id" {
			t.Fatal("Retrieved WAF policy did not have updated ID setting")
		}
	})
}

func TestGetOrCreateWAFPolicyForGateway(t *testing.T) {
	graph := NewAGCResourceGraph()
	gatewayNN := types.NamespacedName{Name: "gw-1", Namespace: "ns-1"}

	var policy *albcontrollerapi_v1.WebApplicationFirewallPolicy

	t.Run("WAF Policy doesn't exist", func(t *testing.T) {
		policy = graph.GetOrCreateWAFPolicyForGateway(gatewayNN)
		if _, ok := graph.WAFPolicies[k8snames.NamespacedName(policy)]; !ok {
			t.Fatal("WAF Policy was not added to graph")
		}
	})

	t.Run("WAF policy exists", func(t *testing.T) {
		// edit the WAF policy, then get it again
		policy.Spec.WebApplicationFirewallPolicy.ID = "id"

		policy = graph.GetOrCreateWAFPolicyForGateway(gatewayNN)
		if policy.Spec.WebApplicationFirewallPolicy.ID != "id" {
			t.Fatal("Retrieved WAF policy did not have updated ID setting")
		}
	})
}

func TestGetOrCreateBackendTLSPolicy(t *testing.T) {
	graph := NewAGCResourceGraph()
	ingress := testutil.MakeIngress("default", "test", testutil.IngressOptions{})
	routeNN := types.NamespacedName{Name: "route-1", Namespace: ingress.Namespace}

	var policy *albcontrollerapi_v1.BackendTLSPolicy

	t.Run("doesn't exist", func(t *testing.T) {
		policy = graph.GetOrCreateBackendTLSPolicy(routeNN)
		if _, ok := graph.BackendTLSPolicies[k8snames.NamespacedName(policy)]; !ok {
			t.Fatal("BackendTLSPolicy was not added to graph")
		}
	})

	t.Run("policy exists", func(t *testing.T) {
		// edit the WAF policy, then get it again
		policy.Spec.Default.Sni = "test"
		policy = graph.GetOrCreateBackendTLSPolicy(routeNN)

		if policy.Spec.Default.Sni != "test" {
			t.Fatal("Retrieved BackendTLSPolicy did not have updated SNI setting")
		}
	})
}

func TestGetOrCreateReferenceGrantForGWSecret(t *testing.T) {
	graph := NewAGCResourceGraph()
	gatewayNamespace := "gateway-ns"
	secretNamespace := "secret-ns"

	grant := graph.GetOrCreateReferenceGrantForGWSecret(gatewayNamespace, secretNamespace)
	if _, ok := graph.ReferenceGrants[k8snames.NamespacedName(grant)]; !ok {
		t.Fatal("ReferenceGrant was not added to graph")
	}

	t.Run("grant already exists", func(t *testing.T) {
		grant.Annotations = map[string]string{"test-annotation": "value"}
		grant = graph.GetOrCreateReferenceGrantForGWSecret(gatewayNamespace, secretNamespace)

		if val, ok := grant.Annotations["test-annotation"]; !ok || val != "value" {
			t.Fatal("Retrieved ReferenceGrant did not have updated annotation")
		}
	})
}

func TestGetorCreateFrontendTLSPolicy(t *testing.T) {
	graph := NewAGCResourceGraph()
	ingress := testutil.MakeIngress("default", "test", testutil.IngressOptions{})
	gwNN := types.NamespacedName{Name: "gw-1", Namespace: ingress.Namespace}
	listenerName := gatewayapi_v1.SectionName("listener-1")

	var policy *albcontrollerapi_v1.FrontendTLSPolicy

	t.Run("doesn't exist", func(t *testing.T) {
		policy = graph.GetOrCreateFrontendTLSPolicy(gwNN, listenerName)
		if _, ok := graph.FrontendTLSPolicies[k8snames.NamespacedName(policy)]; !ok {
			t.Fatal("FrontendTLSPolicy was not added to graph")
		}
	})

	t.Run("policy exists", func(t *testing.T) {
		policy.Spec.Default = &albcontrollerapi_v1.FrontendTLSPolicyConfig{
			FrontendTLSPolicyType: &albcontrollerapi_v1.PolicyType{
				Name: "updated",
			},
		}
		policy = graph.GetOrCreateFrontendTLSPolicy(gwNN, listenerName)

		if policy.Spec.Default.FrontendTLSPolicyType.Name != "updated" {
			t.Fatal("Retrieved FrontendTLSPolicy did not have updated FrontendTLSPolicyType.Name setting")
		}
	})
}
