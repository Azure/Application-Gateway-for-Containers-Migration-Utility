package conversion

import (
	"reflect"
	"testing"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	"github.com/go-test/deep"
	network_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

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

type DummyProvider struct{}

func (DummyProvider) GetAnnotationHandlers() map[string]AnnotationHandler {
	return nil
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

		got, err := Convert(conversionInput, Options{Provider: DummyProvider{}})
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

func TestConvert(t *testing.T) {
	ingressWithTimeout := testutil.MakeIngress("default", "has-timeout", testutil.IngressOptions{}, "appgw.ingress.kubernetes.io/request-timeout", "123")

	ingressWithTwoHostsAndTimeout := ingressWithTimeout
	ingressWithTwoHostsAndTimeout.Spec.Rules = append(ingressWithTwoHostsAndTimeout.Spec.Rules, network_v1.IngressRule{
		Host: "host-2",
		IngressRuleValue: network_v1.IngressRuleValue{
			HTTP: &network_v1.HTTPIngressRuleValue{
				Paths: []network_v1.HTTPIngressPath{
					{
						Path:     "/",
						PathType: ptr.To(network_v1.PathTypeExact),
						Backend: network_v1.IngressBackend{
							Service: &network_v1.IngressServiceBackend{
								Name: "service",
							},
						},
					},
				},
			},
		},
	})

	cases := []testCase{
		// plain ingress
		{
			name: "1 Ingress no annotations",
			testCaseInput: testCaseInput{
				ingresses: []network_v1.Ingress{testutil.MakeIngress("default", "no-rp", testutil.IngressOptions{})},
			},
			testCaseOutput: testCaseOutput{
				httpRoute: []gatewayapi_v1.HTTPRoute{testutil.MakeHTTPRoute("default", "no-rp-all-hosts", testutil.HTTPRouteOptions{})},
				ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
					{Namespace: "default", Name: "no-rp"}: resources.MigrationStatusCompleted,
				},
			},
		},
		// 	{
		// 		name: "1 ingress 1 host 1 route timeout",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{ingressWithTimeout},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			routePolicy: []albcontrollerapi_v1.RoutePolicy{
		// 				testutil.MakeRoutePolicy("default", "has-timeout-all-hosts-policy", "default", "has-timeout-all-hosts", testutil.RoutePolicyOptions{
		// 					Timeout: ptr.To(time.Second * 123),
		// 				}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{testutil.MakeHTTPRoute("default", "has-timeout-all-hosts", testutil.HTTPRouteOptions{})},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-timeout"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},

		// 	// route policy
		// 	{
		// 		name: "1 ingress with two hosts with a route timeout",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{ingressWithTwoHostsAndTimeout},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			routePolicy: []albcontrollerapi_v1.RoutePolicy{
		// 				testutil.MakeRoutePolicy("default", "has-timeout-host-2-policy", "default", "has-timeout-host-2", testutil.RoutePolicyOptions{
		// 					Timeout: ptr.To(time.Second * 123),
		// 				}),
		// 				testutil.MakeRoutePolicy("default", "has-timeout-all-hosts-policy", "default", "has-timeout-all-hosts", testutil.RoutePolicyOptions{
		// 					Timeout: ptr.To(time.Second * 123),
		// 				}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-timeout-all-hosts", testutil.HTTPRouteOptions{}),
		// 				testutil.MakeHTTPRoute("default", "has-timeout-host-2", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-timeout"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with 1 host with cookie affinity",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{testutil.MakeIngress("default", "has-cookie", testutil.IngressOptions{}, AnnotationCookieBasedAffinity, "true")},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			routePolicy: []albcontrollerapi_v1.RoutePolicy{
		// 				testutil.MakeRoutePolicy("default", "has-cookie-all-hosts-policy", "default", "has-cookie-all-hosts", testutil.RoutePolicyOptions{
		// 					ManagedCookie: true,
		// 				}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-cookie-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-cookie"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with 1 host with cookie affinity set to false",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{testutil.MakeIngress("default", "has-cookie", testutil.IngressOptions{}, AnnotationCookieBasedAffinity, "false")},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			routePolicy: []albcontrollerapi_v1.RoutePolicy{},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-cookie-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-cookie"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with 1 host with cookie affinity and a route timeout",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-cookie", testutil.IngressOptions{}, AnnotationCookieBasedAffinity, "true", AnnotationRequestTimeout, "1"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			routePolicy: []albcontrollerapi_v1.RoutePolicy{
		// 				testutil.MakeRoutePolicy("default", "has-cookie-all-hosts-policy", "default", "has-cookie-all-hosts", testutil.RoutePolicyOptions{
		// 					ManagedCookie: true,
		// 					Timeout:       ptr.To(time.Second),
		// 				}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-cookie-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-cookie"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},

		// 	// health check policy
		// 	{
		// 		name: "1 ingress with health probe interval",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-interval", testutil.IngressOptions{}, AnnotationHealthProbeInterval, "10"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						Interval: ptr.To(10 * time.Second),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-interval-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-interval"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with health probe host",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-host", testutil.IngressOptions{}, AnnotationHealthProbeHostname, "host.example.com"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						Hostname: ptr.To("host.example.com"),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-host-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-host"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with health probe port",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-port", testutil.IngressOptions{}, AnnotationHealthProbePort, "8080"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						Port: ptr.To(8080),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-port-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-port"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with health probe status code",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-status-code", testutil.IngressOptions{}, AnnotationHealthProbeStatusCode, "200"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						StatusCodes: []*albcontrollerapi_v1.StatusCodes{{Start: 200, End: 200}},
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-status-code-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-status-code"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with health probe path",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-path", testutil.IngressOptions{}, AnnotationHealthProbePath, "/health"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						Path: ptr.To("/health"),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-path-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-path"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with health probe threshold",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-threshold", testutil.IngressOptions{}, AnnotationHealthProbeUnhealthyThreshold, "5"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						UnhealthyThreshold: ptr.To(5),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-threshold-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-threshold"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with health probe timeout",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-timeout", testutil.IngressOptions{}, AnnotationHealthProbeTimeout, "5"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						Timeout: ptr.To(time.Second * 5),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-timeout-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-timeout"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "1 ingress with health probe fully specified",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "has-health-probe-path", testutil.IngressOptions{},
		// 					AnnotationHealthProbePath, "/health",
		// 					AnnotationHealthProbeInterval, "15",
		// 					AnnotationHealthProbeHostname, "example.com",
		// 					AnnotationHealthProbePort, "9090",
		// 					AnnotationHealthProbeStatusCode, "201",
		// 					AnnotationHealthProbeUnhealthyThreshold, "3",
		// 					AnnotationHealthProbeTimeout, "5",
		// 				),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						Path:               ptr.To("/health"),
		// 						Interval:           ptr.To(time.Second * 15),
		// 						Hostname:           ptr.To("example.com"),
		// 						Port:               ptr.To(9090),
		// 						StatusCodes:        []*albcontrollerapi_v1.StatusCodes{{Start: 201, End: 201}},
		// 						UnhealthyThreshold: ptr.To(3),
		// 						Timeout:            ptr.To(time.Second * 5),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "has-health-probe-path-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "has-health-probe-path"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
		// 	{
		// 		name: "2 ingresses with the same backend but conflicting health probes",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "ingress-1", testutil.IngressOptions{Host: "1"},
		// 					AnnotationHealthProbeInterval, "1",
		// 					AnnotationHealthProbeHostname, "host1.example.com",
		// 					AnnotationHealthProbePort, "8080",
		// 					AnnotationHealthProbeStatusCode, "200-299",
		// 					AnnotationHealthProbeTimeout, "2",
		// 					AnnotationHealthProbeUnhealthyThreshold, "3",
		// 				),
		// 				testutil.MakeIngress("default", "ingress-2", testutil.IngressOptions{Host: "2"},
		// 					AnnotationHealthProbeInterval, "10",
		// 					AnnotationHealthProbeHostname, "host2.example.com",
		// 					AnnotationHealthProbePort, "8080",
		// 					AnnotationHealthProbeStatusCode, "1200-1299",
		// 					AnnotationHealthProbeTimeout, "12",
		// 					AnnotationHealthProbeUnhealthyThreshold, "13",
		// 				),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
		// 				testutil.MakeHealthCheckPolicy(
		// 					"default", "service-policy", "default", "service",
		// 					testutil.HealthCheckPolicyOptions{
		// 						Interval: ptr.To(time.Second * 1),
		// 						Hostname: ptr.To("host1.example.com"),
		// 						Port:     ptr.To(8080),
		// 						StatusCodes: []*albcontrollerapi_v1.StatusCodes{
		// 							{Start: 200, End: 299},
		// 						},
		// 						Timeout:            ptr.To(time.Second * 2),
		// 						UnhealthyThreshold: ptr.To(3),
		// 					}),
		// 			},
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "ingress-1-1", testutil.HTTPRouteOptions{}),
		// 				testutil.MakeHTTPRoute("default", "ingress-2-2", testutil.HTTPRouteOptions{}),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "ingress-1"}: resources.MigrationStatusCompleted,
		// 				{Namespace: "default", Name: "ingress-2"}: resources.MigrationStatusError,
		// 			},
		// 		},
		// 	},

		// 	// WAF Policy
		// 	{
		// 		name: "1 Ingress with WAF policy annotation",
		// 		testCaseInput: testCaseInput{
		// 			ingresses: []network_v1.Ingress{
		// 				testutil.MakeIngress("default", "waf-ingress", testutil.IngressOptions{},
		// 					AnnotationWAFPolicyForPath, "waf-policy-1"),
		// 			},
		// 		},
		// 		testCaseOutput: testCaseOutput{
		// 			httpRoute: []gatewayapi_v1.HTTPRoute{
		// 				testutil.MakeHTTPRoute("default", "waf-ingress-all-hosts", testutil.HTTPRouteOptions{}),
		// 			},
		// 			wafPolicy: []albcontrollerapi_v1.WebApplicationFirewallPolicy{
		// 				testutil.MakeWAFPolicy("default", "service-waf-policy", "default", "service", "waf-policy-1"),
		// 			},
		// 			ingressMigrationStatus: map[types.NamespacedName]resources.MigrationStatus{
		// 				{Namespace: "default", Name: "waf-ingress"}: resources.MigrationStatusCompleted,
		// 			},
		// 		},
		// 	},
	}

	for _, tc := range cases {
		tc.Run(t)
	}
}

func TestConversionManaged(t *testing.T) {
	got, err := Convert(resources.AGICResources{}, Options{ManagedSubnetID: "123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got.ApplicationLoadBalancer == nil {
		t.Fatal("Expected ApplicationLoadBalancer to be created")
	}

	if !reflect.DeepEqual(got.ApplicationLoadBalancer.Spec.Associations, []string{"123"}) {
		t.Fatalf("Expected ManagedSubnetID to be [\"123\"], got: %v", got.ApplicationLoadBalancer.Spec.Associations)
	}
}

func TestConversionBYO(t *testing.T) {
	got, err := Convert(resources.AGICResources{}, Options{BYOResourceID: "123"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if got.ApplicationLoadBalancer != nil {
		t.Fatal("Expected ApplicationLoadBalancer to not be created")
	}

	if anno, ok := got.Gateway.Annotations[k8snames.ALBArmResourceIDAnnotation]; !ok || anno != "123" {
		t.Fatalf("Expected Gateway to have BYOApplicationLoadBalancer annotation set to \"123\", got: %v", anno)
	}
}

// func TestConversionEdgeCases(t *testing.T) {
// 	cases := []testCase{
// 		{
// 			name: "multiple ingresses producing policies targeting the same host & backend",
// 			testCaseInput: testCaseInput{
// 				ingresses: []network_v1.Ingress{
// 					testutil.MakeIngress("default", "ingress-1", testutil.IngressOptions{Host: "same-host", Path: "/health"}, AnnotationHealthProbeInterval, "1"),
// 					testutil.MakeIngress("default", "ingress-2", testutil.IngressOptions{Host: "same-host", Path: "/timeout"}, AnnotationRequestTimeout, "10"),
// 					testutil.MakeIngress("default", "ingress-3", testutil.IngressOptions{Host: "same-host", Path: "/cookie"}, AnnotationCookieBasedAffinity, "true"),
// 				},
// 			},
// 			testCaseOutput: testCaseOutput{
// 				httpRoute: []gatewayapi_v1.HTTPRoute{
// 					testutil.MakeHTTPRoute("default", "ingress-1-same-host", testutil.HTTPRouteOptions{}),
// 					testutil.MakeHTTPRoute("default", "ingress-2-same-host", testutil.HTTPRouteOptions{}),
// 					testutil.MakeHTTPRoute("default", "ingress-3-same-host", testutil.HTTPRouteOptions{}),
// 				},
// 				routePolicy: []albcontrollerapi_v1.RoutePolicy{
// 					testutil.MakeRoutePolicy("default", "ingress-2-same-host-policy", "default", "ingress-2-same-host", testutil.RoutePolicyOptions{
// 						Timeout: ptr.To(10 * time.Second),
// 					}),
// 					testutil.MakeRoutePolicy("default", "ingress-3-same-host-policy", "default", "ingress-3-same-host", testutil.RoutePolicyOptions{
// 						ManagedCookie: true,
// 					}),
// 				},
// 				healthCheckPolicy: []albcontrollerapi_v1.HealthCheckPolicy{
// 					testutil.MakeHealthCheckPolicy("default", "service-policy", "default", "service", testutil.HealthCheckPolicyOptions{Interval: ptr.To(time.Second)}),
// 				},
// 			},
// 		},
// 	}
// 	for _, tc := range cases {
// 		tc.Run(t)
// 	}
// }

// func TestBackendPathPrefix(t *testing.T) {
// 	tc := testCase{
// 		testCaseInput: testCaseInput{
// 			ingresses: []network_v1.Ingress{
// 				testutil.MakeIngress("default", "has-rewrite", testutil.IngressOptions{}, AnnotationBackendPathPrefix, "my-prefix"),
// 			},
// 		},
// 		testCaseOutput: testCaseOutput{
// 			httpRoute: []gatewayapi_v1.HTTPRoute{
// 				testutil.MakeHTTPRoute("default", "has-rewrite-all-hosts", testutil.HTTPRouteOptions{
// 					Paths: []testutil.HTTPRoutePathOption{
// 						{
// 							Path: "/has-rewrite",
// 							Filters: []gatewayapi_v1.HTTPRouteFilter{
// 								{
// 									Type: gatewayapi_v1.HTTPRouteFilterURLRewrite,
// 									URLRewrite: &gatewayapi_v1.HTTPURLRewriteFilter{
// 										Path: &gatewayapi_v1.HTTPPathModifier{
// 											Type:               gatewayapi_v1.PrefixMatchHTTPPathModifier,
// 											ReplacePrefixMatch: ptr.To("my-prefix"),
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				}),
// 			},
// 		},
// 	}
// 	tc.Run(t)
// }

// func TestBackendHostnameRewrite(t *testing.T) {
// 	tc := testCase{
// 		testCaseInput: testCaseInput{
// 			ingresses: []network_v1.Ingress{
// 				testutil.MakeIngress("default", "has-rewrite", testutil.IngressOptions{}, AnnotationBackendHostname, "my-new-hostname"),
// 			},
// 		},
// 		testCaseOutput: testCaseOutput{
// 			httpRoute: []gatewayapi_v1.HTTPRoute{
// 				testutil.MakeHTTPRoute("default", "has-rewrite-all-hosts",
// 					testutil.HTTPRouteOptions{
// 						Paths: []testutil.HTTPRoutePathOption{
// 							{
// 								Path: "/has-rewrite",
// 								Filters: []gatewayapi_v1.HTTPRouteFilter{
// 									{
// 										Type: gatewayapi_v1.HTTPRouteFilterURLRewrite,
// 										URLRewrite: &gatewayapi_v1.HTTPURLRewriteFilter{
// 											Hostname: ptr.To(gatewayapi_v1.PreciseHostname("my-new-hostname")),
// 										},
// 									},
// 								},
// 							},
// 						},
// 					}),
// 			},
// 		},
// 	}
// 	tc.Run(t)
// }
