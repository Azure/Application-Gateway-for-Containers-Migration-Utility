package aggregation

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-test/deep"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"
)

func TestLoadAGICConfigMap(t *testing.T) {
	ctx := t.Context()
	agg := NewAggregator(Options{})

	t.Run("client returns an error", func(t *testing.T) {
		client := fake.NewClientset()
		client.PrependReactor("list", "configmaps", func(k8s_testing.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("test failure")
		})

		_, err := agg.loadAGICConfigMap(ctx, agicHelmLabel, client)
		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		if !strings.Contains(err.Error(), "test failure") {
			t.Fatalf("expected \"test failure\" error, got %v", err)
		}
	})

	t.Run("no configmap", func(t *testing.T) {
		fakeClient := fake.NewClientset()
		_, err := agg.loadAGICConfigMap(ctx, agicHelmLabel, fakeClient)

		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		if !strings.Contains(err.Error(), "no ConfigMap found") {
			t.Fatalf("expected a no ConfigMap found error, got %v", err)
		}
	})

	t.Run("too many configmaps", func(t *testing.T) {
		fakeClient := fake.NewClientset(
			makeAGICConfigMap("cfg1", map[string]string{}),
			makeAGICConfigMap("cfg2", map[string]string{}),
		)
		_, err := agg.loadAGICConfigMap(ctx, agicHelmLabel, fakeClient)

		if err == nil {
			t.Fatalf("expected an error, got none")
		}

		if !strings.Contains(err.Error(), "multiple ConfigMaps found") {
			t.Fatalf("expected a multiple ConfigMaps found error, got %v", err)
		}
	})

	t.Run("fully specified config map", func(t *testing.T) {
		fakeClient := fake.NewClientset(makeAGICConfigMap("fully-specified", map[string]string{
			"AZURE_CLOUD_PROVIDER_LOCATION":     "dummy-location",
			"AZURE_CLIENT_ID":                   "dummy-client-id",
			"APPGW_SUBSCRIPTION_ID":             "dummy-subscription-id",
			"APPGW_RESOURCE_GROUP":              "dummy-resource-group",
			"APPGW_NAME":                        "dummy-appgw-name",
			"APPGW_SUBNET_NAME":                 "dummy-subnet-name",
			"APPGW_SUBNET_PREFIX":               "dummy-subnet-prefix",
			"APPGW_RESOURCE_ID":                 "dummy-resource-id",
			"APPGW_SUBNET_ID":                   "dummy-subnet-id",
			"APPGW_SKU_NAME":                    "dummy-sku",
			"AZURE_AUTH_LOCATION":               "dummy-auth-location",
			"KUBERNETES_WATCHNAMESPACE":         "dummy-namespace",
			"USE_PRIVATE_IP":                    "true",
			"APPGW_VERBOSITY_LEVEL":             "dummy-verbosity",
			"APPGW_ENABLE_SHARED_APPGW":         "true",
			"APPGW_ENABLE_ISTIO_INTEGRATION":    "true",
			"APPGW_ENABLE_SAVE_CONFIG_TO_FILE":  "true",
			"APPGW_ENABLE_PANIC_ON_PUT_ERROR":   "true",
			"APPGW_ENABLE_DEPLOY":               "true",
			"HTTP_SERVICE_PORT":                 "dummy-port",
			"AGIC_POD_NAME":                     "dummy-pod",
			"AGIC_POD_NAMESPACE":                "dummy-pod-namespace",
			"USE_MANAGED_IDENTITY_FOR_POD":      "true",
			"ATTACH_WAF_POLICY_TO_LISTENER":     "true",
			"HOSTED_ON_UNDERLAY":                "true",
			"RECONCILE_PERIOD_SECONDS":          "60",
			"INGRESS_CLASS":                     "dummy-ingress-class",
			"INGRESS_CLASS_RESOURCE_ENABLED":    "true",
			"INGRESS_CLASS_RESOURCE_NAME":       "dummy-ingress-class-resource",
			"INGRESS_CLASS_RESOURCE_DEFAULT":    "true",
			"INGRESS_CLASS_RESOURCE_CONTROLLER": "dummy-controller",
			"MULTI_CLUSTER_MODE":                "true",
			"ADDON_MODE":                        "true",
		}))

		cfg, err := agg.loadAGICConfigMap(ctx, agicHelmLabel, fakeClient)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expect := AGICConfigMap{
			CloudProviderConfigLocation: "dummy-location",
			ClientID:                    "dummy-client-id",
			SubscriptionID:              "dummy-subscription-id",
			ResourceGroupName:           "dummy-resource-group",
			AppGwName:                   "dummy-appgw-name",
			AppGwSubnetName:             "dummy-subnet-name",
			AppGwSubnetPrefix:           "dummy-subnet-prefix",
			AppGwResourceID:             "dummy-resource-id",
			AppGwSubnetID:               "dummy-subnet-id",
			AppGwSKU:                    "dummy-sku",
			AuthLocation:                "dummy-auth-location",
			WatchNamespace:              "dummy-namespace",
			UsePrivateIP:                true,
			VerbosityLevel:              "dummy-verbosity",
			EnableBrownfieldDeployment:  true,
			EnableIstioIntegration:      true,
			EnableSaveConfigToFile:      true,
			EnablePanicOnPutError:       true,
			EnableDeployAppGateway:      true,
			HTTPServicePort:             "dummy-port",
			AGICPodName:                 "dummy-pod",
			AGICPodNamespace:            "dummy-pod-namespace",
			UseManagedIdentityForPod:    true,
			AttachWAFPolicyToListener:   true,
			HostedOnUnderlay:            true,
			ReconcilePeriodSeconds:      "60",
			IngressClass:                "dummy-ingress-class",
			IngressClassResourceEnabled: true,
			IngressClassResourceName:    "dummy-ingress-class-resource",
			IngressClassResourceDefault: true,
			IngressClassControllerName:  "dummy-controller",
			MultiClusterMode:            true,
			AddonMode:                   true,
		}

		diff := deep.Equal(cfg, expect)
		if diff != nil {
			t.Fatalf("Config mismatch: %v", diff)
		}
	})

	t.Run("ingress class name", func(t *testing.T) {
		tests := []struct {
			name     string
			config   AGICConfigMap
			expected string
		}{
			{
				name:     "no ingress class or controller name",
				expected: "",
			},
			{
				name: "only ingress class",
				config: AGICConfigMap{
					IngressClass: "my-ingress-class",
				},
				expected: "my-ingress-class",
			},
			{
				name: "only controller name",
				config: AGICConfigMap{
					IngressClassControllerName: "my-controller-name",
				},
				expected: "my-controller-name",
			},
			{
				name: "both ingress class and controller name",
				config: AGICConfigMap{
					IngressClass:               "my-ingress-class",
					IngressClassControllerName: "my-controller-name",
				},
				// ingress class takes precedence over controller name
				expected: "my-ingress-class",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if got := tc.config.IngressClassName(); got != tc.expected {
					t.Fatalf("expected effective ingress class name %q, got %q", tc.expected, got)
				}
			})
		}
	})

	t.Run("effective ingress controller name", func(t *testing.T) {
		tests := []struct {
			name     string
			config   AGICConfigMap
			expected string
		}{
			{
				name:     "no ingress class resource name",
				expected: "azure-application-gateway",
			},
			{
				name: "with ingress class resource name",
				config: AGICConfigMap{
					IngressClassResourceName: "my-ingress-class-resource",
				},
				expected: "my-ingress-class-resource",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if got := tc.config.IngressControllerName(); got != tc.expected {
					t.Fatalf("expected effective ingress controller name %q, got %q", tc.expected, got)
				}
			})
		}
	})
}

func makeAGICConfigMapList(name string, data map[string]string) *core_v1.ConfigMapList {
	return &core_v1.ConfigMapList{
		Items: []core_v1.ConfigMap{
			*makeAGICConfigMap(name, data),
		},
	}
}

func makeAGICConfigMap(name string, data map[string]string) *core_v1.ConfigMap {
	return &core_v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": "ingress-azure",
			},
		},
		Data: data,
	}
}
