// Package nginx provides annotation handlers for migrating NGINX Ingress Controller
// ingress resources to Gateway API resources for Azure Application Gateway for Containers.
package nginx

import (
	"fmt"
	"regexp"
	"strconv"

	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

const (
	// annotationValueTrue is the string value "true" used for boolean annotations.
	annotationValueTrue = "true"
)

// Canary annotations from ingress-nginx
// Reference: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#canary

// handleCanary handles the nginx.ingress.kubernetes.io/canary annotation.
// When set to "true", this enables canary behavior for the Ingress.
// This annotation alone doesn't change routing behavior; it must be combined with
// canary-weight or canary-by-header annotations.
func (p Provider) handleCanary(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	if annotationCtx.Value != annotationValueTrue {
		// canary=false means this is not a canary Ingress
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	}

	// Mark as canary for reference - actual processing depends on other annotations
	ingressCtx.Metadata["isCanary"] = annotationValueTrue

	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// handleCanaryWeight handles the nginx.ingress.kubernetes.io/canary-weight annotation.
// This specifies the percentage of traffic (0-100 by default) to route to the canary backend.
//
// Gateway API supports backend weights via HTTPBackendRef.Weight, but this requires
// both the main and canary services to be backends of the same HTTPRoute rule.
// Since ingresses are processed independently, we add an issue noting manual configuration
// is required and set the weight on the route's backends.
func (p Provider) handleCanaryWeight(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	weight, err := strconv.ParseInt(annotationCtx.Value, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid canary-weight value %q: %w", annotationCtx.Value, err)
	}

	if weight < 0 {
		return fmt.Errorf("canary-weight must be non-negative, got %d", weight)
	}

	// Get canary-weight-total (default 100)
	weightTotal := int64(100)
	if totalStr, ok := ingressCtx.Ingress.Annotations[AnnotationCanaryWeightTotal]; ok {
		weightTotal, err = strconv.ParseInt(totalStr, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid canary-weight-total value %q: %w", totalStr, err)
		}

		if weightTotal <= 0 {
			return fmt.Errorf("canary-weight-total must be positive, got %d", weightTotal)
		}
	}

	if weight > weightTotal {
		return fmt.Errorf("canary-weight (%d) exceeds canary-weight-total (%d)", weight, weightTotal)
	}

	// Set the weight on all backend refs in this route
	weight32 := int32(weight)

	for i := range routeCtx.Spec.Rules {
		for j := range routeCtx.Spec.Rules[i].BackendRefs {
			routeCtx.Spec.Rules[i].BackendRefs[j].Weight = &weight32
		}
	}

	// Store metadata for potential post-processing
	ingressCtx.Metadata["canaryWeight"] = strconv.FormatInt(weight, 10)
	ingressCtx.Metadata["canaryWeightTotal"] = strconv.FormatInt(weightTotal, 10)

	// Register an issue explaining that manual merging may be required
	annotationCtx.RegisterIssue(resources.NewIssue(
		resources.IssueNGINXCanaryWeightRequiresManualMerge,
		fmt.Errorf("canary weight=%d, weightTotal=%d", weight, weightTotal),
	))

	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

// handleCanaryWeightTotal handles the nginx.ingress.kubernetes.io/canary-weight-total annotation.
// This sets the total weight value (default 100) for canary weight calculations.
// It's processed as part of handleCanaryWeight.
func (p Provider) handleCanaryWeightTotal(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Validation is done in handleCanaryWeight
	// This handler just marks the annotation as processed
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)
	return nil
}

// handleCanaryByHeader handles the nginx.ingress.kubernetes.io/canary-by-header annotation.
// When this header is present with any value (except "never"), traffic routes to the canary.
// When combined with canary-by-header-value or canary-by-header-pattern, more specific
// matching is performed.
//
// This maps directly to Gateway API HTTPHeaderMatch.
func (p Provider) handleCanaryByHeader(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	headerName := annotationCtx.Value
	if headerName == "" {
		return fmt.Errorf("canary-by-header requires a header name")
	}

	// Check if there's a specific value or pattern to match
	headerValue := ingressCtx.Ingress.Annotations[AnnotationCanaryByHeaderValue]
	headerPattern := ingressCtx.Ingress.Annotations[AnnotationCanaryByHeaderPattern]

	var headerMatch gatewayapi_v1.HTTPHeaderMatch

	switch {
	case headerPattern != "":
		// Validate the pattern is valid regex
		if _, err := regexp.Compile(headerPattern); err != nil {
			return fmt.Errorf("invalid canary-by-header-pattern regex %q: %w", headerPattern, err)
		}

		headerMatch = gatewayapi_v1.HTTPHeaderMatch{
			Type:  ptr.To(gatewayapi_v1.HeaderMatchRegularExpression),
			Name:  gatewayapi_v1.HTTPHeaderName(headerName),
			Value: headerPattern,
		}
	case headerValue != "":
		// Exact value match
		headerMatch = gatewayapi_v1.HTTPHeaderMatch{
			Type:  ptr.To(gatewayapi_v1.HeaderMatchExact),
			Name:  gatewayapi_v1.HTTPHeaderName(headerName),
			Value: headerValue,
		}
	default:
		// Any value except "never" - we'll use regex to match anything that's not "never"
		// In practice, most implementations just check for presence with value != "never"
		// Gateway API requires a value, so we'll use a regex that matches common patterns
		headerMatch = gatewayapi_v1.HTTPHeaderMatch{
			Type:  ptr.To(gatewayapi_v1.HeaderMatchRegularExpression),
			Name:  gatewayapi_v1.HTTPHeaderName(headerName),
			Value: "^(?!never$).*$", // Regex: match anything except exactly "never"
		}

		annotationCtx.RegisterIssue(resources.NewIssue(
			resources.IssueNGINXCanaryHeaderApproximated,
			fmt.Errorf("header=%s", headerName),
		))
	}

	// Add the header match to all rules in the route
	for i := range routeCtx.Spec.Rules {
		for j := range routeCtx.Spec.Rules[i].Matches {
			routeCtx.Spec.Rules[i].Matches[j].Headers = append(
				routeCtx.Spec.Rules[i].Matches[j].Headers,
				headerMatch,
			)
		}
	}

	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// handleCanaryByHeaderValue handles the nginx.ingress.kubernetes.io/canary-by-header-value annotation.
// This is processed as part of handleCanaryByHeader.
func (p Provider) handleCanaryByHeaderValue(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Processed in handleCanaryByHeader
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)
	return nil
}

// handleCanaryByHeaderPattern handles the nginx.ingress.kubernetes.io/canary-by-header-pattern annotation.
// This is processed as part of handleCanaryByHeader.
func (p Provider) handleCanaryByHeaderPattern(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Processed in handleCanaryByHeader
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)
	return nil
}

// handleCanaryByCookie handles the nginx.ingress.kubernetes.io/canary-by-cookie annotation.
// Gateway API doesn't have native cookie-based routing, so this is not supported.
func (p Provider) handleCanaryByCookie(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXCanaryByCookieNotSupported, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}
