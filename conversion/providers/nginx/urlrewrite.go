package nginx

import (
	"regexp"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// captureGroupPattern matches regex capture group references like $1, $2, etc.
var captureGroupPattern = regexp.MustCompile(`\$\d+`)

// regexMetaCharsPattern matches common regex metacharacters that indicate regex path patterns
var regexMetaCharsPattern = regexp.MustCompile(`[().*+?|\\^$\[\]]`)

// ========== URL Rewriting Handlers ==========

func (p Provider) handleRewriteTarget(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	rewriteTarget := annotationCtx.Value
	route := routeCtx.HTTPRoute

	// Check if the rewrite-target uses regex capture groups (e.g., /$1, /$2)
	hasCaptureGroups := captureGroupPattern.MatchString(rewriteTarget)

	// Process each rule and selectively apply URL rewrite
	for ruleIdx := range route.Spec.Rules {
		rule := &route.Spec.Rules[ruleIdx]

		// Determine if this rule's path pattern is a regex pattern
		// (contains regex metacharacters like (), .*, etc.)
		isRegexPath := false

		var pathValue string

		if len(rule.Matches) > 0 && rule.Matches[0].Path != nil && rule.Matches[0].Path.Value != nil {
			pathValue = *rule.Matches[0].Path.Value
			isRegexPath = regexMetaCharsPattern.MatchString(pathValue)
		}

		// If rewrite-target has capture groups but this path is not a regex pattern,
		// skip adding the rewrite filter to this rule (the capture groups don't apply)
		if hasCaptureGroups && !isRegexPath {
			continue
		}

		// Determine the replacement path
		replacementPath := rewriteTarget

		if hasCaptureGroups && isRegexPath {
			// For regex patterns with capture groups, we need to determine the prefix
			// being stripped. Common pattern: /prefix(/|$)(.*) with rewrite /$2 means
			// strip /prefix and keep the rest.
			// Gateway API doesn't support capture groups, so we translate to prefix replacement.
			replacementPath = determineReplacementPath(pathValue, rewriteTarget)
		}

		// Add or update the URLRewrite filter for this rule
		found := false

		for filterIdx := range rule.Filters {
			if rule.Filters[filterIdx].Type == gatewayapi_v1.HTTPRouteFilterURLRewrite {
				if rule.Filters[filterIdx].URLRewrite == nil {
					rule.Filters[filterIdx].URLRewrite = &gatewayapi_v1.HTTPURLRewriteFilter{}
				}

				if rule.Filters[filterIdx].URLRewrite.Path == nil {
					rule.Filters[filterIdx].URLRewrite.Path = &gatewayapi_v1.HTTPPathModifier{}
				}

				rule.Filters[filterIdx].URLRewrite.Path.ReplacePrefixMatch = &replacementPath
				rule.Filters[filterIdx].URLRewrite.Path.Type = gatewayapi_v1.PrefixMatchHTTPPathModifier
				found = true

				break
			}
		}

		if !found {
			rule.Filters = append(rule.Filters, gatewayapi_v1.HTTPRouteFilter{
				Type: gatewayapi_v1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayapi_v1.HTTPURLRewriteFilter{
					Path: &gatewayapi_v1.HTTPPathModifier{
						Type:               gatewayapi_v1.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: &replacementPath,
					},
				},
			})
		}

		// Update the path match type to PathPrefix for Gateway API compatibility
		// since regex paths are not fully supported and we're converting to prefix matching
		if isRegexPath && len(rule.Matches) > 0 && rule.Matches[0].Path != nil {
			// Extract the prefix from the regex pattern (before any regex metacharacters)
			prefix := extractPrefixFromRegex(pathValue)
			prefixType := gatewayapi_v1.PathMatchPathPrefix
			rule.Matches[0].Path.Type = &prefixType
			rule.Matches[0].Path.Value = &prefix
		}
	}

	if hasCaptureGroups {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXRewriteTargetCaptureGroups, nil))
		annotationCtx.SetStatus(resources.MigrationStatusWarning)
	} else {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
	}

	return nil
}

// determineReplacementPath converts a NGINX rewrite-target with capture groups
// to a Gateway API compatible replacement path.
// For example: pattern "/v2(/|$)(.*)" with target "/$2" -> replacement "/"
func determineReplacementPath(_, rewriteTarget string) string {
	// If rewrite-target is just a capture group reference like /$2, /$1,
	// the intent is typically to strip the prefix and keep the rest
	// In Gateway API, this translates to ReplacePrefixMatch with "/"
	if strings.TrimSpace(rewriteTarget) == "/$1" ||
		strings.TrimSpace(rewriteTarget) == "/$2" {
		return "/"
	}

	// For other patterns, strip capture group references and use the static part
	// e.g., "/api/$1" -> "/api/"
	replacement := captureGroupPattern.ReplaceAllString(rewriteTarget, "")
	if replacement == "" {
		replacement = "/"
	}

	return replacement
}

// extractPrefixFromRegex extracts the static prefix from a regex path pattern.
// For example: "/v2(/|$)(.*)" -> "/v2"
func extractPrefixFromRegex(regexPath string) string {
	// Find the first regex metacharacter and take everything before it
	loc := regexMetaCharsPattern.FindStringIndex(regexPath)
	if loc == nil {
		return regexPath // No regex chars found, return as-is
	}

	prefix := regexPath[:loc[0]]

	if prefix == "" {
		prefix = "/"
	}

	return prefix
}

func (p Provider) handleUseRegex(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// When use-regex is true, paths should be treated as regex patterns
	// Gateway API supports RegularExpression path match type
	useRegex, err := strconv.ParseBool(annotationCtx.Value)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))
		return err
	}

	if !useRegex {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	}

	// Convert existing path matches to regex type
	route := routeCtx.HTTPRoute
	regexType := gatewayapi_v1.PathMatchRegularExpression

	for i := range route.Spec.Rules {
		for j := range route.Spec.Rules[i].Matches {
			if route.Spec.Rules[i].Matches[j].Path != nil {
				route.Spec.Rules[i].Matches[j].Path.Type = &regexType
			}
		}
	}

	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXUseRegexLimitedSupport, nil))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleAppRoot(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// app-root causes a redirect from / to the specified path
	// Implement as a redirect filter on root path
	route := routeCtx.HTTPRoute

	// Add a rule to redirect root to app-root
	redirectRule := gatewayapi_v1.HTTPRouteRule{
		Matches: []gatewayapi_v1.HTTPRouteMatch{
			{
				Path: &gatewayapi_v1.HTTPPathMatch{
					Type:  ptr.To(gatewayapi_v1.PathMatchExact),
					Value: ptr.To("/"),
				},
			},
		},
		Filters: []gatewayapi_v1.HTTPRouteFilter{
			{
				Type: gatewayapi_v1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayapi_v1.HTTPRequestRedirectFilter{
					Path: &gatewayapi_v1.HTTPPathModifier{
						Type:            gatewayapi_v1.FullPathHTTPPathModifier,
						ReplaceFullPath: &annotationCtx.Value,
					},
					StatusCode: ptr.To(302),
				},
			},
		},
	}

	// Prepend the redirect rule so it takes precedence
	route.Spec.Rules = append([]gatewayapi_v1.HTTPRouteRule{redirectRule}, route.Spec.Rules...)

	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

// InsertOrModifyHTTPRouteFilter inserts or modifies an HTTPRoute filter.
func InsertOrModifyHTTPRouteFilter(
	routeCtx *conversion.HTTPRouteContext,
	filterType gatewayapi_v1.HTTPRouteFilterType,
	modifyFn func(gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter,
) {
	route := routeCtx.HTTPRoute
	rules := route.Spec.Rules

	for ruleIdx, rule := range rules {
		found := false

		for j := range rule.Filters {
			if rule.Filters[j].Type == filterType {
				rules[ruleIdx].Filters[j] = modifyFn(rule.Filters[j])
				found = true

				break
			}
		}

		if !found {
			newFilter := gatewayapi_v1.HTTPRouteFilter{Type: filterType}
			rules[ruleIdx].Filters = append(rules[ruleIdx].Filters, modifyFn(newFilter))
		}
	}
}

func (p Provider) handleXForwardedPrefix(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// X-Forwarded-Prefix annotation adds a header to the request sent to the backend
	// indicating the original path prefix before any rewrites.
	// In Gateway API, this is implemented using RequestHeaderModifier filter.
	prefixValue := annotationCtx.Value

	insertHeaderModifier := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.RequestHeaderModifier == nil {
			filter.RequestHeaderModifier = &gatewayapi_v1.HTTPHeaderFilter{}
		}
		// Add or set the X-Forwarded-Prefix header
		filter.RequestHeaderModifier.Set = append(filter.RequestHeaderModifier.Set,
			gatewayapi_v1.HTTPHeader{
				Name:  "X-Forwarded-Prefix",
				Value: prefixValue,
			},
		)

		return filter
	}

	InsertOrModifyHTTPRouteFilter(routeCtx, gatewayapi_v1.HTTPRouteFilterRequestHeaderModifier, insertHeaderModifier)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}
