package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// IngressExtensionList contains a list of IngressExtension.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type IngressExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IngressExtension `json:"items"`
}

// IngressExtensionConditionType is a type of condition associated with a
// IngressExtension. This type should be used with the IngressExtensionStatus.Conditions
// field.
type IngressExtensionConditionType string

// IngressExtensionConditionReason defines the set of reasons that explain why a
// particular IngressExtension condition type is raised.
type IngressExtensionConditionReason string

const (
	// IngressExtensionConditionAccepted indicates if the IngressExtension is accepted (reconciled) by the controller
	IngressExtensionConditionAccepted IngressExtensionConditionType = "Accepted"

	// IngressExtensionReasonAccepted is used to set the IngressExtensionConditionAccepted to Accepted
	IngressExtensionReasonAccepted IngressExtensionConditionReason = "Accepted"

	// IngressExtensionReasonPartiallyAccepted is used to set the IngressExtensionConditionAccepted to Accepted, but with nonfatal validation errors
	IngressExtensionReasonPartiallyAccepted IngressExtensionConditionReason = "PartiallyAcceptedWithErrors"

	// IngressExtensionConditionErrors indicates if there are validation or build errors on the extension
	IngressExtensionConditionErrors IngressExtensionConditionType = "Errors"

	// IngressExtensionReasonNoErrors indicates there are no validation errors
	IngressExtensionReasonNoErrors IngressExtensionConditionReason = "NoValidationErrors"

	// IngressExtensionReasonHasErrors indicates there are some validation errors
	IngressExtensionReasonHasErrors IngressExtensionConditionReason = "HasValidationErrors"

	// IngressExtensionMessageAcceptedNoErrors is applicable when an IngressExtension has been accepted and no errors are present
	IngressExtensionMessageAcceptedNoErrors = "Accepted and validated successfully"

	// IngressExtensionMessageErrorsPresent is applicable to the Errors condition when errors are present
	IngressExtensionMessageErrorsPresent = "Validation errors encountered, see rules and backend settings status for details"

	// IngressExtensionMessageAcceptedPartially is application to the Accepted condition when errors are present
	IngressExtensionMessageAcceptedPartially = "Partially accepted with some validation errors"

	// IngressExtensionMessageErrorsNoErrors is application to the Errors condition when no errors are present
	IngressExtensionMessageErrorsNoErrors = "No validation errors"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=ingressextension
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Errors",type=string,JSONPath=`.status.conditions[?(@.type=="Errors")].status`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IngressExtension is the schema for the IngressExtension API
type IngressExtension struct {
	// Object's type metadata.
	metav1.TypeMeta `json:",inline"`

	// Object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the IngressExtension specification.
	Spec IngressExtensionSpec `json:"spec,omitempty"`

	// Status describes the current state of the IngressExtension as enacted by the ALB controller
	// +optional

	// +kubebuilder:default={conditions: {{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}}
	Status IngressExtensionStatus `json:"status"`
}

// IngressExtensionSpec defines the desired configuration of IngressExtension
type IngressExtensionSpec struct {
	// Rules define the rules per host
	// +optional
	Rules []IngressRuleSetting `json:"rules,omitempty"`

	// BackendSettings defines a set of configuration options for Ingress service backends
	// +optional
	BackendSettings []IngressBackendSettings `json:"backendSettings,omitempty"`
}

// IngressRuleSetting provides configuration options for rules
type IngressRuleSetting struct {
	// Host is used to match against Ingress rules with the same hostname in order to identify which rules affect these settings
	Host string `json:"host"`

	// AdditionalHostnames specifies more hostnames to listen on
	// +optional
	AdditionalHostnames []string `json:"additionalHostnames,omitempty"`

	// Rewrites defines the rewrites for the rule
	// +optional
	Rewrites []IngressRewrites `json:"rewrites,omitempty"`

	// RequestRedirect defines the redirect behavior for the rule
	// +optional
	RequestRedirect *Redirect `json:"requestRedirect,omitempty"`
}

// IngressRewrites provides the various rewrites supported on a rule
type IngressRewrites struct {
	// Type identifies the type of rewrite
	// +kubebuilder:validation:Enum=RequestHeaderModifier;ResponseHeaderModifier;URLRewrite
	Type RewriteType `json:"type"`

	// RequestHeaderModifier defines a schema that modifies request headers.
	// +optional
	RequestHeaderModifier *HeaderFilter `json:"requestHeaderModifier,omitempty"`

	// RequestHeaderModifier defines a schema that modifies response headers.
	// +optional
	ResponseHeaderModifier *HeaderFilter `json:"responseHeaderModifier,omitempty"`

	// URLRewrite defines a schema that modifies a request during forwarding.
	// +optional
	URLRewrite *URLRewriteFilter `json:"urlRewrite,omitempty"`
}

// HeaderFilter defines a filter that modifies the headers of an HTTP
// request or response. Only one action for a given header name is permitted.
// Filters specifying multiple actions of the same or different type for any one
// header name are invalid and rejected.
// Configuration to set or add multiple values for a header must use RFC 7230
// header value formatting, separating each value with a comma.
type HeaderFilter struct {
	// Set overwrites the request with the given header (name, value)
	// before the action.
	//
	// Input:
	//   GET /foo HTTP/1.1
	//   my-header: foo
	//
	// Config:
	//   set:
	//   - name: "my-header"
	//     value: "bar"
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   my-header: bar
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Set []HTTPHeader `json:"set,omitempty"`

	// Add adds the given header(s) (name, value) to the request
	// before the action. It appends to any existing values associated
	// with the header name.
	//
	// Input:
	//   GET /foo HTTP/1.1
	//   my-header: foo
	//
	// Config:
	//   add:
	//   - name: "my-header"
	//     value: "bar,baz"
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   my-header: foo,bar,baz
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Add []HTTPHeader `json:"add,omitempty"`

	// Remove the given header(s) from the HTTP request before the action. The
	// value of Remove is a list of HTTP header names. Header names
	// are case-insensitive (see
	// https://datatracker.ietf.org/doc/html/rfc2616#section-4.2).
	//
	// Input:
	//   GET /foo HTTP/1.1
	//   my-header1: foo
	//   my-header2: bar
	//   my-header3: baz
	//
	// Config:
	//   remove: ["my-header1", "my-header3"]
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   my-header2: bar
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Remove []string `json:"remove,omitempty"`
}

// HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.
type HTTPHeader struct {
	// Name is the name of the HTTP Header to be matched. Name matching MUST be
	// case insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).
	//
	// If multiple entries specify equivalent header names, the first entry with
	// an equivalent name MUST be considered for a match. Subsequent entries
	// with an equivalent header name MUST be ignored. Due to the
	// case-insensitivity of header names, "foo" and "Foo" are considered
	// equivalent.
	Name HTTPHeaderName `json:"name"`

	// Value is the value of HTTP Header to be matched.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`
}

// HTTPHeaderName is the name of an HTTP header.
//
// Valid values include:
//
// * "Authorization"
// * "Set-Cookie"
//
// Invalid values include:
//
//   - ":method" - ":" is an invalid character. This means that HTTP/2 pseudo
//     headers aren't currently supported by this type.
//   - "/invalid" - "/ " is an invalid character
type HTTPHeaderName HeaderName

// HeaderName is the name of a header or query parameter.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`
// +k8s:deepcopy-gen=false
type HeaderName string

// URLRewriteFilter defines a filter that modifies a request during
// forwarding. At most one of these filters may be used on a rule. This
// MUST NOT be used on the same rule having an sslRedirect.
type URLRewriteFilter struct {
	// Hostname is the value to be used to replace the Host header value during
	// forwarding.
	// +optional
	Hostname *PreciseHostname `json:"hostname,omitempty"`

	// Path defines a path rewrite.
	// +optional
	Path *HTTPPathModifier `json:"path,omitempty"`
}

// Redirect defines a filter that redirects a request. This
// MUST NOT be used on the same rule that also has a URLRewriteFilter.
type Redirect struct {
	// Scheme is the scheme to be used in the value of the `Location` header in
	// the response. When empty, the scheme of the request is used.
	//
	//
	// +optional
	// +kubebuilder:validation:Enum=http;https
	Scheme *string `json:"scheme,omitempty"`

	// Hostname is the hostname to be used in the value of the `Location`
	// header in the response.
	// When empty, the hostname in the `Host` header of the request is used.
	//
	// +optional
	Hostname *PreciseHostname `json:"hostname,omitempty"`

	// Path defines parameters used to modify the path of the incoming request.
	// The modified path is then used to construct the `Location` header. When
	// empty, the request path is used as-is.
	//
	// +optional
	Path *HTTPPathModifier `json:"path,omitempty"`

	// Port is the port to be used in the value of the `Location`
	// header in the response.
	//
	// If no port is specified, the redirect port MUST be derived using the
	// following rules:
	//
	// * If redirect scheme is not-empty, the redirect port MUST be the well-known
	//   port associated with the redirect scheme. Specifically "http" to port 80
	//   and "https" to port 443. If the redirect scheme doesn't have a
	//   well-known port, the listener port of the Gateway SHOULD be used.
	// * If redirect scheme is empty, the redirect port MUST be the Gateway
	//   Listener port.
	//
	// Implementations SHOULD NOT add the port number in the 'Location'
	// header in the following cases:
	//
	// * A Location header that uses HTTP (whether that is determined via
	//   the Listener protocol or the Scheme field) _and_ use port 80.
	// * A Location header that uses HTTPS (whether that is determined via
	//   the Listener protocol or the Scheme field) _and_ use port 443.
	//
	// +optional
	Port *PortNumber `json:"port,omitempty"`

	// StatusCode is the HTTP status code to be used in response.
	//
	// Values may be added to this enum, implementations
	// must ensure that unknown values won't cause a crash.
	//
	// +optional
	// +kubebuilder:default=302
	// +kubebuilder:validation:Enum=301;302;303;307;308
	StatusCode *int `json:"statusCode,omitempty"`
}

// PortNumber defines a network port.
//
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=65535
type PortNumber int32

// PreciseHostname is the fully qualified domain name of a network host. This
// matches the RFC 1123 definition of a hostname with one notable exception that
// numeric IP addresses aren't allowed.
//
// Per RFC1035 and RFC1123, a *label* must consist of lower case
// alphanumeric characters or '-', and must start and end with an alphanumeric
// character. No other punctuation is allowed.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
type PreciseHostname string

// HTTPPathModifier defines configuration for path modifiers.
type HTTPPathModifier struct {
	// Type defines the type of path modifier. More types may be
	// added in a future release of the API.
	//
	// Values may be added to this enum, implementations
	// must ensure unknown values won't cause a crash.
	//
	// Unknown values here must result in the implementation setting the
	// Accepted Condition for the rule to be false
	//
	// +kubebuilder:validation:Enum=ReplaceFullPath;ReplacePrefixMatch
	Type HTTPPathModifierType `json:"type"`

	// ReplaceFullPath specifies the value with which to replace the full path
	// of a request during a rewrite or redirect.
	//
	// +kubebuilder:validation:MaxLength=1024
	// +optional
	ReplaceFullPath *string `json:"replaceFullPath,omitempty"`

	// ReplacePrefixMatch specifies the value with which to replace the prefix
	// match of a request during a rewrite or redirect. For example, a request
	// to "/foo/bar" with a prefix match of "/foo" and a ReplacePrefixMatch
	// of "/xyz" would be modified to "/xyz/bar".
	//
	// This matches the behavior of the PathPrefix match type. This
	// matches full path elements. A path element refers to the list of labels
	// in the path split by the `/` separator. When specified, a trailing `/` is
	// ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all
	// match the prefix `/abc`, but the path `/abcd` wouldn't.
	//
	// ReplacePrefixMatch is only compatible with a `PathPrefix` HTTPRouteMatch.
	// Using any other HTTPRouteMatch type on the same HTTPRouteRule results in
	// the implementation setting the Accepted Condition for the Route to `status: False`.
	//
	// Request Path | Prefix Match | Replace Prefix | Modified Path
	// |------------|--------------|----------------|----------
	// /foo/bar     | /foo         | /xyz           | /xyz/bar
	// /foo/bar     | /foo         | /xyz/          | /xyz/bar
	// /foo/bar     | /foo/        | /xyz           | /xyz/bar
	// /foo/bar     | /foo/        | /xyz/          | /xyz/bar
	// /foo         | /foo         | /xyz           | /xyz
	// /foo/        | /foo         | /xyz           | /xyz/
	// /foo/bar     | /foo         | <empty string> | /bar
	// /foo/        | /foo         | <empty string> | /
	// /foo         | /foo         | <empty string> | /
	// /foo/        | /foo         | /              | /
	// /foo         | /foo         | /              | /
	//
	// +kubebuilder:validation:MaxLength=1024
	// +optional
	ReplacePrefixMatch *string `json:"replacePrefixMatch,omitempty"`
}

// HTTPPathModifierType defines the type of path redirect or rewrite.
type HTTPPathModifierType string

const (
	//  FullPathHTTPPathModifier replaces the full path with the specified value.
	FullPathHTTPPathModifier HTTPPathModifierType = "ReplaceFullPath"

	// PrefixMatchHTTPPathModifier replaces any prefix path with the
	// substitution value. For example, a path with a prefix
	// match of "/foo" and a ReplacePrefixMatch substitution of "/bar"
	// replace "/foo" with "/bar" in matching requests.
	//
	// This matches the behavior of the PathPrefix match type. This
	// matches full path elements. A path element refers to the list of labels
	// in the path split by the `/` separator. When specified, a trailing `/` is
	// ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all
	// match the prefix `/abc`, but the path `/abcd` wouldn't.
	PrefixMatchHTTPPathModifier HTTPPathModifierType = "ReplacePrefixMatch"
)

// RewriteType identifies the rewrite type
type RewriteType string

const (
	// RequestHeaderModifier can be used to add or remove an HTTP
	// header from an HTTP request before it's sent to the upstream target.
	RequestHeaderModifier RewriteType = "RequestHeaderModifier"

	// ResponseHeaderModifier can be used to add or remove an HTTP
	// header from an HTTP response before it's sent to the client.
	ResponseHeaderModifier RewriteType = "ResponseHeaderModifier"

	// URLRewrite can be used to modify a request during forwarding.
	URLRewrite RewriteType = "URLRewrite"
)

// IngressBackendSettings provides extended configuration options for a backend service
type IngressBackendSettings struct {
	// Service is the name of a backend service that this configuration applies to
	Service string `json:"service"`

	// Ports can be used to indicate if the backend service is listening on HTTP or HTTPS
	// +optional
	Ports []IngressBackendPort `json:"ports,omitempty"`

	// TrustedRootCertificate can be used to supply a certificate for the gateway to trust when communicating to the
	// backend on a port specified as https
	// +optional
	TrustedRootCertificate string `json:"trustedRootCertificate,omitempty"`

	// +optional
	// ConnectionDraining bool `json:"connectionDraining"`

	// SessionAffinity allows client requests to be consistently given to the same backend
	// +optional
	*SessionAffinity `json:"sessionAffinity,omitempty"`

	// Timeouts define a set of timeout parameters to be applied to an Ingress
	// +optional
	Timeouts IngressTimeouts `json:"timeouts,omitempty"`

	// LoadBalancing defines the load balancing policy for the backend service
	// +optional
	*BackendLoadBalancingPolicySpec `json:"loadBalancingPolicySpec,omitempty"`
}

// IngressBackendPort describes a port on a backend.
// Only one of Name/Number should be defined.
type IngressBackendPort struct {
	// Port indicates the port  on the backend service
	// +optional
	Port *int32 `json:"port"`

	// Name must refer to a name on a port on the backend service
	// +optional
	Name *string `json:"name"`

	// Protocol should be one of "HTTP", "HTTPS"
	// +kubebuilder:validation:Enum=HTTP;HTTPS
	Protocol Protocol `json:"protocol"`
}

// IngressTimeouts can be used to configure timeout properties for an Ingress
type IngressTimeouts struct {
	// RequestTimeout defines the timeout used by the load balancer when forwarding requests to a backend service
	// +optional
	RequestTimeout metav1.Duration `json:"requestTimeout,omitempty"`
}

// IngressExtensionStatus describes the current state of the IngressExtension
type IngressExtensionStatus struct {
	// Rules have detailed status information regarding each Rule
	// +optional
	Rules []IngressRuleStatus `json:"rules,omitempty"`

	// BackendSettings has detailed status information regarding each BackendSettings
	// +optional
	BackendSettings []IngressBackendSettingStatus `json:"backendSettings,omitempty"`

	// Conditions describe the current conditions of the IngressExtension.
	// +optional
	// Known condition types are:
	//
	// * "Accepted"
	// * "Errors"
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{type: "Accepted", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"},{type: "ValidationErrors", status: "Unknown", reason:"Pending", message:"Waiting for controller", lastTransitionTime: "1970-01-01T00:00:00Z"}}
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// IngressRuleStatus describes the state of a rule
type IngressRuleStatus struct {
	// Host identifies the rule this status describes
	Host string `json:"host"`

	// Errors are a list of errors relating to this setting
	// +optional
	Errors []string `json:"validationErrors"`

	// Valid indicates that there are no validation errors present on this rule
	// +optional
	Valid bool `json:"valid"`
}

// IngressBackendSettingStatus describes the state of a BackendSetting
type IngressBackendSettingStatus struct {
	// Service identifies the BackendSetting this status describes
	Service string `json:"service"`

	// Errors are a list of errors relating to this setting
	// +optional
	Errors []string `json:"validationErrors"`

	// Valid indicates that there are no validation errors present on this BackendSetting
	Valid bool `json:"valid"`
}
