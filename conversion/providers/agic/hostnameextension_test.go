package agic

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	networking_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestHandleHostnameExtension(t *testing.T) {
	converter, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs("hostname-extension", "host-1, host-2, host-3, host-4")
	ingressCtx.Ingress.Spec.TLS = []networking_v1.IngressTLS{
		{
			Hosts:      []string{"host-1", "host-2"},
			SecretName: "secret-1",
		},
		{
			Hosts:      []string{"host-3"},
			SecretName: "secret-2",
		},
	}

	if err := converter.handleHostNameExtension(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
		t.Fatalf("got err: %v", err)
	}

	expectHTTPSListeners := []struct {
		hostname string
		secretNN types.NamespacedName
	}{
		{
			hostname: "host-1",
			secretNN: types.NamespacedName{
				Namespace: ingressCtx.Ingress.Namespace,
				Name:      "secret-1",
			},
		},
		{
			hostname: "host-2",
			secretNN: types.NamespacedName{
				Namespace: ingressCtx.Ingress.Namespace,
				Name:      "secret-1",
			},
		},
		{
			hostname: "host-3",
			secretNN: types.NamespacedName{
				Namespace: ingressCtx.Ingress.Namespace,
				Name:      "secret-2",
			},
		},
	}

	for _, expect := range expectHTTPSListeners {
		found := false
		for _, listener := range gwCtx.HTTPSListeners {
			if listener.Hostname == nil {
				continue
			}

			if *listener.Hostname == gatewayapi_v1.Hostname(expect.hostname) {
				if !listener.Secrets.Has(expect.secretNN) {
					t.Fatalf("listener for %q does not secret %q", expect.hostname, expect.secretNN)
				}
				found = true
			}
		}

		if !found {
			t.Fatalf("could not find HTTPS listener for hostname %q", expect.hostname)
		}
	}

	extendedHostnames := []gatewayapi_v1.Hostname{"host-1", "host-2", "host-3", "host-4"}
	if !sets.New(routeCtx.Spec.Hostnames...).HasAll(extendedHostnames...) {
		t.Fatalf("route hostnames: %v missing all expected hostnames: %v", routeCtx.Spec.Hostnames, extendedHostnames)
	}

	if annoCtx.Status() != resources.MigrationStatusCompleted {
		t.Fatalf("Expected status: %q, got: %q", resources.MigrationStatusCompleted, annoCtx.Status())
	}
}
