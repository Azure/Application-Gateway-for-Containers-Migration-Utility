package nginx

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestNGINXHandleLoadBalanceRoundRobin(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationLoadBalance, "round_robin",
	)

	err := provider.handleLoadBalance(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleLoadBalance returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusIgnored {
		t.Errorf("expected status Ignored for round_robin, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleLoadBalanceUnsupported(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationLoadBalance, "ip_hash",
	)

	err := provider.handleLoadBalance(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleLoadBalance returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported for ip_hash, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleProxyReadTimeout(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		seconds int
	}{
		{"seconds with s suffix", "120s", 120},
		{"seconds without suffix", "60", 60},
		{"minutes with m suffix", "2m", 120},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
				AnnotationProxyReadTimeout, tc.value,
			)

			err := provider.handleProxyReadTimeout(graph, gwCtx, routeCtx, nil, annotationCtx)
			if err != nil {
				t.Errorf("handleProxyReadTimeout returned error: %v", err)
			}

			// Check that RoutePolicy was created with timeout
			policyNN := types.NamespacedName{Name: "route-1-policy", Namespace: "default"}
			policy, exists := graph.RoutePolicies[policyNN]

			if !exists {
				t.Fatal("expected RoutePolicy to be created")
			}

			if policy.Spec.Default.RouteTimeouts == nil {
				t.Fatal("expected RouteTimeouts to be set")
			}
		})
	}
}

func TestNGINXHandleProxyConnectTimeout(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationProxyConnectTimeout, "30s",
	)

	err := provider.handleProxyConnectTimeout(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleProxyConnectTimeout returned error: %v", err)
	}

	// Connect timeout is not directly supported in AGC, should return warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered for partial support")
	}
}

func TestNGINXHandleProxySendTimeout(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationProxySendTimeout, "30s",
	)

	err := provider.handleProxySendTimeout(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleProxySendTimeout returned error: %v", err)
	}

	// Send timeout is not directly supported in AGC, should return warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered for partial support")
	}
}

func TestNGINXHandleProxyBodySize(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationProxyBodySize, "100m",
	)

	err := provider.handleProxyBodySize(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleProxyBodySize returned error: %v", err)
	}

	// Body size is not fully supported in AGC, should return warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered")
	}
}

func TestNGINXHandleProxyBuffering(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationProxyBuffering, "on",
	)

	err := provider.handleProxyBuffering(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleProxyBuffering returned error: %v", err)
	}

	// Buffering is not fully supported in AGC, should return warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered")
	}
}

func TestNGINXHandleProxyReadTimeoutInvalid(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationProxyReadTimeout, "invalid",
	)

	err := provider.handleProxyReadTimeout(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err == nil {
		t.Error("expected error for invalid timeout value")
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered for invalid value")
	}
}

func TestNGINXHandleUpstreamVhostEmpty(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationUpstreamVhost, "",
	)

	err := provider.handleUpstreamVhost(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleUpstreamVhost returned error: %v", err)
	}

	// Empty value should be ignored
	if annotationCtx.Status() != resources.MigrationStatusIgnored {
		t.Errorf("expected status Ignored for empty value, got %s", annotationCtx.Status())
	}
}
