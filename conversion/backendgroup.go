package conversion

import (
	"errors"
	"fmt"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"k8s.io/apimachinery/pkg/types"
)

// BackendGroup represents a collection of services targeted by an IngressRuleGroup
type BackendGroup struct {
	services map[types.NamespacedName]ServiceDetails
}

type ServiceDetails struct {
	NamespacedName types.NamespacedName
	Port           *int32
}

func NewBackendGroup(routeCtx *HTTPRouteContext) BackendGroup {
	services := make(map[types.NamespacedName]ServiceDetails)

	for _, rule := range routeCtx.Spec.Rules {
		for _, backendRef := range rule.BackendRefs {
			if backendRef.Kind != nil && *backendRef.Kind != "Service" {
				continue
			}

			svcNN := types.NamespacedName{
				Namespace: k8snames.NamespaceDerefOr(backendRef.Namespace, routeCtx.Namespace),
				Name:      string(backendRef.Name),
			}

			if svc, ok := services[svcNN]; ok {
				if svc.Port == nil {
					svc.Port = (*int32)(backendRef.Port)
					services[svcNN] = svc
				}
			} else {
				services[svcNN] = ServiceDetails{
					NamespacedName: svcNN,
					Port:           (*int32)(backendRef.Port),
				}
			}
		}
	}

	return BackendGroup{
		services: services,
	}
}

func (bg BackendGroup) Apply(fn func(service ServiceDetails) error) error {
	var errs error

	allAppliesFailed := true

	for _, service := range bg.services {
		if err := fn(service); err != nil {
			errs = errors.Join(errs, err)
		} else {
			allAppliesFailed = false
		}
	}

	if allAppliesFailed {
		return fmt.Errorf("failed to apply health check policies: %w", errs)
	}

	return errs
}
