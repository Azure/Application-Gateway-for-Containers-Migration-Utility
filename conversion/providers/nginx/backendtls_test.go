package nginx

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestNGINXHandleBackendProtocol(t *testing.T) {
	tests := []struct {
		protocol     string
		expectPolicy bool
	}{
		{"HTTP", false},
		{"HTTPS", true},
		{"GRPC", false},
		{"GRPCS", true},
	}

	for _, tc := range tests {
		t.Run(tc.protocol, func(t *testing.T) {
			provider, testGraph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
				AnnotationBackendProtocol, tc.protocol,
			)

			err := provider.handleBackendProtocol(testGraph, gwCtx, routeCtx, nil, annotationCtx)
			if err != nil {
				t.Errorf("handleBackendProtocol returned error: %v", err)
			}

			hasPolicy := len(testGraph.BackendTLSPolicies) > 0
			if hasPolicy != tc.expectPolicy {
				t.Errorf("for protocol %s, expected BackendTLSPolicy=%v, got %v", tc.protocol, tc.expectPolicy, hasPolicy)
			}
		})
	}
}

func TestNGINXHandleBackendProtocolInvalid(t *testing.T) {
	provider, testGraph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationBackendProtocol, "INVALID",
	)

	// Invalid protocol returns an error and sets status to Error
	_ = provider.handleBackendProtocol(testGraph, gwCtx, routeCtx, nil, annotationCtx)

	if annotationCtx.Status() != resources.MigrationStatusError {
		t.Errorf("expected status Error for invalid protocol, got %s", annotationCtx.Status())
	}
}
