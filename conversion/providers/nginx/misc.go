package nginx

import (
	"strings"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// ========== Custom Headers and Configuration Snippet Handlers ==========

func (p Provider) handleConfigurationSnippet(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Configuration snippets contain raw NGINX config - cannot be migrated
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXConfigurationSnippetNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

func (p Provider) handleServerSnippet(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Server snippets contain raw NGINX config - cannot be migrated
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXServerSnippetNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

func (p Provider) handleConnectionProxyHeader(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// connection-proxy-header sets the Connection header for proxied requests
	// This can be done with request header modifier
	insertHeaderFilter := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.RequestHeaderModifier == nil {
			filter.RequestHeaderModifier = &gatewayapi_v1.HTTPHeaderFilter{}
		}

		header := gatewayapi_v1.HTTPHeader{
			Name:  "Connection",
			Value: annotationCtx.Value,
		}
		filter.RequestHeaderModifier.Set = append(filter.RequestHeaderModifier.Set, header)

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterRequestHeaderModifier, insertHeaderFilter)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// ========== Default Backend Handler ==========

func (p Provider) handleDefaultBackend(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// default-backend annotation specifies a service to use for unmatched requests
	// AGC doesn't have direct support - user should create a catch-all HTTPRoute
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXDefaultBackendNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

// ========== Ignored Annotation Handler ==========

func (p Provider) handleIgnoredAnnotation(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	annotationCtx.SetStatus(resources.MigrationStatusIgnored)
	return nil
}

// ========== Server Alias Handler ==========

func (p Provider) handleServerAlias(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// server-alias annotation allows specifying additional hostnames for the server.
	// The value can be a comma-separated or space-separated list of hostnames.
	// These are added to the HTTPRoute's hostnames array.
	route := routeCtx.HTTPRoute

	// Parse the aliases - can be comma or space separated
	aliasValue := annotationCtx.Value
	// Replace commas with spaces to normalize, then split
	aliasValue = strings.ReplaceAll(aliasValue, ",", " ")
	aliases := strings.Fields(aliasValue)

	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}
		// Add the alias to the hostnames array
		route.Spec.Hostnames = append(route.Spec.Hostnames, gatewayapi_v1.Hostname(alias))
	}

	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}
