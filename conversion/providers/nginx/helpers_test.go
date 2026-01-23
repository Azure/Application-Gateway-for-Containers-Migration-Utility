package nginx

import (
	"k8s.io/utils/ptr"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

// setupAnnotationHandlerInputs creates mock objects for unit testing individual annotation handlers.
// Use this for testing a single handler function in isolation.
func setupAnnotationHandlerInputs(key, val string) (
	Provider,
	resources.AGCResourceGraph,
	*conversion.GatewayContext,
	*conversion.HTTPRouteContext,
	*resources.IngressContext,
	*resources.IngressAnnotationContext,
) {
	provider := NewProvider(resources.NewAGICResources(nil, nil))
	graph := resources.NewAGCResourceGraph()
	gwCtx := conversion.NewGatewayContext(ptr.To(testutil.MakeGateway("default", "gw-1")))
	routeCtx := conversion.NewHTTPRouteContext(ptr.To(testutil.MakeHTTPRoute("default", "route-1", testutil.HTTPRouteOptions{})))
	ingressCtx := resources.NewIngressContext(testutil.MakeIngress("default", "in-1", testutil.IngressOptions{}))
	annoCtx := resources.NewIngressAnnotationContext(key, val)

	return provider, graph, gwCtx, routeCtx, ingressCtx, annoCtx
}
