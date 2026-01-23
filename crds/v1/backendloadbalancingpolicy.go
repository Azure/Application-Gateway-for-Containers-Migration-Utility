package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=backendloadbalancingpolicy
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackendLoadBalancingPolicy represents the configuration for backend load balancing.
type BackendLoadBalancingPolicy struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the BackendLoadBalancingPolicy specification.
	Spec BackendLoadBalancingPolicySpec `json:"spec,omitempty"`

	// Status defines the current state of BackendLoadBalancingPolicy.
	Status BackendLoadBalancingPolicyStatus `json:"status,omitempty"`
}

// BackendLoadBalancingPolicySpec defines the specification for BackendLoadBalancingPolicy.
type BackendLoadBalancingPolicySpec struct {
	// TargetRefs identifies a list of API objects to apply policy to.
	TargetRefs []TargetRefSpec `json:"targetRefs"`

	// CircuitBreaker defines the schema for configuring Circuit Breaking
	// +optional
	CircuitBreaker *CircuitBreakerConfig `json:"circuitBreaker,omitempty"`

	// LoadBalancing defines the schema for configuring Load Balancing options
	// +optional
	LoadBalancing *LoadBalancingConfig `json:"loadBalancing,omitempty"`
}

// TargetRefSpec defines the target reference and ports for the backend load balancing policy.
type TargetRefSpec struct {
	// TargetRef identifies an API object to apply policy to.
	TargetRef CustomTargetRef `json:"targetRef"`

	// Ports specifies the list of ports on the target where the policy is applied.
	// +optional
	Ports []BackendLoadBalancingPolicyPort `json:"ports,omitempty"`
}

// BackendLoadBalancingPolicyPort defines the port configuration for the backend load balancing policy.
type BackendLoadBalancingPolicyPort struct {
	// Port is the port to use for connection to the backend
	//
	// +kubebuilder:validation:Minimum=1
	Port int32 `json:"port,omitempty"`
}

// CircuitBreakerConfig defines the configuration for circuit breaking.
type CircuitBreakerConfig struct {
	// The maximum number of connections that will be made to the backend service.
	// Default is 1000000000 (disabled).
	//
	// +optional
	// +kubebuilder:default=1000000000
	// +kubebuilder:validation:Minimum=0
	MaxConnections uint32 `json:"maxConnections,omitempty"`

	// Maximum number of parallel requests that will be made to the backend service.
	// Default is 1000000000 (disabled).
	//
	// +optional
	// +kubebuilder:default=1000000000
	// +kubebuilder:validation:Minimum=0
	MaxRequests uint32 `json:"maxRequests,omitempty"`

	// Maximum number of pending requests that will be made to the backend service.
	// Default is 1000000000 (disabled).
	//
	// +optional
	// +kubebuilder:default=1000000000
	// +kubebuilder:validation:Minimum=0
	MaxPendingRequests uint32 `json:"maxPendingRequests,omitempty"`

	// The maximum number of parallel retries allowed to the backend service.
	// Default is 1000000000 (disabled).
	//
	// +optional
	// +kubebuilder:default=1000000000
	// +kubebuilder:validation:Minimum=0
	MaxRetries uint32 `json:"maxRetries,omitempty"`
}

// LoadBalancingConfig defines the configuration for load balancing.
type LoadBalancingConfig struct {
	// Strategy defines the policy to use when load balancing traffic to the backend service.
	// Default is round-robin.
	//
	// +optional
	Strategy LoadBalancingStrategy `json:"strategy,omitempty"`

	// SlowStart defines the schema for Slow Start specification
	// +optional
	SlowStart *SlowStartConfig `json:"slowStart,omitempty"`

	// LoadAware defines the schema for Load Aware Routing specification
	// +optional
	LoadAware *LoadAwareConfig `json:"loadAware,omitempty"`
}

// LoadBalancingStrategy defines the policy to use when balancing traffic across a service
//
// +kubebuilder:validation:Enum=round-robin;least-request;load-aware
type LoadBalancingStrategy string

const (
	// LoadBalancingRoundRobin is used to set the LoadBalancingStrategy to round-robin
	LoadBalancingRoundRobin LoadBalancingStrategy = "round-robin"

	// LoadBalancingLeastRequest is used to set the LoadBalancingStrategy to least-request
	LoadBalancingLeastRequest LoadBalancingStrategy = "least-request"

	// LoadBalancingRingHash is used to set the LoadBalancingStrategy to ring-hash
	LoadBalancingRingHash LoadBalancingStrategy = "ring-hash"

	// LoadBalancingLoadAware is used to set the LoadBalancingStrategy to load-aware
	LoadBalancingLoadAware LoadBalancingStrategy = "load-aware"
)

// SlowStartConfig defines the configuration for slow start.
type SlowStartConfig struct {
	// The duration of the slow start window.
	// +required
	Window metav1.Duration `json:"window,omitempty"`

	// The speed of traffic increase over the slow start window, must be greater than 0.0.
	// Defaults to 1.0 if unspecified.
	//
	// +optional
	// +kubebuilder:default=`1.0`
	// +kubebuilder:validation:Pattern=`^([0-9]+([.][0-9]+)?|[.][0-9]+)$`
	Aggression string `json:"aggression,omitempty"`

	// The minimum or starting percentage of traffic to send to new endpoints.
	// Defaults to 10% if unspecified.
	//
	// +optional
	// +kubebuilder:default=10
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	StartWeightPercent uint32 `json:"startWeightPercent,omitempty"`
}

// LoadAwareConfig defines the configuration for load aware routing.
type LoadAwareConfig struct {
	// An endpoint must report load metrics continuously for at least this long before the endpoint
	// metrics will be used to influence load balancing decisions. Takes effect both immediately after
	// we establish a connection to an endpoint and after `metricExpirationPeriod` has elapsed.
	// Default is 10 seconds.
	//
	// +optional
	BlackoutPeriod metav1.Duration `json:"blackoutPeriod"`

	// If the endpoint does not report load metrics for this duration, metrics will stop
	// being used to influence load balancing decisions. Default is 3 minutes.
	//
	// +optional
	MetricExpirationPeriod metav1.Duration `json:"metricExpirationPeriod"`

	// The multiplier used to adjust endpoint weights with the error rated calculated based
	// on the reported `rps_fractional` and `eps` load metrics. Must not be a negative value.
	// Default is 1.0.
	//
	// +optional
	// +kubebuilder:default=`1.0`
	// +kubebuilder:validation:Pattern=`^([0-9]+([.][0-9]+)?|[.][0-9]+)$`
	ErrorUtilizationPenalty string `json:"errorUtilizationPenalty,omitempty"`

	// A list of custom metrics reported by endpoints to be used for reporting utilization
	// and influencing load balancing decisions. Utilization will be computed by taking the
	// max of the values of metrics specified in this list.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=10
	NamedMetrics []string `json:"namedMetrics,omitempty"`
}

// BackendLoadBalancingPolicyStatus defines the observed state of BackendLoadBalancingPolicy.
type BackendLoadBalancingPolicyStatus struct {
	// Conditions describe the current conditions of the BackendLoadBalancingPolicy.
	//
	// Implementations should prefer to express BackendLoadBalancingPolicy conditions
	// using the `BackendLoadBalancingPolicyConditionType` and `BackendLoadBalancingPolicyConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe BackendLoadBalancingPolicy state.
	//

	Targets []BackendLoadBalancingPolicyTargetStatus `json:"targets"`
}

// BackendLoadBalancingPolicyTargetStatus defines the observed status for a target ref
type BackendLoadBalancingPolicyTargetStatus struct {
	TargetRef CustomTargetRef `json:"targetRef"`

	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// BackendLoadBalancingPolicyConditionType is a type of condition associated with a
// BackendLoadBalancingPolicy. This type should be used with the BackendLoadBalancingPolicyStatus.Conditions
// field.
type BackendLoadBalancingPolicyConditionType string

// BackendLoadBalancingPolicyConditionReason defines the set of reasons that explain why a
// particular BackendLoadBalancingPolicy condition type is raised.
type BackendLoadBalancingPolicyConditionReason string

const (

	// BackendLoadBalancingPolicyConditionAccepted is used to set the BackendLoadBalancingPolicyConditionType to Accepted
	BackendLoadBalancingPolicyConditionAccepted BackendLoadBalancingPolicyConditionType = "Accepted"

	// BackendLoadBalancingPolicyConditionResolvedRefs is used to set the BackendLoadBalancingPolicyCondition to ResolvedRefs
	BackendLoadBalancingPolicyConditionResolvedRefs BackendLoadBalancingPolicyConditionType = "ResolvedRefs"

	// BackendLoadBalancingPolicyReasonAccepted is used to set the BackendLoadBalancingPolicyConditionReason to Accepted
	// When the given BackendLoadBalancingPolicy is correctly configured
	BackendLoadBalancingPolicyReasonAccepted BackendLoadBalancingPolicyConditionReason = "Accepted"

	// BackendLoadBalancingPolicyReasonResolvedRefs is used to set the BackendLoadBalancingPolicyConditionReason to ResolvedRefs
	// when the given BackendLoadBalancingPolicy has correct references
	BackendLoadBalancingPolicyReasonResolvedRefs BackendLoadBalancingPolicyConditionReason = "ResolvedRefs"

	// BackendLoadBalancingPolicyConditionResolvedRefs is used to set the FrontendCondition to ResolvedRefs
	// This is used with the following reasons :
	// *BackendLoadBalancingPolicyReasonInvalid
	// *BackendLoadBalancingPolicyReasonRefNotPermitted
	// *BackendLoadBalancingPolicyReasonInvalidGroup
	// *BackendLoadBalancingPolicyReasonInvalidKind
	// *BackendLoadBalancingPolicyReasonInvalidName
	// *BackendLoadBalancingPolicyReasonNoTargetReference

	// BackendLoadBalancingPolicyReasonInvalid is the reason when the BackendLoadBalancingPolicy isn't Accepted
	BackendLoadBalancingPolicyReasonInvalid BackendLoadBalancingPolicyConditionReason = "InvalidBackendLoadBalancingPolicy"

	// BackendLoadBalancingPolicyReasonRefNotPermitted is used when the ref isn't permitted
	BackendLoadBalancingPolicyReasonRefNotPermitted BackendLoadBalancingPolicyConditionReason = "RefNotPermitted"

	// BackendLoadBalancingPolicyReasonInvalidGroup is used when the group is invalid
	BackendLoadBalancingPolicyReasonInvalidGroup BackendLoadBalancingPolicyConditionReason = "InvalidGroup"

	// BackendLoadBalancingPolicyReasonInvalidKind is used when the kind/group is invalid
	BackendLoadBalancingPolicyReasonInvalidKind BackendLoadBalancingPolicyConditionReason = "InvalidKind"

	// BackendLoadBalancingPolicyReasonInvalidName is used when the name is invalid
	BackendLoadBalancingPolicyReasonInvalidName BackendLoadBalancingPolicyConditionReason = "InvalidName"

	// BackendLoadBalancingPolicyReasonNoTargetReference is used when there's no target reference
	BackendLoadBalancingPolicyReasonNoTargetReference BackendLoadBalancingPolicyConditionReason = "NoTargetReference"

	// BackendLoadBalancingPolicyReasonInvalidService is used when the Service is invalid
	BackendLoadBalancingPolicyReasonInvalidService BackendLoadBalancingPolicyConditionReason = "InvalidService"

	// BackendLoadBalancingPolicyReasonConflicted is used when the target ref conflicts with a pre-existing policy target
	BackendLoadBalancingPolicyReasonConflicted BackendLoadBalancingPolicyConditionReason = "Conflicted"
)

// BackendLoadBalancingPolicyList contains a list of BackendLoadBalancingPolicy.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BackendLoadBalancingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendLoadBalancingPolicy `json:"items"`
}

// GetNamespacedName returns the policy's name
func (r BackendLoadBalancingPolicy) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.Namespace,
		Name:      r.Name,
	}
}
