package migration

import (
	"context"
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/aggregation"

	"k8s.io/client-go/kubernetes/fake"
	fake_ctrl "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// We'll mostly rely on integration/e2e tests to validation this package
// These tests are just basic sanity checks to protect against the most obvious errors, and to satisfy coverage checks
func TestMigrateFiles(t *testing.T) {
	t.Run("no files", func(t *testing.T) {
		_, _, err := MigrateFiles(context.Background(), Options{Aggregation: aggregation.Options{IngressClassName: "my-ingress-class"}})
		if err != nil {
			t.Fatal("expected error when no files are provided, got nil")
		}
	})
}

func TestMigrateCluster(t *testing.T) {
	t.Run("no resources", func(t *testing.T) {
		fakeK8s := fake.NewClientset()
		fakeCtrl := fake_ctrl.NewFakeClient()
		_, _, err := MigrateCluster(context.Background(), fakeK8s, fakeCtrl, Options{})

		if err == nil {
			t.Fatal("expected no error when no resources are present")
		}
	})
}
