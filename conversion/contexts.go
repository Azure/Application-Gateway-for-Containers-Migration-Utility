package conversion

import (
	"cmp"
	"fmt"
	"maps"
	"slices"

	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

const httpListenerName = gatewayapi_v1.SectionName("http")

type GatewayContext struct {
	*gatewayapi_v1.Gateway
	HTTPListener     *gatewayapi_v1.Listener
	HTTPSListeners   map[ListenerKey]HTTPSListener
	HasHTTPSRedirect bool
	issues           []resources.Issue
}

func NewGatewayContext(gateway *gatewayapi_v1.Gateway) *GatewayContext {
	return &GatewayContext{
		Gateway:        gateway,
		HTTPSListeners: make(map[ListenerKey]HTTPSListener),
	}
}

type HTTPRouteContext struct {
	*gatewayapi_v1.HTTPRoute
	namedParentSections sets.Set[gatewayapi_v1.SectionName]
}

func NewHTTPRouteContext(route *gatewayapi_v1.HTTPRoute) *HTTPRouteContext {
	namedParentSections := sets.New[gatewayapi_v1.SectionName]()

	for _, parentRef := range route.Spec.ParentRefs {
		if parentRef.SectionName == nil {
			continue
		}

		namedParentSections.Insert(*parentRef.SectionName)
	}

	return &HTTPRouteContext{
		HTTPRoute:           route,
		namedParentSections: namedParentSections,
	}
}

func (h *HTTPRouteContext) EnsureHTTPSListener(graph *resources.AGCResourceGraph, gatewayCtx GatewayContext, hostname string, secretName string, secretNamespace string) {
	secretNN := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
	sectionName := gatewayCtx.EnsureHTTPSListener(graph, hostname, secretNN)

	if !h.namedParentSections.Has(sectionName) {
		h.Spec.ParentRefs = append(h.Spec.ParentRefs, gatewayapi_v1.ParentReference{
			Name:        gatewayapi_v1.ObjectName(gatewayCtx.Gateway.Name),
			Namespace:   ptr.To(gatewayapi_v1.Namespace(gatewayCtx.Gateway.Namespace)),
			Kind:        ptr.To(gatewayapi_v1.Kind(k8snames.KindGateway)),
			Group:       ptr.To(gatewayapi_v1.Group(k8snames.GatewayNetworkingGroup)),
			SectionName: ptr.To(sectionName),
		})
	}
}

func (h HTTPRouteContext) GetParentHTTPSListenerNames(gatewayCtx GatewayContext) []gatewayapi_v1.SectionName {
	listenerNames := sets.New[gatewayapi_v1.SectionName]()

	for _, parentRef := range h.Spec.ParentRefs {
		if parentRef.SectionName != nil {
			listenerNames.Insert(*parentRef.SectionName)
		}
	}

	httpsListenerNames := []gatewayapi_v1.SectionName{}

	for _, listener := range gatewayCtx.HTTPSListeners {
		if listenerNames.Has(listener.Name) {
			httpsListenerNames = append(httpsListenerNames, listener.Name)
		}
	}

	return httpsListenerNames
}

func (g *GatewayContext) RegisterIssue(issue resources.Issue) {
	g.issues = append(g.issues, issue)
}

func (g *GatewayContext) build() *gatewayapi_v1.Gateway {
	g.Gateway.Spec.Listeners = []gatewayapi_v1.Listener{}
	if g.HTTPListener != nil {
		g.Gateway.Spec.Listeners = append(g.Gateway.Spec.Listeners, *g.HTTPListener)
	}

	compareFn := func(a, b ListenerKey) int {
		if a.Port != b.Port {
			return cmp.Compare(a.Port, b.Port)
		}

		return cmp.Compare(a.Hostname, b.Hostname)
	}
	for _, key := range slices.SortedFunc(maps.Keys(g.HTTPSListeners), compareFn) {
		g.Gateway.Spec.Listeners = append(g.Gateway.Spec.Listeners, *g.HTTPSListeners[key].Listener)
	}

	return g.Gateway
}

type ListenerKey struct {
	Hostname string
	Port     int32
}

type HTTPSListener struct {
	Secrets sets.Set[types.NamespacedName]
	*gatewayapi_v1.Listener
}

func (hl *HTTPSListener) AddSecret(secretNN types.NamespacedName) {
	if hl.Secrets.Has(secretNN) {
		return
	}
	// TODO: I think this is actually an error?
	hl.Secrets.Insert(secretNN)
	hl.TLS.CertificateRefs = append(hl.TLS.CertificateRefs, gatewayapi_v1.SecretObjectReference{
		Kind:      ptr.To(gatewayapi_v1.Kind(k8snames.KindSecret)),
		Name:      gatewayapi_v1.ObjectName(secretNN.Name),
		Namespace: ptr.To(gatewayapi_v1.Namespace(secretNN.Namespace)),
	})
}

func (g *GatewayContext) EnsureHTTPSListener(graph *resources.AGCResourceGraph, hostname string, secretNN types.NamespacedName) gatewayapi_v1.SectionName {
	if secretNN.Namespace != g.Namespace {
		// create ReferenceGrant
		_ = graph.GetOrCreateReferenceGrantForGWSecret(g.Namespace, secretNN.Namespace)
	}

	listener, ok := g.HTTPSListeners[ListenerKey{Hostname: hostname, Port: 443}]
	if ok {
		listener.AddSecret(secretNN)
		return listener.Name
	}

	var secretNamespace *gatewayapi_v1.Namespace
	if secretNN.Namespace != g.Namespace {
		secretNamespace = ptr.To(gatewayapi_v1.Namespace(secretNN.Namespace))
	}

	listener = HTTPSListener{
		Secrets: sets.New(secretNN),
		Listener: &gatewayapi_v1.Listener{
			Name:     gatewayapi_v1.SectionName(fmt.Sprintf("https-%s", common.NameFromHost(hostname))),
			Port:     443,
			Protocol: gatewayapi_v1.HTTPSProtocolType,
			Hostname: ptr.To(gatewayapi_v1.Hostname(hostname)),
			AllowedRoutes: &gatewayapi_v1.AllowedRoutes{
				Namespaces: &gatewayapi_v1.RouteNamespaces{
					From: ptr.To(gatewayapi_v1.NamespacesFromAll),
				},
			},
			TLS: &gatewayapi_v1.GatewayTLSConfig{
				CertificateRefs: []gatewayapi_v1.SecretObjectReference{
					{
						Kind:      ptr.To(gatewayapi_v1.Kind(k8snames.KindSecret)),
						Name:      gatewayapi_v1.ObjectName(secretNN.Name),
						Namespace: secretNamespace,
						Group:     ptr.To(gatewayapi_v1.Group("")),
					},
				},
			},
		},
	}

	key := ListenerKey{Hostname: hostname, Port: 443}
	g.HTTPSListeners[key] = listener

	return listener.Name
}

func (g *GatewayContext) EnsureHTTPListener() gatewayapi_v1.SectionName {
	if g.HTTPListener != nil {
		return g.HTTPListener.Name
	}

	g.HTTPListener = &gatewayapi_v1.Listener{
		Name:     httpListenerName,
		Port:     80,
		Protocol: gatewayapi_v1.HTTPProtocolType,
		AllowedRoutes: &gatewayapi_v1.AllowedRoutes{
			Namespaces: &gatewayapi_v1.RouteNamespaces{
				From: ptr.To(gatewayapi_v1.NamespacesFromAll),
			},
		},
	}

	return g.HTTPListener.Name
}
