package agic

import (
	"fmt"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

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

			err := fmt.Errorf("error creating backend TLS policy for backend protocol annotation: %w", err)
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueCreatingBackendTLSPolicy, err))

			return err
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

type backendTLSPortSet struct {
	ports sets.Set[albcontrollerapi_v1.BackendTLSPolicyPort]
}

func newPortSet(ports ...albcontrollerapi_v1.BackendTLSPolicyPort) backendTLSPortSet {
	return backendTLSPortSet{
		ports: sets.New(ports...),
	}
}

func (ps *backendTLSPortSet) InsertInts(ports ...int) {
	backendTLSPorts := make([]albcontrollerapi_v1.BackendTLSPolicyPort, 0, len(ports))
	for _, port := range ports {
		backendTLSPorts = append(backendTLSPorts, albcontrollerapi_v1.BackendTLSPolicyPort{
			Port: port,
		})
	}

	ps.ports.Insert(backendTLSPorts...)
}

func (ps *backendTLSPortSet) List() []albcontrollerapi_v1.BackendTLSPolicyPort {
	ports := ps.ports.UnsortedList()
	slices.SortFunc(ports, func(a, b albcontrollerapi_v1.BackendTLSPolicyPort) int {
		return a.Port - b.Port
	})

	return ports
}

func (p Provider) handleAppGWTrustedRootCertificates(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueAppGWTrustedRootCertificatesNotSupported, nil))
	// nolint: staticcheck
	return nil
}
