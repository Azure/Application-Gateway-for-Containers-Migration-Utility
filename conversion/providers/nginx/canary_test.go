package nginx

import (
	"testing"

	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestNGINXHandleCanary(t *testing.T) {
	t.Run("canary=true marks ingress as canary", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanary, "true",
		)

		err := provider.handleCanary(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if ingressCtx.Metadata["isCanary"] != annotationValueTrue {
			t.Error("expected isCanary metadata to be set to 'true'")
		}

		if annotationCtx.Status() != resources.MigrationStatusCompleted {
			t.Errorf("expected status Completed, got %s", annotationCtx.Status())
		}
	})

	t.Run("canary=false does not mark as canary", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanary, "false",
		)

		err := provider.handleCanary(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := ingressCtx.Metadata["isCanary"]; ok {
			t.Error("expected isCanary metadata to not be set")
		}

		if annotationCtx.Status() != resources.MigrationStatusCompleted {
			t.Errorf("expected status Completed, got %s", annotationCtx.Status())
		}
	})
}

func TestNGINXHandleCanaryWeight(t *testing.T) {
	t.Run("valid weight sets backend weight", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryWeight, "20",
		)

		// Add a backend ref to the route
		routeCtx.Spec.Rules = []gatewayapi_v1.HTTPRouteRule{
			{
				BackendRefs: []gatewayapi_v1.HTTPBackendRef{
					{BackendRef: gatewayapi_v1.BackendRef{}},
				},
			},
		}

		err := provider.handleCanaryWeight(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		weight := routeCtx.Spec.Rules[0].BackendRefs[0].Weight
		if weight == nil || *weight != 20 {
			t.Errorf("expected weight 20, got %v", weight)
		}

		if annotationCtx.Status() != resources.MigrationStatusWarning {
			t.Errorf("expected status Warning (for manual merge notice), got %s", annotationCtx.Status())
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issue about manual merge to be registered")
		}
	})

	t.Run("invalid weight returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryWeight, "not-a-number",
		)

		err := provider.handleCanaryWeight(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err == nil {
			t.Error("expected error for invalid weight value")
		}
	})

	t.Run("negative weight returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryWeight, "-5",
		)

		err := provider.handleCanaryWeight(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err == nil {
			t.Error("expected error for negative weight value")
		}
	})

	t.Run("weight exceeding total returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryWeight, "150",
		)

		err := provider.handleCanaryWeight(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err == nil {
			t.Error("expected error when weight exceeds total")
		}
	})

	t.Run("custom weight total is respected", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryWeight, "500",
		)

		// Set custom total
		ingressCtx.Ingress.Annotations[AnnotationCanaryWeightTotal] = "1000"

		// Add a backend ref
		routeCtx.Spec.Rules = []gatewayapi_v1.HTTPRouteRule{
			{
				BackendRefs: []gatewayapi_v1.HTTPBackendRef{
					{BackendRef: gatewayapi_v1.BackendRef{}},
				},
			},
		}

		err := provider.handleCanaryWeight(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		weight := routeCtx.Spec.Rules[0].BackendRefs[0].Weight
		if weight == nil || *weight != 500 {
			t.Errorf("expected weight 500, got %v", weight)
		}
	})
}

func TestNGINXHandleCanaryWeightTotal(t *testing.T) {
	t.Run("weight total marks as completed", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryWeightTotal, "1000",
		)

		err := provider.handleCanaryWeightTotal(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusCompleted {
			t.Errorf("expected status Completed, got %s", annotationCtx.Status())
		}
	})
}

func TestNGINXHandleCanaryByHeader(t *testing.T) {
	t.Run("header with exact value creates exact match", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByHeader, "X-Canary",
		)

		ingressCtx.Ingress.Annotations[AnnotationCanaryByHeaderValue] = "always"

		// Add a match to the route
		routeCtx.Spec.Rules = []gatewayapi_v1.HTTPRouteRule{
			{
				Matches: []gatewayapi_v1.HTTPRouteMatch{{}},
			},
		}

		err := provider.handleCanaryByHeader(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(routeCtx.Spec.Rules[0].Matches[0].Headers) != 1 {
			t.Fatalf("expected 1 header match, got %d", len(routeCtx.Spec.Rules[0].Matches[0].Headers))
		}

		headerMatch := routeCtx.Spec.Rules[0].Matches[0].Headers[0]
		if headerMatch.Name != "X-Canary" {
			t.Errorf("expected header name 'X-Canary', got '%s'", headerMatch.Name)
		}

		if headerMatch.Value != "always" {
			t.Errorf("expected header value 'always', got '%s'", headerMatch.Value)
		}

		if *headerMatch.Type != gatewayapi_v1.HeaderMatchExact {
			t.Errorf("expected Exact match type, got %v", *headerMatch.Type)
		}
	})

	t.Run("header with regex pattern creates regex match", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByHeader, "X-Version",
		)

		ingressCtx.Ingress.Annotations[AnnotationCanaryByHeaderPattern] = "^v[0-9]+$"

		routeCtx.Spec.Rules = []gatewayapi_v1.HTTPRouteRule{
			{
				Matches: []gatewayapi_v1.HTTPRouteMatch{{}},
			},
		}

		err := provider.handleCanaryByHeader(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		headerMatch := routeCtx.Spec.Rules[0].Matches[0].Headers[0]
		if *headerMatch.Type != gatewayapi_v1.HeaderMatchRegularExpression {
			t.Errorf("expected RegularExpression match type, got %v", *headerMatch.Type)
		}

		if headerMatch.Value != "^v[0-9]+$" {
			t.Errorf("expected pattern '^v[0-9]+$', got '%s'", headerMatch.Value)
		}
	})

	t.Run("header without value uses regex for any-except-never", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByHeader, "X-Enable-Canary",
		)

		routeCtx.Spec.Rules = []gatewayapi_v1.HTTPRouteRule{
			{
				Matches: []gatewayapi_v1.HTTPRouteMatch{{}},
			},
		}

		err := provider.handleCanaryByHeader(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		headerMatch := routeCtx.Spec.Rules[0].Matches[0].Headers[0]
		if *headerMatch.Type != gatewayapi_v1.HeaderMatchRegularExpression {
			t.Errorf("expected RegularExpression match type, got %v", *headerMatch.Type)
		}

		// Should register an approximation issue
		if len(annotationCtx.Issues) == 0 {
			t.Error("expected approximation issue to be registered")
		}
	})

	t.Run("empty header name returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByHeader, "",
		)

		err := provider.handleCanaryByHeader(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err == nil {
			t.Error("expected error for empty header name")
		}
	})

	t.Run("invalid regex pattern returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByHeader, "X-Test",
		)

		ingressCtx.Ingress.Annotations[AnnotationCanaryByHeaderPattern] = "[invalid"

		err := provider.handleCanaryByHeader(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err == nil {
			t.Error("expected error for invalid regex pattern")
		}
	})
}

func TestNGINXHandleCanaryByHeaderValue(t *testing.T) {
	t.Run("marks as completed", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByHeaderValue, "always",
		)

		err := provider.handleCanaryByHeaderValue(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusCompleted {
			t.Errorf("expected status Completed, got %s", annotationCtx.Status())
		}
	})
}

func TestNGINXHandleCanaryByHeaderPattern(t *testing.T) {
	t.Run("marks as completed", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByHeaderPattern, "^v[0-9]+$",
		)

		err := provider.handleCanaryByHeaderPattern(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusCompleted {
			t.Errorf("expected status Completed, got %s", annotationCtx.Status())
		}
	})
}

func TestNGINXHandleCanaryByCookie(t *testing.T) {
	t.Run("marks as not supported", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationCanaryByCookie, "canary-cookie",
		)

		err := provider.handleCanaryByCookie(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusNotSupported {
			t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issue to be registered")
		}
	})
}

func TestNGINXHandleCanaryByHeaderWithMultipleRules(t *testing.T) {
	t.Run("adds header match to all rules", func(t *testing.T) {
		provider := NewProvider(resources.NewAGICResources(nil, nil))
		graph := resources.NewAGCResourceGraph()

		gwCtx, routeCtx, ingressCtx := setupCanaryTestWithMultipleRules()
		annotationCtx := resources.NewIngressAnnotationContext(AnnotationCanaryByHeader, "X-Canary")
		ingressCtx.Ingress.Annotations[AnnotationCanaryByHeaderValue] = "true"

		err := provider.handleCanaryByHeader(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i, rule := range routeCtx.Spec.Rules {
			for j, match := range rule.Matches {
				if len(match.Headers) != 1 {
					t.Errorf("rule[%d].match[%d]: expected 1 header match, got %d", i, j, len(match.Headers))
				}
			}
		}
	})
}

func setupCanaryTestWithMultipleRules() (*conversion.GatewayContext, *conversion.HTTPRouteContext, *resources.IngressContext) {
	gwCtx := conversion.NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1")))
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})))

	// Add multiple rules with multiple matches
	routeCtx.Spec.Rules = []gatewayapi_v1.HTTPRouteRule{
		{
			Matches: []gatewayapi_v1.HTTPRouteMatch{{}, {}},
		},
		{
			Matches: []gatewayapi_v1.HTTPRouteMatch{{}},
		},
	}

	ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{}))

	return gwCtx, routeCtx, ingressCtx
}
