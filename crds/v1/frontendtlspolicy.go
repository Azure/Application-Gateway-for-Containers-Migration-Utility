package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=frontendtlspolicies
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FrontendTLSPolicy is the schema for the FrontendTLSPolicy API
type FrontendTLSPolicy struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the FrontendTLSPolicy specification.
	Spec FrontendTLSPolicySpec `json:"spec,omitempty"`

	// Status defines the current state of FrontendTLSPolicy.
	//
	// +kubebuilder:default={conditions: {{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}}
	Status FrontendTLSPolicyStatus `json:"status,omitempty"`
}

// FrontendTLSPolicySpec defines the desired state of FrontendTLSPolicy
type FrontendTLSPolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	TargetRef CustomTargetRef `json:"targetRef"`

	// Default defines default policy configuration for the targeted resource.
	//
	// +optional
	Default *FrontendTLSPolicyConfig `json:"default,omitempty"`

	// Override defines policy configuration that should override policy
	// configuration attached below the targeted resource in the hierarchy.
	//
	// Note: Override is currently not supported and result in a validation error.
	// Support for Override will be added in a future release.
	//
	// +optional
	Override *FrontendTLSPolicyConfig `json:"override,omitempty"`
}

// FrontendTLSPolicyConfig defines the policy specification for the Frontend TLS
// Policy.
type FrontendTLSPolicyConfig struct {
	// Verify provides the options to verify the peer certificate.
	//
	// +optional
	Verify *MTLSPolicyVerify `json:"verify,omitempty"`

	// Type is the type of the policy.
	//
	// +optional
	// +kubebuilder:default={name: "2023-06", type: "predefined"}
	FrontendTLSPolicyType *PolicyType `json:"policyType"`
}

// MTLSPolicyVerify defines the schema for the MTLSPolicyVerify API.
type MTLSPolicyVerify struct {
	// CaCertificateRef is the CA certificate used to verify peer certificate.
	//
	// +kubebuilder:validation:Required
	CaCertificateRef *gatewayapi_v1.SecretObjectReference `json:"caCertificateRef,omitempty"`

	// SubjectAltNames is the list of subject alternative names used to verify peer
	// certificate.
	//
	// +optional
	SubjectAltNames []string `json:"subjectAltNames,omitempty"`
}

// PolicyType is the type of the policy.
type PolicyType struct {
	// Name is the name of the policy.
	Name FrontendTLSPolicyTypeName `json:"name"`

	// FrontendTLSPolicyType specifies the frontend TLS policy type
	FrontendTLSPolicyType FrontendTLSPolicyType `json:"type"`
}

// FrontendTLSPolicyType is the type of the Frontend TLS Policy.
type FrontendTLSPolicyType string

const (
	// PredefinedFrontendTLSPolicyType is the type of the predefined Frontend TLS Policy.
	PredefinedFrontendTLSPolicyType FrontendTLSPolicyType = "predefined"
)

// FrontendTLSPolicyTypeName is the name of the Frontend TLS Policy.
type FrontendTLSPolicyTypeName string

const (
	// PredefinedPolicy202306 is the name of the predefined Frontend TLS Policy for the policy "2023-06".
	PredefinedPolicy202306 FrontendTLSPolicyTypeName = "2023-06"

	// PredefinedPolicy202306Strict is the name of the predefined Frontend TLS Policy for the policy "2023-06-S".
	// This is a strict version of the policy "2023-06".
	PredefinedPolicy202306Strict FrontendTLSPolicyTypeName = "2023-06-S"
)

// FrontendTLSPolicyStatus defines the observed state of FrontendTLSPolicy.
type FrontendTLSPolicyStatus struct {
	// Conditions describe the current conditions of the FrontendTLSPolicy.
	//
	// Implementations should prefer to express FrontendTLSPolicy conditions
	// using the `FrontendTLSPolicyConditionType` and `FrontendTLSPolicyConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe FrontendTLSPolicy state.
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

// +kubebuilder:object:root=true

// FrontendTLSPolicyList lists the FrontendTLSPolicy objects.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FrontendTLSPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FrontendTLSPolicy `json:"items"`
}

// GetNamespacedName returns the policy's name
func (f FrontendTLSPolicy) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: f.Namespace,
		Name:      f.Name,
	}
}

// GetTargetRefs returns the target reference of the FrontendTLSPolicy.
func (f FrontendTLSPolicy) GetTargetRefs() []*CustomTargetRef {
	return []*CustomTargetRef{&f.Spec.TargetRef}
}
