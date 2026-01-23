// Package migration provides the high level interface for migrating AGIC resources to ALB Controller resources.
package migration

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	ctrl_client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/aggregation"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/reporting"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

type Options struct {
	Aggregation  aggregation.Options
	Conversion   conversion.Options
	ProviderName string
}

func MigrateCluster(ctx context.Context, k8sClient kubernetes.Interface, ctrlClient ctrl_client.Client, opts Options) (resources.AGCResourceGraph, reporting.MigrationReport, error) {
	// Validate provider name - default to AGIC if not specified
	providerName := opts.ProviderName
	if providerName == "" {
		providerName = ProviderNameAGIC
	}

	if providerName != ProviderNameAGIC && providerName != ProviderNameNGINX {
		return resources.AGCResourceGraph{}, reporting.MigrationReport{}, fmt.Errorf("unsupported provider %q for cluster mode, supported providers: %s, %s", providerName, ProviderNameAGIC, ProviderNameNGINX)
	}

	// collect the resources
	aggregator := aggregation.NewAggregator(opts.Aggregation)

	agicResources, err := aggregator.AggregateFromCluster(ctx, k8sClient, ctrlClient, providerName)
	if err != nil {
		return resources.AGCResourceGraph{}, reporting.MigrationReport{}, err
	}

	return migrate(agicResources, reporting.AGICResourcesSourceCluster, opts)
}

func MigrateFiles(_ context.Context, opts Options, files ...string) (resources.AGCResourceGraph, reporting.MigrationReport, error) {
	// collect the resources
	aggregator := aggregation.NewAggregator(opts.Aggregation)

	agicResources, err := aggregator.AggregateFromFiles(files...)
	if err != nil {
		return resources.AGCResourceGraph{}, reporting.MigrationReport{}, err
	}

	return migrate(agicResources, reporting.AGICResourcesSourceFiles, opts)
}

func migrate(agicResources resources.AGICResources, source reporting.AGICResourcesSource, opts Options) (resources.AGCResourceGraph, reporting.MigrationReport, error) {
	provider, err := makeProvider(opts.ProviderName, agicResources)
	if err != nil {
		return resources.AGCResourceGraph{}, reporting.MigrationReport{}, err
	}

	opts.Conversion.Provider = provider

	result, err := conversion.Convert(agicResources, opts.Conversion)
	if err != nil {
		return result, reporting.MigrationReport{}, err
	}

	report, err := reporting.NewReport(agicResources, result, source, opts.Conversion)

	return result, report, err
}
