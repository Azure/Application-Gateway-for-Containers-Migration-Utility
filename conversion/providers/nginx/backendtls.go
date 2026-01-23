package nginx

import (
	"fmt"
	"strings"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	crds_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// ========== Backend Protocol Handlers ==========

func (p Provider) handleBackendProtocol(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	protocol := strings.ToLower(annotationCtx.Value)

	switch protocol {
	case "http":
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	case "https":
		backends := conversion.NewBackendGroup(routeCtx)
		if err := backends.Apply(func(service conversion.ServiceDetails) error {
			policy := output.GetOrCreateBackendTLSPolicy(service.NamespacedName)
			ports := newPortSet(policy.Spec.Default.Ports...)
			if service.Port != nil {
				ports.InsertInts(int(*service.Port))
			}

			policy.Spec.Default.Ports = ports.List()

			annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
			return nil
		}); err != nil {
			annotationCtx.SetStatus(resources.MigrationStatusError)

			err = fmt.Errorf("failed to create backend TLS policy for backend protocol annotation: %w", err)
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueCreatingBackendTLSPolicy, err))

			return err
		}

		annotationCtx.SetStatus(resources.MigrationStatusCompleted)

		return nil
	case "grpc", "grpcs":
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueGRPCNotSupportedByTool, nil))
		if protocol == "grpcs" {
			backends := conversion.NewBackendGroup(routeCtx)
			if err := backends.Apply(func(service conversion.ServiceDetails) error {
				policy := output.GetOrCreateBackendTLSPolicy(service.NamespacedName)
				ports := newPortSet(policy.Spec.Default.Ports...)
				if service.Port != nil {
					ports.InsertInts(int(*service.Port))
				}
				policy.Spec.Default.Ports = ports.List()
				annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
				return nil
			}); err != nil {
				annotationCtx.SetStatus(resources.MigrationStatusError)

				err = fmt.Errorf("failed to create backend TLS policy for GRPCS backend protocol: %w", err)
				annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueCreatingBackendTLSPolicy, err))

				return err
			}
		}

		annotationCtx.SetStatus(resources.MigrationStatusCompleted)

		return nil
	default:
		annotationCtx.SetStatus(resources.MigrationStatusError)
		err := fmt.Errorf("unknown backend protocol %q", annotationCtx.Value)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueUnsupportBackendProtocol, err))

		return err
	}
}

// backendTLSPortSet manages backend TLS policy ports.
type backendTLSPortSet struct {
	ports map[int]crds_v1.BackendTLSPolicyPort
}

// NewPortSet creates a new port set.
func newPortSet(ports ...crds_v1.BackendTLSPolicyPort) backendTLSPortSet {
	ps := backendTLSPortSet{
		ports: make(map[int]crds_v1.BackendTLSPolicyPort),
	}

	for _, p := range ports {
		ps.ports[p.Port] = p
	}

	return ps
}

// InsertInts adds integer ports to the set.
func (ps *backendTLSPortSet) InsertInts(ports ...int) {
	for _, port := range ports {
		ps.ports[port] = crds_v1.BackendTLSPolicyPort{Port: port}
	}
}

// List returns a list of ports.
func (ps *backendTLSPortSet) List() []crds_v1.BackendTLSPolicyPort {
	result := make([]crds_v1.BackendTLSPolicyPort, 0, len(ps.ports))

	for _, p := range ps.ports {
		result = append(result, p)
	}

	return result
}
