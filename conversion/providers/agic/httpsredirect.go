package agic

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

const (
	httpRouteRedirectname = "https-redirect"
)

func (p Provider) handleSSLRedirect(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	if gwCtx.HasHTTPSRedirect {
		return nil
	}

	httpListenerName := gwCtx.EnsureHTTPListener()

	redirectRoute := &gatewayapi_v1.HTTPRoute{
		TypeMeta: v1.TypeMeta{
			Kind:       k8snames.KindHTTPRoute,
			APIVersion: gatewayapi_v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      httpRouteRedirectname,
			Namespace: gwCtx.Namespace,
		},
		Spec: gatewayapi_v1.HTTPRouteSpec{
			CommonRouteSpec: gatewayapi_v1.CommonRouteSpec{
				ParentRefs: []gatewayapi_v1.ParentReference{
					{
						Group:       (*gatewayapi_v1.Group)(&gatewayapi_v1.GroupVersion.Group),
						Kind:        (*gatewayapi_v1.Kind)(ptr.To(k8snames.KindGateway)),
						Name:        gatewayapi_v1.ObjectName(gwCtx.Name),
						Namespace:   (*gatewayapi_v1.Namespace)(&gwCtx.Namespace),
						SectionName: ptr.To(httpListenerName),
					},
				},
			},
			Rules: []gatewayapi_v1.HTTPRouteRule{
				{
					Matches: []gatewayapi_v1.HTTPRouteMatch{
						{
							Path: &gatewayapi_v1.HTTPPathMatch{
								Type:  ptr.To(gatewayapi_v1.PathMatchPathPrefix),
								Value: ptr.To("/"),
							},
						},
					},
					Filters: []gatewayapi_v1.HTTPRouteFilter{
						{
							Type: gatewayapi_v1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayapi_v1.HTTPRequestRedirectFilter{
								Scheme: ptr.To("https"),
							},
						},
					},
				},
			},
		},
	}
	output.HTTPRoutes[k8snames.NamespacedName(redirectRoute)] = redirectRoute
	annotationCtx.AddDestination(resources.NewK8sResourceID(redirectRoute))

	gwCtx.HasHTTPSRedirect = true

	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}
