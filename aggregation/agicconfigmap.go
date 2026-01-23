package aggregation

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// nolint: revive
const (
	SKUWAFv2      = "WAF_v2"
	SKUStandardv2 = "Standard_v2"
	SKUBasic      = "Basic"
)

// AGICConfigMap is our representation of the ConfigMap used by AGIC.
// I've copied the comments from the AGIC source code.
type AGICConfigMap struct {
	// CloudProviderConfigLocation is an environment variable for Azure cloud provider location
	CloudProviderConfigLocation string `json:"AZURE_CLOUD_PROVIDER_LOCATION,omitempty"`

	// ClientID is an environment variable which stores the client id provided through user assigned identity
	ClientID string `json:"AZURE_CLIENT_ID,omitempty"`

	// SubscriptionID is the name of the APPGW_SUBSCRIPTION_ID
	SubscriptionID string `json:"APPGW_SUBSCRIPTION_ID,omitempty"`

	// ResourceGroupName is the name of the APPGW_RESOURCE_GROUP
	ResourceGroupName string `json:"APPGW_RESOURCE_GROUP,omitempty"`

	// AppGwName is the name of the APPGW_NAME
	AppGwName string `json:"APPGW_NAME,omitempty"`

	// AppGwSubnetName is the name of the APPGW_SUBNET_NAME
	AppGwSubnetName string `json:"APPGW_SUBNET_NAME,omitempty"`

	// AppGwSubnetPrefix is the name of the APPGW_SUBNET_PREFIX
	AppGwSubnetPrefix string `json:"APPGW_SUBNET_PREFIX,omitempty"`

	// AppGwResourceID is the name of the APPGW_RESOURCE_ID
	AppGwResourceID string `json:"APPGW_RESOURCE_ID,omitempty"`

	// AppGwSubnetID is the name of the APPGW_SUBNET_ID
	AppGwSubnetID string `json:"APPGW_SUBNET_ID,omitempty"`

	// AppGwSKU is the sku of the AGW
	AppGwSKU string `json:"APPGW_SKU_NAME,omitempty"`

	// AuthLocation is the name of the AZURE_AUTH_LOCATION
	AuthLocation string `json:"AZURE_AUTH_LOCATION,omitempty"`

	// WatchNamespace is the name of the KUBERNETES_WATCHNAMESPACE
	WatchNamespace string `json:"KUBERNETES_WATCHNAMESPACE,omitempty"`

	// UsePrivateIP is the name of the USE_PRIVATE_IP
	UsePrivateIP bool `json:"USE_PRIVATE_IP,omitempty"`

	// VerbosityLevel sets the level of klog verbosity should the CLI argument be blank
	VerbosityLevel string `json:"APPGW_VERBOSITY_LEVEL,omitempty"`

	// EnableBrownfieldDeployment is a feature flag enabling observation of {Managed,Prohibited}Target CRDs
	EnableBrownfieldDeployment bool `json:"APPGW_ENABLE_SHARED_APPGW,omitempty"`

	// EnableIstioIntegration is a feature flag enabling observation of Istio specific CRDs
	EnableIstioIntegration bool `json:"APPGW_ENABLE_ISTIO_INTEGRATION,omitempty"`

	// EnableSaveConfigToFile is a feature flag, which enables saving the App Gwy config to disk.
	EnableSaveConfigToFile bool `json:"APPGW_ENABLE_SAVE_CONFIG_TO_FILE,omitempty"`

	// EnablePanicOnPutError is a feature flag.
	EnablePanicOnPutError bool `json:"APPGW_ENABLE_PANIC_ON_PUT_ERROR,omitempty"`

	// EnableDeployAppGateway is a feature flag.
	EnableDeployAppGateway bool `json:"APPGW_ENABLE_DEPLOY,omitempty"`

	// HTTPServicePort is an environment variable name.
	HTTPServicePort string `json:"HTTP_SERVICE_PORT,omitempty"`

	// AGICPodName is an environment variable name.
	AGICPodName string `json:"AGIC_POD_NAME,omitempty"`

	// AGICPodNamespace is an environment variable name.
	AGICPodNamespace string `json:"AGIC_POD_NAMESPACE,omitempty"`

	// UseManagedIdentityForPod is an environment variable name.
	UseManagedIdentityForPod bool `json:"USE_MANAGED_IDENTITY_FOR_POD,omitempty"`

	// AttachWAFPolicyToListener is an environment variable name.
	AttachWAFPolicyToListener bool `json:"ATTACH_WAF_POLICY_TO_LISTENER,omitempty"`

	// HostedOnUnderlay  is an environment variable name.
	HostedOnUnderlay bool `json:"HOSTED_ON_UNDERLAY,omitempty"`

	// ReconcilePeriodSeconds is an environment variable to control reconcile period for the AGIC.
	ReconcilePeriodSeconds string `json:"RECONCILE_PERIOD_SECONDS,omitempty"`

	// IngressClass is an environment variable
	IngressClass string `json:"INGRESS_CLASS,omitempty"`

	// IngressClassResourceEnabled is an environment variable to enable V1 Ingress class.
	IngressClassResourceEnabled bool `json:"INGRESS_CLASS_RESOURCE_ENABLED,omitempty"`

	// IngressClassResourceName is an environment variable which specifies the name of the ingress class object to watch.
	IngressClassResourceName string `json:"INGRESS_CLASS_RESOURCE_NAME,omitempty"`

	// IngressClassResourceDefault is an environment variable to enable AGIC as default ingress.
	IngressClassResourceDefault bool `json:"INGRESS_CLASS_RESOURCE_DEFAULT,omitempty"`

	// IngressClassControllerName is an environment variable to specify controller class.
	IngressClassControllerName string `json:"INGRESS_CLASS_RESOURCE_CONTROLLER,omitempty"`

	// MultiClusterMode is an environment variable to control whether AGIC monitors Ingresses or MutliClusterIngresses
	MultiClusterMode bool `json:"MULTI_CLUSTER_MODE,omitempty"`

	// AddonMode is an environment variable to inform if the controller is running as an addon.
	AddonMode bool `json:"ADDON_MODE,omitempty"`
}

const DefaultIngressClassName = "azure/application-gateway"

// IngressClassName returns the IngressClass name from this AGIC config.
// Refer to pkg/environment/environment.go in the AGIC source code for the logic.
func (cm AGICConfigMap) IngressClassName() string {
	if cm.IngressClass != "" {
		return cm.IngressClass
	}

	if cm.IngressClassControllerName != "" {
		return cm.IngressClassControllerName
	}

	return ""
}

// IngressControllerName returns the effective IngressClassController name for this AGIC installation.
// Refer to pkg/environment/environment.go in the AGIC source code for the logic.
func (cm AGICConfigMap) IngressControllerName() string {
	if cm.IngressClassResourceName != "" {
		return cm.IngressClassResourceName
	}

	return "azure-application-gateway"
}

// LoadAGICConfigMap reads the AGIC ConfigMap from the cluster and returns it as an AGICConfigMap struct.
func (a Aggregator) loadAGICConfigMap(ctx context.Context, agicLabel string, client kubernetes.Interface) (AGICConfigMap, error) {
	maps, err := client.CoreV1().ConfigMaps("").List(ctx, meta_v1.ListOptions{LabelSelector: agicLabel})
	if err != nil {
		return AGICConfigMap{}, fmt.Errorf("failed to list ConfigMaps with label selector %q: %w", agicLabel, err)
	}

	if len(maps.Items) == 0 {
		return AGICConfigMap{}, fmt.Errorf("no ConfigMap found with label selector %q", agicLabel)
	}

	if len(maps.Items) > 1 {
		return AGICConfigMap{}, fmt.Errorf("multiple ConfigMaps found with label selector %q", agicLabel)
	}

	config := AGICConfigMap{}

	// build the map of tags to struct fields
	reflectType := reflect.TypeOf(config)
	fieldMap := make(map[string]reflect.StructField)

	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		tag := field.Tag.Get("json")

		if tag != "" {
			tag = strings.Split(tag, ",")[0]
			fieldMap[tag] = field
		}
	}

	for k, v := range maps.Items[0].Data {
		field, ok := fieldMap[k]
		if !ok {
			a.log.Warn("no matching struct field for ConfigMap", "key", k)
			continue
		}
		// we have a match, set the field
		reflectValue := reflect.ValueOf(&config).Elem().FieldByName(field.Name)
		if !reflectValue.IsValid() || !reflectValue.CanSet() {
			a.log.Warn("cannot set field", "field", field.Name)
			continue
		}

		switch reflectValue.Kind() {
		case reflect.String:
			reflectValue.SetString(v)
		case reflect.Bool:
			boolVal, _ := strconv.ParseBool(v)
			reflectValue.SetBool(boolVal)
		default:
			a.log.Warn("unsupported type for field", "type", reflectValue.Kind(), "field", field.Name)
		}
	}

	return config, nil
}
