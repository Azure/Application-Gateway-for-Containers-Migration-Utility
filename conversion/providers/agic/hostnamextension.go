package agic

import (
	"strings"

	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/conversion"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func (p Provider) handleHostNameExtension(
	graph resources.AGCResourceGraph,
	gatewayCtx *conversion.GatewayContext,
	routeCtx *conversion.HTTPRouteContext,
	ingressCtx *resources.IngressContext,
	annotationCtx *resources.IngressAnnotationContext,
) error {
	hostnameSecrets := make(map[string]string)

	for _, tlsEntry := range ingressCtx.Ingress.Spec.TLS {
		for _, hostname := range tlsEntry.Hosts {
			hostnameSecrets[hostname] = tlsEntry.SecretName
		}
	}

	var hostnames []gatewayapi_v1.Hostname

	for split := range strings.SplitSeq(annotationCtx.Value, ",") {
		hostname := strings.TrimSpace(split)
		if len(hostname) == 0 {
			continue
		}

		hostnames = append(hostnames, gatewayapi_v1.Hostname(hostname))
		secretName, ok := hostnameSecrets[hostname]

		if ok {
			routeCtx.EnsureHTTPSListener(&graph, *gatewayCtx, hostname, secretName, ingressCtx.Ingress.Namespace)
		}
	}

	routeCtx.Spec.Hostnames = append(routeCtx.Spec.Hostnames, hostnames...)

	annotationCtx.SetStatus(resources.MigrationStatusCompleted)

	return nil
}
