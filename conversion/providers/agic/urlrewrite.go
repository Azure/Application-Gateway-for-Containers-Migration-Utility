package agic

import (
	"errors"
	"fmt"
	"strings"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

const (
	agicHeaderDelete = "delete"
	agicHeaderSet    = "set"
)

func (p Provider) handleBackendPathPrefix(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	insertPathRewrite := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.URLRewrite == nil {
			filter.URLRewrite = &gatewayapi_v1.HTTPURLRewriteFilter{}
		}

		if filter.URLRewrite.Path == nil {
			filter.URLRewrite.Path = &gatewayapi_v1.HTTPPathModifier{}
		}

		filter.URLRewrite.Path.ReplacePrefixMatch = &annotationCtx.Value
		filter.URLRewrite.Path.Type = gatewayapi_v1.PrefixMatchHTTPPathModifier

		return filter
	}

	return handleHTTPFilter(
		output,
		routeCtx,
		annotationCtx,
		gatewayapi_v1.HTTPRouteFilterURLRewrite,
		insertPathRewrite,
	)
}

func (p Provider) handleBackendHostnameRewrite(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	insertHostnameFilter := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
		if filter.URLRewrite == nil {
			filter.URLRewrite = &gatewayapi_v1.HTTPURLRewriteFilter{}
		}

		filter.URLRewrite.Hostname = ptr.To(gatewayapi_v1.PreciseHostname(annotationCtx.Value))

		return filter
	}

	return handleHTTPFilter(
		output,
		routeCtx,
		annotationCtx,
		gatewayapi_v1.HTTPRouteFilterURLRewrite,
		insertHostnameFilter,
	)
}

func (p Provider) handleRewriteRuleSetCustomResource(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	var errs error

	rulesetNN := types.NamespacedName{
		Name:      annotationCtx.Value,
		Namespace: ingressCtx.Ingress.Namespace,
	}

	rulesetCtx, ok := p.input.AppGWRewrites[rulesetNN]
	if !ok {
		err := fmt.Errorf("AzureApplicationGatewayRewrite %s not found", rulesetNN)
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueCouldNotFindAppGWRewriteCustomResource, err))

		return err
	}

	rules := rulesetCtx.Object.Spec.RewriteRules

	var firstRuleSequence *int

	ruleSequenceWarningIssued := false

	for _, rule := range rules {
		if len(rule.Conditions) > 0 {
			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueRewriteRuleSetConditionsNotSupported, nil))
			p.log.Warn("rewrite rule conditions are not supported, rewrites will be applied without conditions")
		}

		if firstRuleSequence == nil {
			seq := rule.RuleSequence
			firstRuleSequence = &seq
		} else if !ruleSequenceWarningIssued && rule.RuleSequence != *firstRuleSequence {
			ruleSequenceWarningIssued = true

			annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueRewriteRuleSetRuleSequenceNotSupported, nil))
			p.log.Warn("Rewrite rule sequences are not supported, rewrites will not be applied in a specific order")
		}

		for _, requestHeaderCfg := range rule.Actions.RequestHeaderConfigurations {
			if err := handleRequestHeaderModifier(
				output,
				routeCtx,
				ingressCtx,
				annotationCtx,
				requestHeaderCfg,
			); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		for _, responseHeaderCfg := range rule.Actions.ResponseHeaderConfigurations {
			if err := handleResponseHeaderModifier(
				output,
				routeCtx,
				annotationCtx,
				responseHeaderCfg,
			); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		if urlCfg := rule.Actions.UrlConfiguration; urlCfg != nil {
			if urlCfg.Reroute {
				annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueRewriteRuleSetRerouteNotSupported, nil))
				p.log.Warn("URL reroute on rewrite rules is not supported, path rewrite will be applied but requests will not be rerouted to a different backend")
			}

			if urlCfg.ModifiedPath != "" {
				insertPathRewrite := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
					if filter.URLRewrite == nil {
						filter.URLRewrite = &gatewayapi_v1.HTTPURLRewriteFilter{}
					}

					if filter.URLRewrite.Path == nil {
						filter.URLRewrite.Path = &gatewayapi_v1.HTTPPathModifier{}
					}

					filter.URLRewrite.Path.Type = gatewayapi_v1.FullPathHTTPPathModifier
					// TODO: verify this is correct, I'm not sure what values AGIC/AppGW actually takes, it may take regex
					filter.URLRewrite.Path.ReplaceFullPath = &urlCfg.ModifiedPath

					return filter
				}

				if err := handleHTTPFilter(
					output,
					routeCtx,
					annotationCtx,
					gatewayapi_v1.HTTPRouteFilterURLRewrite,
					insertPathRewrite,
				); err != nil {
					errs = errors.Join(errs, err)
				}
			}
		}
	}

	if errs == nil && annotationCtx.Status() == resources.MigrationStatusNotStarted {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
	}

	return errs
}

func handleRequestHeaderModifier(
	output resources.AGCResourceGraph,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
	requestHeaderCfg appgwrewrite.HeaderConfiguration,
) error {
	switch strings.ToLower(requestHeaderCfg.ActionType) {
	case agicHeaderSet:
		insertHeaderFilter := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
			if filter.RequestHeaderModifier == nil {
				filter.RequestHeaderModifier = &gatewayapi_v1.HTTPHeaderFilter{}
			}

			header := gatewayapi_v1.HTTPHeader{
				Name:  gatewayapi_v1.HTTPHeaderName(requestHeaderCfg.HeaderName),
				Value: requestHeaderCfg.HeaderValue,
			}
			filter.RequestHeaderModifier.Set = append(filter.RequestHeaderModifier.Set, header)

			return filter
		}

		return handleHTTPFilter(
			output,
			routeCtx,
			annotationCtx,
			gatewayapi_v1.HTTPRouteFilterRequestHeaderModifier,
			insertHeaderFilter,
		)
	case agicHeaderDelete:
		insertHeaderFilter := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
			if filter.RequestHeaderModifier == nil {
				filter.RequestHeaderModifier = &gatewayapi_v1.HTTPHeaderFilter{}
			}

			filter.RequestHeaderModifier.Remove = append(filter.RequestHeaderModifier.Remove, requestHeaderCfg.HeaderName)

			return filter
		}

		return handleHTTPFilter(
			output,
			routeCtx,
			annotationCtx,
			gatewayapi_v1.HTTPRouteFilterRequestHeaderModifier,
			insertHeaderFilter,
		)
	default:
		return fmt.Errorf("unsupported rewrite request header action type: %s", requestHeaderCfg.ActionType)
	}
}

func handleResponseHeaderModifier(
	output resources.AGCResourceGraph,
	routeCtx *conversion.HTTPRouteContext,
	annotationCtx *resources.IngressAnnotationContext,
	responseHeaderCfg appgwrewrite.HeaderConfiguration,
) error {
	switch strings.ToLower(responseHeaderCfg.ActionType) {
	case agicHeaderSet:
		insertHeaderFilter := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
			if filter.ResponseHeaderModifier == nil {
				filter.ResponseHeaderModifier = &gatewayapi_v1.HTTPHeaderFilter{}
			}

			header := gatewayapi_v1.HTTPHeader{
				Name:  gatewayapi_v1.HTTPHeaderName(responseHeaderCfg.HeaderName),
				Value: responseHeaderCfg.HeaderValue,
			}
			filter.ResponseHeaderModifier.Set = append(filter.ResponseHeaderModifier.Set, header)

			return filter
		}

		return handleHTTPFilter(
			output,
			routeCtx,
			annotationCtx,
			gatewayapi_v1.HTTPRouteFilterResponseHeaderModifier,
			insertHeaderFilter,
		)
	case agicHeaderDelete:
		insertHeaderFilter := func(filter gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter {
			if filter.ResponseHeaderModifier == nil {
				filter.ResponseHeaderModifier = &gatewayapi_v1.HTTPHeaderFilter{}
			}

			filter.ResponseHeaderModifier.Remove = append(filter.ResponseHeaderModifier.Remove, responseHeaderCfg.HeaderName)

			return filter
		}

		return handleHTTPFilter(
			output,
			routeCtx,
			annotationCtx,
			gatewayapi_v1.HTTPRouteFilterResponseHeaderModifier,
			insertHeaderFilter,
		)
	default:
		return fmt.Errorf("unsupported rewrite response header action type: %s", responseHeaderCfg.ActionType)
	}
}

func handleHTTPFilter(
	_ resources.AGCResourceGraph,
	routeCtx *conversion.HTTPRouteContext,
	annotationCtx *resources.IngressAnnotationContext,
	filterType gatewayapi_v1.HTTPRouteFilterType,
	modifyFn func(gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter,
) error {
	insertOrModifyHTTPRouteFilter(routeCtx, filterType, modifyFn)
	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}

func insertOrModifyHTTPRouteFilter(
	routeCtx *conversion.HTTPRouteContext,
	filterType gatewayapi_v1.HTTPRouteFilterType,
	modifyFn func(gatewayapi_v1.HTTPRouteFilter) gatewayapi_v1.HTTPRouteFilter,
) {
	rules := routeCtx.Spec.Rules

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
