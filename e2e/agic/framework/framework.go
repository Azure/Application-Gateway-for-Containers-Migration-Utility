// Package framework provides a test framework for E2E tests.
// nolint: gosec // warnings relating to executing commands and reading files from variables not important in these tests
package framework

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi_v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	k8s_yaml "sigs.k8s.io/yaml"

	"github.com/go-test/deep"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	agicmigration "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility"
	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/output"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/reporting"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

type TestCase struct {
	// input files for the migration tool
	InputFiles []string

	ExtraFiles []string // these are extra resources to be applied to the cluster, but not part of the migration input

	ParentDir string

	// expected resources after migration
	ExpectedResources resources.AGCResourceGraph
	// expected report after migration
	ExpectedReport reporting.MigrationReport
}

const (
	// BYOAGC is placeholder value to use for a BYO AGC resource
	BYOAGC = "/subscriptions/11111111-1111-1111-bbbb-bbbbbbbbbbbb/resourceGroups/onebox-rg/providers/Microsoft.ServiceNetworking/trafficControllers/test-tc"

	// ManagedSubnet is a placeholder of a subnet ID to use for managed scenarios
	ManagedSubnet = "/subscriptions/11111111-1111-1111-bbbb-bbbbbbbbbbbb/resourceGroups/onebox-rg/providers/Microsoft.Network/virtualNetworks/onebox-vnet/subnets/alb-subnet-1"
)

func NewTestCase(t *testing.T, inputFiles []string, extraFiles []string, expectedResourceFiles []string, expectedReportFile string) TestCase {
	t.Helper()
	expectedResources := loadResources(t, expectedResourceFiles)
	reportBytes, err := os.ReadFile(expectedReportFile)

	if err != nil {
		t.Fatalf("failed to read expected report file: %s, %v", expectedReportFile, err)
	}

	expectedReport := loadReport(t, string(reportBytes))

	return TestCase{
		InputFiles:        inputFiles,
		ExpectedResources: expectedResources,
		ExpectedReport:    expectedReport,
		ExtraFiles:        extraFiles,
	}
}

func NewTestCaseFromCurrentDir(t *testing.T) TestCase {
	pattern := "input/*.yaml"
	inputFiles, err := filepath.Glob(pattern)

	if err != nil {
		t.Fatalf("error globbing %s: %v", pattern, err)
	}

	pattern = "output/*.yaml"
	expectedResourceFiles, err := filepath.Glob(pattern)

	if err != nil {
		t.Fatalf("error globbing %s: %v", pattern, err)
	}

	pattern = "extras/*.yaml"
	extraFiles, err := filepath.Glob(pattern)

	if err != nil {
		t.Fatalf("error globbing %s: %v", pattern, err)
	}

	expectedReportFile := "report.yaml"

	return NewTestCase(t, inputFiles, extraFiles, expectedResourceFiles, expectedReportFile)
}

func (tc TestCase) Run(t *testing.T, flags ...string) string {
	if len(flags) == 0 {
		flags = []string{"--byo-resource-id", BYOAGC}
	}

	report, result, outDir := tc.RunMigration(t, flags...)

	// deep uses global settings, so we're going to save and reset them so that no one gets a nasty surprise
	// if the integrate these tests into a larger suite down the line.
	originalSlicesEmpty := deep.NilSlicesAreEmpty
	originalMapsEmpty := deep.NilMapsAreEmpty

	// Restore it after test
	t.Cleanup(func() {
		deep.NilSlicesAreEmpty = originalSlicesEmpty
		deep.NilMapsAreEmpty = originalMapsEmpty
	})
	// treat empty slices and maps as equal to nil
	deep.NilSlicesAreEmpty = true
	deep.NilMapsAreEmpty = true

	tc.compareReport(t, report)
	tc.compareResources(t, result)
	tc.compareReportToResources(t, result, report)

	return outDir
}

func (tc TestCase) compareReport(t *testing.T, got reporting.MigrationReport) {
	tc.ExpectedReport.Metadata.GeneratedAt = got.GeneratedAt
	tc.ExpectedReport.Metadata.ToolVersion = agicmigration.GetVersion()

	t.Run("Report", func(t *testing.T) {
		printedReport := false

		if diff := deep.Equal(got, tc.ExpectedReport); diff != nil {
			for _, line := range diff {
				if !printedReport {
					if yamlStr, err := yaml.Marshal(got); err == nil {
						t.Logf("Got report:\n%s", string(yamlStr))
					}

					printedReport = true
				}

				t.Errorf("Report mismatch: %s", line)
			}
		}
	})
}

func (tc TestCase) compareResources(t *testing.T, got resources.AGCResourceGraph) {
	t.Run("Resources", func(t *testing.T) {
		if diff := deep.Equal(got, tc.ExpectedResources); diff != nil {
			for _, line := range diff {
				t.Errorf("Resource mismatch: %s", line)
			}
		}
	})
}

func (tc TestCase) compareReportToResources(t *testing.T, resources resources.AGCResourceGraph, report reporting.MigrationReport) {
	t.Run("Report vs Resources", func(t *testing.T) {
		// check that all resources in the report are in the resources
		unaccountedForHTTPRoutes := sets.New[types.NamespacedName]()
		for route := range resources.HTTPRoutes {
			unaccountedForHTTPRoutes.Insert(route)
		}

		for _, ingress := range report.AGICResources.Ingresses {
			for _, route := range ingress.MigratedIngressResources.HTTPRoutes {
				if _, ok := resources.HTTPRoutes[route]; !ok {
					t.Errorf("Report lists HTTPRoute %s as migrated, but it is missing from output resources", route.String())
				} else {
					unaccountedForHTTPRoutes.Delete(route)
				}
			}
		}

		if unaccountedForHTTPRoutes.Len() > 0 {
			t.Errorf("Migrated resources includes %d unexpected HTTPRoutes: %v", unaccountedForHTTPRoutes.Len(), unaccountedForHTTPRoutes.UnsortedList())
		}

		for _, route := range report.AGCResources.HTTPRoutes {
			if _, ok := resources.HTTPRoutes[route]; !ok {
				t.Errorf("Report lists HTTPRoute %s, but it is missing from resources", route.String())
			}
		}

		for _, policy := range report.AGCResources.HealthCheckPolicies {
			if _, ok := resources.HealthCheckPolicies[policy]; !ok {
				t.Errorf("Report lists HealthCheckPolicy %s, but it is missing from resources", policy.String())
			}
		}

		for _, policy := range report.AGCResources.RoutePolicies {
			if _, ok := resources.RoutePolicies[policy]; !ok {
				t.Errorf("Report lists RoutePolicy %s, but it is missing from resources", policy.String())
			}
		}
	})
}

func (tc TestCase) RunMigration(t *testing.T, flags ...string) (reporting.MigrationReport, resources.AGCResourceGraph, string) {
	t.Helper()

	stdout, outputDir := execute(t, tc.InputFiles, flags)
	files, err := filepath.Glob(filepath.Join(outputDir, "*.yaml"))

	if err != nil {
		t.Fatalf("failed to list yaml files in dir: %s, %v", outputDir, err)
	}

	t.Logf("Migration output files: %v", files)

	return parseReportFromStdout(t, stdout), loadResources(t, files), outputDir
}

func execute(t *testing.T, files []string, flags []string) (string, string) {
	t.Helper()
	outputDir := t.TempDir()

	cmdPath := filepath.Join(getRepoRoot(t), "cmd")
	args := append([]string{"run", cmdPath, "files"}, files...)
	args = append(args, []string{"--output-dir", outputDir}...)
	args = append(args, flags...)
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Migration command failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), outputDir
}

func getRepoRoot(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to get module root: %v", err)
	}

	return strings.TrimSpace(out.String())
}

func parseReportFromStdout(t *testing.T, stdout string) reporting.MigrationReport {
	t.Helper()

	var reportContent string
	if idx := strings.Index(stdout, output.MigrationReportHeader); idx != -1 {
		reportContent = stdout[idx+len(output.MigrationReportHeader):]
	} else {
		t.Fatalf("Failed to find migration report in output")
	}

	return loadReport(t, reportContent)
}

func loadReport(t *testing.T, reportContent string) reporting.MigrationReport {
	t.Helper()

	report, err := reporting.NewReportFromYAML(reportContent)
	if err != nil {
		t.Fatalf("Failed to parse migration report: %v", err)
	}

	return report
}

// nolint: errcheck
func loadResources(t *testing.T, files []string) resources.AGCResourceGraph {
	t.Helper()

	result := resources.NewAGCResourceGraph()

	type kindHelper struct {
		factory       func() any // factory returns an instance of the kind
		graphInserter func(any)  // graphInserter adds the instance to the resource graph
	}

	kindMapping := map[string]kindHelper{
		k8snames.KindApplicationLoadBalancer: {
			factory: func() any { return &albcontrollerapi_v1.ApplicationLoadBalancer{} },
			graphInserter: func(obj any) {
				result.ApplicationLoadBalancer = obj.(*albcontrollerapi_v1.ApplicationLoadBalancer)
			},
		},
		k8snames.KindBackendTLSPolicy: {
			factory: func() any { return &albcontrollerapi_v1.BackendTLSPolicy{} },
			graphInserter: func(obj any) {
				policy := obj.(*albcontrollerapi_v1.BackendTLSPolicy)
				result.BackendTLSPolicies[k8snames.NamespacedName(policy)] = policy
			},
		},
		k8snames.KindFrontendTLSPolicy: {
			factory: func() any { return &albcontrollerapi_v1.FrontendTLSPolicy{} },
			graphInserter: func(obj any) {
				policy := obj.(*albcontrollerapi_v1.FrontendTLSPolicy)
				result.FrontendTLSPolicies[k8snames.NamespacedName(policy)] = policy
			},
		},
		k8snames.KindWebApplicationFirewallPolicy: {
			factory: func() any { return &albcontrollerapi_v1.WebApplicationFirewallPolicy{} },
			graphInserter: func(obj any) {
				policy := obj.(*albcontrollerapi_v1.WebApplicationFirewallPolicy)
				result.WAFPolicies[k8snames.NamespacedName(policy)] = policy
			},
		},
		k8snames.KindGateway: {
			factory: func() any { return &gatewayapi_v1.Gateway{} },
			graphInserter: func(obj any) {
				gateway := obj.(*gatewayapi_v1.Gateway)
				result.Gateway = gateway
			},
		},
		k8snames.KindHealthCheckPolicy: {
			factory: func() any { return &albcontrollerapi_v1.HealthCheckPolicy{} },
			graphInserter: func(obj any) {
				policy := obj.(*albcontrollerapi_v1.HealthCheckPolicy)
				result.HealthCheckPolicies[k8snames.NamespacedName(policy)] = policy
			},
		},
		k8snames.KindHTTPRoute: {
			factory: func() any { return &gatewayapi_v1.HTTPRoute{} },
			graphInserter: func(obj any) {
				route := obj.(*gatewayapi_v1.HTTPRoute)
				result.HTTPRoutes[k8snames.NamespacedName(route)] = route
			},
		},
		k8snames.KindReferenceGrant: {
			factory: func() any { return &gatewayapi_v1beta1.ReferenceGrant{} },
			graphInserter: func(obj any) {
				grant := obj.(*gatewayapi_v1beta1.ReferenceGrant)
				result.ReferenceGrants[k8snames.NamespacedName(grant)] = grant
			},
		},
		k8snames.KindRoutePolicy: {
			factory: func() any { return &albcontrollerapi_v1.RoutePolicy{} },
			graphInserter: func(obj any) {
				policy := obj.(*albcontrollerapi_v1.RoutePolicy)
				result.RoutePolicies[k8snames.NamespacedName(policy)] = policy
			},
		},
	}

	for _, file := range files {
		unstructured := make(map[string]any)
		f, err := os.Open(file)

		if err != nil {
			t.Fatalf("failed to open file: %s, %v", file, err)
		}

		defer func() { _ = f.Close() }()

		if err := yaml.NewDecoder(f).Decode(&unstructured); err != nil {
			t.Fatalf("failed to decode yaml file: %s, %v", file, err)
		}

		kind, ok := unstructured["Kind"].(string)

		if !ok {
			kind, ok = unstructured["kind"].(string)
			if !ok {
				t.Fatalf("missing Kind in yaml file: %s", file)
			}
		}

		if _, err := f.Seek(0, 0); err != nil {
			t.Fatalf("failed to seek to the start of the file: %s, %v", file, err)
		}

		helper, ok := kindMapping[kind]
		if !ok {
			t.Fatalf("unknown kind %q in yaml file: %s", kind, file)
		}

		bytes, err := io.ReadAll(f)
		if err != nil {
			t.Fatalf("failed to read yaml file: %s, %v", file, err)
		}

		target := helper.factory()

		if err := k8s_yaml.Unmarshal(bytes, target); err != nil {
			t.Fatalf("failed to unmarshal k8s yaml: %s, %v", file, err)
		}

		helper.graphInserter(target)
	}

	return result
}
