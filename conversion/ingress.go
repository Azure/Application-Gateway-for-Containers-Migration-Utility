package conversion

import (
	"errors"
	"fmt"
	"slices"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	networking_v1 "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

func (c converter) handleIngress(graph *resources.AGCResourceGraph, gw *GatewayContext, ingressCtx *resources.IngressContext) error {
	for _, tls := range ingressCtx.Ingress.Spec.TLS {
		if tls.SecretName == "" {
			c.log.Error("ingress TLS section missing secret name", "ingress", k8snames.NamespacedName(&ingressCtx.Ingress))
			continue
		}

		if ingressCtx.Ingress.Namespace != gw.Namespace {
			_ = graph.GetOrCreateReferenceGrantForGWSecret(gw.Namespace, ingressCtx.Ingress.Namespace)
		}
	}

	var errs error

	routes := make(map[string]*gatewayapi_v1.HTTPRoute)

	for _, rule := range ingressCtx.Ingress.Spec.Rules {
		host := rule.Host
		route, ok := routes[host]

		if !ok {
			route = initRoute(graph, gw, *ingressCtx, host)
			routes[host] = route
		}

		ingressCtx.HTTPRoutes = append(ingressCtx.HTTPRoutes, k8snames.NamespacedName(route))

		if rule.HTTP == nil {
			continue
		}

		for _, path := range rule.HTTP.Paths {
			if path.Backend.Resource != nil {
				c.log.Warn("ignoring unsupported Ingress backend resource reference", "ingress", k8snames.NamespacedName(&ingressCtx.Ingress))
				continue
			}

			if path.Backend.Service.Port.Name != "" && path.Backend.Service.Port.Number == 0 {
				// TODO: implement support for this
				c.log.Warn("ignoring unsupported Ingress backend service port name reference", "ingress", k8snames.NamespacedName(&ingressCtx.Ingress))
				continue
			}

			backendRef := gatewayapi_v1.HTTPBackendRef{
				BackendRef: gatewayapi_v1.BackendRef{
					BackendObjectReference: gatewayapi_v1.BackendObjectReference{
						Name: gatewayapi_v1.ObjectName(path.Backend.Service.Name),
						Port: ptr.To(gatewayapi_v1.PortNumber(path.Backend.Service.Port.Number)),
					},
				},
			}

			pathType := gatewayapi_v1.PathMatchExact

			if path.PathType != nil {
				switch *path.PathType {
				case networking_v1.PathTypeExact, networking_v1.PathTypeImplementationSpecific:
					pathType = gatewayapi_v1.PathMatchExact
				case networking_v1.PathTypePrefix:
					pathType = gatewayapi_v1.PathMatchPathPrefix
				default:
					c.log.Warn("Unknown path type for Ingress, defaulting to Exact", "path-type", *path.PathType, "ingress", k8snames.NamespacedName(&ingressCtx.Ingress))

					pathType = gatewayapi_v1.PathMatchExact
				}
			}

			route.Spec.Rules = append(route.Spec.Rules, gatewayapi_v1.HTTPRouteRule{
				Matches: []gatewayapi_v1.HTTPRouteMatch{
					{
						Path: &gatewayapi_v1.HTTPPathMatch{
							Type:  &pathType,
							Value: &path.Path,
						},
					},
				},
				BackendRefs: []gatewayapi_v1.HTTPBackendRef{
					backendRef,
				},
			})
		}

		err := c.handleIngressAnnotations(graph, gw, ingressCtx, NewHTTPRouteContext(route))
		errs = errors.Join(errs, err)

		graph.HTTPRoutes[k8snames.NamespacedName(route)] = route
	}

	return errs
}

func (c converter) handleIngressAnnotations(graph *resources.AGCResourceGraph, gw *GatewayContext, ingressCtx *resources.IngressContext, routeCtx *HTTPRouteContext) error {
	handlers := c.provider.GetAnnotationHandlers()

	var errs error

	for key, annoCtx := range ingressCtx.Annotations {
		if key == resources.LastAppliedConfiguration {
			_ = c.handleIgnoredAnnotation(*graph, gw, routeCtx, ingressCtx, annoCtx)
			continue
		}

		if handler, exists := handlers[key]; exists {
			if err := handler(*graph, gw, routeCtx, ingressCtx, annoCtx); err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to migrate Ingress %s annotation %s=%s: %w", k8snames.NamespacedName(&ingressCtx.Ingress), key, annoCtx.Value, err))
			}
		} else {
			annoCtx.RegisterIssue(resources.NewIssue(resources.IssueUnsupportedAnnotationGeneric, nil))
			c.log.Debug("unsupported annotation", "annotation", fmt.Sprintf("%s=%v", key, annoCtx.Value))
		}
	}

	return errs
}

func (converter) handleIgnoredAnnotation(
	_ resources.AGCResourceGraph,
	_ *GatewayContext,
	_ *HTTPRouteContext,
	_ *resources.IngressContext,
	annoCtx *resources.IngressAnnotationContext,
) error {
	annoCtx.SetStatus(resources.MigrationStatusIgnored)
	return nil
}

func initRoute(graph *resources.AGCResourceGraph, gw *GatewayContext, ingress resources.IngressContext, host string) *gatewayapi_v1.HTTPRoute {
	route := &gatewayapi_v1.HTTPRoute{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindHTTPRoute,
			APIVersion: gatewayapi_v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      common.RouteName(ingress.Ingress.Name, host),
			Namespace: ingress.Ingress.Namespace,
		},
	}

	if host != "" {
		route.Spec.Hostnames = []gatewayapi_v1.Hostname{gatewayapi_v1.Hostname(host)}
	}

	var parentNamespace *gatewayapi_v1.Namespace
	if gw.Namespace != route.Namespace {
		parentNamespace = ptr.To(gatewayapi_v1.Namespace(gw.Namespace))
	}

	if len(ingress.Ingress.Spec.TLS) > 0 {
		var secretNN types.NamespacedName

		var wildcardSecretNN types.NamespacedName

		for _, tls := range ingress.Ingress.Spec.TLS {
			if len(tls.Hosts) == 0 {
				wildcardSecretNN = types.NamespacedName{Name: tls.SecretName, Namespace: ingress.Ingress.Namespace}
			}

			if slices.Contains(tls.Hosts, host) {
				secretNN = types.NamespacedName{Name: tls.SecretName, Namespace: ingress.Ingress.Namespace}
			}
		}

		if secretNN == (types.NamespacedName{}) {
			secretNN = wildcardSecretNN
		}

		if secretNN != (types.NamespacedName{}) {
			route.Spec.ParentRefs = append(route.Spec.ParentRefs, gatewayapi_v1.ParentReference{
				Name:        gatewayapi_v1.ObjectName(gw.Gateway.Name),
				Namespace:   parentNamespace,
				SectionName: ptr.To(gw.EnsureHTTPSListener(graph, host, secretNN)),
			})
		}
	} else {
		route.Spec.ParentRefs = []gatewayapi_v1.ParentReference{
			{
				Name:        gatewayapi_v1.ObjectName(gw.Gateway.Name),
				Namespace:   parentNamespace,
				SectionName: ptr.To(gw.EnsureHTTPListener()),
			},
		}
	}

	return route
}
