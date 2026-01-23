package resources

import (
	"testing"

	network_v1 "k8s.io/api/networking/v1"

	appgwrewrite "github.com/Azure/application-gateway-kubernetes-ingress/pkg/apis/azureapplicationgatewayrewrite/v1beta1"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"
)

func TestNewAGICResources(t *testing.T) {
	ingress1 := testutil.MakeIngress("default", "ingress-1", testutil.IngressOptions{}, "anno-1", "value-1")
	ingress2 := testutil.MakeIngress("default", "ingress-2", testutil.IngressOptions{}, "anno-2", "value-2")
	agicResources := NewAGICResources([]network_v1.Ingress{ingress1, ingress2}, []appgwrewrite.AzureApplicationGatewayRewrite{})

	if len(agicResources.IngressContexts) != 2 {
		t.Fatalf("Expected 2 ingress contexts, got: %d", len(agicResources.IngressContexts))
	}

	if got1, exists := agicResources.IngressContexts[k8snames.NamespacedName(&ingress1)]; !exists {
		t.Fatalf("Expected ingress context for ingress-1 to exist")
	} else if got1.Annotations["anno-1"].Value != "value-1" {
		t.Fatalf("Expected annotation value 'value-1', got: %s", got1.Annotations["anno-1"].Value)
	}

	if got2, exists := agicResources.IngressContexts[k8snames.NamespacedName(&ingress2)]; !exists {
		t.Fatalf("Expected ingress context for ingress-2 to exist")
	} else if got2.Annotations["anno-2"].Value != "value-2" {
		t.Fatalf("Expected annotation value 'value-2', got: %s", got2.Annotations["anno-2"].Value)
	}
}

func TestAGICResources_Ingresses(t *testing.T) {
	resources := NewAGICResources([]network_v1.Ingress{
		testutil.MakeIngress("default", "ingress-1", testutil.IngressOptions{}),
		testutil.MakeIngress("default", "ingress-2", testutil.IngressOptions{}),
	}, []appgwrewrite.AzureApplicationGatewayRewrite{})
	got := resources.Ingresses()

	if len(got) != 2 {
		t.Fatalf("Expected 2 ingresses, got: %d", len(got))
	}

	if got[0].Name != "ingress-1" && got[1].Name != "ingress-1" {
		t.Fatalf("Expected ingress-1 to be in the list, got: %+v", got)
	}

	if got[0].Name != "ingress-2" && got[1].Name != "ingress-2" {
		t.Fatalf("Expected ingress-2 to be in the list, got: %+v", got)
	}
}
