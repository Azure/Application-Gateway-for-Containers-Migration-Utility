package k8snames

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func NamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

func NamespaceDerefOr(namespace *gatewayapi_v1.Namespace, defaultNamespace string) string {
	if namespace != nil && *namespace != "" {
		return string(*namespace)
	}

	return defaultNamespace
}
