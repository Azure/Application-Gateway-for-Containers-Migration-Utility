package nginx

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestProviderGetAnnotationHandlers(t *testing.T) {
	provider := NewProvider(resources.AGICResources{})
	handlers := provider.GetAnnotationHandlers()

	// Verify key handlers exist
	expectedHandlers := []string{
		AnnotationRewriteTarget,
		AnnotationAffinity,
		AnnotationSSLRedirect,
		AnnotationBackendProtocol,
		AnnotationEnableModSecurity,
		AnnotationProxyReadTimeout,
		AnnotationLoadBalance,
		AnnotationAppRoot,
		AnnotationUpstreamVhost,
		AnnotationPermanentRedirect,
		AnnotationXForwardedPrefix,
		AnnotationServerAlias,
	}

	for _, key := range expectedHandlers {
		if _, ok := handlers[key]; !ok {
			t.Errorf("expected handler for %s to exist", key)
		}
	}
}

func TestNGINXHandleConfigurationSnippet(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationConfigurationSnippet, "proxy_set_header X-Custom 'test';",
	)

	err := provider.handleConfigurationSnippet(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleConfigurationSnippet returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered")
	}
}

func TestNGINXHandleServerAlias(t *testing.T) {
	tests := []struct {
		name            string
		aliasValue      string
		expectedAliases []string
	}{
		{
			name:            "single alias",
			aliasValue:      "alias.example.com",
			expectedAliases: []string{"alias.example.com"},
		},
		{
			name:            "comma separated aliases",
			aliasValue:      "alias1.example.com,alias2.example.com",
			expectedAliases: []string{"alias1.example.com", "alias2.example.com"},
		},
		{
			name:            "space separated aliases",
			aliasValue:      "alias1.example.com alias2.example.com alias3.example.com",
			expectedAliases: []string{"alias1.example.com", "alias2.example.com", "alias3.example.com"},
		},
		{
			name:            "mixed separators",
			aliasValue:      "alias1.example.com, alias2.example.com alias3.example.com",
			expectedAliases: []string{"alias1.example.com", "alias2.example.com", "alias3.example.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
				AnnotationServerAlias, tc.aliasValue,
			)

			err := provider.handleServerAlias(graph, gwCtx, routeCtx, nil, annotationCtx)
			if err != nil {
				t.Errorf("handleServerAlias returned error: %v", err)
			}

			// Check that the aliases were added to the HTTPRoute hostnames
			route := routeCtx.HTTPRoute

			for _, expectedAlias := range tc.expectedAliases {
				found := false

				for _, hostname := range route.Spec.Hostnames {
					if string(hostname) == expectedAlias {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("expected alias %s to be in hostnames, got %v", expectedAlias, route.Spec.Hostnames)
				}
			}

			if annotationCtx.Status() != resources.MigrationStatusCompleted {
				t.Errorf("expected status Completed, got %s", annotationCtx.Status())
			}
		})
	}
}

func TestNGINXHandleDefaultBackend(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationDefaultBackend, "default-service",
	)

	err := provider.handleDefaultBackend(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleDefaultBackend returned error: %v", err)
	}

	// Default backend is not supported and should register an issue
	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered")
	}

	if annotationCtx.Issues[0].Code != resources.IssueNGINXDefaultBackendNotSupported {
		t.Errorf("expected issue code %d, got %d", resources.IssueNGINXDefaultBackendNotSupported, annotationCtx.Issues[0].Code)
	}
}

func TestNGINXHandleServerSnippet(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationServerSnippet, "location /health { return 200; }",
	)

	err := provider.handleServerSnippet(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleServerSnippet returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered")
	}

	if annotationCtx.Issues[0].Code != resources.IssueNGINXServerSnippetNotSupported {
		t.Errorf("expected issue code %d, got %d", resources.IssueNGINXServerSnippetNotSupported, annotationCtx.Issues[0].Code)
	}
}

func TestNGINXHandleConnectionProxyHeader(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationConnectionProxyHeader, "keep-alive",
	)

	err := provider.handleConnectionProxyHeader(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleConnectionProxyHeader returned error: %v", err)
	}

	// Check that a request header modifier filter was added
	route := routeCtx.HTTPRoute
	if len(route.Spec.Rules) == 0 || len(route.Spec.Rules[0].Filters) == 0 {
		t.Fatal("expected filter to be added")
	}

	filter := route.Spec.Rules[0].Filters[0]
	if filter.RequestHeaderModifier == nil {
		t.Fatal("expected RequestHeaderModifier to be set")
	}

	// Check that the Connection header is set
	found := false

	for _, header := range filter.RequestHeaderModifier.Set {
		if string(header.Name) == "Connection" && header.Value == "keep-alive" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected Connection header to be set to 'keep-alive'")
	}

	if annotationCtx.Status() != resources.MigrationStatusCompleted {
		t.Errorf("expected status Completed, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleIgnoredAnnotation(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		"nginx.ingress.kubernetes.io/some-ignored-annotation", "some-value",
	)

	err := provider.handleIgnoredAnnotation(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleIgnoredAnnotation returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusIgnored {
		t.Errorf("expected status Ignored, got %s", annotationCtx.Status())
	}
}
