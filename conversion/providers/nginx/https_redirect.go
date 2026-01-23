package nginx

import (
	"strconv"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

// ========== SSL Redirect Handlers ==========

func (p Provider) handleSSLRedirect(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	_ *conversion.HTTPRouteContext,
	_ *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// Check if redirect is enabled
	enableRedirect, err := strconv.ParseBool(annotationCtx.Value)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))
		annotationCtx.SetStatus(resources.MigrationStatusError)

		return err
	}

	if !enableRedirect {
		annotationCtx.SetStatus(resources.MigrationStatusIgnored)
		return nil
	}

	if gwCtx.HasHTTPSRedirect {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	}

	httpListenerName := gwCtx.EnsureHTTPListener()

	redirectRoute := &gatewayapi_v1.HTTPRoute{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "HTTPRoute",
			APIVersion: gatewayapi_v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "https-redirect",
			Namespace: gwCtx.Gateway.Namespace,
		},
		Spec: gatewayapi_v1.HTTPRouteSpec{
			CommonRouteSpec: gatewayapi_v1.CommonRouteSpec{
				ParentRefs: []gatewayapi_v1.ParentReference{
					{
						Group:       (*gatewayapi_v1.Group)(&gatewayapi_v1.GroupVersion.Group),
						Kind:        (*gatewayapi_v1.Kind)(ptr.To("Gateway")),
						Name:        gatewayapi_v1.ObjectName(gwCtx.Gateway.Name),
						Namespace:   (*gatewayapi_v1.Namespace)(&gwCtx.Gateway.Namespace),
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

func (p Provider) handleForceSSLRedirect(
	output resources.AGCResourceGraph,
	gwCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	// force-ssl-redirect works the same as ssl-redirect but forces even on HTTP scheme
	forceRedirect, err := strconv.ParseBool(annotationCtx.Value)
	if err != nil {
		annotationCtx.RegisterIssue(resources.NewIssue(resources.IssueInvalidAnnotationValue, err))
		return err
	}

	if !forceRedirect {
		annotationCtx.SetStatus(resources.MigrationStatusCompleted)
		return nil
	}

	return p.handleSSLRedirect(output, gwCtx, routeCtx, ingressCtx, annotationCtx)
}
