package v1

import gatewayapi_v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

// CustomTargetRef is a reference to a custom resource that isn't part of the
// Kubernetes core API.
type CustomTargetRef struct {
	gatewayapi_v1alpha2.NamespacedPolicyTargetReference `json:",inline"`

	// SectionNames is the name of the section within the target resource. When
	// unspecified, this targetRef targets the entire resource. In the following
	// resources, SectionNames is interpreted as the following:
	//
	// * Gateway: Listener Name
	// * Service: Port Name
	//
	// If a SectionNames is specified, but doesn't exist on the targeted object,
	// the Policy fails to attach, and the policy implementation will record
	// a `ResolvedRefs` or similar Condition in the Policy's status.
	//
	//
	// +optional
	SectionNames []string `json:"sectionNames"`
}
