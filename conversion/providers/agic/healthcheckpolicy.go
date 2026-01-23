package agic

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

func (c Provider) handleHealthProbePath(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	group := conversion.NewBackendGroup(routeCtx)

	return group.Apply(func(service conversion.ServiceDetails) error {
		policy := output.GetOrCreateHealthCheckPolicy(service.NamespacedName)
		if policy.Spec.Default.HTTP == nil {
			policy.Spec.Default.HTTP = &albcontrollerapi_v1.HTTPSpecifiers{}
		}

		if policy.Spec.Default.HTTP.Path != "" && policy.Spec.Default.HTTP.Path != annotationCtx.Value {
			err := fmt.Errorf("HealthCheckPolicy %q already has path defined from a different Ingress", k8snames.NamespacedName(policy))
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueHealthCheckConflict, err))

			return err
		}

		policy.Spec.Default.HTTP.Path = annotationCtx.Value
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

		return nil
	})
}

func (c Provider) handleHealthProbeInterval(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	group := conversion.NewBackendGroup(routeCtx)
	seconds, err := strconv.Atoi(annotationCtx.Value)

	if err != nil {
		err := fmt.Errorf("could not convert health probe interval annotation value to int: %v", err)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))

		return err
	}

	return group.Apply(func(service conversion.ServiceDetails) error {
		metaDuration := meta_v1.Duration{Duration: time.Second * time.Duration(seconds)}
		policy := output.GetOrCreateHealthCheckPolicy(service.NamespacedName)

		if policy.Spec.Default.Interval.Duration != 0 && metaDuration != policy.Spec.Default.Interval {
			err := fmt.Errorf("HealthCheckPolicy %q already has an interval defined from a different Ingress", k8snames.NamespacedName(policy))
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueHealthCheckConflict, err))

			return err
		}

		policy.Spec.Default.Interval = metaDuration
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

		return nil
	})
}

func (c Provider) handleHealthProbeTimeout(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	group := conversion.NewBackendGroup(routeCtx)
	seconds, err := strconv.Atoi(annotationCtx.Value)

	if err != nil {
		err := fmt.Errorf("could not convert health probe timeout annotation value to int: %v", err)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))

		return err
	}

	return group.Apply(func(service conversion.ServiceDetails) error {
		policy := output.GetOrCreateHealthCheckPolicy(service.NamespacedName)
		timeout := meta_v1.Duration{Duration: time.Second * time.Duration(seconds)}

		if policy.Spec.Default.Timeout.Duration != 0 && timeout != policy.Spec.Default.Timeout {
			err := fmt.Errorf("HealthCheckPolicy %q already has a timeout defined from a different Ingress", k8snames.NamespacedName(policy))
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueHealthCheckConflict, err))

			return err
		}

		policy.Spec.Default.Timeout = timeout
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

		return nil
	})
}

func (c Provider) handleHealthProbeUnhealthyThreshold(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	group := conversion.NewBackendGroup(routeCtx)
	threshold, err := strconv.ParseInt(annotationCtx.Value, 10, 32)

	if err != nil {
		err := fmt.Errorf("could not convert health probe unhealthy threshold annotation value to int: %v", err)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))

		return err
	}

	return group.Apply(func(service conversion.ServiceDetails) error {
		policy := output.GetOrCreateHealthCheckPolicy(service.NamespacedName)
		if policy.Spec.Default.UnhealthyThreshold != 0 && int32(threshold) != policy.Spec.Default.UnhealthyThreshold {
			err := fmt.Errorf("HealthCheckPolicy %q already has an unhealthy threshold defined from a different Ingress", k8snames.NamespacedName(policy))
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueHealthCheckConflict, err))

			return err
		}

		policy.Spec.Default.UnhealthyThreshold = int32(threshold) // #nosec G115
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

		return nil
	})
}

func (c Provider) handleHealthProbePort(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	group := conversion.NewBackendGroup(routeCtx)
	port, err := annotationCtx.ValueInt32()

	if err != nil {
		err := fmt.Errorf("could not convert health probe port annotation value to int: %v", err)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))

		return err
	}

	return group.Apply(func(service conversion.ServiceDetails) error {
		policy := output.GetOrCreateHealthCheckPolicy(service.NamespacedName)

		if policy.Spec.Default.Port != 0 && policy.Spec.Default.Port != port {
			err := fmt.Errorf("HealthCheckPolicy %q already has a port defined from a different Ingress", k8snames.NamespacedName(policy))
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueHealthCheckConflict, err))

			return err
		}

		policy.Spec.Default.Port = port
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

		return nil
	})
}

func (c Provider) handleHealthProbeHostname(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	group := conversion.NewBackendGroup(routeCtx)

	return group.Apply(func(service conversion.ServiceDetails) error {
		policy := output.GetOrCreateHealthCheckPolicy(service.NamespacedName)
		if policy.Spec.Default.HTTP == nil {
			policy.Spec.Default.HTTP = &albcontrollerapi_v1.HTTPSpecifiers{}
		}

		if policy.Spec.Default.HTTP.Host != "" && policy.Spec.Default.HTTP.Host != annotationCtx.Value {
			err := fmt.Errorf("HealthCheckPolicy %q already has a hostname defined from a different Ingress", k8snames.NamespacedName(policy))
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueHealthCheckConflict, err))

			return err
		}

		policy.Spec.Default.HTTP.Host = annotationCtx.Value
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

		return nil
	})
}

func (c Provider) handleHealthProbeStatusCode(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	group := conversion.NewBackendGroup(routeCtx)
	statusCodes, err := parseStatusCodes(annotationCtx.Value)

	if err != nil {
		annotationCtx.SetStatus(resources.MigrationStatusError)

		return fmt.Errorf("could not parse health probe status codes: %v", err)
	}

	return group.Apply(func(service conversion.ServiceDetails) error {
		policy := output.GetOrCreateHealthCheckPolicy(service.NamespacedName)
		if policy.Spec.Default.HTTP == nil {
			policy.Spec.Default.HTTP = &albcontrollerapi_v1.HTTPSpecifiers{}
		}

		if policy.Spec.Default.HTTP.Match == nil {
			policy.Spec.Default.HTTP.Match = &albcontrollerapi_v1.HTTPMatch{}
		}

		if policy.Spec.Default.HTTP.Match.StatusCodes != nil && !reflect.DeepEqual(policy.Spec.Default.HTTP.Match.StatusCodes, statusCodes) {
			err := fmt.Errorf("HealthCheckPolicy %q already has status codes defined", k8snames.NamespacedName(policy))
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueHealthCheckConflict, err))

			return err
		}

		policy.Spec.Default.HTTP.Match.StatusCodes = statusCodes
		annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

		return nil
	})
}

// parseStatusCodes parses a comma-separated list of status codes and ranges (e.g., "200,201-204,301")
func parseStatusCodes(value string) ([]*albcontrollerapi_v1.StatusCodes, error) {
	statusCodes := []*albcontrollerapi_v1.StatusCodes{}

	for rangeStr := range strings.SplitSeq(value, ",") {
		bounds := strings.Split(rangeStr, "-")
		if len(bounds) == 1 {
			intVal, err := strconv.ParseInt(strings.TrimSpace(bounds[0]), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("could not convert status code value to int: %v", err)
			}

			statusCodes = append(statusCodes, &albcontrollerapi_v1.StatusCodes{
				Start: int32(intVal), // #nosec G115
				End:   int32(intVal), // #nosec G115
			})
		} else if len(bounds) == 2 {
			startVal, err := strconv.ParseInt(strings.TrimSpace(bounds[0]), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("could not convert status code range start value to int: %v", err)
			}

			endVal, err := strconv.ParseInt(strings.TrimSpace(bounds[1]), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("could not convert status code range end value to int: %v", err)
			}

			statusCodes = append(statusCodes, &albcontrollerapi_v1.StatusCodes{
				Start: int32(startVal), // #nosec G115
				End:   int32(endVal),   // #nosec G115
			})
		}
	}

	return statusCodes, nil
}
