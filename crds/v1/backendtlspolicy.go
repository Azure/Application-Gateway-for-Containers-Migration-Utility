package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// BackendTLSPolicySpec defines the desired state of BackendTLSPolicy.
type BackendTLSPolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	TargetRef CustomTargetRef `json:"targetRef"`

	// Override defines policy configuration that should override policy
	// configuration attached below the targeted resource in the hierarchy.
	//
	// Note: Override is currently not supported and result in a validation error.
	// Support for Override will be added in a future release.
	//
	// +optional
	Override *BackendTLSPolicyConfig `json:"override,omitempty"`

	// Default defines default policy configuration for the targeted resource.
	//
	// +optional
	Default *BackendTLSPolicyConfig `json:"default,omitempty"`
}

// BackendTLSPolicyConfig defines the policy specification for the Backend TLS
// Policy.
type BackendTLSPolicyConfig struct {
	CommonTLSPolicy `json:",inline"`

	// Sni is the server name to use for the TLS connection to the backend.
	//
	// +optional
	Sni string `json:"sni,omitempty"`

	// Ports specifies the list of ports where the policy is applied.
	Ports []BackendTLSPolicyPort `json:"ports,omitempty"`

	// ClientCertificateRef is the reference to the client certificate to
	// use for the TLS connection to the backend.
	//
	// +optional
	ClientCertificateRef *gatewayapi_v1.SecretObjectReference `json:"clientCertificateRef,omitempty"`
}

// BackendTLSPolicyPort defines the port to use for the TLS connection to the backend
type BackendTLSPolicyPort struct {
	// Port is the port to use for the TLS connection to the backend
	//
	// +kubebuilder:validation:Minimum=1
	Port int `json:"port,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=backendtlspolicies
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackendTLSPolicy is the schema for the BackendTLSPolicies API.
type BackendTLSPolicy struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the BackendTLSPolicy specification.
	Spec BackendTLSPolicySpec `json:"spec,omitempty"`

	// Status defines the current state of BackendTLSPolicy.
	//
	// +kubebuilder:default={conditions: {{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}}
	Status BackendTLSPolicyStatus `json:"status,omitempty"`
}

// BackendTLSPolicyStatus defines the observed state of BackendTLSPolicy.
type BackendTLSPolicyStatus struct {
	// Conditions describe the current conditions of the BackendTLSPolicy.
	//
	// Implementations should prefer to express BackendTLSPolicy conditions
	// using the `BackendTLSPolicyConditionType` and `BackendTLSPolicyConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe BackendTLSPolicy state.
	//
	// +optional
	// Known condition types are:
	//
	// * "Accepted"
	// * "ResolvedRefs"
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// BackendTLSPolicyList lists the BackendTLSPolicy objects.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackendTLSPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendTLSPolicy `json:"items"`
}

// GetNamespacedName returns the policy's name
func (b BackendTLSPolicy) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: b.Namespace,
		Name:      b.Name,
	}
}
