package resources

import (
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"
)

// nolint: revive
const (

	// Standard annotations
	LastAppliedConfiguration = "kubectl.kubernetes.io/last-applied-configuration"
)

// IngressAnnotationContext tracks the migration status of a single Ingress annotation
type IngressAnnotationContext struct {
	Key                  string
	Value                string
	status               MigrationStatus
	DestinationResources sets.Set[K8sResourceID]
	Issues               []Issue
}

func NewIngressAnnotationContext(key, value string) *IngressAnnotationContext {
	return &IngressAnnotationContext{
		Key:                  key,
		Value:                value,
		status:               MigrationStatusNotStarted,
		DestinationResources: sets.New[K8sResourceID](),
	}
}

func (iac *IngressAnnotationContext) RegisterIssue(issue Issue) {
	iac.Issues = append(iac.Issues, issue)
	iac.SetStatus(issue.MigrationStatus())
}

func (iac *IngressAnnotationContext) AddDestination(resource K8sResourceID) {
	iac.DestinationResources.Insert(resource)
	iac.SetStatus(MigrationStatusCompleted)
}

func (iac IngressAnnotationContext) Status() MigrationStatus {
	return iac.status
}

// SetStatus updates the context status.
// Note that the status cannot be set to Completed once it has been set to Warning or Error.
func (iac *IngressAnnotationContext) SetStatus(newStatus MigrationStatus) {
	if iac.status == MigrationStatusNotStarted {
		iac.status = newStatus
		return
	}

	switch iac.status {
	case MigrationStatusError:
		return
	case MigrationStatusWarning:
		if newStatus == MigrationStatusError {
			iac.status = newStatus
		}
	default:
		iac.status = newStatus
	}
}

func (iac IngressAnnotationContext) ValueInt32() (int32, error) {
	val, err := strconv.ParseInt(iac.Value, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(val), nil // #nosec G115
}
