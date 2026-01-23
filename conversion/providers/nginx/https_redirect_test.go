package nginx

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestNGINXHandleSSLRedirect(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSSLRedirect, "true",
	)

	err := provider.handleSSLRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleSSLRedirect returned error: %v", err)
	}

	if !gwCtx.HasHTTPSRedirect {
		t.Error("expected HasHTTPSRedirect to be true")
	}

	// Check that an HTTPRoute was created for the redirect
	redirectNN := types.NamespacedName{Name: "https-redirect", Namespace: "default"}
	_, exists := graph.HTTPRoutes[redirectNN]

	if !exists {
		t.Fatal("expected https-redirect HTTPRoute to be created")
	}
}

func TestNGINXHandleSSLRedirectFalse(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSSLRedirect, "false",
	)

	err := provider.handleSSLRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleSSLRedirect returned error: %v", err)
	}

	if gwCtx.HasHTTPSRedirect {
		t.Error("expected HasHTTPSRedirect to be false")
	}

	if annotationCtx.Status() != resources.MigrationStatusIgnored {
		t.Errorf("expected status Ignored when ssl-redirect is false, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleSSLRedirectInvalidValue(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSSLRedirect, "not-a-boolean",
	)

	err := provider.handleSSLRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err == nil {
		t.Error("expected error for invalid boolean value")
	}

	if annotationCtx.Status() != resources.MigrationStatusError {
		t.Errorf("expected status Error, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered")
	}
}

func TestNGINXHandleSSLRedirectAlreadyExists(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSSLRedirect, "true",
	)
	gwCtx.HasHTTPSRedirect = true

	err := provider.handleSSLRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleSSLRedirect returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusCompleted {
		t.Errorf("expected status Completed when redirect already exists, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleSSLRedirectCreatesCorrectHTTPRoute(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSSLRedirect, "true",
	)

	_ = provider.handleSSLRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)

	redirectNN := types.NamespacedName{Name: "https-redirect", Namespace: "default"}
	route, exists := graph.HTTPRoutes[redirectNN]

	if !exists {
		t.Fatal("expected https-redirect HTTPRoute to be created")
	}

	// Verify route has correct structure
	if len(route.Spec.Rules) == 0 {
		t.Fatal("expected route to have rules")
	}

	rule := route.Spec.Rules[0]
	if len(rule.Filters) == 0 {
		t.Fatal("expected rule to have filters")
	}

	filter := rule.Filters[0]
	if filter.Type != gatewayapi_v1.HTTPRouteFilterRequestRedirect {
		t.Errorf("expected RequestRedirect filter type, got %s", filter.Type)
	}

	if filter.RequestRedirect == nil || filter.RequestRedirect.Scheme == nil {
		t.Fatal("expected redirect scheme to be set")
	}

	if *filter.RequestRedirect.Scheme != "https" {
		t.Errorf("expected scheme 'https', got '%s'", *filter.RequestRedirect.Scheme)
	}
}

func TestNGINXHandleForceSSLRedirect(t *testing.T) {
	t.Run("true value creates redirect", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationForceSSLRedirect, "true",
		)

		err := provider.handleForceSSLRedirect(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleForceSSLRedirect returned error: %v", err)
		}

		if !gwCtx.HasHTTPSRedirect {
			t.Error("expected HasHTTPSRedirect to be true")
		}

		redirectNN := types.NamespacedName{Name: "https-redirect", Namespace: "default"}
		_, exists := graph.HTTPRoutes[redirectNN]

		if !exists {
			t.Fatal("expected https-redirect HTTPRoute to be created")
		}
	})

	t.Run("false value completes without creating redirect", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationForceSSLRedirect, "false",
		)

		err := provider.handleForceSSLRedirect(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleForceSSLRedirect returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusCompleted {
			t.Errorf("expected status Completed, got %s", annotationCtx.Status())
		}

		if gwCtx.HasHTTPSRedirect {
			t.Error("expected HasHTTPSRedirect to be false for 'false' value")
		}
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationForceSSLRedirect, "invalid",
		)

		err := provider.handleForceSSLRedirect(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err == nil {
			t.Error("expected error for invalid boolean value")
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issue to be registered")
		}
	})
}

func TestNGINXHandleSSLRedirectSetsDestination(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSSLRedirect, "true",
	)

	_ = provider.handleSSLRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)

	if annotationCtx.DestinationResources.Len() == 0 {
		t.Error("expected destination to be added")
	}
}
