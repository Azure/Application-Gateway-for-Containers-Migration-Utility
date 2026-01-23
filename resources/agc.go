package resources

import (
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapi_v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

// AGCResourceGraph represents the output result of the conversion process.
type AGCResourceGraph struct {
	ApplicationLoadBalancer *albcontrollerapi_v1.ApplicationLoadBalancer                               `yaml:"applicationLoadBalancer"`
	Gateway                 *gatewayapi_v1.Gateway                                                     `yaml:"gateway"`
	HTTPRoutes              map[types.NamespacedName]*gatewayapi_v1.HTTPRoute                          `yaml:"httpRoutes"`
	ReferenceGrants         map[types.NamespacedName]*gatewayapi_v1beta1.ReferenceGrant                `yaml:"referenceGrants"`
	BackendTLSPolicies      map[types.NamespacedName]*albcontrollerapi_v1.BackendTLSPolicy             `yaml:"backendTLSPolicies"`
	FrontendTLSPolicies     map[types.NamespacedName]*albcontrollerapi_v1.FrontendTLSPolicy            `yaml:"frontendTLSPolicies"`
	HealthCheckPolicies     map[types.NamespacedName]*albcontrollerapi_v1.HealthCheckPolicy            `yaml:"healthCheckPolicies"`
	RoutePolicies           map[types.NamespacedName]*albcontrollerapi_v1.RoutePolicy                  `yaml:"routePolicies"`
	WAFPolicies             map[types.NamespacedName]*albcontrollerapi_v1.WebApplicationFirewallPolicy `yaml:"wafPolicies"`
}

func NewAGCResourceGraph() AGCResourceGraph {
	return AGCResourceGraph{
		ReferenceGrants:     make(map[types.NamespacedName]*gatewayapi_v1beta1.ReferenceGrant),
		HTTPRoutes:          make(map[types.NamespacedName]*gatewayapi_v1.HTTPRoute),
		HealthCheckPolicies: make(map[types.NamespacedName]*albcontrollerapi_v1.HealthCheckPolicy),
		BackendTLSPolicies:  make(map[types.NamespacedName]*albcontrollerapi_v1.BackendTLSPolicy),
		FrontendTLSPolicies: make(map[types.NamespacedName]*albcontrollerapi_v1.FrontendTLSPolicy),
		RoutePolicies:       make(map[types.NamespacedName]*albcontrollerapi_v1.RoutePolicy),
		WAFPolicies:         make(map[types.NamespacedName]*albcontrollerapi_v1.WebApplicationFirewallPolicy),
	}
}

func (a *AGCResourceGraph) GetOrCreateReferenceGrantForGWSecret(gatewayNamespace, secretNamespace string) *gatewayapi_v1beta1.ReferenceGrant {
	grant := referenceGrantForGatewayToKind(gatewayNamespace, secretNamespace, k8snames.KindSecret)
	if val, ok := a.ReferenceGrants[types.NamespacedName{Name: grant.Name, Namespace: grant.Namespace}]; ok {
		return val
	}

	a.ReferenceGrants[types.NamespacedName{Name: grant.Name, Namespace: grant.Namespace}] = grant

	return grant
}

func referenceGrantForGatewayToKind(gatewayNamespace, secretNamespace, kind string) *gatewayapi_v1beta1.ReferenceGrant {
	name := fmt.Sprintf("gateway-in-%s-%s-grant", gatewayNamespace, kind)

	return &gatewayapi_v1beta1.ReferenceGrant{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindReferenceGrant,
			APIVersion: gatewayapi_v1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: secretNamespace,
		},
		Spec: gatewayapi_v1beta1.ReferenceGrantSpec{
			From: []gatewayapi_v1beta1.ReferenceGrantFrom{
				{
					Namespace: gatewayapi_v1beta1.Namespace(gatewayNamespace),
					Group:     gatewayapi_v1.GroupName,
					Kind:      k8snames.KindGateway,
				},
			},
			To: []gatewayapi_v1beta1.ReferenceGrantTo{
				{
					Kind: gatewayapi_v1.Kind(kind),
				},
			},
		},
	}
}

func (a *AGCResourceGraph) GetOrCreateRoutePolicy(routeName types.NamespacedName) *albcontrollerapi_v1.RoutePolicy {
	policy := routePolicy(routeName)
	return getOrInsert(a.RoutePolicies, &policy)
}

func routePolicy(route types.NamespacedName) albcontrollerapi_v1.RoutePolicy {
	name := fmt.Sprintf("%s-policy", route.Name)

	return albcontrollerapi_v1.RoutePolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindRoutePolicy,
			APIVersion: k8snames.ALBAPIVersion,
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: route.Namespace,
		},
		Spec: albcontrollerapi_v1.RoutePolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     v1alpha2.Group(k8snames.GatewayNetworkingGroup),
					Kind:      k8snames.KindHTTPRoute,
					Name:      v1alpha2.ObjectName(route.Name),
					Namespace: (*v1alpha2.Namespace)(&route.Namespace),
				},
			},
			Default: &albcontrollerapi_v1.RoutePolicyConfig{},
		},
	}
}

func (a *AGCResourceGraph) GetOrCreateFrontendTLSPolicy(gatewayName types.NamespacedName, listener gatewayapi_v1.SectionName) *albcontrollerapi_v1.FrontendTLSPolicy {
	policy := frontendTLSPolicy(gatewayName, listener)
	return getOrInsert(a.FrontendTLSPolicies, &policy)
}

func frontendTLSPolicy(gatewayNN types.NamespacedName, listener gatewayapi_v1.SectionName) albcontrollerapi_v1.FrontendTLSPolicy {
	name := fmt.Sprintf("%s-%s-policy", gatewayNN.Name, listener)

	return albcontrollerapi_v1.FrontendTLSPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindFrontendTLSPolicy,
			APIVersion: k8snames.ALBAPIVersion,
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: gatewayNN.Namespace,
		},
		Spec: albcontrollerapi_v1.FrontendTLSPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group: gatewayapi_v1.GroupName,
					Kind:  k8snames.KindGateway,
					Name:  v1alpha2.ObjectName(gatewayNN.Name),
				},
				SectionNames: []string{string(listener)},
			},
			Default: &albcontrollerapi_v1.FrontendTLSPolicyConfig{},
		},
	}
}

func (a *AGCResourceGraph) GetOrCreateHealthCheckPolicy(serviceName types.NamespacedName) *albcontrollerapi_v1.HealthCheckPolicy {
	policy := healthCheckPolicy(serviceName)
	return getOrInsert(a.HealthCheckPolicies, &policy)
}

func healthCheckPolicy(service types.NamespacedName) albcontrollerapi_v1.HealthCheckPolicy {
	name := fmt.Sprintf("%s-policy", service.Name)

	return albcontrollerapi_v1.HealthCheckPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindHealthCheckPolicy,
			APIVersion: k8snames.ALBAPIVersion,
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: service.Namespace,
		},
		Spec: albcontrollerapi_v1.HealthCheckPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     "",
					Kind:      k8snames.KindService,
					Name:      v1alpha2.ObjectName(service.Name),
					Namespace: (*v1alpha2.Namespace)(&service.Namespace),
				},
			},
			Default: &albcontrollerapi_v1.HealthCheckPolicyConfig{},
		},
	}
}

func (a *AGCResourceGraph) GetOrCreateWAFPolicyForGateway(gateway types.NamespacedName) *albcontrollerapi_v1.WebApplicationFirewallPolicy {
	targetRef := v1alpha2.NamespacedPolicyTargetReference{
		Group:     k8snames.GatewayNetworkingGroup,
		Kind:      k8snames.KindGateway,
		Name:      v1alpha2.ObjectName(gateway.Name),
		Namespace: (*v1alpha2.Namespace)(&gateway.Namespace),
	}
	policy := wafPolicy("gateway", targetRef)

	return getOrInsert(a.WAFPolicies, &policy)
}

func (a *AGCResourceGraph) GetOrCreateWAFPolicyForRoute(route types.NamespacedName) *albcontrollerapi_v1.WebApplicationFirewallPolicy {
	targetRef := v1alpha2.NamespacedPolicyTargetReference{
		Group:     k8snames.GatewayNetworkingGroup,
		Kind:      k8snames.KindHTTPRoute,
		Name:      v1alpha2.ObjectName(route.Name),
		Namespace: (*v1alpha2.Namespace)(&route.Namespace),
	}
	policy := wafPolicy("httproute", targetRef)

	return getOrInsert(a.WAFPolicies, &policy)
}

func wafPolicy(kind string, targetRef v1alpha2.NamespacedPolicyTargetReference) albcontrollerapi_v1.WebApplicationFirewallPolicy {
	name := fmt.Sprintf("%s-%s-%s-policy", kind, *targetRef.Namespace, targetRef.Name)

	return albcontrollerapi_v1.WebApplicationFirewallPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindWebApplicationFirewallPolicy,
			APIVersion: k8snames.ALBAPIVersion,
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: string(*targetRef.Namespace),
		},
		Spec: albcontrollerapi_v1.WebApplicationFirewallPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: targetRef,
			},
			WebApplicationFirewallPolicy: &albcontrollerapi_v1.WebApplicationFirewallConfig{},
		},
	}
}

func (a *AGCResourceGraph) GetOrCreateBackendTLSPolicy(serviceName types.NamespacedName) *albcontrollerapi_v1.BackendTLSPolicy {
	policy := backendTLSPolicy(serviceName)
	return getOrInsert(a.BackendTLSPolicies, &policy)
}

func backendTLSPolicy(service types.NamespacedName) albcontrollerapi_v1.BackendTLSPolicy {
	name := fmt.Sprintf("%s-policy", service.Name)

	return albcontrollerapi_v1.BackendTLSPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindBackendTLSPolicy,
			APIVersion: k8snames.ALBAPIVersion,
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: service.Namespace,
		},
		Spec: albcontrollerapi_v1.BackendTLSPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     "",
					Kind:      k8snames.KindService,
					Name:      v1alpha2.ObjectName(service.Name),
					Namespace: (*v1alpha2.Namespace)(&service.Namespace),
				},
			},
			Default: &albcontrollerapi_v1.BackendTLSPolicyConfig{},
		},
	}
}

type policy interface {
	GetNamespacedName() types.NamespacedName
}

func getOrInsert[V policy](m map[types.NamespacedName]V, val V) V {
	key := val.GetNamespacedName()
	if val, ok := m[key]; ok {
		return val
	}

	m[key] = val

	return val
}
