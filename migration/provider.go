package migration

import (
	"fmt"
	"strings"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion/providers/agic"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion/providers/nginx"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

// Provider name constants
// todo: integrate this with the (existing or a new) provider interface
const (
	ProviderNameAGIC  = "agic"
	ProviderNameNGINX = "nginx"
)

func makeProvider(name string, inputs resources.AGICResources) (conversion.Provider, error) {
	// If no provider specified, try to auto-detect
	if name == "" {
		name = detectProvider(inputs)
	}

	switch strings.ToLower(name) {
	case "", ProviderNameAGIC:
		return agic.NewProvider(inputs), nil
	case ProviderNameNGINX:
		return nginx.NewProvider(inputs), nil
	}

	return nil, fmt.Errorf("unknown provider: %q, supported providers are %q, %q", name, ProviderNameAGIC, ProviderNameNGINX)
}

// detectProvider attempts to auto-detect the ingress controller provider
// based on the ingress class annotations in the input resources.
// Returns:
//   - "nginx" if any ingress has class "nginx"
//   - "agic" if any ingress has class "azure/application-gateway"
//   - empty string if no match found (will default to "agic")
func detectProvider(inputs resources.AGICResources) string {
	for _, ingressCtx := range inputs.IngressContexts {
		ingress := ingressCtx.Ingress

		// Check ingressClassName field (preferred in newer Kubernetes)
		if ingress.Spec.IngressClassName != nil {
			className := strings.ToLower(*ingress.Spec.IngressClassName)
			if className == ProviderNameNGINX {
				return ProviderNameNGINX
			}

			if className == "azure-application-gateway" || strings.Contains(className, ProviderNameAGIC) {
				return ProviderNameAGIC
			}
		}

		// Check kubernetes.io/ingress.class annotation (legacy)
		if className, ok := ingress.Annotations["kubernetes.io/ingress.class"]; ok {
			className = strings.ToLower(className)
			if className == ProviderNameNGINX {
				return ProviderNameNGINX
			}

			if className == "azure/application-gateway" || strings.Contains(className, ProviderNameAGIC) {
				return ProviderNameAGIC
			}
		}
	}

	return ""
}
