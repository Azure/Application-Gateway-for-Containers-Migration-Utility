package output

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi_v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/reporting"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestHandle(t *testing.T) {
	buffer := &memWriter{}
	bufferWriterOpts := Options{
		writer: buffer,
	}

	t.Run("1 of each resource type", func(t *testing.T) {
		result := resources.AGCResourceGraph{
			Gateway: ptr.To(testutil.MakeGateway("default", "gw-1")),
			HTTPRoutes: map[types.NamespacedName]*gatewayapi_v1.HTTPRoute{
				{Name: "route-1", Namespace: "default"}: ptr.To(testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})),
			},
			ReferenceGrants: map[types.NamespacedName]*gatewayapi_v1beta1.ReferenceGrant{
				{Name: "rg-1", Namespace: "default"}: ptr.To(testutil.MakeReferenceGrant("default", "rg-1")),
			},
			HealthCheckPolicies: map[types.NamespacedName]*albcontrollerapi_v1.HealthCheckPolicy{
				{Name: "hcp-1", Namespace: "default"}: ptr.To(
					testutil.MakeHealthCheckPolicy(
						"default", "hcp-1", "default", "has-health-probe-interval",
						testutil.HealthCheckPolicyOptions{
							Interval: ptr.To(30 * time.Second),
						},
					)),
			},
			RoutePolicies: map[types.NamespacedName]*albcontrollerapi_v1.RoutePolicy{
				{Name: "rp-1", Namespace: "default"}: ptr.To(testutil.MakeRoutePolicy("default", "rp-1", "default", "route-1", testutil.RoutePolicyOptions{
					Timeout: ptr.To(time.Second),
				})),
			},
			FrontendTLSPolicies: map[types.NamespacedName]*albcontrollerapi_v1.FrontendTLSPolicy{
				{Name: "ftp-1", Namespace: "default"}: ptr.To(testutil.MakeFrontendTLSPolicy("default", "ftp-1")),
			},
			BackendTLSPolicies: map[types.NamespacedName]*albcontrollerapi_v1.BackendTLSPolicy{
				{Name: "btp-1", Namespace: "default"}: ptr.To(testutil.MakeBackendTLSPolicy("default", "btp-1", "default", "svc-1", 443)),
			},
			ApplicationLoadBalancer: ptr.To(testutil.MakeALB("default", "alb-1", "subnet-1")),
		}

		if err := Write(bufferWriterOpts, reporting.MigrationReport{}, result); err != nil {
			t.Fatalf("Expected nil error, got: %v", err)
		}

		expect := []string{
			"applicationloadbalancer-default-alb-1.yaml",
			"backendtlspolicy-default-btp-1.yaml",
			"frontendtlspolicy-default-ftp-1.yaml",
			"gateway-default-gw-1.yaml",
			"httproute-default-route-1.yaml",
			"referencegrant-default-rg-1.yaml",
			"healthcheckpolicy-default-hcp-1.yaml",
			"routepolicy-default-rp-1.yaml",
		}

		for _, e := range expect {
			if buffer.data[e] == nil {
				t.Fatalf("Expected file %s to be written", e)
			}
		}
	})
}

func TestOptions_getWriter(t *testing.T) {
	t.Run("stdout writer", func(t *testing.T) {
		opts := Options{}
		writer := opts.getWriter()

		if _, ok := writer.(stdoutWriter); !ok {
			t.Fatalf("Expected stdoutWriter, got %T", writer)
		}
	})

	t.Run("file writer", func(t *testing.T) {
		opts := Options{
			OutputFile: "/tmp",
		}
		writer := opts.getWriter()

		if _, ok := writer.(*fileWriter); !ok {
			t.Fatalf("Expected fileWriter, got %T", writer)
		}
	})

	t.Run("dir writer", func(t *testing.T) {
		opts := Options{
			OutputDir: "/tmp",
		}
		writer := opts.getWriter()

		if _, ok := writer.(dirWriter); !ok {
			t.Fatalf("Expected dirWriter, got %T", writer)
		}
	})
}

func TestConverMap_errors(t *testing.T) {
	t.Run("wrong key type", func(t *testing.T) {
		objects := map[string]gatewayapi_v1.Gateway{
			"wrong": {},
		}
		if _, err := convertMap(objects); err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("non-map input", func(t *testing.T) {
		if _, err := convertMap("not-a-map"); err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
}

func TestMarshalYAML(t *testing.T) {
	testCases := []struct {
		name                       string
		inputFactory               func() any
		expectSubStrings           []string
		expectSubStringsAreRegex   bool
		expectSubStringsNotPresent []string
		expectErr                  bool
	}{
		{
			name: "RoutePolicy with nil sectionNames",
			inputFactory: func() any {
				return ptr.To(testutil.MakeRoutePolicy("default", "name", "default", "route-1", testutil.RoutePolicyOptions{}))
			},
			expectSubStrings:           []string{"kind: RoutePolicy"},
			expectSubStringsNotPresent: []string{"sectionNames", "creationTimestamp", "status"},
		},
		{
			name: "Invalid object",
			inputFactory: func() any {
				return true
			},
			expectErr: true,
		},
		{
			name: "HealthCheck start/end order",
			inputFactory: func() any {
				opts := testutil.HealthCheckPolicyOptions{
					StatusCodes: []*albcontrollerapi_v1.StatusCodes{
						{
							Start: 1,
							End:   2,
						},
					},
				}
				return ptr.To(testutil.MakeHealthCheckPolicy("default", "health-1", "default", "svc-1", opts))
			},
			expectSubStrings:         []string{"\\s+- start: 1\\s+end: 2"},
			expectSubStringsAreRegex: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := marshalYAML(tc.inputFactory())
			if err != nil != tc.expectErr {
				t.Fatalf("expected error: %t but got %v", tc.expectErr, err)
			}

			for _, subStr := range tc.expectSubStrings {
				if tc.expectSubStringsAreRegex {
					ok, err := regexp.Match(subStr, got)
					if err != nil {
						t.Fatalf("regex err: %v", err)
					}

					if !ok {
						t.Fatalf("expected yaml to match: %q", subStr)
					}
				} else if !strings.Contains(string(got), subStr) {
					t.Fatalf("expected yaml to contain %q", subStr)
				}
			}

			for _, subStr := range tc.expectSubStringsNotPresent {
				if strings.Contains(string(got), subStr) {
					t.Fatalf("expected yaml to not contain %q", subStr)
				}
			}
		})
	}
}

type memWriter struct {
	data map[string][]byte
}

func (bw *memWriter) write(name string, data []byte) error {
	if bw.data == nil {
		bw.data = make(map[string][]byte)
	}

	bw.data[name] = data

	return nil
}
