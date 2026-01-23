package nginx

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestNGINXHandlePermanentRedirect(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationPermanentRedirect, "https://example.com/new-path",
	)

	err := provider.handlePermanentRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handlePermanentRedirect returned error: %v", err)
	}

	// Check that redirect filter was added
	route := routeCtx.HTTPRoute
	if len(route.Spec.Rules) == 0 || len(route.Spec.Rules[0].Filters) == 0 {
		t.Fatal("expected filters to be added to the route")
	}

	filter := route.Spec.Rules[0].Filters[0]
	if filter.Type != gatewayapi_v1.HTTPRouteFilterRequestRedirect {
		t.Errorf("expected filter type RequestRedirect, got %s", filter.Type)
	}

	if filter.RequestRedirect == nil {
		t.Fatal("expected RequestRedirect to be set")
	}

	if filter.RequestRedirect.StatusCode == nil || *filter.RequestRedirect.StatusCode != 301 {
		t.Error("expected status code 301")
	}
}

func TestNGINXHandlePermanentRedirectEmpty(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationPermanentRedirect, "",
	)

	err := provider.handlePermanentRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handlePermanentRedirect returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusIgnored {
		t.Errorf("expected status Ignored for empty value, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandlePermanentRedirectInvalidURL(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationPermanentRedirect, "://invalid-url",
	)

	err := provider.handlePermanentRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handlePermanentRedirect returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported for invalid URL, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered for invalid URL")
	}
}

func TestNGINXHandlePermanentRedirectWithPort(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationPermanentRedirect, "https://example.com:8443/path",
	)

	err := provider.handlePermanentRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handlePermanentRedirect returned error: %v", err)
	}

	route := routeCtx.HTTPRoute
	filter := route.Spec.Rules[0].Filters[0]

	if filter.RequestRedirect.Port == nil {
		t.Error("expected port to be set")
	} else if *filter.RequestRedirect.Port != 8443 {
		t.Errorf("expected port 8443, got %d", *filter.RequestRedirect.Port)
	}
}

func TestNGINXHandlePermanentRedirectCode(t *testing.T) {
	t.Run("valid code 301", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationPermanentRedirectCode, "301",
		)

		err := provider.handlePermanentRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handlePermanentRedirectCode returned error: %v", err)
		}

		route := routeCtx.HTTPRoute
		if len(route.Spec.Rules[0].Filters) == 0 {
			t.Fatal("expected filter to be added")
		}

		filter := route.Spec.Rules[0].Filters[0]
		if filter.RequestRedirect == nil || filter.RequestRedirect.StatusCode == nil {
			t.Fatal("expected status code to be set")
		}

		if *filter.RequestRedirect.StatusCode != 301 {
			t.Errorf("expected status code 301, got %d", *filter.RequestRedirect.StatusCode)
		}
	})

	t.Run("valid code 308", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationPermanentRedirectCode, "308",
		)

		err := provider.handlePermanentRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handlePermanentRedirectCode returned error: %v", err)
		}

		filter := routeCtx.HTTPRoute.Spec.Rules[0].Filters[0]
		if *filter.RequestRedirect.StatusCode != 308 {
			t.Errorf("expected status code 308, got %d", *filter.RequestRedirect.StatusCode)
		}
	})

	t.Run("invalid code returns error", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationPermanentRedirectCode, "302",
		)

		err := provider.handlePermanentRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handlePermanentRedirectCode returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusNotSupported {
			t.Errorf("expected status NotSupported for invalid code, got %s", annotationCtx.Status())
		}
	})

	t.Run("empty value is ignored", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationPermanentRedirectCode, "",
		)

		err := provider.handlePermanentRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handlePermanentRedirectCode returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusIgnored {
			t.Errorf("expected status Ignored, got %s", annotationCtx.Status())
		}
	})
}

func TestNGINXHandleTemporalRedirect(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationTemporalRedirect, "https://example.com/temp-path",
	)

	err := provider.handleTemporalRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleTemporalRedirect returned error: %v", err)
	}

	// Check that redirect filter was added
	route := routeCtx.HTTPRoute
	if len(route.Spec.Rules) == 0 || len(route.Spec.Rules[0].Filters) == 0 {
		t.Fatal("expected filters to be added to the route")
	}

	filter := route.Spec.Rules[0].Filters[0]
	if filter.Type != gatewayapi_v1.HTTPRouteFilterRequestRedirect {
		t.Errorf("expected filter type RequestRedirect, got %s", filter.Type)
	}

	if filter.RequestRedirect == nil {
		t.Fatal("expected RequestRedirect to be set")
	}

	if filter.RequestRedirect.StatusCode == nil || *filter.RequestRedirect.StatusCode != 302 {
		t.Error("expected status code 302")
	}
}

func TestNGINXHandleTemporalRedirectEmpty(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationTemporalRedirect, "",
	)

	err := provider.handleTemporalRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleTemporalRedirect returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusIgnored {
		t.Errorf("expected status Ignored for empty value, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleTemporalRedirectInvalidURL(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationTemporalRedirect, "://invalid-url",
	)

	err := provider.handleTemporalRedirect(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleTemporalRedirect returned error: %v", err)
	}

	if annotationCtx.Status() != resources.MigrationStatusNotSupported {
		t.Errorf("expected status NotSupported for invalid URL, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleTemporalRedirectCode(t *testing.T) {
	t.Run("valid code 302", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationTemporalRedirectCode, "302",
		)

		err := provider.handleTemporalRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleTemporalRedirectCode returned error: %v", err)
		}

		filter := routeCtx.HTTPRoute.Spec.Rules[0].Filters[0]
		if *filter.RequestRedirect.StatusCode != 302 {
			t.Errorf("expected status code 302, got %d", *filter.RequestRedirect.StatusCode)
		}
	})

	t.Run("valid code 303", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationTemporalRedirectCode, "303",
		)

		err := provider.handleTemporalRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleTemporalRedirectCode returned error: %v", err)
		}

		filter := routeCtx.HTTPRoute.Spec.Rules[0].Filters[0]
		if *filter.RequestRedirect.StatusCode != 303 {
			t.Errorf("expected status code 303, got %d", *filter.RequestRedirect.StatusCode)
		}
	})

	t.Run("valid code 307", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationTemporalRedirectCode, "307",
		)

		err := provider.handleTemporalRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleTemporalRedirectCode returned error: %v", err)
		}

		filter := routeCtx.HTTPRoute.Spec.Rules[0].Filters[0]
		if *filter.RequestRedirect.StatusCode != 307 {
			t.Errorf("expected status code 307, got %d", *filter.RequestRedirect.StatusCode)
		}
	})

	t.Run("invalid code 301 for temporal", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationTemporalRedirectCode, "301",
		)

		err := provider.handleTemporalRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleTemporalRedirectCode returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusNotSupported {
			t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
		}
	})

	t.Run("empty value is ignored", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationTemporalRedirectCode, "",
		)

		err := provider.handleTemporalRedirectCode(graph, gwCtx, routeCtx, nil, annotationCtx)
		if err != nil {
			t.Errorf("handleTemporalRedirectCode returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusIgnored {
			t.Errorf("expected status Ignored, got %s", annotationCtx.Status())
		}
	})
}

func TestNGINXHandleFromToWWWRedirect(t *testing.T) {
	t.Run("enabled with host creates redirect", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationFromToWWWRedirect, "true",
		)

		// Create ingress with host
		ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{
			Host: "example.com",
		}))

		err := provider.handleFromToWWWRedirect(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleFromToWWWRedirect returned error: %v", err)
		}

		route := routeCtx.HTTPRoute
		if len(route.Spec.Rules[0].Filters) == 0 {
			t.Fatal("expected filter to be added")
		}

		filter := route.Spec.Rules[0].Filters[0]
		if filter.RequestRedirect == nil {
			t.Fatal("expected redirect filter")
		}

		if filter.RequestRedirect.Hostname == nil || string(*filter.RequestRedirect.Hostname) != "www.example.com" {
			if filter.RequestRedirect.Hostname != nil {
				t.Errorf("expected hostname www.example.com, got %s", *filter.RequestRedirect.Hostname)
			} else {
				t.Error("expected hostname to be set")
			}
		}
	})

	t.Run("false value is ignored", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationFromToWWWRedirect, "false",
		)

		err := provider.handleFromToWWWRedirect(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleFromToWWWRedirect returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusIgnored {
			t.Errorf("expected status Ignored, got %s", annotationCtx.Status())
		}
	})

	t.Run("no host registers issue", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationFromToWWWRedirect, "true",
		)

		// Create ingress without host
		ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{}))

		err := provider.handleFromToWWWRedirect(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleFromToWWWRedirect returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusNotSupported {
			t.Errorf("expected status NotSupported, got %s", annotationCtx.Status())
		}
	})

	t.Run("host already has www is ignored", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
			AnnotationFromToWWWRedirect, "true",
		)

		ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{
			Host: "www.example.com",
		}))

		err := provider.handleFromToWWWRedirect(graph, gwCtx, routeCtx, ingressCtx, annotationCtx)
		if err != nil {
			t.Errorf("handleFromToWWWRedirect returned error: %v", err)
		}

		if annotationCtx.Status() != resources.MigrationStatusIgnored {
			t.Errorf("expected status Ignored for www host, got %s", annotationCtx.Status())
		}
	})
}
