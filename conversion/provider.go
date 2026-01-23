package conversion

import "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"

// Provider defines the interface for source systems that can be migrated from (e.g. AGIC).
type Provider interface {
	// GetAnnotationHandlers returns a map of annotation keys to their corresponding handler functions.
	// The keys are the full annotation names (e.g., "appgw.ingress.kubernetes.io/ssl-redirect").
	GetAnnotationHandlers() map[string]AnnotationHandler
}

// AnnotationHandler processes a single ingress annotation and updates the Gateway API resource graph.
// It receives the resource graph, gateway context, HTTP route context, ingress context, and annotation context.
// Returns an error if the annotation value is invalid or conversion fails.
type AnnotationHandler func(resources.AGCResourceGraph, *GatewayContext, *HTTPRouteContext, *resources.IngressContext, *resources.IngressAnnotationContext) error
