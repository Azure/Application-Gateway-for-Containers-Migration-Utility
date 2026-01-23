package resources

import (
	network_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
)

const IngressClassAnnotation = "kubernetes.io/ingress.class"

// IngressContext tracks the migration of a single Ingress and its annotations
type IngressContext struct {
	Ingress     network_v1.Ingress
	Status      MigrationStatus
	HTTPRoutes  []types.NamespacedName
	Annotations map[string]*IngressAnnotationContext
	Metadata    map[string]string // Additional metadata for cross-annotation processing
}

func NewIngressContext(ingress network_v1.Ingress) *IngressContext {
	annotations := make(map[string]*IngressAnnotationContext)

	for key, val := range ingress.Annotations {
		if key == IngressClassAnnotation {
			continue
		}

		annotations[key] = NewIngressAnnotationContext(key, val)
	}

	return &IngressContext{
		Ingress:     ingress,
		Status:      MigrationStatusNotStarted,
		Annotations: annotations,
		Metadata:    make(map[string]string),
	}
}

// MigrationComplete marks the ingress migration as complete or failed.
// If errors is non-nil, the status is set to Error and the error details
// will be included in the migration report for this ingress.
func (ic *IngressContext) MigrationComplete(errors error) {
	// Note: Error details are captured in annotation-level issues.
	// TODO: add an Errors field to IngressContext for ingress-level errors.
	if errors != nil {
		ic.Status = MigrationStatusError
	} else {
		ic.Status = MigrationStatusCompleted
	}
}
