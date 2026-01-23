package agic

import (
	"testing"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestParseStatusCodes(t *testing.T) {
	type testCase struct {
		input          string
		expectedOutput []*albcontrollerapi_v1.StatusCodes
		expectError    bool
	}

	testCases := []testCase{
		{
			input: "200",
			expectedOutput: []*albcontrollerapi_v1.StatusCodes{
				{Start: 200, End: 200},
			},
			expectError: false,
		},
		{
			input: "200,201-204,301",
			expectedOutput: []*albcontrollerapi_v1.StatusCodes{
				{Start: 200, End: 200},
				{Start: 201, End: 204},
				{Start: 301, End: 301},
			},
			expectError: false,
		},
		{
			input:       "invalid",
			expectError: true,
		},
		{
			input:       "200-abc",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			output, err := parseStatusCodes(tc.input)
			if tc.expectError != (err != nil) {
				t.Fatalf("For input %q, expected error: %v, got: %v", tc.input, tc.expectError, err)
			}

			if !tc.expectError {
				if len(output) != len(tc.expectedOutput) {
					t.Fatalf("For input %q, expected output length %d, got %d", tc.input, len(tc.expectedOutput), len(output))
				}

				for i := range output {
					if output[i].Start != tc.expectedOutput[i].Start || output[i].End != tc.expectedOutput[i].End {
						t.Fatalf("For input %q, expected output %v, got %v", tc.input, tc.expectedOutput, output)
					}
				}
			}
		})
	}
}

func TestHandleHealthProbeHostname(t *testing.T) {
	t.Run("new health check", func(t *testing.T) {
		healthCheckHostname := "health-host"
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeHostname, healthCheckHostname)
		if err := provider.handleHealthProbeHostname(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		if len(graph.HealthCheckPolicies) != 1 {
			t.Fatalf("expected 1 health check policy, got %d", len(graph.HealthCheckPolicies))
		}

		for _, policy := range graph.HealthCheckPolicies {
			if policy.Spec.Default.HTTP == nil {
				t.Fatal("expected HTTP config to be set")
			}
			if policy.Spec.Default.HTTP.Host != healthCheckHostname {
				t.Fatalf("expected health check with host %q, got %q", healthCheckHostname, policy.Spec.Default.HTTP.Host)
			}
		}
	})

	t.Run("conflict with existing health check", func(t *testing.T) {
		healthCheckHostname := "health-host"
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeHostname, healthCheckHostname)
		if err := provider.handleHealthProbeHostname(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		conflictingAnno := resources.NewIngressAnnotationContext(AnnotationHealthProbeHostname, "conflict")
		if err := provider.handleHealthProbeHostname(graph, gwCtx, routeCtx, ingressCtx, conflictingAnno); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if len(conflictingAnno.Issues) == 0 {
			t.Fatal("expected the annotation to register an issue but found none")
		}
	})
}

func TestHandleHealthProbeStatusCode(t *testing.T) {
	t.Run("new health check with status codes", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeStatusCode, "200,201-204")
		if err := provider.handleHealthProbeStatusCode(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		if len(graph.HealthCheckPolicies) != 1 {
			t.Fatalf("expected 1 health check policy, got %d", len(graph.HealthCheckPolicies))
		}

		for _, policy := range graph.HealthCheckPolicies {
			if policy.Spec.Default.HTTP == nil {
				t.Fatal("expected HTTP config to be set")
			}
			if policy.Spec.Default.HTTP.Match == nil {
				t.Fatal("expected HTTP Match config to be set")
			}
			if len(policy.Spec.Default.HTTP.Match.StatusCodes) != 2 {
				t.Fatalf("expected 2 status code ranges, got %d", len(policy.Spec.Default.HTTP.Match.StatusCodes))
			}
		}
	})

	t.Run("conflict with existing status codes", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeStatusCode, "200")
		if err := provider.handleHealthProbeStatusCode(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		conflictingAnno := resources.NewIngressAnnotationContext(AnnotationHealthProbeStatusCode, "201")
		if err := provider.handleHealthProbeStatusCode(graph, gwCtx, routeCtx, ingressCtx, conflictingAnno); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if len(conflictingAnno.Issues) == 0 {
			t.Fatal("expected the annotation to register an issue but found none")
		}

		if conflictingAnno.Issues[0].Code != resources.IssueHealthCheckConflict {
			t.Fatalf("expected IssueHealthCheckConflict, got %v", conflictingAnno.Issues[0].Code)
		}
	})

	t.Run("invalid status code format", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeStatusCode, "invalid")
		if err := provider.handleHealthProbeStatusCode(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("expected MigrationStatusError, got %v", annoCtx.Status())
		}
	})
}

func TestHandleHealthProbePath(t *testing.T) {
	t.Run("valid path", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbePath, "/health")
		if err := provider.handleHealthProbePath(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		for _, policy := range graph.HealthCheckPolicies {
			if policy.Spec.Default.HTTP == nil || policy.Spec.Default.HTTP.Path != "/health" {
				t.Fatal("expected health check path to be /health")
			}
		}
	})
}

func TestHandleHealthProbeInterval(t *testing.T) {
	t.Run("valid interval", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeInterval, "30")
		if err := provider.handleHealthProbeInterval(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		for _, policy := range graph.HealthCheckPolicies {
			if policy.Spec.Default.Interval.Seconds() != 30 {
				t.Fatalf("expected interval of 30s, got %v", policy.Spec.Default.Interval.Seconds())
			}
		}
	})

	t.Run("invalid interval", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeInterval, "not-a-number")
		if err := provider.handleHealthProbeInterval(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("expected MigrationStatusError, got %v", annoCtx.Status())
		}
	})
}

func TestHandleHealthProbeTimeout(t *testing.T) {
	t.Run("valid timeout", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeTimeout, "5")
		if err := provider.handleHealthProbeTimeout(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		for _, policy := range graph.HealthCheckPolicies {
			if policy.Spec.Default.Timeout.Seconds() != 5 {
				t.Fatalf("expected timeout of 5s, got %v", policy.Spec.Default.Timeout.Seconds())
			}
		}
	})

	t.Run("invalid timeout", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeTimeout, "invalid")
		if err := provider.handleHealthProbeTimeout(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("expected MigrationStatusError, got %v", annoCtx.Status())
		}
	})
}

func TestHandleHealthProbeUnhealthyThreshold(t *testing.T) {
	t.Run("valid threshold", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeUnhealthyThreshold, "3")
		if err := provider.handleHealthProbeUnhealthyThreshold(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		for _, policy := range graph.HealthCheckPolicies {
			if policy.Spec.Default.UnhealthyThreshold != 3 {
				t.Fatalf("expected unhealthy threshold of 3, got %d", policy.Spec.Default.UnhealthyThreshold)
			}
		}
	})

	t.Run("invalid threshold", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbeUnhealthyThreshold, "abc")
		if err := provider.handleHealthProbeUnhealthyThreshold(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("expected MigrationStatusError, got %v", annoCtx.Status())
		}
	})
}

func TestHandleHealthProbePort(t *testing.T) {
	t.Run("valid port", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbePort, "8080")
		if err := provider.handleHealthProbePort(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err != nil {
			t.Fatal(err)
		}

		for _, policy := range graph.HealthCheckPolicies {
			if policy.Spec.Default.Port != 8080 {
				t.Fatalf("expected port 8080, got %d", policy.Spec.Default.Port)
			}
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx := setupAnnotationHandlerInputs(AnnotationHealthProbePort, "not-a-port")
		if err := provider.handleHealthProbePort(graph, gwCtx, routeCtx, ingressCtx, annoCtx); err == nil {
			t.Fatal("expected an error but got nil")
		}

		if annoCtx.Status() != resources.MigrationStatusError {
			t.Fatalf("expected MigrationStatusError, got %v", annoCtx.Status())
		}
	})
}
