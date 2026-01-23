package resources

type IssueLibraryEntry struct {
	Code           IssueCode
	Level          IssueLevel
	Description    string
	Recommendation string
}

type IssueCode int

// revive:disable:exported
const (
	IssueUnsupportBackendProtocol IssueCode = iota
	IssueCreatingBackendTLSPolicy
	IssueAppGWTrustedRootCertificatesNotSupported
	IssueInvalidAnnotationValue
	IssueHealthCheckConflict
	IssueNoGateway
	IssueCouldNotFindRoute
	IssueCouldNotFindAppGWRewriteCustomResource
	IssueRewriteRuleSetHasErrors
	IssueRewriteRuleSetRuleSequenceNotSupported
	IssueRewriteRuleSetConditionsNotSupported
	IssueRewriteRuleSetRerouteNotSupported
	IssueFrontendTLSPolicyProfileConflict
	IssueFrontendTLSPolicyProfileCipherWarning
	IssueNoHTTPSListenersForSSLProfile
	IssueUnsupportedAnnotationGeneric
	IssueHostnameExtensionsNotSupportedForHTTPS
	IssueWAFPotentialIncompatibility
)

type IssueLevel string

const (
	IssueLevelNotSupported IssueLevel = "NotSupported"
	IssueLevelWarning      IssueLevel = "Warning"
	IssueLevelError        IssueLevel = "Error"
)

func (l IssueLevel) MigrationStatus() MigrationStatus {
	switch l {
	case IssueLevelNotSupported:
		return MigrationStatusNotSupported
	case IssueLevelWarning:
		return MigrationStatusWarning
	default:
		return MigrationStatusError
	}
}

type Issue struct {
	Code  IssueCode
	Error error
}

func (i Issue) Entry() IssueLibraryEntry {
	return IssueLibrary[i.Code]
}

func NewIssue(code IssueCode, err error) Issue {
	return Issue{
		Code:  code,
		Error: err,
	}
}

func (i Issue) MigrationStatus() MigrationStatus {
	return i.Entry().Level.MigrationStatus()
}

const RecommendationPleaseReviewTheErrorMessage = "Please review the error message on the related AGIC resources in the Migration Report."

var IssueLibrary = map[IssueCode]IssueLibraryEntry{
	IssueUnsupportBackendProtocol: {
		Code:           IssueUnsupportBackendProtocol,
		Level:          IssueLevelNotSupported,
		Description:    "The specified backend protocol is not supported for migration.",
		Recommendation: RecommendationPleaseReviewTheErrorMessage,
	},
	IssueFrontendTLSPolicyProfileConflict: {
		Code:           IssueFrontendTLSPolicyProfileConflict,
		Level:          IssueLevelError,
		Description:    "There are conflicting FrontendTLSPolicy SSL profiles targeting the same listener.",
		Recommendation: "Review the generated FrontendTLSPolicies and ensure they are appropriate.",
	},
	IssueCreatingBackendTLSPolicy: {
		Code:  IssueCreatingBackendTLSPolicy,
		Level: IssueLevelError,

		Description:    "There was an error creating a BackendTLSPolicy",
		Recommendation: RecommendationPleaseReviewTheErrorMessage,
	},
	IssueAppGWTrustedRootCertificatesNotSupported: {
		Code:        IssueAppGWTrustedRootCertificatesNotSupported,
		Level:       IssueLevelNotSupported,
		Description: "Application Gateway trusted root certificates are not supported for migration.",
		Recommendation: "To setup a trusted root certificate to be used by the Gateway to verify the backends TLS " +
			"certificate, you will need to create a Secret in the cluster containing the trusted root certificate " +
			"and configure it on a BackendTLSPolicy's caCertificateRef field manually.",
	},
	IssueInvalidAnnotationValue: {
		Code:           IssueInvalidAnnotationValue,
		Level:          IssueLevelError,
		Description:    "The annotation has an invalid value.",
		Recommendation: "Review the annotations value to ensure it is valid.",
	},
	IssueHealthCheckConflict: {
		Code:           IssueHealthCheckConflict,
		Level:          IssueLevelError,
		Description:    "There are conflicting health probe settings targeting the same backend service.",
		Recommendation: "Review the generated HealthCheckPolicies and ensure they are appropriate.",
	},
	IssueNoGateway: {
		Code:           IssueNoGateway,
		Level:          IssueLevelError,
		Description:    "No Gateway was generated during migration.",
		Recommendation: "Ensure that the Ingress(es) were correctly configured and given to the migration tool.",
	},
	IssueCouldNotFindRoute: {
		Code:           IssueCouldNotFindRoute,
		Level:          IssueLevelError,
		Description:    "Could not find converted HTTPRoute when processing WAF policy annotation.",
		Recommendation: "This may indicate a bug with the migration tool, please review the tool logs for errors.",
	},
	IssueCouldNotFindAppGWRewriteCustomResource: {
		Code:        IssueCouldNotFindAppGWRewriteCustomResource,
		Level:       IssueLevelError,
		Description: "Could not find AzureApplicationGatewayRewrite object when processing rewrite rule set annotation.",
		Recommendation: "Ensure that the referenced AzureApplicationGatewayRewrite custom resource exists " +
			"and that it was given to the migration tool.",
	},
	IssueRewriteRuleSetHasErrors: {
		Code:           IssueRewriteRuleSetHasErrors,
		Level:          IssueLevelError,
		Description:    "The referenced AzureApplicationGatewayRewrite object has errors.",
		Recommendation: RecommendationPleaseReviewTheErrorMessage,
	},
	IssueRewriteRuleSetRuleSequenceNotSupported: {
		Code:        IssueRewriteRuleSetRuleSequenceNotSupported,
		Level:       IssueLevelError,
		Description: "Rewrite rule sequences are not supported in Application Gateway for Containers.",
		Recommendation: "Rewrite rules will not be applied in a specific order, please review the  " +
			"on the generated Application Gateway for Containers resources to ensure they meet your requirements.",
	},
	IssueRewriteRuleSetConditionsNotSupported: {
		Code:        IssueRewriteRuleSetConditionsNotSupported,
		Level:       IssueLevelWarning,
		Description: "Rewrite rule conditions are not supported in Application Gateway for Containers.",
		Recommendation: "Rewrite rules will be applied without conditions, please review the Filters on the generated " +
			"HTTPRoutes to ensure they meet your requirements.",
	},
	IssueRewriteRuleSetRerouteNotSupported: {
		Code:        IssueRewriteRuleSetRerouteNotSupported,
		Level:       IssueLevelWarning,
		Description: "URL reroute on rewrite rules is not supported in Application Gateway for Containers.",
		Recommendation: "Path rewrites will be applied but requests will not be rerouted to a different backend. " +
			"Please review the generated HTTPRoutes to ensure they meet your requirements.",
	},
	IssueFrontendTLSPolicyProfileCipherWarning: {
		Code:        IssueFrontendTLSPolicyProfileCipherWarning,
		Level:       IssueLevelWarning,
		Description: "AGC SSL profiles are not exact matches for Application Gateway SSL profiles.",
		Recommendation: "Review the SSL Profile on the FrontendTLSPolicy and ensure it aligns with your requirements. " +
			"Read more at https://learn.microsoft.com/en-us/azure/application-gateway/for-containers/tls-policy?tabs=tls-policy-gateway-api#predefined-tls-policy",
	},
	IssueNoHTTPSListenersForSSLProfile: {
		Code:        IssueNoHTTPSListenersForSSLProfile,
		Level:       IssueLevelError,
		Description: "No HTTPS listeners were found for the Ingress's SSL profile annotation.",
		Recommendation: "Ensure that the Ingress is correctly configured with a TLS section " +
			"to use the SSL profile annotation.",
	},
	IssueUnsupportedAnnotationGeneric: {
		Code:        IssueUnsupportedAnnotationGeneric,
		Level:       IssueLevelNotSupported,
		Description: "The annotation is not supported for migration.",
		Recommendation: "Please review the annotation and its value to determine if manual changes to the migrated " +
			"Application Gateway for Containers resources are required.",
	},
	IssueHostnameExtensionsNotSupportedForHTTPS: {
		Code:        IssueHostnameExtensionsNotSupportedForHTTPS,
		Level:       IssueLevelNotSupported,
		Description: "The hostname extension annotation is not supported for migration with HTTPS listeners.",
		Recommendation: "Review the Gateway HTTPS listener(s) and related HTTPRoutes are correct, and make modifications as necessary; you may consider using a wildcard " +
			"hostname on the Gateway.",
	},
	IssueWAFPotentialIncompatibility: {
		Code:  IssueWAFPotentialIncompatibility,
		Level: IssueLevelWarning,
		Description: "Application Gateway and Application Gateway for Containers has different support for WAF policy Rule Sets. " +
			"This tool does not have access to your WAF policy to verify if its Rule Sets are supported.",
		Recommendation: "AGC supports WAF policies using Default Rule Sets 2.1 and Bot Manager Rulesets 1.0 or greater. " +
			"Please verify your WAF policy rulesets in the Azure Portal under your WAF policy resource. " +
			"Read more at https://aka.ms/agc/waf.",
	},
}
