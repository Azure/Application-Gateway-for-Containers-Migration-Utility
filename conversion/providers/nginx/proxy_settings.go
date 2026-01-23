package nginx

import (
	"strconv"
	"strings"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	crds_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// ========== Proxy Settings Handlers ==========

func (p Provider) handleProxyConnectTimeout(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC doesn't have connect timeout - only route timeout
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXProxySettingsPartialSupport, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleProxySendTimeout(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC doesn't have send timeout - only route timeout
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXProxySettingsPartialSupport, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleProxyReadTimeout(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// proxy-read-timeout can be mapped to route timeout
	// NGINX format can be "60s" or just "60"
	value := annotationCtx.Value

	var seconds int

	var err error

	switch {
	case strings.HasSuffix(value, "s"):
		seconds, err = strconv.Atoi(strings.TrimSuffix(value, "s"))
	case strings.HasSuffix(value, "m"):
		minutes, parseErr := strconv.Atoi(strings.TrimSuffix(value, "m"))
		if parseErr != nil {
			err = parseErr
		} else {
			seconds = minutes * 60
		}
	default:
		seconds, err = strconv.Atoi(value)
	}

	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))
		return err
	}

	policy, err := p.GetOrCreateRoutePolicy(output, routeCtx.HTTPRoute, annotationCtx)
	if err != nil {
		return err
	}

	policy.Spec.Default.RouteTimeouts = &crds_v1.RouteTimeouts{
		RouteTimeout: meta_v1.Duration{Duration: time.Duration(seconds) * time.Second},
	}
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))

	return nil
}

func (p Provider) handleProxyBodySize(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC doesn't have direct support for request body size limits
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXProxySettingsPartialSupport, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

func (p Provider) handleProxyBuffering(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC doesn't have direct support for proxy buffering settings
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXProxySettingsPartialSupport, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

// ========== Load Balancing Handlers ==========

// handleLoadBalance handles the load-balance annotation.
// AGC uses round-robin by default, so if the annotation is set to round_robin,
// we can safely ignore it. Other algorithms are not supported.
func (p Provider) handleLoadBalance(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	algorithm := strings.ToLower(strings.TrimSpace(annotationCtx.Value))

	// round_robin is the default in AGC, so we can safely ignore this annotation
	if algorithm == "round_robin" {
		p.log.Info("Load balance algorithm 'round_robin' is the default in AGC; annotation ignored")
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)

		return nil
	}

	// Other algorithms are not supported
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXLoadBalanceNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

// handleUpstreamVhost handles the upstream-vhost annotation.
// This sets the Host header sent to the backend, which maps to Gateway API's
// URLRewrite filter with Hostname field.
func (p Provider) handleUpstreamVhost(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	hostname := strings.TrimSpace(annotationCtx.Value)
	if hostname == "" {
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	// Use URLRewrite filter to set the Host header to the backend
	insertHostRewrite := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.URLRewrite == nil {
			filter.URLRewrite = &gatewayapi_v1.HTTPURLRewriteFilter{}
		}

		hostnameValue := gatewayapi_v1.PreciseHostname(hostname)
		filter.URLRewrite.Hostname = &hostnameValue

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterURLRewrite, insertHostRewrite)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}
