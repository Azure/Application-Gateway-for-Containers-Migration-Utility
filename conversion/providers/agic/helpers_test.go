package agic

import (
	"testing"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	"github.com/go-test/deep"
	network_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func setupAnnotationHandlerInputs(key, val string) (
	Provider,
	resources.AGCResourceGraph,
	*conversion.GatewayContext,
	*conversion.HTTPRouteContext,
	*resources.IngressContext,
	*resources.IngressAnnotationContext,
) {
	provider := NewProvider(resources.NewAGICResources(nil, nil))
	graph := resources.NewAGCResourceGraph()
	gwCtx := conversion.NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1")))
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})))
	ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{}))
	annoCtx := resources.NewIngressAnnotationContext(key, val)

	return provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx
}

type testCase struct {
	name string
	testCaseInput
	testCaseOutput
	expectErr bool
}

type testCaseInput struct {
	ingresses    []network_v1.Ingress
	agicRewrites []appgwrewrite.AzureApplicationGatewayRewrite
}

type testCaseOutput struct {
	routePolicy            []albcontrollerapi_v1.RoutePolicy
	healthCheckPolicy      []albcontrollerapi_v1.HealthCheckPolicy
	backendTLSPolicy       []albcontrollerapi_v1.BackendTLSPolicy
	httpRoute              []gatewayapi_v1.HTTPRoute
	ingressMigrationStatus map[types.NamespacedName]resources.MigrationStatus
}

func (tc *testCase) AddIngress(ingress network_v1.Ingress) {
	tc.ingresses = append(tc.ingresses, ingress)
}

func (tc *testCase) AddRoutePolicy(policy albcontrollerapi_v1.RoutePolicy) {
	tc.routePolicy = append(tc.routePolicy, policy)
}

func (tc *testCase) Run(t *testing.T) {
	run := func(t *testing.T) {
		conversionInput := resources.NewAGICResources(tc.ingresses, tc.agicRewrites)

		got, err := conversion.Convert(conversionInput, conversion.Options{Provider: NewProvider(conversionInput)})
		if err != nil != tc.expectErr {
			t.Fatalf("Expected error: %t, got error: %v", tc.expectErr, err)
		}

		expectHTTPRoutes := make(map[types.NamespacedName]gatewayapi_v1.HTTPRoute)
		for _, route := range tc.httpRoute {
			expectHTTPRoutes[k8snames.NamespacedName(&route)] = route
		}

		expectRoutePolicies := make(map[types.NamespacedName]*albcontrollerapi_v1.RoutePolicy)
		for _, policy := range tc.routePolicy {
			expectRoutePolicies[k8snames.NamespacedName(&policy)] = &policy
		}

		expectHealthCheckPolicies := make(map[types.NamespacedName]*albcontrollerapi_v1.HealthCheckPolicy)
		for _, policy := range tc.healthCheckPolicy {
			expectHealthCheckPolicies[k8snames.NamespacedName(&policy)] = &policy
		}

		expectBackendTLSPolicies := make(map[types.NamespacedName]*albcontrollerapi_v1.BackendTLSPolicy)
		for _, policy := range tc.backendTLSPolicy {
			expectBackendTLSPolicies[k8snames.NamespacedName(&policy)] = &policy
		}

		if diff := deep.Equal(got.BackendTLSPolicies, expectBackendTLSPolicies); diff != nil {
			for _, line := range diff {
				t.Errorf("BackendTLSPolicy mismatch: %s", line)
			}
		}

		if diff := deep.Equal(got.HealthCheckPolicies, expectHealthCheckPolicies); diff != nil {
			for _, line := range diff {
				t.Errorf("HealthCheckPolicy mismatch: %s", line)
			}
		}

		if diff := deep.Equal(got.RoutePolicies, expectRoutePolicies); diff != nil {
			for _, line := range diff {
				t.Errorf("RoutePolicy mismatch: %s", line)
			}
		}

		// just compare the HTTP route names are correct, the full spec is tested in ingress2gateway
		if len(got.HTTPRoutes) != len(expectHTTPRoutes) {
			t.Errorf("Expected %d HTTPRoutes, got %d", len(expectHTTPRoutes), len(got.HTTPRoutes))
		}

		for routeNN := range expectHTTPRoutes {
			if _, ok := got.HTTPRoutes[routeNN]; !ok {
				t.Errorf("Missing expected HTTPRoute %q", routeNN)
			}
		}

		if tc.testCaseInput.agicRewrites != nil {
			for route, expectRoute := range expectHTTPRoutes {
				gotRoute, ok := got.HTTPRoutes[route]
				if !ok {
					t.Fatalf("Expected HTTPRoute %q not found in output", route)
				}

				for i := range gotRoute.Spec.Rules {
					if diff := deep.Equal(gotRoute.Spec.Rules[i].Filters, expectRoute.Spec.Rules[i].Filters); diff != nil {
						t.Fatalf("HTTPRoute %q mismatch: %v", route, diff)
					}
				}
			}
		}

		// check the ingress migration status
		for ingressNN, expectStatus := range tc.ingressMigrationStatus {
			ingressCtx, ok := conversionInput.IngressContexts[ingressNN]
			if !ok {
				t.Errorf("Expected Ingress %q migration status to be set to %q, but Ingress not found in input", ingressNN, expectStatus)
				continue
			}

			if ingressCtx.Status != expectStatus {
				t.Errorf("Expected Ingress %q migration status to be %q, got %q", ingressNN, expectStatus, ingressCtx.Status)
			}
		}

		for _, policy := range tc.routePolicy {
			routeNN := types.NamespacedName{
				Name:      string(policy.Spec.TargetRef.Name),
				Namespace: string(*policy.Spec.TargetRef.Namespace),
			}

			if _, ok := got.HTTPRoutes[routeNN]; !ok {
				t.Errorf("HTTPRoute %q referenced by RoutePolicy %q doesn't exist", routeNN, k8snames.NamespacedName(&policy))
			}
		}
	}

	if tc.name != "" {
		t.Run(tc.name, run)
	} else {
		run(t)
	}
}
