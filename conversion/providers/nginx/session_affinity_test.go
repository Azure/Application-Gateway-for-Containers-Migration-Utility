package nginx

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	crds_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestNGINXHandleAffinity(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAffinity, "cookie",
	)

	err := provider.handleAffinity(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAffinity returned error: %v", err)
	}

	// Check that a RoutePolicy was created
	policyNN := types.NamespacedName{Name: "route-1-policy", Namespace: "default"}
	policy, exists := graph.RoutePolicies[policyNN]

	if !exists {
		t.Fatal("expected RoutePolicy to be created")
	}

	if policy.Spec.Default.SessionAffinity == nil {
		t.Fatal("expected SessionAffinity to be set")
	}

	if policy.Spec.Default.SessionAffinity.AffinityType != crds_v1.AffinityTypeManagedCookie {
		t.Errorf("expected AffinityType managed-cookie, got %s", policy.Spec.Default.SessionAffinity.AffinityType)
	}
}

func TestNGINXHandleAffinityNotCookie(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAffinity, "ip_hash",
	)

	err := provider.handleAffinity(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAffinity returned error: %v", err)
	}

	// Should be not supported
	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleAuthTLSSecret(t *testing.T) {
	t.Run("with HTTPS listener creates FrontendTLSPolicy", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSTestInputs(
			AnnotationAuthTLSSecret, "my-namespace/my-ca-secret",
		)

		err := provider.handleAuthTLSSecret(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleAuthTLSSecret returned error: %v", err)
		}

		if len(graph.FrontendTLSPolicies) != 1 {
			t.Fatalf("expected 1 FrontendTLSPolicy, got %d", len(graph.FrontendTLSPolicies))
		}

		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.Verify == nil {
				t.Fatal("expected Verify to be set")
			}

			if policy.Spec.Default.Verify.CaCertificateRef == nil {
				t.Fatal("expected CaCertificateRef to be set")
			}

			if string(policy.Spec.Default.Verify.CaCertificateRef.Name) != "my-ca-secret" {
				t.Errorf("expected secret name 'my-ca-secret', got %s", policy.Spec.Default.Verify.CaCertificateRef.Name)
			}

			if policy.Spec.Default.Verify.CaCertificateRef.Namespace == nil || string(*policy.Spec.Default.Verify.CaCertificateRef.Namespace) != "my-namespace" {
				t.Error("expected namespace 'my-namespace'")
			}
		}
	})

	t.Run("without namespace uses ingress namespace", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSTestInputs(
			AnnotationAuthTLSSecret, "simple-secret",
		)

		err := provider.handleAuthTLSSecret(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleAuthTLSSecret returned error: %v", err)
		}

		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.Verify.CaCertificateRef.Namespace == nil {
				t.Fatal("expected namespace to be set")
			}

			if string(*policy.Spec.Default.Verify.CaCertificateRef.Namespace) != ingressCtx.Ingress.Namespace {
				t.Errorf("expected namespace %s, got %s", ingressCtx.Ingress.Namespace, *policy.Spec.Default.Verify.CaCertificateRef.Namespace)
			}
		}
	})

	t.Run("without HTTPS listener registers issue", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationAuthTLSSecret, "my-secret",
		)
		gwCtx.HTTPSListeners = nil

		err := provider.handleAuthTLSSecret(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleAuthTLSSecret returned error: %v", err)
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issue to be registered")
		}

		if annotationCtx.Issues[0].Code != resources.IssueNoHTTPSListenersForSSLProfile {
			t.Errorf("expected issue code %d, got %d", resources.IssueNoHTTPSListenersForSSLProfile, annotationCtx.Issues[0].Code)
		}
	})
}

func TestNGINXHandleAuthTLSVerifyClient(t *testing.T) {
	t.Run("on value creates FrontendTLSPolicy with Verify", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSTestInputs(
			AnnotationAuthTLSVerifyClient, "on",
		)

		err := provider.handleAuthTLSVerifyClient(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleAuthTLSVerifyClient returned error: %v", err)
		}

		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.Verify == nil {
				t.Fatal("expected Verify to be set for 'on' value")
			}
		}
	})

	t.Run("optional value creates FrontendTLSPolicy with Verify", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSTestInputs(
			AnnotationAuthTLSVerifyClient, "optional",
		)

		err := provider.handleAuthTLSVerifyClient(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleAuthTLSVerifyClient returned error: %v", err)
		}

		for _, policy := range graph.FrontendTLSPolicies {
			if policy.Spec.Default.Verify == nil {
				t.Fatal("expected Verify to be set for 'optional' value")
			}
		}
	})

	t.Run("off value completes without creating Verify", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSTestInputs(
			AnnotationAuthTLSVerifyClient, "off",
		)

		err := provider.handleAuthTLSVerifyClient(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleAuthTLSVerifyClient returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusCompleted {
			t.Errorf("expected status Completed for 'off' value, got %s", annotationCtx.Status())
		}
	})

	t.Run("optional_no_ca registers issue", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSTestInputs(
			AnnotationAuthTLSVerifyClient, "optional_no_ca",
		)

		err := provider.handleAuthTLSVerifyClient(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleAuthTLSVerifyClient returned error: %v", err)
		}

		// The issue has level Warning, so status ends up as Warning
		if annotationCtx.Status() != resources.MigrationStatusWarning {
			t.Errorf("expected status Warning for 'optional_no_ca', got %s", annotationCtx.Status())
		}

		if len(annotationCtx.Issues) == 0 {
			t.Error("expected issue to be registered")
		}
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupFrontendTLSTestInputs(
			AnnotationAuthTLSVerifyClient, "invalid-value",
		)

		err := provider.handleAuthTLSVerifyClient(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err == nil {
			t.Error("expected error for invalid value")
		}
	})
}

func TestNGINXHandleAuthTLSVerifyDepth(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAuthTLSVerifyDepth, "3",
	)

	err := provider.handleAuthTLSVerifyDepth(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAuthTLSVerifyDepth returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered")
	}
}

func TestNGINXHandleAuthTLSErrorPage(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAuthTLSErrorPage, "/error-page",
	)

	err := provider.handleAuthTLSErrorPage(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAuthTLSErrorPage returned error: %v", err)
	}

	// The issue level is Warning, so final status is Warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered")
	}
}

func TestNGINXHandleAuthTLSPassCertToUpstream(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAuthTLSPassCertToUpstream, "true",
	)

	err := provider.handleAuthTLSPassCertToUpstream(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAuthTLSPassCertToUpstream returned error: %v", err)
	}

	// The issue level is Warning, so final status is Warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered")
	}
}

func TestNGINXHandleAffinityMode(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAffinityMode, "balanced",
	)

	err := provider.handleAffinityMode(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAffinityMode returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered")
	}

	if annotationCtx.Issues[0].Code != resources.IssueNGINXAffinityModeNotSupported {
		t.Errorf("expected issue code %d, got %d", resources.IssueNGINXAffinityModeNotSupported, annotationCtx.Issues[0].Code)
	}
}

func TestNGINXHandleAffinityCanaryBehavior(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAffinityCanaryBehavior, "sticky",
	)

	err := provider.handleAffinityCanaryBehavior(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleAffinityCanaryBehavior returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleSessionCookieName(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSessionCookieName, "my-custom-cookie",
	)

	err := provider.handleSessionCookieName(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleSessionCookieName returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleSessionCookiePath(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSessionCookiePath, "/app",
	)

	err := provider.handleSessionCookiePath(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleSessionCookiePath returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleSessionCookieExpires(t *testing.T) {
	t.Run("valid value sets cookie duration", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationSessionCookieExpires, "3600",
		)

		err := provider.handleSessionCookieExpires(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleSessionCookieExpires returned error: %v", err)
		}

		policyNN := types.NamespacedName{Name: "route-1-policy", Namespace: "default"}
		policy, exists := graph.RoutePolicies[policyNN]

		if !exists {
			t.Fatal("expected RoutePolicy to be created")
		}

		if policy.Spec.Default.SessionAffinity == nil {
			t.Fatal("expected SessionAffinity to be set")
		}

		expectedDuration := time.Second * 3600
		if policy.Spec.Default.SessionAffinity.CookieDuration.Duration != expectedDuration {
			t.Errorf("expected cookie duration %v, got %v", expectedDuration, policy.Spec.Default.SessionAffinity.CookieDuration.Duration)
		}
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationSessionCookieExpires, "not-a-number",
		)

		err := provider.handleSessionCookieExpires(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err == nil {
			t.Error("expected error for invalid value")
		}
	})
}

func TestNGINXHandleSessionCookieMaxAge(t *testing.T) {
	// Should behave the same as expires
	provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSessionCookieMaxAge, "7200",
	)

	err := provider.handleSessionCookieMaxAge(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
	if err != nil {
		t.Errorf("handleSessionCookieMaxAge returned error: %v", err)
	}

	policyNN := types.NamespacedName{Name: "route-1-policy", Namespace: "default"}
	policy, exists := graph.RoutePolicies[policyNN]

	if !exists {
		t.Fatal("expected RoutePolicy to be created")
	}

	expectedDuration := time.Second * 7200
	if policy.Spec.Default.SessionAffinity.CookieDuration.Duration != expectedDuration {
		t.Errorf("expected cookie duration %v, got %v", expectedDuration, policy.Spec.Default.SessionAffinity.CookieDuration.Duration)
	}
}

func TestNGINXHandleSessionCookieSameSite(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationSessionCookieSameSite, "Strict",
	)

	err := provider.handleSessionCookieSameSite(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleSessionCookieSameSite returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}
}

func TestNGINXGetOrCreateRoutePolicy(t *testing.T) {
	provider, graph, _, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationAffinity, "cookie",
	)

	policy, err := provider.GetOrCreateRoutePolicy(graph, routeCtx.HTTPRoute, annotationCtx)
	if err != nil {
		t.Errorf("GetOrCreateRoutePolicy returned error: %v", err)
	}

	if policy == nil {
		t.Fatal("expected policy to be created")
	}

	// Call again to verify idempotent behavior
	policy2, err := provider.GetOrCreateRoutePolicy(graph, routeCtx.HTTPRoute, annotationCtx)
	if err != nil {
		t.Errorf("GetOrCreateRoutePolicy returned error: %v", err)
	}

	if policy != policy2 {
		t.Error("expected same policy to be returned on second call")
	}
}

// Helper function for FrontendTLSPolicy tests with HTTPS listeners
func setupFrontendTLSTestInputs(key, val string) (
	Provider,
	resources.AGCResourceGraph,
	*conversion.GatewayContext,
	*conversion.HTTPRouteContext,
	*resources.IngressContext,
	*resources.IngressAnnotationContext,
) {
	provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(key, val)

	// Add HTTPS listener to gateway context
	gwCtx.HTTPSListeners = map[conversion.ListenerKey]conversion.HTTPSListener{
		{Hostname: "*", Port: 443}: {Listener: &gatewayapi_v1.Listener{Name: "https"}},
	}

	// Update route to reference HTTPS listener
	routeCtx.HTTPRoute.Spec.ParentRefs = []gatewayapi_v1.ParentReference{
		{
			Name:        gatewayapi_v1.ObjectName(gwCtx.Gateway.Name),
			SectionName: ptr.To(gatewayapi_v1.SectionName("https")),
		},
	}

	return provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx
}
