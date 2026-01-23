package agic

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	network_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestRewriteRuleSetCRD(t *testing.T) {
	t.Run("1 ingress with url rewrite annotation with no ruleset defined", func(t *testing.T) {
		tc := testCase{
			testCaseInput: testCaseInput{
				ingresses: []network_v1.Ingress{
					testutil.MakeIngress("default", "has-rewrite", testutil.IngressOptions{}, AnnotationRewriteRuleSetCustomResource, "my-ruleset"),
				},
				agicRewrites: []appgwrewrite.AzureApplicationGatewayRewrite{},
			},
			testCaseOutput: testCaseOutput{
				httpRoute: []gatewayapi_v1.HTTPRoute{
					testutil.MakeHTTPRoute("default", "has-rewrite-all-hosts", testutil.HTTPRouteOptions{}),
				},
			},
		}
		tc.Run(t)
	})

	t.Run("1 ingress with url rewrite annotation with all supported rule types defined", func(t *testing.T) {
		tc := testCase{
			testCaseInput: testCaseInput{
				ingresses: []network_v1.Ingress{
					testutil.MakeIngress("default", "has-rewrite", testutil.IngressOptions{}, AnnotationRewriteRuleSetCustomResource, "my-ruleset"),
				},
				agicRewrites: []appgwrewrite.AzureApplicationGatewayRewrite{
					testutil.MakeRewriteRuleSet("default", "my-ruleset", testutil.RewriteRuleSetOptions{
						ReqHeadersSet:     map[string]string{"req-set-me": "req-set-value"},
						ReqHeadersDelete:  []string{"req-delete-me"},
						RespHeadersSet:    map[string]string{"resp-set-me": "resp-set-value"},
						RespHeadersDelete: []string{"resp-delete-me"},
						PathRewrite:       "rewrite-path",
					}),
				},
			},
			testCaseOutput: testCaseOutput{
				httpRoute: []gatewayapi_v1.HTTPRoute{
					testutil.MakeHTTPRoute("default", "has-rewrite-all-hosts",
						testutil.HTTPRouteOptions{
							Paths: []testutil.HTTPRoutePathOption{
								{
									Path: "/has-rewrite",
									Filters: []gatewayapi_v1.HTTPRouteFilter{
										{
											Type: gatewayapi_v1.HTTPRouteFilterRequestHeaderModifier,
											RequestHeaderModifier: &gatewayapi_v1.HTTPHeaderFilter{
												Set: []gatewayapi_v1.HTTPHeader{
													{
														Name:  "req-set-me",
														Value: "req-set-value",
													},
												},
												Remove: []string{"req-delete-me"},
											},
										},
										{
											Type: gatewayapi_v1.HTTPRouteFilterResponseHeaderModifier,
											ResponseHeaderModifier: &gatewayapi_v1.HTTPHeaderFilter{
												Set: []gatewayapi_v1.HTTPHeader{
													{
														Name:  "resp-set-me",
														Value: "resp-set-value",
													},
												},
												Remove: []string{"resp-delete-me"},
											},
										},
										{
											Type: gatewayapi_v1.HTTPRouteFilterURLRewrite,
											URLRewrite: &gatewayapi_v1.HTTPURLRewriteFilter{
												Path: &gatewayapi_v1.HTTPPathModifier{
													Type:            gatewayapi_v1.FullPathHTTPPathModifier,
													ReplaceFullPath: ptr.To("rewrite-path"),
												},
											},
										},
									}},
							},
						},
					),
				},
			},
		}
		tc.Run(t)
	})

	t.Run("different rule sequences get a warning", func(t *testing.T) {
		rewriteRules := []appgwrewrite.RewriteRule{
			{
				RuleSequence: 1,
			},
			{
				RuleSequence: 2,
			},
			{
				RuleSequence: 3,
			},
		}
		rewrite := testutil.MakeRewriteRuleSet("default", "my-ruleset", testutil.RewriteRuleSetOptions{}, rewriteRules...)
		provider := NewProvider(resources.AGICResources{
			AppGWRewrites: map[types.NamespacedName]*resources.AppGWRewriteContext{
				k8snames.NamespacedName(&rewrite): resources.NewAppGWRewriteContext(rewrite),
			},
		})
		annoCtx := resources.NewIngressAnnotationContext("key", "my-ruleset")
		graph := resources.NewAGCResourceGraph()
		err := provider.handleRewriteRuleSetCustomResource(
			graph,
			conversion.NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1"))),
			&conversion.HTTPRouteContext{},
			resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{})),
			annoCtx,
		)

		if err != nil {
			t.Fatalf("Expected err to be nil but go: %v", err)
		}

		if len(annoCtx.Issues) != 1 {
			t.Fatalf("Expected exactly 1 issue, got %d", annoCtx.Issues)
		}

		if annoCtx.Issues[0].Code != resources.IssueRewriteRuleSetRuleSequenceNotSupported {
			t.Fatalf("Expected Issue to be IssueRewriteRuleSetRuleSequenceNotSupported, got %d: %s",
				annoCtx.Issues[0].Code, resources.IssueLibrary[annoCtx.Issues[0].Code].Description)
		}
	})
}

func TestHandleBackendPathPrefix(t *testing.T) {
	t.Run("adds path prefix rewrite filter", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationBackendPathPrefix, "/api/v1")

		if err := provider.handleBackendPathPrefix(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		if len(routeCtx.Spec.Rules) == 0 || len(routeCtx.Spec.Rules[0].Filters) == 0 {
			t.Fatal("expected filter to be added to route")
		}

		filter := routeCtx.Spec.Rules[0].Filters[0]
		if filter.Type != gatewayapi_v1.HTTPRouteFilterURLRewrite {
			t.Fatalf("expected URLRewrite filter, got %v", filter.Type)
		}

		if filter.URLRewrite == nil || filter.URLRewrite.Path == nil {
			t.Fatal("expected URLRewrite.Path to be set")
		}

		if filter.URLRewrite.Path.Type != gatewayapi_v1.PrefixMatchHTTPPathModifier {
			t.Fatalf("expected PrefixMatchHTTPPathModifier, got %v", filter.URLRewrite.Path.Type)
		}

		if filter.URLRewrite.Path.ReplacePrefixMatch == nil || *filter.URLRewrite.Path.ReplacePrefixMatch != "/api/v1" {
			t.Fatalf("expected ReplacePrefixMatch to be /api/v1, got %v", filter.URLRewrite.Path.ReplacePrefixMatch)
		}

		if annoCtx.Status() != resources.MigrationStatusCompleted {
			t.Fatalf("expected MigrationStatusCompleted, got %v", annoCtx.Status())
		}
	})
}

func TestHandleBackendHostnameRewrite(t *testing.T) {
	t.Run("adds hostname rewrite filter", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationBackendHostname, "backend.example.com")

		if err := provider.handleBackendHostnameRewrite(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		if len(routeCtx.Spec.Rules) == 0 || len(routeCtx.Spec.Rules[0].Filters) == 0 {
			t.Fatal("expected filter to be added to route")
		}

		filter := routeCtx.Spec.Rules[0].Filters[0]
		if filter.Type != gatewayapi_v1.HTTPRouteFilterURLRewrite {
			t.Fatalf("expected URLRewrite filter, got %v", filter.Type)
		}

		if filter.URLRewrite == nil || filter.URLRewrite.Hostname == nil {
			t.Fatal("expected URLRewrite.Hostname to be set")
		}

		if string(*filter.URLRewrite.Hostname) != "backend.example.com" {
			t.Fatalf("expected hostname to be backend.example.com, got %v", *filter.URLRewrite.Hostname)
		}

		if annoCtx.Status() != resources.MigrationStatusCompleted {
			t.Fatalf("expected MigrationStatusCompleted, got %v", annoCtx.Status())
		}
	})
}
