package aggregation

import (
	"context"
	"fmt"
	"os"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	k8s_yaml "k8s.io/apimachinery/pkg/util/yaml"
	ctrl_client "sigs.k8s.io/controller-runtime/pkg/client"
)

func collectRewriteRuleSetsFromFiles(filepath ...string) ([]appgwrewrite.AzureApplicationGatewayRewrite, error) {
	var rewrites []appgwrewrite.AzureApplicationGatewayRewrite

	for _, path := range filepath {
		var ruleset appgwrewrite.AzureApplicationGatewayRewrite

		bytes, err := os.ReadFile(path) // nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %s: %w", path, err)
		}

		if err := k8s_yaml.Unmarshal(bytes, &ruleset); err != nil {
			return nil, fmt.Errorf("failed to unmarshal AzureApplicationGatewayRewrite: %s: %w", path, err)
		}

		if ruleset.Kind != "AzureApplicationGatewayRewrite" || ruleset.APIVersion != "appgw.ingress.azure.io/v1beta1" {
			continue
		}

		rewrites = append(rewrites, ruleset)
	}

	return rewrites, nil
}

func collectRewriteRuleSetsFromCluster(ctx context.Context, client ctrl_client.Client) ([]appgwrewrite.AzureApplicationGatewayRewrite, error) {
	var rewriteRuleSets appgwrewrite.AzureApplicationGatewayRewriteList
	if err := client.List(ctx, &rewriteRuleSets); err != nil {
		return nil, fmt.Errorf("failed to list AzureApplicationGatewayRewrite resources: %w", err)
	}

	return rewriteRuleSets.Items, nil
}
