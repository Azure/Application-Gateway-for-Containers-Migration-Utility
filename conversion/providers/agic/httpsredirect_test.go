package agic

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestHandleSSLRedirect(t *testing.T) {
	t.Run("when the converter has already produced a TLS redirect, it is a no-op", func(t *testing.T) {
		provider := NewProvider(resources.AGICResources{})
		gwContext := &conversion.GatewayContext{
			HasHTTPSRedirect: true,
		}
		output := resources.NewAGCResourceGraph()
		err := provider.handleSSLRedirect(output, gwContext, &conversion.HTTPRouteContext{}, &resources.IngressContext{}, &resources.IngressAnnotationContext{})
		if err != nil {
			t.Fatalf("expected no error when TLS redirect is already produced, got: %v", err)
		}
	})

	t.Run("when there is no HTTP listener, it is created", func(t *testing.T) {
		provider := NewProvider(resources.AGICResources{})
		gwContext := conversion.NewGatewayContext(&gatewayapi_v1.Gateway{
			ObjectMeta: v1.ObjectMeta{
				Name:      "gw-1",
				Namespace: "gw-namespace",
			},
		})

		output := resources.NewAGCResourceGraph()
		annoCtx := resources.NewIngressAnnotationContext(AnnotationSSLRedirect, "true")

		err := provider.handleSSLRedirect(output, gwContext, &conversion.HTTPRouteContext{}, &resources.IngressContext{}, annoCtx)
		if err != nil {
			t.Fatalf("expected no error when creating HTTP listener, got: %v", err)
		}

		if gwContext.HTTPListener == nil {
			t.Fatalf("expected HTTP listener to be created, but it was not found")
		}

		// check there is an HTTPRoute with a redirect filter
		route, ok := output.HTTPRoutes[types.NamespacedName{Name: httpRouteRedirectname, Namespace: "gw-namespace"}]
		if !ok {
			t.Fatalf("expected HTTPRoute for TLS redirect to be created, but it was not found")
		}

		if route.Spec.Rules[0].Filters[0].Type != gatewayapi_v1.HTTPRouteFilterType(gatewayapi_v1.HTTPRouteFilterRequestRedirect) {
			t.Fatalf("expected HTTPRoute to have a RequestRedirect filter, but it did not")
		}
		if *route.Spec.Rules[0].Filters[0].RequestRedirect.Scheme != "https" {
			t.Fatalf("expected RequestRedirect filter to have HTTPS scheme, but got: %q", *route.Spec.Rules[0].Filters[0].RequestRedirect.Scheme)
		}
	})
}
