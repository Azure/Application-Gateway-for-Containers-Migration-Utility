package conversion

import (
	"testing"

	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestNewBackendGroup(t *testing.T) {
	agicResources := resources.NewAGICResources(nil, nil)
	svc := core_v1.Service{
		Spec: core_v1.ServiceSpec{
			Ports: []core_v1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
				{
					Name: "https",
					Port: 443,
				},
			},
		},
	}
	agicResources.Services = map[types.NamespacedName]*resources.ServiceContext{
		{Namespace: "default", Name: "service-1"}: {
			Service: svc,
		},
		{Namespace: "default", Name: "service-2"}: {
			Service: svc,
		},
	}

	t.Run("A group with a single backend", func(t *testing.T) {
		// conv := newConverter(agicResources, nil)
		routeCtx := &HTTPRouteContext{
			HTTPRoute: ptr.To(testutil.MakeHTTPRoute("default", "ingress-1", testutil.HTTPRouteOptions{})),
		}
		group := NewBackendGroup(routeCtx)

		if len(group.services) != 1 {
			t.Fatalf("expected 1 service in backend group, got %d", len(group.services))
		}

		expectedNN := types.NamespacedName{Namespace: "default", Name: "service-1"}
		svcDetails, ok := group.services[expectedNN]

		if !ok {
			t.Fatalf("expected service %v in backend group, not found", expectedNN)
		}

		if svcDetails.Port == nil || *svcDetails.Port != 80 {
			t.Fatalf("expected port 80 for service %v, got %v", expectedNN, svcDetails.Port)
		}
	})

	t.Run("A group with multiple backends", func(t *testing.T) {
		// conv := newConverter(agicResources)
		// Just a note: I don't think we'll ever have a HTTPRoute with a backend in a different namespace,
		httpRouteContext := &HTTPRouteContext{
			HTTPRoute: ptr.To(testutil.MakeHTTPRoute("default", "ingress-1", testutil.HTTPRouteOptions{
				Paths: []testutil.HTTPRoutePathOption{
					{
						Path: "/path-1",
						Backends: []testutil.HTTPRoutePathBackendOption{
							{
								BackendName: "service-1",
							},
						},
					},
					{
						Path: "/path-2",
						Backends: []testutil.HTTPRoutePathBackendOption{
							{
								BackendName:      "service-2",
								BackendNamespace: "service-2-ns",
								BackendPort:      8080,
							},
							{
								BackendName:      "service-3",
								BackendNamespace: "service-3-ns",
								BackendPort:      8081,
							},
						},
					},
				},
			})),
		}

		group := NewBackendGroup(httpRouteContext)
		if len(group.services) != 3 {
			t.Fatalf("expected 3 services in backend group, got %d", len(group.services))
		}

		expectedNN1 := types.NamespacedName{Namespace: "default", Name: "service-1"}
		svcDetails1, ok := group.services[expectedNN1]

		if !ok {
			t.Fatalf("expected service %v in backend group, not found", expectedNN1)
		}

		if svcDetails1.Port == nil || *svcDetails1.Port != 80 {
			t.Fatalf("expected port 80 for service %v, got %v", expectedNN1, svcDetails1.Port)
		}

		expectedNN2 := types.NamespacedName{Namespace: "service-2-ns", Name: "service-2"}
		svcDetails2, ok := group.services[expectedNN2]

		if !ok {
			t.Fatalf("expected service %v in backend group, not found", expectedNN2)
		}

		if svcDetails2.Port == nil || *svcDetails2.Port != 8080 {
			t.Fatalf("expected port 8080 for service %v, got %v", expectedNN2, svcDetails2.Port)
		}

		expectedNN3 := types.NamespacedName{Namespace: "service-3-ns", Name: "service-3"}
		svcDetails3, ok := group.services[expectedNN3]

		if !ok {
			t.Fatalf("expected service %v in backend group, not found", expectedNN3)
		}

		if svcDetails3.Port == nil || *svcDetails3.Port != 8081 {
			t.Fatalf("expected port 8081 for service %v, got %v", expectedNN3, svcDetails3.Port)
		}
	})
}
