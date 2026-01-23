package resources

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestNewIngressAnnotationContext(t *testing.T) {
	key := "test-key"
	value := "test-value"
	got := NewIngressAnnotationContext(key, value)

	expect := &IngressAnnotationContext{
		Key:                  key,
		Value:                value,
		status:               MigrationStatusNotStarted,
		DestinationResources: sets.New[K8sResourceID](),
	}

	if deep.Equal(got, expect) != nil {
		t.Fatalf("Expected annotation context: %+v, got: %+v", expect, got)
	}
}

func TestIngressAnnotation_SetDestination(t *testing.T) {
	key := "test-key"
	value := "test-value"
	annotationCtx := NewIngressAnnotationContext(key, value)

	resourceID := K8sResourceID{
		GroupVersionKind: schema.GroupVersionKind{
			Group:   "group",
			Version: "version",
			Kind:    "kind",
		},
		NamespacedName: types.NamespacedName{
			Namespace: "namespace",
			Name:      "name",
		},
	}

	annotationCtx.AddDestination(resourceID)

	if !annotationCtx.DestinationResources.Equal(sets.New(resourceID)) {
		t.Fatalf("Expected destination resources to contain: %+v, got: %+v", resourceID, annotationCtx.DestinationResources)
	}

	if annotationCtx.Status() != MigrationStatusCompleted {
		t.Fatalf("Expected status: %s, got: %s", MigrationStatusCompleted, annotationCtx.Status())
	}
}

func TestIngressAnnotation_SetStatus(t *testing.T) {
	testCases := []struct {
		before MigrationStatus
		set    MigrationStatus
		after  MigrationStatus
	}{
		{
			before: MigrationStatusNotStarted,
			set:    MigrationStatusCompleted,
			after:  MigrationStatusCompleted,
		},
		{
			before: MigrationStatusWarning,
			set:    MigrationStatusCompleted,
			after:  MigrationStatusWarning,
		},
		{
			before: MigrationStatusError,
			set:    MigrationStatusWarning,
			after:  MigrationStatusError,
		},
		{
			before: MigrationStatusCompleted,
			set:    MigrationStatusWarning,
			after:  MigrationStatusWarning,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("from %s set %s", tc.before, tc.set), func(t *testing.T) {
			annoCtx := NewIngressAnnotationContext("key", "val")
			annoCtx.status = tc.before
			annoCtx.SetStatus(tc.set)

			if annoCtx.Status() != tc.after {
				t.Fatalf("Expect status %s, got %s", tc.after, annoCtx.Status())
			}
		})
	}
}
