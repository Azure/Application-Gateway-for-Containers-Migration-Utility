package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=routepolicies
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RoutePolicy is the schema for the RoutePolicy API.
type RoutePolicy struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the RoutePolicy specification.
	Spec RoutePolicySpec `json:"spec,omitempty"`

	// Status defines the current state of RoutePolicy.
	//
	// +kubebuilder:default={conditions: {{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}}
	Status RoutePolicyStatus `json:"status,omitempty"`
}

// RoutePolicySpec defines the desired state of RoutePolicy.
type RoutePolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	TargetRef CustomTargetRef `json:"targetRef"`

	// Override defines policy configuration that should override policy
	// configuration attached below the targeted resource in the hierarchy.
	//
	// Note: Override is currently not supported and result in a validation error.
	// Support for Override will be added in a future release.
	//
	// +optional
	Override *RoutePolicyConfig `json:"override,omitempty"`

	// Default defines default policy configuration for the targeted resource.
	//
	// +optional
	Default *RoutePolicyConfig `json:"default,omitempty"`
}

// RoutePolicyStatus defines the observed state of RoutePolicy.
type RoutePolicyStatus struct {
	// Conditions describe the current conditions of the RoutePolicy.
	//
	// Implementations should prefer to express RoutePolicy conditions
	// using the `RoutePolicyConditionType` and `RoutePolicyConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe RoutePolicy state.
	//
	// +optional
	// Known condition types are:
	//
	// * "Accepted"
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// RoutePolicyConfig defines the schema for RoutePolicy specification.
// This allows the specification of the following attributes:
// * Timeouts
// * Session Affinity
type RoutePolicyConfig struct {
	// Custom Timeouts
	// Timeout for the target resource.
	// +optional
	RouteTimeouts *RouteTimeouts `json:"timeouts,omitempty"`

	// SessionAffinity defines the schema for Session Affinity specification
	// +optional
	SessionAffinity *SessionAffinity `json:"sessionAffinity,omitempty"`
}

// RouteTimeouts defines the schema for Timeouts specification.
type RouteTimeouts struct {
	// RouteTimeout is the timeout for the route.
	//
	// +optional
	RouteTimeout metav1.Duration `json:"routeTimeout,omitempty"`
}

// SessionAffinity defines the schema for Session Affinity specification.
type SessionAffinity struct {
	// +kubebuilder:validation:Required
	AffinityType AffinityType `json:"affinityType,omitempty"`

	// +optional
	CookieName string `json:"cookieName,omitempty"`

	// +optional
	CookieDuration metav1.Duration `json:"cookieDuration,omitempty"`
}

// AffinityType defines the affinity type for the Service
//
// +kubebuilder:validation:Enum=application-cookie;managed-cookie
type AffinityType string

const (
	// AffinityTypeApplicationCookie is a session affinity type for an application cookie
	AffinityTypeApplicationCookie AffinityType = "application-cookie"

	// AffinityTypeManagedCookie is a session affinity type for a managed cookie
	AffinityTypeManagedCookie AffinityType = "managed-cookie"
)

// +kubebuilder:object:root=true

// RoutePolicyConditionType is a type of condition associated with a
// RoutePolicy. This type should be used with the RoutePolicyStatus.Conditions
// field.
type RoutePolicyConditionType string

// RoutePolicyConditionReason defines the set of reasons that explain why a
// particular RoutePolicy condition type is raised.
type RoutePolicyConditionReason string

const (

	// RoutePolicyConditionAccepted is used to set the RoutePolicyConditionType to Accepted
	RoutePolicyConditionAccepted RoutePolicyConditionType = "Accepted"

	// RoutePolicyConditionResolvedRefs is used to set the RoutePolicyCondition to ResolvedRefs
	RoutePolicyConditionResolvedRefs RoutePolicyConditionType = "ResolvedRefs"

	// RoutePolicyReasonAccepted is used to set the RoutePolicyConditionReason to Accepted
	// When the given RoutePolicy is correctly configured
	RoutePolicyReasonAccepted RoutePolicyConditionReason = "Accepted"

	// RoutePolicyReasonAcceptedWithTimeoutConflict is used to set the RoutePolicyConditionReason to AcceptedWithTimeoutConflict
	// When the given RoutePolicy is correctly configured but has a timeout conflict with the target route
	RoutePolicyReasonAcceptedWithTimeoutConflict RoutePolicyConditionReason = "AcceptedWithTimeoutConflict"

	// RoutePolicyConditionResolvedRefs is used to set the FrontendCondition to ResolvedRefs
	// This is used with the following reasons :
	// *RoutePolicyReasonInvalid
	// *RoutePolicyReasonRefNotPermitted
	// *RoutePolicyReasonInvalidGroup
	// *RoutePolicyReasonInvalidKind
	// *RoutePolicyReasonInvalidName
	// *RoutePolicyReasonNoTargetReference
	// *RoutePolicyReasonInvalidHTTPRoute
	// *RoutePolicyReasonSectionNamesNotPermitted
	// *RoutePolicyReasonOverrideNotSupported

	// RoutePolicyReasonInvalid is the reason when the RoutePolicy isn't Accepted
	RoutePolicyReasonInvalid RoutePolicyConditionReason = "InvalidRoutePolicy"

	// RoutePolicyReasonRefNotPermitted is used when the ref isn't permitted
	RoutePolicyReasonRefNotPermitted RoutePolicyConditionReason = "RefNotPermitted"

	// RoutePolicyReasonInvalidGroup is used when the group is invalid
	RoutePolicyReasonInvalidGroup RoutePolicyConditionReason = "InvalidGroup"

	// RoutePolicyReasonInvalidKind is used when the kind/group is invalid
	RoutePolicyReasonInvalidKind RoutePolicyConditionReason = "InvalidKind"

	// RoutePolicyReasonInvalidName is used when the name is invalid
	RoutePolicyReasonInvalidName RoutePolicyConditionReason = "InvalidName"

	// RoutePolicyReasonNoTargetReference is used when there's no target reference
	RoutePolicyReasonNoTargetReference RoutePolicyConditionReason = "NoTargetReference"

	// RoutePolicyReasonInvalidHTTPRoute is used when the HTTPRoute is invalid
	RoutePolicyReasonInvalidHTTPRoute RoutePolicyConditionReason = "InvalidHTTPRoute"

	// RoutePolicyReasonInvalidGRPCRoute is used when the GRPCRoute is invalid
	RoutePolicyReasonInvalidGRPCRoute RoutePolicyConditionReason = "InvalidGRPCRoute"

	// RoutePolicyReasonSectionNamesNotPermitted is used when the section names aren't permitted
	RoutePolicyReasonSectionNamesNotPermitted RoutePolicyConditionReason = "SectionNamesNotPermitted"

	// RoutePolicyReasonOverrideNotSupported is used when the override isn't supported
	RoutePolicyReasonOverrideNotSupported RoutePolicyConditionReason = "OverrideNotSupported"
)

// RoutePolicyList contains a list of RoutePolicy.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RoutePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoutePolicy `json:"items"`
}

// GetNamespacedName returns the policy's name
func (r RoutePolicy) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}
