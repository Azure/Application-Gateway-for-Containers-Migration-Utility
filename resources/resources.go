// Package resources defines the structures representing AGIC and AGC resources.
package resources

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// K8sResourceID uniquely identifies a Kubernetes resource.
type K8sResourceID struct {
	schema.GroupVersionKind
	types.NamespacedName
}

type Object interface {
	GetObjectKind() schema.ObjectKind
	GetName() string
	GetNamespace() string
}

func NewK8sResourceID(obj Object) K8sResourceID {
	return K8sResourceID{
		GroupVersionKind: obj.GetObjectKind().GroupVersionKind(),
		NamespacedName: types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      obj.GetName(),
		},
	}
}
