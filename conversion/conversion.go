// Package conversion is responsible for the process of converting AGIC resources to AGC resources.
package conversion

import (
	"log/slog"
	"sort"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/logging"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

type converter struct {
	log      *slog.Logger
	input    resources.AGICResources
	provider Provider
}

type Options struct {
	BYOResourceID    string
	ManagedSubnetID  string
	GatewayNamespace string
	GatewayWAFID     string
	Provider         Provider
}

func newConverter(input resources.AGICResources, provider Provider) converter {
	return converter{
		log:      logging.New("converter"),
		input:    input,
		provider: provider,
	}
}

// Convert receives the collection of AGIC resources and produces a set of AGC resources.
// It updates the status of the input resources as it goes.
func Convert(input resources.AGICResources, opts Options) (resources.AGCResourceGraph, error) {
	converter := newConverter(input, opts.Provider)

	// create a sorted list of Ingress namespaced names to ensure deterministic processing order
	var sortedIngressNNs []types.NamespacedName
	for ingressNN := range input.IngressContexts {
		sortedIngressNNs = append(sortedIngressNNs, ingressNN)
	}

	sort.Slice(sortedIngressNNs, func(i, j int) bool {
		if sortedIngressNNs[i].Namespace != sortedIngressNNs[j].Namespace {
			return sortedIngressNNs[i].Namespace < sortedIngressNNs[j].Namespace
		}

		return sortedIngressNNs[i].Name < sortedIngressNNs[j].Name
	})

	gwNamespace := opts.GatewayNamespace
	if gwNamespace == "" {
		if len(sortedIngressNNs) > 0 {
			gwNamespace = sortedIngressNNs[0].Namespace
		} else {
			gwNamespace = "default"
		}
	}

	graph := resources.NewAGCResourceGraph()
	gwContext := &GatewayContext{
		Gateway: &gatewayapi_v1.Gateway{
			TypeMeta: meta_v1.TypeMeta{
				Kind:       k8snames.KindGateway,
				APIVersion: gatewayapi_v1.GroupVersion.String(),
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:        "alb-gateway",
				Namespace:   gwNamespace,
				Annotations: make(map[string]string),
			},
			Spec: gatewayapi_v1.GatewaySpec{
				GatewayClassName: k8snames.ALBExternalGatewayClassName,
				Listeners:        []gatewayapi_v1.Listener{},
			},
		},
		HTTPSListeners: make(map[ListenerKey]HTTPSListener),
	}

	if opts.BYOResourceID != "" {
		gwContext.Gateway.Annotations[k8snames.ALBArmResourceIDAnnotation] = opts.BYOResourceID
	} else if opts.ManagedSubnetID != "" {
		// create the ALB object
		alb := albcontrollerapi_v1.ApplicationLoadBalancer{
			TypeMeta: meta_v1.TypeMeta{
				Kind:       k8snames.KindApplicationLoadBalancer,
				APIVersion: albcontrollerapi_v1.GroupVersion.String(),
			},
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      "alb",
				Namespace: gwNamespace,
			},
			Spec: albcontrollerapi_v1.AlbSpec{
				Associations: []string{opts.ManagedSubnetID},
			},
		}
		gwContext.Gateway.Annotations[k8snames.ALBK8sResourceNameAnnotation] = alb.Name
		gwContext.Gateway.Annotations[k8snames.ALBK8sResourceNamespaceAnnotation] = alb.Namespace
		graph.ApplicationLoadBalancer = &alb
	}

	// the explicitly set WAF ID takes precedence
	if opts.GatewayWAFID != "" {
		converter.handleWAFPolicyForGateway(graph, gwContext, opts.GatewayWAFID)
	} else if input.WAFPolicyID != "" {
		converter.handleWAFPolicyForGateway(graph, gwContext, input.WAFPolicyID)
	}

	for _, ingressNN := range sortedIngressNNs {
		converter.log.Debug("Processing Ingress", "ingress", ingressNN)
		err := converter.handleIngress(&graph, gwContext, input.IngressContexts[ingressNN])
		ingressCtx := input.IngressContexts[ingressNN]
		ingressCtx.MigrationComplete(err)
	}

	graph.Gateway = gwContext.build()

	return graph, nil
}
