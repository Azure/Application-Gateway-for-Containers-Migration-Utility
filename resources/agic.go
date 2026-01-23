package resources

import (
	"sort"

	networking_v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
)

// AGICResources covers all AGIC resources in scope for migration
type AGICResources struct {
	IngressContexts map[types.NamespacedName]*IngressContext
	Services        map[types.NamespacedName]*ServiceContext
	AppGWRewrites   map[types.NamespacedName]*AppGWRewriteContext
	WAFPolicyID     string
}

func NewAGICResources(ingresses []networking_v1.Ingress, rewrites []appgwrewrite.AzureApplicationGatewayRewrite) AGICResources {
	ingressMap := make(map[types.NamespacedName]*IngressContext, len(ingresses))
	rewriteMap := make(map[types.NamespacedName]*AppGWRewriteContext, len(rewrites))

	for _, ing := range ingresses {
		ingressMap[types.NamespacedName{Namespace: ing.Namespace, Name: ing.Name}] = NewIngressContext(ing)
	}

	for _, rewrite := range rewrites {
		rewriteMap[types.NamespacedName{Namespace: rewrite.Namespace, Name: rewrite.Name}] = NewAppGWRewriteContext(rewrite)
	}

	return AGICResources{
		IngressContexts: ingressMap,
		AppGWRewrites:   rewriteMap,
	}
}

func (a AGICResources) Ingresses() []networking_v1.Ingress {
	ingresses := make([]networking_v1.Ingress, 0, len(a.IngressContexts))
	for _, ingress := range a.IngressContexts {
		ingresses = append(ingresses, ingress.Ingress)
	}

	sort.Slice(ingresses, func(i, j int) bool {
		if ingresses[i].Namespace != ingresses[j].Namespace {
			return ingresses[i].Namespace < ingresses[j].Namespace
		}

		return ingresses[i].Name < ingresses[j].Name
	})

	return ingresses
}
