package k8snames

const (
	ALBAPIVersion                     = "alb.networking.azure.io/v1"
	KindHTTPRoute                     = "HTTPRoute"
	KindSecret                        = "Secret"
	KindGateway                       = "Gateway"
	KindService                       = "Service"
	KindIngress                       = "Ingress"
	KindBackendTLSPolicy              = "BackendTLSPolicy"
	KindHealthCheckPolicy             = "HealthCheckPolicy"
	KindRoutePolicy                   = "RoutePolicy"
	KindFrontendTLSPolicy             = "FrontendTLSPolicy"
	KindApplicationLoadBalancer       = "ApplicationLoadBalancer"
	KindWebApplicationFirewallPolicy  = "WebApplicationFirewallPolicy"
	KindReferenceGrant                = "ReferenceGrant"
	ALBExternalGatewayClassName       = "azure-alb-external"
	ALBK8sResourceNamespaceAnnotation = "alb.networking.azure.io/alb-namespace"
	ALBK8sResourceNameAnnotation      = "alb.networking.azure.io/alb-name"
	ALBArmResourceIDAnnotation        = "alb.networking.azure.io/alb-id"
	GatewayNetworkingGroup            = "gateway.networking.k8s.io"
)
