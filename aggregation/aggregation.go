// Package aggregation is responsible for collecting all of the AGIC resources relevant to migration.
package aggregation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"

	networking_v1 "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl_client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/logging"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

type Aggregator struct {
	log *slog.Logger
	Options
}

type Options struct {
	IngressClassName string
	ClusterLabel     string
}

func NewAggregator(options Options) Aggregator {
	log := logging.New("aggregator")

	return Aggregator{
		log:     log,
		Options: options,
	}
}

func (a Aggregator) ingressClassName(cfg *AGICConfigMap) string {
	if a.Options.IngressClassName != "" {
		return a.Options.IngressClassName
	}

	if cfg != nil && cfg.IngressClassName() != "" {
		return cfg.IngressClassName()
	}

	return DefaultIngressClassName
}

func (a Aggregator) AggregateFromCluster(
	ctx context.Context,
	k8sClient kubernetes.Interface,
	ctrlClient ctrl_client.Client,
	provider string,
) (resources.AGICResources, error) {
	var cfg AGICConfigMap

	var filter ingressFilterFunc

	var clusterLabel string

	// TODO: replace this hack with a provider aggregation interface
	if provider == "agic" {
		clusterLabel = a.Options.ClusterLabel
		if clusterLabel == "" {
			var err error

			clusterLabel, err = findClusterLabel(ctx, k8sClient)
			if err != nil {
				return resources.AGICResources{}, err
			}
		}

		a.log.Debug("AGIC", "label", clusterLabel)

		var err error

		cfg, err = a.loadAGICConfigMap(ctx, clusterLabel, k8sClient)

		if err != nil {
			return resources.AGICResources{}, fmt.Errorf("failed to load & parse AGIC ConfigMap: %w", err)
		}

		filter = a.ingressClassNameFilter(a.ingressClassName(&cfg))
	} else {
		filter = a.ingressClassNameFilter("nginx")
	}

	ingresses, err := a.collectIngressesFromCluster(ctx, ctrlClient, filter)
	if err != nil {
		return resources.AGICResources{}, fmt.Errorf("failed to collect Ingresses from cluster: %w", err)
	}

	var wafPolicyID string

	var rewriteRuleSets []appgwrewrite.AzureApplicationGatewayRewrite

	if provider == "agic" {
		rewriteRuleSets, err = collectRewriteRuleSetsFromCluster(ctx, ctrlClient)
		if err != nil {
			return resources.AGICResources{}, fmt.Errorf("failed to collect AzureApplicationGatewayRewrite from cluster: %w", err)
		}

		if cfg.AppGwSKU == SKUWAFv2 || cfg.AppGwSKU == "" {
			wafPolicyID, err = scrapeAGICLogsForWAFPolicyID(ctx, clusterLabel, k8sClient)
			if err != nil {
				// we'll log the error but continue
				a.log.Error("Failed to identify WAF Policy ID from Ingress Controller pod log", "error", err)
			} else if wafPolicyID != "" {
				a.log.Info("Found WAF policy", "ID", wafPolicyID)
			}
		}
	}

	agicResources := resources.NewAGICResources(ingressMapToSlice(ingresses), rewriteRuleSets)
	agicResources.WAFPolicyID = wafPolicyID
	a.logResult(agicResources)

	return agicResources, nil
}

func (a Aggregator) AggregateFromFiles(filepaths ...string) (resources.AGICResources, error) {
	a.log.Debug("collecting Ingresses from files", "paths", filepaths)

	ingresses, err := a.collectIngressesFromFiles(a.ingressClassNameFilter(a.ingressClassName(nil)), filepaths...)

	if err != nil {
		return resources.AGICResources{}, fmt.Errorf("failed to collect Ingresses from files: %w", err)
	}

	rewriteRuleSets, err := collectRewriteRuleSetsFromFiles(filepaths...)
	if err != nil {
		return resources.AGICResources{}, fmt.Errorf("failed to collect AzureApplicationGatewayRewrite from files: %w", err)
	}

	result := resources.NewAGICResources(ingressMapToSlice(ingresses), rewriteRuleSets)
	a.logResult(result)

	return result, nil
}

func (a Aggregator) logResult(result resources.AGICResources) {
	a.log.Debug("collected ingresses", "count", len(result.IngressContexts))

	if len(result.AppGWRewrites) > 0 {
		a.log.Debug("collected azureapplicationgatewayrewrites", "count", len(result.AppGWRewrites))
	}
}

func ingressMapToSlice(ingressMap map[types.NamespacedName]*networking_v1.Ingress) []networking_v1.Ingress {
	var ingresses []networking_v1.Ingress

	for _, ingress := range ingressMap {
		ingresses = append(ingresses, *ingress)
	}

	return ingresses
}

func findClusterLabel(ctx context.Context, client kubernetes.Interface) (string, error) {
	testLabel := func(label string) (bool, error) {
		pods, err := client.CoreV1().Pods("").List(ctx, meta_v1.ListOptions{
			LabelSelector: label,
		})
		if err != nil {
			return false, fmt.Errorf("failed to list ingress controller pods: %w", err)
		}

		if len(pods.Items) == 0 {
			return false, fmt.Errorf("no pods found for label selector: %q", label)
		}

		return true, nil
	}

	ok, helmErr := testLabel(agicHelmLabel)
	if ok {
		return agicHelmLabel, nil
	}

	ok, addonErr := testLabel(agicAddonLabel)
	if ok {
		return agicAddonLabel, nil
	}

	err := errors.New("failed to find determine label being used by AGIC")

	return "", errors.Join(err, helmErr, addonErr)
}
