package migration

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion/providers/agic"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion/providers/nginx"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name       string
		className  *string
		annotation string
		expected   string
	}{
		{
			name:      "nginx via ingressClassName",
			className: ptr.To("nginx"),
			expected:  ProviderNameNGINX,
		},
		{
			name:      "agic via ingressClassName",
			className: ptr.To("azure-application-gateway"),
			expected:  ProviderNameAGIC,
		},
		{
			name:       "nginx via annotation",
			annotation: "nginx",
			expected:   ProviderNameNGINX,
		},
		{
			name:       "agic via annotation",
			annotation: "azure/application-gateway",
			expected:   ProviderNameAGIC,
		},
		{
			name:     "no match returns empty",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ingress := testutil.MakeIngress("default", "test", testutil.IngressOptions{})
			if tc.className != nil {
				ingress.Spec.IngressClassName = tc.className
			}
			if tc.annotation != "" {
				ingress.Annotations["kubernetes.io/ingress.class"] = tc.annotation
			}

			inputs := resources.AGICResources{
				IngressContexts: map[types.NamespacedName]*resources.IngressContext{
					{Name: "test", Namespace: "default"}: resources.NewIngressContext(ingress),
				},
			}

			got := detectProvider(inputs)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestMakeProvider(t *testing.T) {
	inputs := resources.AGICResources{}

	t.Run("agic provider", func(t *testing.T) {
		provider, err := makeProvider("agic", inputs)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := provider.(agic.Provider); !ok {
			t.Fatal("expected agic.Provider")
		}
	})

	t.Run("nginx provider", func(t *testing.T) {
		provider, err := makeProvider("nginx", inputs)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := provider.(nginx.Provider); !ok {
			t.Fatal("expected nginx.Provider")
		}
	})

	t.Run("empty defaults to agic", func(t *testing.T) {
		provider, err := makeProvider("", inputs)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := provider.(agic.Provider); !ok {
			t.Fatal("expected agic.Provider")
		}
	})

	t.Run("unknown provider returns error", func(t *testing.T) {
		_, err := makeProvider("unknown", inputs)
		if err == nil {
			t.Fatal("expected error for unknown provider")
		}
	})
}
