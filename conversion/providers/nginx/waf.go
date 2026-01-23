package nginx

import (
	"strconv"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

// ========== ModSecurity/WAF Handlers ==========

func (p Provider) handleEnableModSecurity(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	enabled, err := strconv.ParseBool(annotationCtx.Value)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))
		return err
	}

	if !enabled {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	}

	// Create a WAF policy placeholder for the route
	// User will need to configure the actual Azure WAF Policy ARM ID
	policy := output.GetOrCreateWAFPolicyForRoute(k8snames.NamespacedName(routeCtx.HTTPRoute))
	policy.Spec.WebApplicationFirewallPolicy.ID = "<AZURE_WAF_POLICY_ARM_ID>"

	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXModSecurityConversion, nil))
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleEnableOWASPCoreRules(
	output resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Similar to enable-modsecurity - creates a WAF policy placeholder
	enabled, err := strconv.ParseBool(annotationCtx.Value)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))
		return err
	}

	if !enabled {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	}

	policy := output.GetOrCreateWAFPolicyForRoute(k8snames.NamespacedName(routeCtx.HTTPRoute))
	if policy.Spec.WebApplicationFirewallPolicy.ID == "" {
		policy.Spec.WebApplicationFirewallPolicy.ID = "<AZURE_WAF_POLICY_ARM_ID>"
	}

	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXModSecurityConversion, nil))
	annotationCtx.AddDestination(resources.NewK8sResourceID(policy))
	annotationCtx.SetStatus(resources.MigrationStatusWarning)

	return nil
}

func (p Provider) handleModSecuritySnippet(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// ModSecurity snippets are custom rules that can't be directly migrated
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXModSecurityConversion, nil))
	annotationCtx.SetStatus(resources.MigrationStatusNotSupported)

	return nil
}

// handleModSecurityTransactionID handles the modsecurity-transaction-id annotation.
// AGC WAF automatically provides request correlation via the 'trackingId' field in logs.
func (p Provider) handleModSecurityTransactionID(
	_ resources.AGCResourceGraph,
	_ *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// AGC WAF uses trackingId automatically - inform the user via a warning note.
	// The RegisterIssue call sets status to Warning, which is appropriate for this
	// informational message that the user should be aware of.
	annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueNGINXModSecurityTransactionID, nil))
	return nil
}
