// Package reporting is responsible for generating migration reports.
package reporting

import (
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"k8s.io/apimachinery/pkg/types"

	agicmigration "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

type MigrationReport struct {
	Metadata      `yaml:"metadata"`
	AGICResources AGICResources `yaml:"inputAGICResources"`
	AGCResources  AGCResources  `yaml:"outputAGCResourceSummary"`
}

type Metadata struct {
	GeneratedAt         time.Time `yaml:"generatedAt"`
	ToolVersion         string    `yaml:"toolVersion"`
	AGICResourcesSource `yaml:"agicResourcesSource"`
	AGCBYOResourceID    string `yaml:"agcByoResourceID,omitempty"`
	ManagedSubnetID     string `yaml:"managedSubnetID,omitempty"`
}

type AGICResources struct {
	Ingresses []Ingress `yaml:"ingresses"`
}

type AppGWRewrites struct {
	types.NamespacedName `yaml:",inline"`
	Result               resources.MigrationStatus `yaml:"result"`
	Issues               []Issue                   `yaml:"issues,omitempty"`
}

type AGCResources struct {
	ApplicationLoadBalancer       *types.NamespacedName  `yaml:"applicationLoadBalancer,omitempty"`
	Gateway                       types.NamespacedName   `yaml:"gateway"`
	HTTPRoutes                    []types.NamespacedName `yaml:"httpRoutes"`
	ReferenceGrants               []types.NamespacedName `yaml:"referenceGrants,omitempty"`
	HealthCheckPolicies           []types.NamespacedName `yaml:"healthCheckPolicies"`
	RoutePolicies                 []types.NamespacedName `yaml:"routePolicies"`
	WebApplicationFirewalPolicies []types.NamespacedName `yaml:"webApplicationFirewallPolicies"`
	FrontendTLSPolicies           []types.NamespacedName `yaml:"frontendTLSPolicies"`
	BackendTLSPolicies            []types.NamespacedName `yaml:"backendTLSPolicies"`
}

type Issue struct {
	ImpactedResources []string `yaml:"impactedResources,omitempty"`
	Description       string   `yaml:"description"`
	Recommendation    string   `yaml:"recommendation,omitempty"`
	Message           string   `yaml:"message,omitempty"`
}

type AGICResourcesSource string

const (
	// AGICResourcesSourceCluster indicates that the AGIC resources have been sourced from a Kubernetes cluster.
	AGICResourcesSourceCluster AGICResourcesSource = "cluster"

	// AGICResourcesSourceFiles indicates that the AGIC resources have been sourced from files on disk.
	AGICResourcesSourceFiles AGICResourcesSource = "files"
)

// NewReport returns a new MigrationReport
func NewReport(
	agicResources resources.AGICResources,
	agcResources resources.AGCResourceGraph,
	source AGICResourcesSource,
	opts conversion.Options,
) (MigrationReport, error) {
	report := MigrationReport{
		Metadata: NewMetadata(source, opts),
	}

	// convert from the agic and agc resources into a structured report
	for _, ingress := range agicResources.IngressContexts {
		report.AGICResources.Ingresses = append(report.AGICResources.Ingresses, newIngress(ingress))
	}

	if agcResources.ApplicationLoadBalancer != nil {
		report.AGCResources.ApplicationLoadBalancer = &types.NamespacedName{
			Name:      agcResources.ApplicationLoadBalancer.Name,
			Namespace: agcResources.ApplicationLoadBalancer.Namespace,
		}
	}

	report.AGCResources.Gateway = types.NamespacedName{
		Name:      agcResources.Gateway.Name,
		Namespace: agcResources.Gateway.Namespace,
	}

	for name := range agcResources.HTTPRoutes {
		report.AGCResources.HTTPRoutes = append(report.AGCResources.HTTPRoutes, name)
	}

	for name := range agcResources.ReferenceGrants {
		report.AGCResources.ReferenceGrants = append(report.AGCResources.ReferenceGrants, name)
	}

	for name := range agcResources.HealthCheckPolicies {
		report.AGCResources.HealthCheckPolicies = append(report.AGCResources.HealthCheckPolicies, name)
	}

	for name := range agcResources.RoutePolicies {
		report.AGCResources.RoutePolicies = append(report.AGCResources.RoutePolicies, name)
	}

	for name := range agcResources.WAFPolicies {
		report.AGCResources.WebApplicationFirewalPolicies = append(report.AGCResources.WebApplicationFirewalPolicies, name)
	}

	for name := range agcResources.BackendTLSPolicies {
		report.AGCResources.BackendTLSPolicies = append(report.AGCResources.BackendTLSPolicies, name)
	}

	for name := range agcResources.FrontendTLSPolicies {
		report.AGCResources.FrontendTLSPolicies = append(report.AGCResources.FrontendTLSPolicies, name)
	}

	report.AGCResources.sortContents()
	report.AGICResources.sortContents()

	return report, nil
}

func NewReportFromYAML(reportStr string) (MigrationReport, error) {
	var report MigrationReport
	err := yaml.Unmarshal([]byte(reportStr), &report)

	if err != nil {
		return MigrationReport{}, err
	}

	return report, nil
}

func NewMetadata(source AGICResourcesSource, opts conversion.Options) Metadata {
	return Metadata{
		GeneratedAt:         time.Now().UTC(),
		ToolVersion:         agicmigration.GetVersion(),
		AGICResourcesSource: source,
		AGCBYOResourceID:    opts.BYOResourceID,
		ManagedSubnetID:     opts.ManagedSubnetID,
	}
}

func (agc AGCResources) sortContents() {
	for _, s := range []*[]types.NamespacedName{
		&agc.HTTPRoutes, &agc.HealthCheckPolicies, &agc.RoutePolicies, &agc.WebApplicationFirewalPolicies,
		&agc.FrontendTLSPolicies, &agc.BackendTLSPolicies,
	} {
		slices.SortFunc(*s, namespacedNameCmp)
	}
}

func (agic AGICResources) sortContents() {
	slices.SortFunc(agic.Ingresses, func(a, b Ingress) int {
		return namespacedNameCmp(a.NamespacedName, b.NamespacedName)
	})

	for _, ingress := range agic.Ingresses {
		ingress.sortContents()
	}
}

func namespacedNameCmp(a, b types.NamespacedName) int {
	if namespaceCmp := strings.Compare(a.Namespace, b.Namespace); namespaceCmp != 0 {
		return namespaceCmp
	}

	return strings.Compare(a.Name, b.Name)
}
