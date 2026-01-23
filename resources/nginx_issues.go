package resources

// TODO: this needs to be made part of the provider interface
const (
	// NGINX-specific issues
	IssueNGINXAuthTLSNotFullySupported IssueCode = 1000 + iota
	IssueNGINXAffinityTypeNotSupported
	IssueNGINXAffinityModeNotSupported
	IssueNGINXAffinityCanaryBehaviorNotSupported
	IssueNGINXUseRegexLimitedSupport
	IssueNGINXConfigurationSnippetNotSupported
	IssueNGINXServerSnippetNotSupported
	IssueNGINXModSecurityConversion
	IssueNGINXSSLPolicyConversion
	IssueNGINXDefaultBackendNotSupported
	IssueNGINXProxySettingsPartialSupport
	IssueNGINXLoadBalanceNotSupported
	IssueNGINXRedirectURLInvalid
	IssueNGINXRedirectCodeInvalid
	IssueNGINXFromToWWWRedirectNoHost
	IssueNGINXFromToWWWRedirectPartial
	IssueNGINXCanaryWeightRequiresManualMerge
	IssueNGINXCanaryHeaderApproximated
	IssueNGINXCanaryByCookieNotSupported
	IssueNGINXModSecurityTransactionID
	IssueNGINXRewriteTargetCaptureGroups
	IssueGRPCNotSupportedByTool
)

func init() {
	registerNGINXIssues(IssueLibrary)
}

func registerNGINXIssues(issueLibrary map[IssueCode]IssueLibraryEntry) {
	// NGINX-specific issues
	issueLibrary[IssueNGINXAuthTLSNotFullySupported] = IssueLibraryEntry{
		Code:        IssueNGINXAuthTLSNotFullySupported,
		Level:       IssueLevelWarning,
		Description: "NGINX auth-tls-* annotations have partial support in AGC.",
		Recommendation: "Client certificate verification in AGC is configured via FrontendTLSPolicy. " +
			"Review the generated policy and manually adjust the verification settings as needed.",
	}

	issueLibrary[IssueNGINXAffinityTypeNotSupported] = IssueLibraryEntry{
		Code:        IssueNGINXAffinityTypeNotSupported,
		Level:       IssueLevelNotSupported,
		Description: "Only cookie-based session affinity is supported in AGC.",
		Recommendation: "AGC uses managed cookie-based session affinity. Other affinity types like 'ip_hash' are not supported. " +
			"Review your session affinity requirements and consider using cookie-based affinity.",
	}

	issueLibrary[IssueNGINXAffinityModeNotSupported] = IssueLibraryEntry{
		Code:        IssueNGINXAffinityModeNotSupported,
		Level:       IssueLevelWarning,
		Description: "NGINX affinity-mode annotation is not directly supported in AGC.",
		Recommendation: "AGC uses managed cookie-based affinity. The 'balanced' mode behavior may differ. " +
			"Review your session affinity requirements and test the migrated configuration.",
	}

	issueLibrary[IssueNGINXAffinityCanaryBehaviorNotSupported] = IssueLibraryEntry{
		Code:        IssueNGINXAffinityCanaryBehaviorNotSupported,
		Level:       IssueLevelNotSupported,
		Description: "NGINX affinity-canary-behavior annotation is not supported in AGC.",
		Recommendation: "Canary affinity behavior must be configured differently in AGC. " +
			"Review AGC documentation for traffic splitting options.",
	}

	issueLibrary[IssueNGINXUseRegexLimitedSupport] = IssueLibraryEntry{
		Code:        IssueNGINXUseRegexLimitedSupport,
		Level:       IssueLevelWarning,
		Description: "NGINX use-regex annotation has limited support in Gateway API.",
		Recommendation: "Gateway API supports RegularExpression path matching. Complex regex patterns may need adjustment. " +
			"Review the generated HTTPRoute path matches.",
	}

	issueLibrary[IssueNGINXConfigurationSnippetNotSupported] = IssueLibraryEntry{
		Code:        IssueNGINXConfigurationSnippetNotSupported,
		Level:       IssueLevelNotSupported,
		Description: "NGINX configuration-snippet annotation cannot be migrated to AGC.",
		Recommendation: "Configuration snippets contain raw NGINX config that has no AGC equivalent. " +
			"Review the snippet content and manually implement equivalent functionality using AGC features.",
	}

	issueLibrary[IssueNGINXServerSnippetNotSupported] = IssueLibraryEntry{
		Code:        IssueNGINXServerSnippetNotSupported,
		Level:       IssueLevelNotSupported,
		Description: "NGINX server-snippet annotation cannot be migrated to AGC.",
		Recommendation: "Server snippets contain raw NGINX config that has no AGC equivalent. " +
			"Review the snippet content and manually implement equivalent functionality using AGC features.",
	}

	issueLibrary[IssueNGINXModSecurityConversion] = IssueLibraryEntry{
		Code:  IssueNGINXModSecurityConversion,
		Level: IssueLevelWarning,
		Description: "NGINX ModSecurity annotations are converted to AGC WebApplicationFirewallPolicy. " +
			"OWASP Core Rule Set configuration may differ between ModSecurity and AGC WAF.",
		Recommendation: "A WebApplicationFirewallPolicy CR has been generated. You must create an Azure WAF Policy " +
			"with appropriate rule sets and reference its ARM ID in the generated policy. " +
			"Read more at https://aka.ms/agc/waf.",
	}

	issueLibrary[IssueNGINXSSLPolicyConversion] = IssueLibraryEntry{
		Code:        IssueNGINXSSLPolicyConversion,
		Level:       IssueLevelWarning,
		Description: "NGINX SSL cipher and protocol annotations are converted to AGC FrontendTLSPolicy.",
		Recommendation: "AGC uses predefined TLS policies. Custom cipher suites are not supported. " +
			"Review the generated FrontendTLSPolicy and select an appropriate predefined policy.",
	}

	issueLibrary[IssueNGINXDefaultBackendNotSupported] = IssueLibraryEntry{
		Code:           IssueNGINXDefaultBackendNotSupported,
		Level:          IssueLevelNotSupported,
		Description:    "NGINX default-backend annotation is not directly supported in AGC.",
		Recommendation: "Configure a catch-all HTTPRoute rule with a PathPrefix of '/' to handle unmatched requests.",
	}

	issueLibrary[IssueNGINXProxySettingsPartialSupport] = IssueLibraryEntry{
		Code:        IssueNGINXProxySettingsPartialSupport,
		Level:       IssueLevelWarning,
		Description: "NGINX proxy timeout and buffer settings have partial support in AGC.",
		Recommendation: "AGC RoutePolicy supports route timeouts. Other proxy settings may not have direct equivalents. " +
			"Review the generated RoutePolicy and AGC documentation.",
	}

	issueLibrary[IssueNGINXLoadBalanceNotSupported] = IssueLibraryEntry{
		Code:        IssueNGINXLoadBalanceNotSupported,
		Level:       IssueLevelNotSupported,
		Description: "The specified load balancing algorithm is not supported by AGC.",
		Recommendation: "AGC uses round-robin load balancing by default. Algorithms like 'ip_hash', 'least_conn', " +
			"or 'ewma' are not available. Consider using session affinity (cookie-based) if sticky sessions are required.",
	}

	issueLibrary[IssueNGINXRedirectURLInvalid] = IssueLibraryEntry{
		Code:        IssueNGINXRedirectURLInvalid,
		Level:       IssueLevelNotSupported,
		Description: "The redirect URL specified in the annotation is invalid or cannot be parsed.",
		Recommendation: "Ensure the redirect URL is a valid URL with proper scheme (http/https), hostname, and optional path. " +
			"Example: https://example.com/new-path",
	}

	issueLibrary[IssueNGINXRedirectCodeInvalid] = IssueLibraryEntry{
		Code:        IssueNGINXRedirectCodeInvalid,
		Level:       IssueLevelNotSupported,
		Description: "The redirect status code specified is not valid or not supported.",
		Recommendation: "Use valid HTTP redirect status codes. For permanent redirects: 301 or 308. " +
			"For temporary redirects: 302, 303, or 307.",
	}

	issueLibrary[IssueNGINXFromToWWWRedirectNoHost] = IssueLibraryEntry{
		Code:           IssueNGINXFromToWWWRedirectNoHost,
		Level:          IssueLevelNotSupported,
		Description:    "Cannot apply from-to-www-redirect because no host is specified in the Ingress rules.",
		Recommendation: "Ensure the Ingress has at least one rule with a host specified to enable www redirection.",
	}

	issueLibrary[IssueNGINXFromToWWWRedirectPartial] = IssueLibraryEntry{
		Code:  IssueNGINXFromToWWWRedirectPartial,
		Level: IssueLevelWarning,
		Description: "The from-to-www-redirect annotation is partially supported. In NGINX, this creates a " +
			"separate server block for the non-www domain that redirects to www. In Gateway API, you may need " +
			"to create a separate HTTPRoute for the non-www hostname.",
		Recommendation: "Create an additional HTTPRoute with the non-www hostname that " +
			"redirects to the www hostname using a RequestRedirect filter.",
	}

	issueLibrary[IssueNGINXCanaryWeightRequiresManualMerge] = IssueLibraryEntry{
		Code:  IssueNGINXCanaryWeightRequiresManualMerge,
		Level: IssueLevelWarning,
		Description: "NGINX canary weight-based traffic splitting requires the canary and non-canary backends " +
			"to be in the same HTTPRoute rule. Since each Ingress creates a separate HTTPRoute, manual merging " +
			"is required to achieve the same traffic splitting behavior.",
		Recommendation: "Merge the canary and non-canary HTTPRoutes by adding both backends to the same rule " +
			"with their respective weights. The canary weight has been set on the generated route.",
	}

	issueLibrary[IssueNGINXCanaryHeaderApproximated] = IssueLibraryEntry{
		Code:  IssueNGINXCanaryHeaderApproximated,
		Level: IssueLevelWarning,
		Description: "NGINX canary-by-header without a specific value routes traffic when the header is present " +
			"with any value except 'never'. This has been approximated using a regex header match.",
		Recommendation: "Review the generated HTTPRoute header match. If precise behavior is required, " +
			"consider using canary-by-header-value for exact matching.",
	}

	issueLibrary[IssueNGINXCanaryByCookieNotSupported] = IssueLibraryEntry{
		Code:        IssueNGINXCanaryByCookieNotSupported,
		Level:       IssueLevelNotSupported,
		Description: "Gateway API does not support cookie-based routing. The canary-by-cookie annotation cannot be migrated.",
		Recommendation: "Consider using header-based canary routing (canary-by-header) instead, or implement " +
			"cookie-based routing at the application level.",
	}

	issueLibrary[IssueNGINXModSecurityTransactionID] = IssueLibraryEntry{
		Code:  IssueNGINXModSecurityTransactionID,
		Level: IssueLevelWarning,
		Description: "Application Gateway for Containers WAF automatically uses 'trackingId' in Access and Firewall logs " +
			"for request correlation. The modsecurity-transaction-id annotation is not needed.",
		Recommendation: "Use the 'trackingId' field in AGC Access and Firewall logs to correlate WAF events with requests. " +
			"See https://aka.ms/agc/logs for details.",
	}

	issueLibrary[IssueNGINXRewriteTargetCaptureGroups] = IssueLibraryEntry{
		Code:  IssueNGINXRewriteTargetCaptureGroups,
		Level: IssueLevelWarning,
		Description: "NGINX rewrite-target uses regex capture groups (e.g., /$1, /$2). Gateway API's URLRewrite filter " +
			"only supports 'ReplaceFullPath' or 'ReplacePrefixMatch' - it cannot use capture groups for dynamic replacement. " +
			"The regex path has been converted to a prefix match, and the rewrite approximates stripping that prefix.",
		Recommendation: "Review the generated HTTPRoute. For the common pattern of stripping a path prefix " +
			"(e.g., /api/v1/(.*) rewritten to /$1), the conversion should work correctly.",
	}

	issueLibrary[IssueGRPCNotSupportedByTool] = IssueLibraryEntry{
		Code:           IssueGRPCNotSupportedByTool,
		Level:          IssueLevelError,
		Description:    "The migration tool does not the creation of GRPCRoutes",
		Recommendation: "Replace the generated HTTPRoute create a Gateway API GRPCRoute object that targets your backend.",
	}
}
