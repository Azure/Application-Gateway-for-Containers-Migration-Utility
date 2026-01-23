package resources

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestNewIngressContext(t *testing.T) {
	ingress := testutil.MakeIngress("default", "test-ingress", testutil.IngressOptions{}, "test", "value")
	ingress.Annotations[IngressClassAnnotation] = "this should be ignored"
	got := NewIngressContext(ingress)

	expectAnnotations := map[string]*IngressAnnotationContext{
		"test": NewIngressAnnotationContext("test", "value"),
	}

	if !reflect.DeepEqual(got.Annotations, expectAnnotations) {
		t.Fatalf("Expected annotations: %+v, got: %+v", expectAnnotations, got.Annotations)
	}

	if !reflect.DeepEqual(got.Ingress, ingress) {
		t.Fatalf("Expected ingress: %+v, got: %+v", ingress, got.Ingress)
	}

	if got.Status != MigrationStatusNotStarted {
		t.Fatalf("Expected MigrationStatusNotStarted, got: %v", got.Status)
	}
}

func TestIngressContext_MigrationComplete(t *testing.T) {
	ingress := testutil.MakeIngress("", "test-ingress", testutil.IngressOptions{}, "test", "value")
	ic := NewIngressContext(ingress)

	t.Run("No errors", func(t *testing.T) {
		ic.MigrationComplete(nil)

		if ic.Status != MigrationStatusCompleted {
			t.Fatalf("Expected status to be MigrationStatusCompleted, got: %v", ic.Status)
		}
	})

	t.Run("With errors", func(t *testing.T) {
		ic.MigrationComplete(errors.New("failed"))

		if ic.Status != MigrationStatusError {
			t.Fatalf("Expected status to be MigrationStatusError, got: %v", ic.Status)
		}
	})
}
