package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=healthcheckpolicy
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HealthCheckPolicy is the schema for the HealthCheckPolicy API.
type HealthCheckPolicy struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the HealthCheckPolicy specification.
	Spec HealthCheckPolicySpec `json:"spec,omitempty"`

	// Status defines the current state of HealthCheckPolicy.
	//
	// +kubebuilder:default={conditions: {{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}}
	Status HealthCheckPolicyStatus `json:"status,omitempty"`
}

// HealthCheckPolicySpec defines the desired state of HealthCheckPolicy.
type HealthCheckPolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	TargetRef CustomTargetRef `json:"targetRef"`

	// Override defines policy configuration that should override policy
	// configuration attached below the targeted resource in the hierarchy.
	//
	// Note: Override is currently not supported and will result in a validation error.
	// Support for Override will be added in a future release.
	//
	// +optional
	Override *HealthCheckPolicyConfig `json:"override,omitempty"`

	// Default defines default policy configuration for the targeted resource.
	//
	// +optional
	Default *HealthCheckPolicyConfig `json:"default,omitempty"`
}

// HealthCheckPolicyStatus defines the observed state of HealthCheckPolicy.
type HealthCheckPolicyStatus struct {
	// Conditions describe the current conditions of the HealthCheckPolicy.
	//
	// Implementations should prefer to express HealthCheckPolicy conditions
	// using the `HealthCheckPolicyConditionType` and `HealthCheckPolicyConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe HealthCheckPolicy state.
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

// HealthCheckPolicyConfig defines the schema for HealthCheck check specification.
type HealthCheckPolicyConfig struct {
	// Interval is the number of seconds between HealthCheck checks.
	//
	// +optional
	Interval metav1.Duration `json:"interval,omitempty"`

	// Timeout is the number of seconds after which the HealthCheck check is
	// considered failed.
	//
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`

	// Port is the port to use for HealthCheck checks.
	//
	// +optional
	Port int32 `json:"port,omitempty"`

	// UnhealthyThreshold is the number of consecutive failed HealthCheck checks.
	//
	// +optional
	UnhealthyThreshold int32 `json:"unhealthyThreshold,omitempty"`

	// HealthyThreshold is the number of consecutive successful HealthCheck checks.
	//
	// +optional
	HealthyThreshold int32 `json:"healthyThreshold,omitempty"`

	// UseTLS indicates whether health check should enforce TLS.
	// By default, health check will use the same protocol as the
	// service if the same port is used for health check. If the port
	// is different, health check will be plaintext.
	//
	// +optional
	UseTLS bool `json:"useTLS,omitempty"`

	// HTTP defines the HTTP constraint specification for the HealthCheck of a
	// target resource.
	//
	// +optional
	HTTP *HTTPSpecifiers `json:"http,omitempty"`

	// GRPC configures a gRPC v1 HealthCheck (https://github.com/grpc/grpc-proto/blob/master/grpc/health/v1/health.proto)
	// against the target resource.
	GRPC *GRPCSpecifiers `json:"grpc,omitempty"`
}

// Protocol defines the protocol used for certain properties.
// Valid Protocol values are:
//
// * HTTP
// * HTTPS
// * TCP
type Protocol string

const (
	// ProtocolHTTP implies that the service uses HTTP.
	ProtocolHTTP Protocol = "HTTP"

	// ProtocolHTTPS implies that the service uses HTTPS.
	ProtocolHTTPS Protocol = "HTTPS"

	// ProtocolTCP implies that the service uses plain TCP.
	ProtocolTCP Protocol = "TCP"
)

// GRPCSpecifiers defines the schema for GRPC HealthCheck.
type GRPCSpecifiers struct {
	// Authority if present is used as the value of the Authority header in the health check.
	//
	// +optional
	Authority string `json:"authority,omitempty"`

	// Service allows the configuration of a Health check registered under a different service name.
	//
	// +optional
	Service string `json:"service,omitempty"`
}

// HTTPSpecifiers defines the schema for HTTP HealthCheck check specification.
type HTTPSpecifiers struct {
	// Host is the host header value to use for HealthCheck checks.
	//
	// +optional
	Host string `json:"host,omitempty"`

	// Path is the path to use for HealthCheck checks.
	//
	// +optional
	Path string `json:"path,omitempty"`

	// Match defines the HTTP matchers to use for HealthCheck checks.
	//
	// +optional
	Match *HTTPMatch `json:"match,omitempty"`
}

// HTTPMatch defines the HTTP matchers to use for HealthCheck checks.
type HTTPMatch struct {
	// Body defines the HTTP body matchers to use for HealthCheck checks.
	//
	// +optional
	Body string `json:"body,omitempty"`

	// StatusCodes defines the HTTP status code matchers to use for HealthCheck checks.
	//
	// +optional
	StatusCodes []*StatusCodes `json:"statusCodes,omitempty"`
}

// StatusCodes defines the HTTP status code matchers to use for HealthCheck checks.
type StatusCodes struct {
	// Start defines the start of the range of status codes to use for HealthCheck checks.
	// This is inclusive.
	//
	// +optional
	Start int32 `json:"start,omitempty"`

	// End defines the end of the range of status codes to use for HealthCheck checks.
	// This is inclusive.
	//
	// +optional
	End int32 `json:"end,omitempty"`
}

// HealthCheckPolicyConditionType is a type of condition associated with a
// HealthCheckPolicy. This type should be used with the HealthCheckPolicyStatus.Conditions
// field.
type HealthCheckPolicyConditionType string

// HealthCheckPolicyConditionReason defines the set of reasons that explain why a
// particular HealthCheckPolicy condition type is raised.
type HealthCheckPolicyConditionReason string

const (
	// HealthCheckPolicyConditionAccepted is used to set the HealthCheckPolicyConditionType to Accepted.
	HealthCheckPolicyConditionAccepted HealthCheckPolicyConditionType = "Accepted"

	// HealthCheckPolicyConditionResolvedRefs is used to set the HealthCheckPolicyCondition to ResolvedRefs.
	HealthCheckPolicyConditionResolvedRefs HealthCheckPolicyConditionType = "ResolvedRefs"

	// HealthCheckPolicyReasonAccepted is used to set the HealthCheckPolicyConditionReason to Accepted.
	// When the given HealthCheckPolicy is correctly configured.
	HealthCheckPolicyReasonAccepted HealthCheckPolicyConditionReason = "Accepted"

	// HealthCheckPolicyConditionResolvedRefs is used to set the FrontendCondition to ResolvedRefs.
	// This is used with the following reasons:
	// * HealthCheckPolicyReasonInvalid
	// * HealthCheckPolicyReasonRefNotPermitted
	// * HealthCheckPolicyReasonInvalidGroup
	// * HealthCheckPolicyReasonInvalidKind
	// * HealthCheckPolicyReasonInvalidName
	// * HealthCheckPolicyReasonNoTargetReference
	// * HealthCheckPolicyReasonInvalidService
	// * HealthCheckPolicyReasonInvalidPort
	// * HealthCheckPolicyReasonSectionNamesNotPermitted
	// * HealthCheckPolicyReasonOverrideNotSupported

	// HealthCheckPolicyReasonInvalid is the reason when the HealthCheckPolicy isn't Accepted.
	HealthCheckPolicyReasonInvalid HealthCheckPolicyConditionReason = "InvalidHealthCheckPolicy"

	// HealthCheckPolicyReasonUnsupportedStatusCodes is used when the HealthCheckPolicy match StatusCodes are not supported.
	HealthCheckPolicyReasonUnsupportedStatusCodes HealthCheckPolicyConditionReason = "UnsupportedStatusCodes"

	// HealthCheckPolicyReasonResolvedRefs is used when the targetRef was resolved successfully.
	HealthCheckPolicyReasonResolvedRefs HealthCheckPolicyConditionReason = "ResolvedRefs"

	// HealthCheckPolicyReasonRefNotPermitted is used when the ref isn't permitted.
	HealthCheckPolicyReasonRefNotPermitted HealthCheckPolicyConditionReason = "RefNotPermitted"

	// HealthCheckPolicyReasonInvalidGroup is used when the group is invalid.
	HealthCheckPolicyReasonInvalidGroup HealthCheckPolicyConditionReason = "InvalidGroup"

	// HealthCheckPolicyReasonInvalidKind is used when the kind/group is invalid.
	HealthCheckPolicyReasonInvalidKind HealthCheckPolicyConditionReason = "InvalidKind"

	// HealthCheckPolicyReasonInvalidName is used when the name is invalid.
	HealthCheckPolicyReasonInvalidName HealthCheckPolicyConditionReason = "InvalidName"

	// HealthCheckPolicyReasonNoTargetReference is used when there's no target reference.
	HealthCheckPolicyReasonNoTargetReference HealthCheckPolicyConditionReason = "NoTargetReference"

	// HealthCheckPolicyReasonInvalidService is used when the Service is invalid.
	HealthCheckPolicyReasonInvalidService HealthCheckPolicyConditionReason = "InvalidService"

	// HealthCheckPolicyReasonInvalidPort is used when the port is invalid.
	HealthCheckPolicyReasonInvalidPort HealthCheckPolicyConditionReason = "InvalidPort"

	// HealthCheckPolicyReasonSectionNamesNotPermitted is used when the section names aren't permitted.
	HealthCheckPolicyReasonSectionNamesNotPermitted HealthCheckPolicyConditionReason = "SectionNamesNotPermitted"

	// HealthCheckPolicyReasonOverrideNotSupported is used when the override isn't supported.
	HealthCheckPolicyReasonOverrideNotSupported HealthCheckPolicyConditionReason = "OverrideNotSupported"

	// BackendTLSPolicyConditionNotFound is used when the BackendTLSPolicy is not found for the service.
	BackendTLSPolicyConditionNotFound HealthCheckPolicyConditionReason = "BackendTLSPolicyNotFound"
)

// +kubebuilder:object:root=true

// HealthCheckPolicyList contains a list of HealthCheckPolicy.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type HealthCheckPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HealthCheckPolicy `json:"items"`
}

// GetNamespacedName returns the policy's name.
func (h HealthCheckPolicy) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: h.Namespace,
		Name:      h.Name,
	}
}
