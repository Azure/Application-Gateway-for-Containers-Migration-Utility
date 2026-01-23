package nginx

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// ========== Redirect Handlers ==========

// handlePermanentRedirect handles the permanent-redirect annotation.
// This sets a 301 redirect to the specified URL using Gateway API's RequestRedirect filter.
func (p Provider) handlePermanentRedirect(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	redirectURL := strings.TrimSpace(annotationCtx.Value)
	if redirectURL == "" {
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	// Parse the redirect URL to extract components
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXRedirectURLInvalid, fmt.Errorf("invalid URL %q: %w", redirectURL, err)))
		annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

		return nil
	}

	// Build the RequestRedirect filter
	insertRedirect := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.RequestRedirect == nil {
			filter.RequestRedirect = &gatewayapi_v1.HTTPRequestRedirectFilter{}
		}

		// Set status code to 301 (Permanent)
		filter.RequestRedirect.StatusCode = ptr.To(301)

		// Set scheme if present
		if parsedURL.Scheme != "" {
			filter.RequestRedirect.Scheme = &parsedURL.Scheme
		}

		// Set hostname if present
		if parsedURL.Hostname() != "" {
			hostname := gatewayapi_v1.PreciseHostname(parsedURL.Hostname())
			filter.RequestRedirect.Hostname = &hostname
		}

		// Set port if explicitly specified
		if parsedURL.Port() != "" {
			port, _ := strconv.Atoi(parsedURL.Port())
			portNum := gatewayapi_v1.PortNumber(port)
			filter.RequestRedirect.Port = &portNum
		}

		// Set path if present
		if parsedURL.Path != "" && parsedURL.Path != "/" {
			filter.RequestRedirect.Path = &gatewayapi_v1.HTTPPathModifier{
				Type:            gatewayapi_v1.FullPathHTTPPathModifier,
				ReplaceFullPath: &parsedURL.Path,
			}
		}

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterRequestRedirect, insertRedirect)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// handlePermanentRedirectCode handles the permanent-redirect-code annotation.
// This allows customizing the status code for permanent redirects (default 301).
// This annotation works in conjunction with permanent-redirect.
func (p Provider) handlePermanentRedirectCode(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	codeStr := strings.TrimSpace(annotationCtx.Value)
	if codeStr == "" {
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	code, err := strconv.Atoi(codeStr)
	if err != nil || (code != 301 && code != 308) {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXRedirectCodeInvalid, fmt.Errorf("invalid permanent redirect code %q: valid codes are 301, 308", codeStr)))
		annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

		return nil
	}

	// Update the existing RequestRedirect filter's status code
	insertRedirect := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.RequestRedirect == nil {
			filter.RequestRedirect = &gatewayapi_v1.HTTPRequestRedirectFilter{}
		}

		filter.RequestRedirect.StatusCode = ptr.To(code)

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterRequestRedirect, insertRedirect)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// handleTemporalRedirect handles the temporal-redirect annotation.
// This sets a 302 redirect to the specified URL using Gateway API's RequestRedirect filter.
func (p Provider) handleTemporalRedirect(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	redirectURL := strings.TrimSpace(annotationCtx.Value)
	if redirectURL == "" {
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	// Parse the redirect URL to extract components
	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXRedirectURLInvalid, fmt.Errorf("invalid URL %q: %w", redirectURL, err)))
		annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

		return nil
	}

	// Build the RequestRedirect filter
	insertRedirect := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.RequestRedirect == nil {
			filter.RequestRedirect = &gatewayapi_v1.HTTPRequestRedirectFilter{}
		}

		// Set status code to 302 (Temporary)
		filter.RequestRedirect.StatusCode = ptr.To(302)

		// Set scheme if present
		if parsedURL.Scheme != "" {
			filter.RequestRedirect.Scheme = &parsedURL.Scheme
		}

		// Set hostname if present
		if parsedURL.Hostname() != "" {
			hostname := gatewayapi_v1.PreciseHostname(parsedURL.Hostname())
			filter.RequestRedirect.Hostname = &hostname
		}

		// Set port if explicitly specified
		if parsedURL.Port() != "" {
			port, _ := strconv.Atoi(parsedURL.Port())
			portNum := gatewayapi_v1.PortNumber(port)
			filter.RequestRedirect.Port = &portNum
		}

		// Set path if present
		if parsedURL.Path != "" && parsedURL.Path != "/" {
			filter.RequestRedirect.Path = &gatewayapi_v1.HTTPPathModifier{
				Type:            gatewayapi_v1.FullPathHTTPPathModifier,
				ReplaceFullPath: &parsedURL.Path,
			}
		}

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterRequestRedirect, insertRedirect)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// handleTemporalRedirectCode handles the temporal-redirect-code annotation.
// This allows customizing the status code for temporal redirects (default 302).
// This annotation works in conjunction with temporal-redirect.
func (p Provider) handleTemporalRedirectCode(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	codeStr := strings.TrimSpace(annotationCtx.Value)
	if codeStr == "" {
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	code, err := strconv.Atoi(codeStr)
	if err != nil || (code != 302 && code != 303 && code != 307) {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXRedirectCodeInvalid, fmt.Errorf("invalid temporal redirect code %q: valid codes are 302, 303, 307", codeStr)))
		annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

		return nil
	}

	// Update the existing RequestRedirect filter's status code
	insertRedirect := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.RequestRedirect == nil {
			filter.RequestRedirect = &gatewayapi_v1.HTTPRequestRedirectFilter{}
		}

		filter.RequestRedirect.StatusCode = ptr.To(code)

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterRequestRedirect, insertRedirect)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// handleFromToWWWRedirect handles the from-to-www-redirect annotation.
// When set to "true", redirects requests from example.com to www.example.com.
func (p Provider) handleFromToWWWRedirect(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	enabled := strings.ToLower(strings.TrimSpace(annotationCtx.Value))
	if enabled != "true" {
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	// Get the host from the ingress to determine the www redirect target
	var host string
	if len(ingressCtx.Ingress.Spec.Rules) > 0 && ingressCtx.Ingress.Spec.Rules[0].Host != "" {
		host = ingressCtx.Ingress.Spec.Rules[0].Host
	}

	if host == "" {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXFromToWWWRedirectNoHost, nil))
		annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

		return nil
	}

	// Create the www redirect hostname
	var wwwHost string

	if strings.HasPrefix(host, "www.") {
		// Already has www prefix, nothing to do
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	wwwHost = "www." + host

	// Note: The from-to-www-redirect in NGINX creates an additional server block
	// that handles requests to the non-www domain and redirects to www.
	// In Gateway API, this would require a separate HTTPRoute with the non-www hostname.
	// Since we're modifying the existing route, we'll add a note about this limitation.
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXFromToWWWRedirectPartial, fmt.Errorf("redirect from %s to %s requires a separate HTTPRoute", host, wwwHost)))

	// Build the RequestRedirect filter to redirect to www
	insertRedirect := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.RequestRedirect == nil {
			filter.RequestRedirect = &gatewayapi_v1.HTTPRequestRedirectFilter{}
		}

		// Set status code to 301 (Permanent redirect for www)
		filter.RequestRedirect.StatusCode = ptr.To(301)

		// Set the www hostname
		hostname := gatewayapi_v1.PreciseHostname(wwwHost)
		filter.RequestRedirect.Hostname = &hostname

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterRequestRedirect, insertRedirect)
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}
