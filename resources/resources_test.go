package resources

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestNewK8sResourceID(t *testing.T) {
	obj := object{
		gvk: schema.GroupVersionKind{
			Group:   "group",
			Version: "version",
			Kind:    "kind",
		},
		NamespacedName: types.NamespacedName{
			Namespace: "namespace",
			Name:      "name",
		},
	}
	expect := K8sResourceID{
		GroupVersionKind: obj.gvk,
		NamespacedName:   obj.NamespacedName,
	}

	if got := NewK8sResourceID(obj); got != expect {
		t.Fatalf("Expected: %+v, got: %+v", expect, got)
	}
}

type object struct {
	gvk schema.GroupVersionKind
	types.NamespacedName
}

func (o object) GetObjectKind() schema.ObjectKind {
	return o
}

func (o object) GroupVersionKind() schema.GroupVersionKind {
	return o.gvk
}

func (o object) SetGroupVersionKind(schema.GroupVersionKind) {}

func (o object) GetName() string {
	return o.Name
}

func (o object) GetNamespace() string {
	return o.Namespace
}
