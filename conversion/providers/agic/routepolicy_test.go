package agic

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestHandleRequestTimeout(t *testing.T) {
	t.Run("valid timeout", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationRequestTimeout, "30")
		if err := provider.handleRequestTimeout(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		if len(graph.RoutePolicies) != 1 {
			t.Fatalf("expected 1 route policy, got %d", len(graph.RoutePolicies))
		}

		for _, policy := range graph.RoutePolicies {
			if policy.Spec.Default == nil || policy.Spec.Default.RouteTimeouts == nil {
				t.Fatal("expected RouteTimeouts to be set")
			}
			if policy.Spec.Default.RouteTimeouts.RouteTimeout.Seconds() != 30 {
				t.Fatalf("expected timeout of 30s, got %v", policy.Spec.Default.RouteTimeouts.RouteTimeout.Seconds())
			}
		}
	})

	t.Run("invalid timeout", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationRequestTimeout, "not-a-number")
		if err := provider.handleRequestTimeout(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("expected MigrationStatusError, got %v", annoCtx.Status())
		}
	})
}

func TestHandleCookieBasedAffinity(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationCookieBasedAffinity, "true")
		if err := provider.handleCookieBasedAffinity(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		if len(graph.RoutePolicies) != 1 {
			t.Fatalf("expected 1 route policy, got %d", len(graph.RoutePolicies))
		}

		for _, policy := range graph.RoutePolicies {
			if policy.Spec.Default == nil || policy.Spec.Default.SessionAffinity == nil {
				t.Fatal("expected SessionAffinity to be set")
			}
			if policy.Spec.Default.SessionAffinity.AffinityType != "managed-cookie" {
				t.Fatalf("expected AffinityType managed-cookie, got %v", policy.Spec.Default.SessionAffinity.AffinityType)
			}
		}
	})

	t.Run("disabled", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationCookieBasedAffinity, "false")
		if err := provider.handleCookieBasedAffinity(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		if len(graph.RoutePolicies) != 0 {
			t.Fatalf("expected 0 route policies, got %d", len(graph.RoutePolicies))
		}

		if annoCtx.Status() != resources.MigrationStatusCompleted {
			t.Fatalf("expected MigrationStatusCompleted, got %v", annoCtx.Status())
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationCookieBasedAffinity, "invalid")
		if err := provider.handleCookieBasedAffinity(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("expected MigrationStatusError, got %v", annoCtx.Status())
		}
	})
}
