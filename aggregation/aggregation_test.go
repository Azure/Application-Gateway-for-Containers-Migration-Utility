package aggregation

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	network_v1 "k8s.io/api/networking/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/logging"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"

	"sigs.k8s.io/controller-runtime/pkg/client"
	fake_ctrl "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestAggregateFromCluster(t *testing.T) {
	aggregator := makeAggregator()
	className := DefaultIngressClassName

	scheme := runtime.NewScheme()
	if err := fake.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	if err := appgwrewrite.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	fakeK8s := fake.NewClientset()
	fakeCtrlClient := fake_ctrl.NewClientBuilder().WithScheme(scheme).Build()

	t.Run("error loading AGIC config", func(t *testing.T) {
		_, err := aggregator.AggregateFromCluster(t.Context(), fakeK8s, fakeCtrlClient, "agic")
		if err == nil || !strings.Contains(err.Error(), "failed to load & parse AGIC ConfigMap") {
			t.Fatalf("expected an error loading AGIC config, got %v", err)
		}
	})

	fakeK8s.PrependReactor("list", "configmaps", func(k8s_testing.Action) (bool, runtime.Object, error) {
		return true, makeAGICConfigMapList(className, nil), nil
	})

	t.Run("ingress2gateway returns an error", func(t *testing.T) {
		ctrlClient := fake_ctrl.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
				if _, ok := list.(*network_v1.IngressList); ok {
					return fmt.Errorf("test failure")
				}
				return nil
			},
		}).Build()

		_, err := aggregator.AggregateFromCluster(t.Context(), fakeK8s, ctrlClient, "agic")
		if err == nil || !strings.Contains(err.Error(), "failed to collect Ingresses from cluster") {
			t.Fatalf("expected an error reading ingresses, got %v", err)
		}
	})

	t.Run("no ingresses", func(t *testing.T) {
		got, err := aggregator.AggregateFromCluster(t.Context(), fakeK8s, fakeCtrlClient, "agic")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(got.IngressContexts) != 0 {
			t.Fatalf("expected no ingresses, got %d", len(got.IngressContexts))
		}
	})

	t.Run("wrong ingress class name", func(t *testing.T) {
		ctrlClient := fake_ctrl.NewClientBuilder().WithScheme(scheme).WithLists(testutil.MakeIngressList("wrong-classname", 1)).Build()
		got, err := aggregator.AggregateFromCluster(t.Context(), fakeK8s, ctrlClient, "agic")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(got.IngressContexts) != 0 {
			t.Fatalf("expected 0 ingress, got %d", len(got.IngressContexts))
		}
	})

	// I'm not attempting to test the ingress2gateway library here, only the code here
	t.Run("2 ingresses", func(t *testing.T) {
		ctrlClient := fake_ctrl.NewClientBuilder().WithScheme(scheme).WithLists(testutil.MakeIngressList(className, 2)).Build()
		got, err := aggregator.AggregateFromCluster(t.Context(), fakeK8s, ctrlClient, "agic")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(got.IngressContexts) != 2 {
			t.Fatalf("expected 2 ingress, got %d", len(got.IngressContexts))
		}
	})

	t.Run("Fails to list AppGW Rewrite CRD", func(t *testing.T) {
		ctrlClient := fake_ctrl.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, list client.ObjectList, _ ...client.ListOption) error {
				if _, ok := list.(*appgwrewrite.AzureApplicationGatewayRewriteList); ok {
					return fmt.Errorf("test failure")
				}
				return nil
			},
		}).Build()

		_, err := aggregator.AggregateFromCluster(t.Context(), fakeK8s, ctrlClient, "agic")
		if err == nil || !strings.Contains(err.Error(), "failed to collect AzureApplicationGatewayRewrite from cluster") {
			t.Fatalf("expected an error reading AppGW Rewrite CRD, got %v", err)
		}
	})

	t.Run("2 AppGW Rewrite CRD", func(t *testing.T) {
		client := fake_ctrl.NewClientBuilder().WithScheme(scheme).WithLists(testutil.MakeAzureApplicationGatewayRewriteList(className, 2)).Build()
		got, err := aggregator.AggregateFromCluster(t.Context(), fakeK8s, client, "agic")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(got.AppGWRewrites) != 2 {
			t.Fatalf("expected 2 AppGW Rewrites, got %d", len(got.AppGWRewrites))
		}
	})
}

func TestFindClusterLabel(t *testing.T) {
	labels := []string{
		agicAddonLabel,
		agicHelmLabel,
	}
	for _, label := range labels {
		t.Run(fmt.Sprintf("using label %q", label), func(t *testing.T) {
			split := strings.Split(label, "=")
			pod := testutil.MakeAGICPod(split[0], split[1])
			fakeClient := fake.NewClientset(&pod)
			got, err := findClusterLabel(t.Context(), fakeClient)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if got != label {
				t.Fatalf("Expected label %q, got %q", label, got)
			}
		})
	}

	t.Run("using neither label", func(t *testing.T) {
		pod := testutil.MakeAGICPod()
		fakeClient := fake.NewClientset(&pod)

		_, err := findClusterLabel(t.Context(), fakeClient)
		if err == nil {
			t.Fatal("Expected an error but got none")
		}
	})
}

func TestAggregateFromFiles(t *testing.T) {
	aggregator := makeAggregator()
	className := DefaultIngressClassName

	// unfortunately ingress2gateway doesn't use an interface for interacting with the filesystem, so we make these
	// temp files
	writeIngress := func(t *testing.T, name string, class string) string {
		t.Helper()

		yml := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ` + name + `
spec:
  ingressClassName: ` + class + `
  rules:
  - http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80
`
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("ingress-%s-*.yaml", name))

		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		defer func() { _ = tmpFile.Close() }()

		if _, err := tmpFile.WriteString(yml); err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}

		t.Cleanup(func() {
			_ = os.Remove(tmpFile.Name())
		})

		return tmpFile.Name()
	}

	t.Run("file doesn't exist", func(t *testing.T) {
		_, err := aggregator.AggregateFromFiles(className, "nonexistent-file.yaml")
		if err == nil || !strings.Contains(err.Error(), "failed to collect Ingresses from files") {
			t.Fatalf("expected an error reading files, got %v", err)
		}
	})

	t.Run("ingress classname mismatch", func(t *testing.T) {
		file := writeIngress(t, "ingress1", "wrong-classname")
		got, err := aggregator.AggregateFromFiles(file)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(got.IngressContexts) != 0 {
			t.Fatalf("Expected 0 ingresses, got %d", len(got.IngressContexts))
		}
	})

	t.Run("two ingresses", func(t *testing.T) {
		file1 := writeIngress(t, "ingress1", className)
		file2 := writeIngress(t, "ingress2", className)
		got, err := aggregator.AggregateFromFiles(file1, file2)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(got.IngressContexts) != 2 {
			t.Fatalf("expected 2 ingress, got %d", len(got.IngressContexts))
		}

		ingress1, ingress1Present := got.IngressContexts[types.NamespacedName{Name: "ingress1"}]
		ingress2, ingress2Present := got.IngressContexts[types.NamespacedName{Name: "ingress2"}]

		if !ingress1Present || !ingress2Present {
			t.Fatalf("expected both ingresses to be present, got ingress1Present=%t, ingress2Present=%t", ingress1Present, ingress2Present)
		}

		if ingress1.Ingress.Name != "ingress1" || ingress2.Ingress.Name != "ingress2" {
			t.Fatalf("ingress names do not match, got %s and %s", ingress1.Ingress.Name, ingress2.Ingress.Name)
		}
	})
}

func makeAggregator() Aggregator {
	return Aggregator{
		log: logging.New("test"),
		Options: Options{
			ClusterLabel: agicHelmLabel,
		},
	}
}
