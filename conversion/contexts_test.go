package conversion

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestNewHTTPRouteContext(t *testing.T) {
	route := ptr.To(testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{}))
	route.Spec.ParentRefs = []gatewayapi_v1.ParentReference{
		{
			Name:        "referenced",
			SectionName: ptr.To(gatewayapi_v1.SectionName("referenced-section")),
		},
		{
			Name: "unreferenced",
		},
	}

	got := NewHTTPRouteContext(route)
	if got.HTTPRoute != route {
		t.Fatalf("httpRouteContext has unexpected route")
	}

	expectSections := sets.New[gatewayapi_v1.SectionName]("referenced-section")
	if !got.namedParentSections.Equal(expectSections) {
		t.Fatalf("got sections: %v, expected: %v", got.namedParentSections, expectSections)
	}
}

func TestHTTPRouteContextEnsureHTTPSListener(t *testing.T) {
	route := testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})
	routeContext := NewHTTPRouteContext(&route)
	graph, gwContext, _ := ingressTestSetup()

	routeContext.EnsureHTTPSListener(&graph, *gwContext, "host-1", "secret-1", "secret-1-ns")

	t.Run("a parent ref is correctly added", func(t *testing.T) {
		if len(route.Spec.ParentRefs) == 0 {
			t.Fatal("no parent refs were added to the route")
		}

		ref := route.Spec.ParentRefs[len(route.Spec.ParentRefs)-1]
		foundSectionName := false

		for _, listener := range gwContext.HTTPSListeners {
			if listener.Name == *ref.SectionName {
				if listener.Hostname == nil || *listener.Hostname != "host-1" {
					t.Fatalf("expected gateway listener to have hostname host-1, got: %v", listener.Hostname)
				}

				foundSectionName = true

				break
			}
		}

		if !foundSectionName {
			t.Fatal("could not find listener on gateway context")
		}
	})

	t.Run("re-adding the same listener doesn't add a duplicate ref", func(t *testing.T) {
		listenerCount := len(routeContext.Spec.ParentRefs)
		routeContext.EnsureHTTPSListener(&graph, *gwContext, "host-1", "secret-1", "secret-1-ns")
		updatedListenerCount := len(gwContext.Spec.Listeners)

		if listenerCount != updatedListenerCount {
			t.Fatalf("exepcted %d listeners, got %d", listenerCount, updatedListenerCount)
		}
	})
}

func TestGatewayContextEnsureHTTPSListener(t *testing.T) {
	graph, gwContext, _ := ingressTestSetup()

	gwContext.EnsureHTTPSListener(&graph, "host-1", types.NamespacedName{Name: "secret-1", Namespace: "secret-1-ns"})

	key := ListenerKey{Hostname: "host-1", Port: 443}
	listener, ok := gwContext.HTTPSListeners[key]

	if !ok {
		t.Fatalf("could not find listener for key %+v", key)
	}

	if listener.Hostname == nil || *listener.Hostname != "host-1" {
		t.Fatal("expected gw to get listener for hostname: host-1")
	}

	if listener.TLS == nil {
		t.Fatal("expect listener to have TLS")
	}

	if len(listener.TLS.CertificateRefs) != 1 {
		t.Fatalf("expected listener certificate refs to have len 1, got: %d", len(listener.TLS.CertificateRefs))
	}

	certRef := listener.TLS.CertificateRefs[0]
	if certRef.Name != "secret-1" {
		t.Fatalf("expected listener certificate ref to have name \"secret-1\", got: %v", certRef.Namespace)
	}

	if certRef.Namespace == nil || *certRef.Namespace != "secret-1-ns" {
		t.Fatalf("expected listener certificate ref to have name \"secret-1\", got: %v", certRef.Namespace)
	}
}
