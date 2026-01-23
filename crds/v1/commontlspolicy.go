package v1

import gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

// CommonTLSPolicy is the schema for the CommonTLSPolicy API.
type CommonTLSPolicy struct {
	// Verify provides the options to verify the peer certificate.
	//
	// +optional
	Verify *CommonTLSPolicyVerify `json:"verify,omitempty"`
}

// CommonTLSPolicyVerify defines the schema for the CommonTLSPolicyVerify API.
type CommonTLSPolicyVerify struct {
	// CaCertificateRef is the CA certificate used to verify peer certificate.
	//
	// +kubebuilder:validation:Required
	CaCertificateRef *gatewayapi_v1.SecretObjectReference `json:"caCertificateRef,omitempty"`

	// SubjectAltName is the subject alternative name used to verify peer
	// certificate.
	//
	// +optional
	SubjectAltName string `json:"subjectAltName,omitempty"`
}
