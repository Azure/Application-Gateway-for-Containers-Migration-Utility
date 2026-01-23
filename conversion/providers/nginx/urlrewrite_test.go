package nginx

import (
	"testing"

	"k8s.io/utils/ptr"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestNGINXHandleRewriteTarget_WithCaptureGroupsAndRegexPath(t *testing.T) {
	// Test scenario: rewrite-target with capture groups ($2) and a regex path
	// This simulates the common NGINX pattern: path /v2(/|$)(.*) with rewrite /$2
	provider := NewProvider(resources.NewAGICResources(nil, nil))
	graph := resources.NewAGCResourceGraph()
	gwCtx := conversion.NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1")))

	// Create a route with a regex-like path that contains metacharacters
	route := testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{
		Paths: []testutil.HTTPRoutePathOption{
			{
				Path:     "/v2(/|$)(.*)",
				Backends: []testutil.HTTPRoutePathBackendOption{{BackendName: "backend-v2"}},
			},
			{
				Path:     "/",
				Backends: []testutil.HTTPRoutePathBackendOption{{BackendName: "backend-v1"}},
			},
		},
	})
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(route))
	annotationCtx := resources.NewIngressAnnotationContext(AnnotationRewriteTarget, "/$2")

	err := provider.handleRewriteTarget(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleRewriteTarget returned error: %v", err)
	}

	// First rule (regex path) should have a URL rewrite filter
	if len(routeCtx.HTTPRoute.Spec.Rules[0].Filters) == 0 {
		t.Fatal("expected URL rewrite filter on regex path rule")
	}

	filter := routeCtx.HTTPRoute.Spec.Rules[0].Filters[0]
	if filter.Type != gatewayapi_v1.HTTPRouteFilterURLRewrite {
		t.Errorf("expected filter type URLRewrite, got %s", filter.Type)
	}

	// Check that the path type was correctly set to ReplacePrefixMatch
	if filter.URLRewrite == nil || filter.URLRewrite.Path == nil {
		t.Fatal("expected URLRewrite.Path to be set")
	}

	if filter.URLRewrite.Path.Type != gatewayapi_v1.PrefixMatchHTTPPathModifier {
		t.Errorf("expected path type ReplacePrefixMatch, got %s", filter.URLRewrite.Path.Type)
	}

	// The replacement should be "/" since /$2 strips the prefix
	if filter.URLRewrite.Path.ReplacePrefixMatch == nil || *filter.URLRewrite.Path.ReplacePrefixMatch != "/" {
		t.Errorf("expected ReplacePrefixMatch to be '/', got %v", filter.URLRewrite.Path.ReplacePrefixMatch)
	}

	// The path should be converted from regex to prefix
	if routeCtx.HTTPRoute.Spec.Rules[0].Matches[0].Path.Value == nil ||
		*routeCtx.HTTPRoute.Spec.Rules[0].Matches[0].Path.Value != "/v2" {
		t.Errorf("expected path to be converted to '/v2', got %v",
			routeCtx.HTTPRoute.Spec.Rules[0].Matches[0].Path.Value)
	}

	// Second rule (simple path "/") should NOT have a URL rewrite filter
	// because capture groups don't apply to non-regex paths
	if len(routeCtx.HTTPRoute.Spec.Rules[1].Filters) != 0 {
		t.Errorf("expected no filters on simple path rule, got %d", len(routeCtx.HTTPRoute.Spec.Rules[1].Filters))
	}

	// Should have warning status due to capture groups
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered for capture groups")
	}
}

func TestNGINXHandleRewriteTarget_SimpleRewrite(t *testing.T) {
	// Test scenario: simple rewrite-target without capture groups (e.g., /api)
	provider := NewProvider(resources.NewAGICResources(nil, nil))
	graph := resources.NewAGCResourceGraph()
	gwCtx := conversion.NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1")))

	route := testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{
		Paths: []testutil.HTTPRoutePathOption{
			{
				Path:     "/old-path",
				Backends: []testutil.HTTPRoutePathBackendOption{{BackendName: "backend"}},
			},
		},
	})
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(route))
	annotationCtx := resources.NewIngressAnnotationContext(AnnotationRewriteTarget, "/new-path")

	err := provider.handleRewriteTarget(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleRewriteTarget returned error: %v", err)
	}

	// Should have URL rewrite filter
	if len(routeCtx.HTTPRoute.Spec.Rules[0].Filters) == 0 {
		t.Fatal("expected URL rewrite filter")
	}

	filter := routeCtx.HTTPRoute.Spec.Rules[0].Filters[0]
	if filter.URLRewrite.Path.ReplacePrefixMatch == nil || *filter.URLRewrite.Path.ReplacePrefixMatch != "/new-path" {
		t.Errorf("expected ReplacePrefixMatch to be '/new-path', got %v", filter.URLRewrite.Path.ReplacePrefixMatch)
	}

	// Should complete without warnings for simple rewrites
	if annotationCtx.Status() != resources.MigrationStatusCompleted {
		t.Errorf("expected status Completed, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleRewriteTarget_PathModifierType(t *testing.T) {
	// Test that the path modifier type is correctly set to ReplacePrefixMatch (not PathPrefix)
	provider := NewProvider(resources.NewAGICResources(nil, nil))
	graph := resources.NewAGCResourceGraph()
	gwCtx := conversion.NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1")))

	route := testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(route))
	annotationCtx := resources.NewIngressAnnotationContext(AnnotationRewriteTarget, "/api")

	err := provider.handleRewriteTarget(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleRewriteTarget returned error: %v", err)
	}

	if len(routeCtx.HTTPRoute.Spec.Rules[0].Filters) == 0 {
		t.Fatal("expected URL rewrite filter")
	}

	filter := routeCtx.HTTPRoute.Spec.Rules[0].Filters[0]
	if filter.URLRewrite == nil || filter.URLRewrite.Path == nil {
		t.Fatal("expected URLRewrite.Path to be set")
	}

	// This is the key assertion - the type must be PrefixMatchHTTPPathModifier (ReplacePrefixMatch)
	// not PathMatchPathPrefix (PathPrefix) which was the bug
	if filter.URLRewrite.Path.Type != gatewayapi_v1.PrefixMatchHTTPPathModifier {
		t.Errorf("expected path type PrefixMatchHTTPPathModifier (ReplacePrefixMatch), got %s", filter.URLRewrite.Path.Type)
	}
}

func TestNGINXHandleAppRoot(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAppRoot, "/app",
	)

	originalRulesCount := len(routeCtx.HTTPRoute.Spec.Rules)

	err := provider.handleAppRoot(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAppRoot returned error: %v", err)
	}

	// Check that a redirect rule was added
	route := routeCtx.HTTPRoute
	if len(route.Spec.Rules) != originalRulesCount+1 {
		t.Fatalf("expected %d rules, got %d", originalRulesCount+1, len(route.Spec.Rules))
	}

	// The new redirect rule should be first
	firstRule := route.Spec.Rules[0]
	if len(firstRule.Filters) == 0 {
		t.Fatal("expected filters on the redirect rule")
	}

	if firstRule.Filters[0].Type != gatewayapi_v1.HTTPRouteFilterRequestRedirect {
		t.Errorf("expected RequestRedirect filter, got %s", firstRule.Filters[0].Type)
	}
}

func TestNGINXHandleXForwardedPrefix(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationXForwardedPrefix, "/api/v1",
	)

	err := provider.handleXForwardedPrefix(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleXForwardedPrefix returned error: %v", err)
	}

	// Check that the HTTPRoute has a RequestHeaderModifier filter
	route := routeCtx.HTTPRoute
	if len(route.Spec.Rules) == 0 {
		t.Fatal("expected at least one rule")
	}

	foundHeaderFilter := false

	for _, rule := range route.Spec.Rules {
		for _, filter := range rule.Filters {
			if filter.Type == gatewayapi_v1.HTTPRouteFilterRequestHeaderModifier {
				foundHeaderFilter = true

				if filter.RequestHeaderModifier == nil {
					t.Fatal("expected RequestHeaderModifier to be set")
				}
				// Check that X-Forwarded-Prefix header is set
				found := false

				for _, header := range filter.RequestHeaderModifier.Set {
					if header.Name == "X-Forwarded-Prefix" && header.Value == "/api/v1" {
						found = true
						break
					}
				}

				if !found {
					t.Error("expected X-Forwarded-Prefix header to be set with value /api/v1")
				}
			}
		}
	}

	if !foundHeaderFilter {
		t.Error("expected RequestHeaderModifier filter to be added")
	}

	if annotationCtx.Status() != resources.MigrationStatusCompleted {
		t.Errorf("expected status Completed, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleUpstreamVhost(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationUpstreamVhost, "backend.example.com",
	)

	err := provider.handleUpstreamVhost(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleUpstreamVhost returned error: %v", err)
	}

	// Check that URL rewrite filter with hostname was added
	route := routeCtx.HTTPRoute
	if len(route.Spec.Rules) == 0 || len(route.Spec.Rules[0].Filters) == 0 {
		t.Fatal("expected filters to be added to the route")
	}

	filter := route.Spec.Rules[0].Filters[0]
	if filter.Type != gatewayapi_v1.HTTPRouteFilterURLRewrite {
		t.Errorf("expected filter type URLRewrite, got %s", filter.Type)
	}

	if filter.URLRewrite == nil || filter.URLRewrite.Hostname == nil {
		t.Fatal("expected URLRewrite.Hostname to be set")
	}

	if string(*filter.URLRewrite.Hostname) != "backend.example.com" {
		t.Errorf("expected hostname backend.example.com, got %s", *filter.URLRewrite.Hostname)
	}
}

func TestNGINXHandleUseRegexTrue(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationUseRegex, "true",
	)

	err := provider.handleUseRegex(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleUseRegex returned error: %v", err)
	}

	// use-regex has limited support, should return warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered for limited support")
	}
}

func TestNGINXHandleUseRegexFalse(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationUseRegex, "false",
	)

	err := provider.handleUseRegex(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleUseRegex returned error: %v", err)
	}

	// false value should complete without issues
	if annotationCtx.Status() != resources.MigrationStatusCompleted {
		t.Errorf("expected status Completed, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleUseRegexInvalid(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationUseRegex, "invalid",
	)

	err := provider.handleUseRegex(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err == nil {
		t.Error("expected error for invalid bool value")
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered for invalid value")
	}
}
