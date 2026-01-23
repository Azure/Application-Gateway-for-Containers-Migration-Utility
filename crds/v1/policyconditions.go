package v1

type (
	// PolicyConditionType is the type of conditions used for different ALB related policy CRDs.
	PolicyConditionType string

	// PolicyConditionReason is the type of reason used for different ALB related policy CRDs.
	PolicyConditionReason string
)

// Policy Condition Types
const (
	PolicyConditionTypeAccepted     PolicyConditionType = "Accepted"
	PolicyConditionTypeResolvedRefs PolicyConditionType = "ResolvedRefs"
	PolicyConditionTypeProgrammed   PolicyConditionType = "Programmed"
	PolicyConditionTypeDeployment   PolicyConditionType = "Deployment"
)

// Reasons applicable to each condition type
const (
	PolicyReasonPending PolicyConditionReason = "Pending"
	PolicyReasonInvalid PolicyConditionReason = "Invalid"
)

// Accepted Condition Reasons
const (
	PolicyReasonAccepted             PolicyConditionReason = "Accepted"
	PolicyReasonOverrideNotSupported PolicyConditionReason = "OverrideNotSupported"
)

// ResolvedRefs Condition Reasons
const (
	PolicyReasonResolvedRefs             PolicyConditionReason = "ResolvedRefs"
	PolicyReasonConflicted               PolicyConditionReason = "Conflicted"
	PolicyReasonRefNotPermitted          PolicyConditionReason = "RefNotPermitted"
	PolicyReasonInvalidCertificateRef    PolicyConditionReason = "InvalidCertificateRef"
	PolicyReasonNoTargetReference        PolicyConditionReason = "NoTargetReference"
	PolicyReasonInvalidGroup             PolicyConditionReason = "InvalidGroup"
	PolicyReasonInvalidKind              PolicyConditionReason = "InvalidKind"
	PolicyReasonInvalidName              PolicyConditionReason = "InvalidName"
	PolicyReasonInvalidTargetRef         PolicyConditionReason = "InvalidTargetReference"
	PolicyReasonSectionNamesNotPermitted PolicyConditionReason = "SectionNamesNotPermitted"
	PolicyReasonInvalidService           PolicyConditionReason = "InvalidService"
)

// Deployment Condition Reasons
const (
	PolicyReasonDeployed         PolicyConditionReason = "Deployed"
	PolicyReasonDeploymentFailed PolicyConditionReason = "DeploymentFailed"
	PolicyReasonNoDeployment     PolicyConditionReason = "NoDeployment"
)

// Programmed Condition Reasons
const (
	PolicyReasonProgrammed        PolicyConditionReason = "Programmed"
	PolicyReasonProgrammingFailed PolicyConditionReason = "ProgrammingFailed"
	PolicyReasonOperationFailed   PolicyConditionReason = "OperationFailed"
)
