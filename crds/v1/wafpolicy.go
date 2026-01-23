package v1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=webapplicationfirewallpolicy,shortName=wafpolicy
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Deployment",type=string,JSONPath=`.status.conditions[?(@.type=="Deployment")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebApplicationFirewallPolicy is the schema for the Application Gateway for Containers Security Policy child resource.
type WebApplicationFirewallPolicy struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specifications for Application Gateway for Containers Security Policy child resource.
	Spec WebApplicationFirewallPolicySpec `json:"spec,omitempty"`

	// Status defines the current state of Application Gateway for Containers Security Policy child resource.
	//
	// +kubebuilder:default={conditions: {{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}}
	Status WebApplicationFirewallPolicyStatus `json:"status,omitempty"`
}

// GetWAFPolicyID returns the ID of the WebApplicationFirewallPolicy
func (f WebApplicationFirewallPolicy) GetWAFPolicyID() (string, bool) {
	if f.Spec.WebApplicationFirewallPolicy == nil || f.Spec.WebApplicationFirewallPolicy.ID == "" {
		return "", false
	}

	return strings.ToLower(f.Spec.WebApplicationFirewallPolicy.ID), true
}

// WebApplicationFirewallPolicySpec defines the desired state of WebApplicationFirewallPolicy.
type WebApplicationFirewallPolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	TargetRef CustomTargetRef `json:"targetRef"`

	// WebApplicationFirewallPolicy is used to specify a WebApplicationPolicy resource
	WebApplicationFirewallPolicy *WebApplicationFirewallConfig `json:"webApplicationFirewall,omitempty"`
}

// WebApplicationFirewallConfig defines the web application firewall policy configuration for the Application Gateway for Containers Security Policy child resource.
type WebApplicationFirewallConfig struct {
	ID string `json:"id"`
}

// WebApplicationFirewallPolicyStatus defines the observed state of Application Gateway for Containers Security Policy child resource.
type WebApplicationFirewallPolicyStatus struct {
	// Conditions describe the current conditions of the Application Gateway for Containers Security Policy child resource.

	// Implementations should prefer to express Application Gateway for Containers Security Policy child resource conditions
	// using the `WAFPolicyConditionType` and `WAFPolicyConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe Application Gateway for Containers Security Policy child resource state.

	// +optional
	// Known condition types are:
	//
	// * "Accepted"
	// * "Deployment"
	// * "ResolvedRefs"
	// * "Programmed"
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// WebApplicationFirewallPolicyList lists the WebApplicationFirewallPolicy objects.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type WebApplicationFirewallPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WebApplicationFirewallPolicy `json:"items"`
}

// GetNamespacedName returns the policy's name
func (f WebApplicationFirewallPolicy) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: f.Namespace,
		Name:      f.Name,
	}
}
