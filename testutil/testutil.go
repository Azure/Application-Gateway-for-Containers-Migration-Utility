// Package testutil has test helpers for AGIC Migration packages.
package testutil

import (
	"strconv"
	"time"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"
	network_v1 "k8s.io/api/networking/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	core_v1 "k8s.io/api/core/v1"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayapi_v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

func MakeALB(namespace, name, subnetID string) albcontrollerapi_v1.ApplicationLoadBalancer {
	return albcontrollerapi_v1.ApplicationLoadBalancer{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindApplicationLoadBalancer,
			APIVersion: albcontrollerapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: albcontrollerapi_v1.AlbSpec{
			Associations: []string{
				subnetID,
			},
		},
	}
}

func MakeIngressList(ingressClass string, count int) *network_v1.IngressList {
	items := make([]network_v1.Ingress, 0, count)
	for i := range count {
		items = append(items, network_v1.Ingress{
			ObjectMeta: meta_v1.ObjectMeta{
				Name: "ingress-" + strconv.Itoa(i),
			},
			Spec: network_v1.IngressSpec{
				IngressClassName: &ingressClass,
			},
		})
	}

	return &network_v1.IngressList{
		Items: items,
	}
}

type IngressOptions struct {
	Host              string
	Path              string
	Service           string
	ServicePortNumber int
}

func MakeIngress(namespace string, name string, options IngressOptions, annotations ...string) network_v1.Ingress {
	ingress := network_v1.Ingress{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindIngress,
			APIVersion: network_v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Spec: network_v1.IngressSpec{
			IngressClassName: ptr.To("azure/ingress-class"),
			Rules: []network_v1.IngressRule{
				{
					IngressRuleValue: network_v1.IngressRuleValue{
						HTTP: &network_v1.HTTPIngressRuleValue{
							Paths: []network_v1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: ptr.To(network_v1.PathTypePrefix),
									Backend: network_v1.IngressBackend{
										Service: &network_v1.IngressServiceBackend{
											Name: "service",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if options.Path != "" {
		ingress.Spec.Rules[0].HTTP.Paths[0].Path = options.Path
	}

	if options.Host != "" {
		ingress.Spec.Rules[0].Host = options.Host
	}

	if options.Service != "" {
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name = options.Service
	}

	if options.ServicePortNumber != 0 {
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number = int32(options.ServicePortNumber)
	}

	if len(annotations)%2 != 0 {
		panic("annotations must be key/value pairs")
	}

	for i := 0; i < len(annotations); i += 2 {
		ingress.Annotations[annotations[i]] = annotations[i+1]
	}

	return ingress
}

func MakeAzureApplicationGatewayRewriteList(namespace string, count int) *appgwrewrite.AzureApplicationGatewayRewriteList {
	var items []appgwrewrite.AzureApplicationGatewayRewrite
	for i := range count {
		items = append(items, MakeAzureApplicationGatewayRewrite(namespace, "rewrite-ruleset-"+strconv.Itoa(i)))
	}

	return &appgwrewrite.AzureApplicationGatewayRewriteList{
		Items: items,
	}
}

func MakeAzureApplicationGatewayRewrite(namespace string, name string, rules ...appgwrewrite.RewriteRule) appgwrewrite.AzureApplicationGatewayRewrite {
	return appgwrewrite.AzureApplicationGatewayRewrite{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "AzureApplicationGatewayRewrite",
			APIVersion: appgwrewrite.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appgwrewrite.AzureApplicationGatewayRewriteSpec{
			RewriteRules: rules,
		},
	}
}

type HTTPRouteOptions struct {
	Paths []HTTPRoutePathOption
}

type HTTPRoutePathOption struct {
	Path     string
	Filters  []gatewayapi_v1.HTTPRouteFilter
	Backends []HTTPRoutePathBackendOption
}

type HTTPRoutePathBackendOption struct {
	BackendName      string
	BackendNamespace string
	BackendPort      int
}

func MakeHTTPRoute(namespace string, name string, options HTTPRouteOptions) gatewayapi_v1.HTTPRoute {
	route := gatewayapi_v1.HTTPRoute{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindHTTPRoute,
			APIVersion: gatewayapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayapi_v1.HTTPRouteSpec{
			Rules: []gatewayapi_v1.HTTPRouteRule{},
		},
	}

	if options.Paths == nil {
		route.Spec.Rules = []gatewayapi_v1.HTTPRouteRule{
			{
				Matches: []gatewayapi_v1.HTTPRouteMatch{
					{
						Path: &gatewayapi_v1.HTTPPathMatch{
							Type:  ptr.To(gatewayapi_v1.PathMatchPathPrefix),
							Value: ptr.To("/"),
						},
					},
				},
				BackendRefs: []gatewayapi_v1.HTTPBackendRef{
					{
						BackendRef: gatewayapi_v1.BackendRef{
							BackendObjectReference: gatewayapi_v1.BackendObjectReference{
								Name: "service-1",
								Port: ptr.To(gatewayapi_v1.PortNumber(80)),
							},
						},
					},
				},
			},
		}
	} else {
		for _, p := range options.Paths {
			var backends []gatewayapi_v1.HTTPBackendRef

			for _, backendOption := range p.Backends {
				backendPort := 80
				if backendOption.BackendPort != 0 {
					backendPort = backendOption.BackendPort
				}

				backendRef := gatewayapi_v1.HTTPBackendRef{
					BackendRef: gatewayapi_v1.BackendRef{
						BackendObjectReference: gatewayapi_v1.BackendObjectReference{
							Name: gatewayapi_v1.ObjectName(backendOption.BackendName),
							Port: ptr.To(gatewayapi_v1.PortNumber(backendPort)),
							Kind: ptr.To(gatewayapi_v1.Kind("Service")),
						},
					},
				}
				if backendOption.BackendNamespace != "" {
					backendRef.BackendRef.Namespace = ptr.To(gatewayapi_v1.Namespace(backendOption.BackendNamespace))
				}

				backends = append(backends, backendRef)
			}

			rule := gatewayapi_v1.HTTPRouteRule{
				Matches: []gatewayapi_v1.HTTPRouteMatch{
					{
						Path: &gatewayapi_v1.HTTPPathMatch{
							Type:  ptr.To(gatewayapi_v1.PathMatchPathPrefix),
							Value: ptr.To(p.Path),
						},
					},
				},
				BackendRefs: backends,
			}

			rule.Filters = append(rule.Filters, p.Filters...)
			route.Spec.Rules = append(route.Spec.Rules, rule)
		}
	}

	return route
}

func MakeReferenceGrant(namespace string, name string) gatewayapi_v1beta1.ReferenceGrant {
	return gatewayapi_v1beta1.ReferenceGrant{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindReferenceGrant,
			APIVersion: gatewayapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func MakeGateway(namespace string, name string) gatewayapi_v1.Gateway {
	return gatewayapi_v1.Gateway{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindGateway,
			APIVersion: gatewayapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayapi_v1.GatewaySpec{
			GatewayClassName: "azure/gateway-class",
			Listeners: []gatewayapi_v1.Listener{
				{
					Name:     "http",
					Protocol: gatewayapi_v1.HTTPProtocolType,
					Port:     gatewayapi_v1.PortNumber(80),
				},
			},
		},
	}
}

type RoutePolicyOptions struct {
	Timeout       *time.Duration
	ManagedCookie bool
}

func MakeRoutePolicy(namespace string, name string, routeNamespace string, routeName string, options RoutePolicyOptions) albcontrollerapi_v1.RoutePolicy {
	policy := albcontrollerapi_v1.RoutePolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindRoutePolicy,
			APIVersion: albcontrollerapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: albcontrollerapi_v1.RoutePolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     gatewayapi_v1.GroupName,
					Kind:      k8snames.KindHTTPRoute,
					Name:      gatewayapi_v1.ObjectName(routeName),
					Namespace: ptr.To(gatewayapi_v1.Namespace(routeNamespace)),
				},
			},
			Default: &albcontrollerapi_v1.RoutePolicyConfig{},
		},
	}

	if options.Timeout != nil {
		policy.Spec.Default.RouteTimeouts = &albcontrollerapi_v1.RouteTimeouts{
			RouteTimeout: meta_v1.Duration{Duration: *options.Timeout},
		}
	}

	if options.ManagedCookie {
		policy.Spec.Default.SessionAffinity = &albcontrollerapi_v1.SessionAffinity{
			AffinityType: albcontrollerapi_v1.AffinityTypeManagedCookie,
		}
	}

	return policy
}

func MakeBackendTLSPolicy(namespace string, name string, serviceNamespace string, serviceName string, ports ...int) albcontrollerapi_v1.BackendTLSPolicy {
	policy := albcontrollerapi_v1.BackendTLSPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindBackendTLSPolicy,
			APIVersion: albcontrollerapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: albcontrollerapi_v1.BackendTLSPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     "",
					Kind:      k8snames.KindService,
					Name:      gatewayapi_v1.ObjectName(serviceName),
					Namespace: ptr.To(gatewayapi_v1.Namespace(serviceNamespace)),
				},
			},
			Default: &albcontrollerapi_v1.BackendTLSPolicyConfig{},
		},
	}
	for _, port := range ports {
		policy.Spec.Default.Ports = append(policy.Spec.Default.Ports, albcontrollerapi_v1.BackendTLSPolicyPort{
			Port: port,
		})
	}

	return policy
}

func MakeFrontendTLSPolicy(namespace string, name string) albcontrollerapi_v1.FrontendTLSPolicy {
	return albcontrollerapi_v1.FrontendTLSPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindFrontendTLSPolicy,
			APIVersion: albcontrollerapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: albcontrollerapi_v1.FrontendTLSPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     "",
					Kind:      k8snames.KindGateway,
					Name:      gatewayapi_v1.ObjectName("gateway-1"),
					Namespace: ptr.To(gatewayapi_v1.Namespace("default")),
				},
			},
			Default: &albcontrollerapi_v1.FrontendTLSPolicyConfig{},
		},
	}
}

type HealthCheckPolicyOptions struct {
	Interval           *time.Duration
	Timeout            *time.Duration
	Hostname           *string
	Path               *string
	StatusCodes        []*albcontrollerapi_v1.StatusCodes
	Port               *int
	UnhealthyThreshold *int
}

func MakeHealthCheckPolicy(namespace string, name string, serviceNamespace string, serviceName string, options HealthCheckPolicyOptions) albcontrollerapi_v1.HealthCheckPolicy {
	policy := albcontrollerapi_v1.HealthCheckPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindHealthCheckPolicy,
			APIVersion: albcontrollerapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: albcontrollerapi_v1.HealthCheckPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     "",
					Kind:      k8snames.KindService,
					Name:      gatewayapi_v1.ObjectName(serviceName),
					Namespace: ptr.To(gatewayapi_v1.Namespace(serviceNamespace)),
				},
			},
			Default: &albcontrollerapi_v1.HealthCheckPolicyConfig{},
		},
	}

	if options.Interval != nil {
		policy.Spec.Default.Interval = meta_v1.Duration{Duration: *options.Interval}
	}

	if options.Timeout != nil {
		policy.Spec.Default.Timeout = meta_v1.Duration{Duration: *options.Timeout}
	}

	if options.Path != nil {
		if policy.Spec.Default.HTTP == nil {
			policy.Spec.Default.HTTP = &albcontrollerapi_v1.HTTPSpecifiers{}
		}

		policy.Spec.Default.HTTP.Path = *options.Path
	}

	if options.Hostname != nil {
		if policy.Spec.Default.HTTP == nil {
			policy.Spec.Default.HTTP = &albcontrollerapi_v1.HTTPSpecifiers{}
		}

		policy.Spec.Default.HTTP.Host = *options.Hostname
	}

	if options.StatusCodes != nil {
		if policy.Spec.Default.HTTP == nil {
			policy.Spec.Default.HTTP = &albcontrollerapi_v1.HTTPSpecifiers{}
		}

		if policy.Spec.Default.HTTP.Match == nil {
			policy.Spec.Default.HTTP.Match = &albcontrollerapi_v1.HTTPMatch{}
		}

		policy.Spec.Default.HTTP.Match.StatusCodes = options.StatusCodes
	}

	if options.Port != nil {
		policy.Spec.Default.Port = int32(*options.Port) // #nosec G115
	}

	if options.UnhealthyThreshold != nil {
		policy.Spec.Default.UnhealthyThreshold = int32(*options.UnhealthyThreshold) // #nosec G115
	}

	return policy
}

func MakeWAFPolicy(namespace string, name string, routeNamespace string, routeName string, id string) albcontrollerapi_v1.WebApplicationFirewallPolicy {
	return albcontrollerapi_v1.WebApplicationFirewallPolicy{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       k8snames.KindWebApplicationFirewallPolicy,
			APIVersion: albcontrollerapi_v1.GroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: albcontrollerapi_v1.WebApplicationFirewallPolicySpec{
			TargetRef: albcontrollerapi_v1.CustomTargetRef{
				NamespacedPolicyTargetReference: v1alpha2.NamespacedPolicyTargetReference{
					Group:     gatewayapi_v1.GroupName,
					Kind:      k8snames.KindHTTPRoute,
					Name:      gatewayapi_v1.ObjectName(routeName),
					Namespace: ptr.To(gatewayapi_v1.Namespace(routeNamespace)),
				},
			},
			WebApplicationFirewallPolicy: &albcontrollerapi_v1.WebApplicationFirewallConfig{
				ID: id,
			},
		},
	}
}

type RewriteRuleSetOptions struct {
	ReqHeadersSet     map[string]string
	ReqHeadersDelete  []string
	RespHeadersSet    map[string]string
	RespHeadersDelete []string
	PathRewrite       string
}

func MakeRewriteRuleSet(namespace string, name string, opts RewriteRuleSetOptions, rules ...appgwrewrite.RewriteRule) appgwrewrite.AzureApplicationGatewayRewrite {
	rewriteObj := appgwrewrite.AzureApplicationGatewayRewrite{
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "AzureApplicationGatewayRewrite",
			APIVersion: appgwrewrite.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appgwrewrite.AzureApplicationGatewayRewriteSpec{
			RewriteRules: rules,
		},
	}

	for header, value := range opts.ReqHeadersSet {
		rewriteObj.Spec.RewriteRules = append(rewriteObj.Spec.RewriteRules, appgwrewrite.RewriteRule{
			Name: "set-req-header-" + header,
			Actions: appgwrewrite.Actions{
				RequestHeaderConfigurations: []appgwrewrite.HeaderConfiguration{
					{
						ActionType:  "set",
						HeaderName:  header,
						HeaderValue: value,
					},
				},
			},
		})
	}

	for _, header := range opts.ReqHeadersDelete {
		rewriteObj.Spec.RewriteRules = append(rewriteObj.Spec.RewriteRules, appgwrewrite.RewriteRule{
			Name: "delete-req-header-" + header,
			Actions: appgwrewrite.Actions{
				RequestHeaderConfigurations: []appgwrewrite.HeaderConfiguration{
					{
						ActionType: "delete",
						HeaderName: header,
					},
				},
			},
		})
	}

	for header, value := range opts.RespHeadersSet {
		rewriteObj.Spec.RewriteRules = append(rewriteObj.Spec.RewriteRules, appgwrewrite.RewriteRule{
			Name: "set-resp-header-" + header,
			Actions: appgwrewrite.Actions{
				ResponseHeaderConfigurations: []appgwrewrite.HeaderConfiguration{
					{
						ActionType:  "set",
						HeaderName:  header,
						HeaderValue: value,
					},
				},
			},
		})
	}

	for _, header := range opts.RespHeadersDelete {
		rewriteObj.Spec.RewriteRules = append(rewriteObj.Spec.RewriteRules, appgwrewrite.RewriteRule{
			Name: "delete-resp-header-" + header,
			Actions: appgwrewrite.Actions{
				ResponseHeaderConfigurations: []appgwrewrite.HeaderConfiguration{
					{
						ActionType: "delete",
						HeaderName: header,
					},
				},
			},
		})
	}

	if opts.PathRewrite != "" {
		rewriteObj.Spec.RewriteRules = append(rewriteObj.Spec.RewriteRules, appgwrewrite.RewriteRule{
			Name: "path-rewrite",
			Actions: appgwrewrite.Actions{
				UrlConfiguration: &appgwrewrite.UrlConfiguration{
					ModifiedPath: opts.PathRewrite,
				},
			},
		})
	}

	return rewriteObj
}

type BackendDescriptor struct {
	Path        string
	ServiceName string
	Port        network_v1.ServiceBackendPort
}

func NewBackendDescriptor(path string, serviceName string, portName string, portNumber int) BackendDescriptor {
	return BackendDescriptor{
		Path:        path,
		ServiceName: serviceName,
		Port: network_v1.ServiceBackendPort{
			Name:   portName,
			Number: int32(portNumber),
		},
	}
}

func MakeAGICPod(labels ...string) core_v1.Pod {
	if len(labels)%2 != 0 {
		panic("labels must be key/value pairs")
	}

	pod := core_v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "default",
			Labels:    make(map[string]string),
		},
		Spec: core_v1.PodSpec{
			Containers: []core_v1.Container{
				{
					Name:  "ingress-azure",
					Image: "mcr.microsoft.com/azure-application-gateway/kubernetes-ingress:latest",
				},
			},
		},
		Status: core_v1.PodStatus{
			Phase: core_v1.PodRunning,
		},
	}

	for i := 0; i < len(labels); i += 2 {
		pod.ObjectMeta.Labels[labels[i]] = labels[i+1]
	}

	return pod
}
