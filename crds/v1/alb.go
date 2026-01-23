package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// AlbSpec defines the specifications for the Application Gateway for Containers resource.
type AlbSpec struct {
	// Associations are subnet resource IDs the Application Gateway for Containers resource are associated with.
	Associations []string `json:"associations,omitempty"`
}

const (
	// ALBFinalizerDeploymentExists should be added as a finalizer to the
	// ApplicationLoadBalancer whenever am ARM deployment exists.
	ALBFinalizerDeploymentExists = "alb-deployment-exists-finalizer.alb.networking.azure.io"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=applicationloadbalancer
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Deployment",type=string,JSONPath=`.status.conditions[?(@.type=="Deployment")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationLoadBalancer is the schema for the Application Gateway for Containers resource.
type ApplicationLoadBalancer struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specifications for Application Gateway for Containers resource.
	Spec AlbSpec `json:"spec,omitempty"`

	// Status defines the current state of Application Gateway for Containers resource.
	//
	// +kubebuilder:default={conditions: {{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}}
	Status AlbStatus `json:"status,omitempty"`
}

// AlbStatus defines the observed state of Application Gateway for Containers resource.
type AlbStatus struct {
	// Conditions describe the current conditions of the Application Gateway for Containers resource.

	// Implementations should prefer to express Application Gateway for Containers resource conditions
	// using the `AlbConditionType` and `AlbConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe Application Gateway for Containers resource state.

	// +optional
	// Known condition types are:
	//
	// * "Accepted"
	// * "Ready"
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// AlbConditionType is a type of condition associated with an
// Application Gateway for Containers resource. This type should be used with the AlbStatus.Conditions
// field.
type AlbConditionType string

// AlbConditionReason defines the set of reasons that explain
// why a particular condition type are raised by the Application Gateway for Containers resource.
type AlbConditionReason string

const (
	// AlbConditionTypeAccepted indicates whether the Application Gateway for Containers resource
	// are accepted by the controller.
	AlbConditionTypeAccepted AlbConditionType = "Accepted"

	// AlbConditionTypeDeployment indicates the deployment status of the Application Gateway for Containers resource.
	AlbConditionTypeDeployment AlbConditionType = "Deployment"

	// AlbReasonAccepted indicates that the Application Gateway for Containers resource
	// are accepted by the controller.
	AlbReasonAccepted AlbConditionReason = "Accepted"

	// AlbReasonDeploymentReady indicates the Application Gateway for Containers resource
	// deployment status.
	AlbReasonDeploymentReady AlbConditionReason = "Ready"

	// AlbReasonInProgress indicates whether the Application Gateway for Containers resource
	// is in the process of being created, updated, or deleted.
	AlbReasonInProgress AlbConditionReason = "InProgress"
)

// +kubebuilder:object:root=true

// ApplicationLoadBalancerList lists the ApplicationLoadBalancer objects.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ApplicationLoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationLoadBalancer `json:"items"`
}
