package agic

import (
	"fmt"
	"strconv"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func (p Provider) handleRequestTimeout(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	seconds, err := strconv.Atoi(annotationCtx.Value)
	if err != nil {
		err := fmt.Errorf("could not convert request timeout annotation value to an int: %v", err)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))

		return err
	}

	policy, err := p.GetOrCreateRoutePolicy(output, routeCtx.HTTPRoute, annotationCtx)
	if err != nil {
		return err
	}

	if policy.Spec.Default != nil && policy.Spec.Default.RouteTimeouts != nil {
		p.log.Warn(
			"overwriting route timeout on RoutePolicy",
			"value",
			policy.Spec.Default.RouteTimeouts,
			"RoutePolicy",
			k8snames.NamespacedName(policy),
		)
	}

	policy.Spec.Default.RouteTimeouts = &albcontrollerapi_v1.RouteTimeouts{
		RouteTimeout: meta_v1.Duration{Duration: time.Second * time.Duration(seconds)},
	}
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

	return nil
}

func (p Provider) handleCookieBasedAffinity(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	val, err := strconv.ParseBool(annotationCtx.Value)
	if err != nil {
		err := fmt.Errorf(
			"could not convert cookie based affinity annotation value %q to bool: %v",
			annotationCtx.Value,
			err,
		)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))

		return err
	}

	if !val {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		p.log.Info(
			"Annotation explicitly set to false; will not migrate",
			"annotation", annotationCtx.Key,
			"ingress", types.NamespacedName{Name: ingressCtx.Ingress.Name, Namespace: ingressCtx.Ingress.Namespace},
		)

		return nil
	}

	policy, err := p.GetOrCreateRoutePolicy(output, routeCtx.HTTPRoute, annotationCtx)
	if err != nil {
		return err
	}

	if policy.Spec.Default != nil && policy.Spec.Default.SessionAffinity != nil {
		// I don't think this should ever happen
		p.log.Warn(
			"overwriting SessionAffinity on RoutePolicy",
			"value",
			policy.Spec.Default.SessionAffinity,
			"RoutePolicy",
			k8snames.NamespacedName(policy),
		)
	}

	policy.Spec.Default.SessionAffinity = &albcontrollerapi_v1.SessionAffinity{
		AffinityType: albcontrollerapi_v1.AffinityTypeManagedCookie,
	}
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

	return nil
}

func (p Provider) GetOrCreateRoutePolicy(
	output resources.AGCResourceGraph,
	httpRoute *gatewayapi_v1.HTTPRoute,
	_ *resources.IngressAnnotationContext,
) (*albcontrollerapi_v1.RoutePolicy, error) {
	return output.GetOrCreateRoutePolicy(k8snames.NamespacedName(httpRoute)), nil
}
