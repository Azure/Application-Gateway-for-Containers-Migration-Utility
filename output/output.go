// Package output is responsible for publishing the migration results.
package output

import (
	"fmt"

	"os"
	"reflect"

	"gopkg.in/yaml.v2" // we use v2 to get the same indentation behaviour as sigs.k8s.io/yaml
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"

	albcontrollerapi_v1 "github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/crds/v1"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/reporting"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

const MigrationReportHeader = "# Migration Report\n"

// Options configures the output options when writing the results.
// It doesn't affect the how the Migration Report is handled.
type Options struct {
	DryRun           bool
	OutputDir        string
	OutputFile       string
	SingleOutputFile string
	writer           writer // lazy hack to allow for dependency injection in unit tests
}

func (o Options) getWriter() writer {
	if o.writer != nil {
		return o.writer
	}

	if o.OutputDir != "" {
		return dirWriter{basePath: o.OutputDir}
	}

	if o.OutputFile != "" {
		return &fileWriter{filePath: o.OutputFile}
	}

	return stdoutWriter{}
}

// Write will produce the output based on the given inputs.
func Write(opts Options, report reporting.MigrationReport, result resources.AGCResourceGraph) error {
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0o750); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", opts.OutputDir, err)
		}
	}

	reportStr, err := reportToYAML(report)

	if err != nil {
		return fmt.Errorf("failed to convert report to YAML: %w", err)
	}

	writer := opts.getWriter()

	if !opts.DryRun {
		if result.ApplicationLoadBalancer != nil {
			bytes, err := marshalYAML(result.ApplicationLoadBalancer)
			if err != nil {
				return fmt.Errorf("failed to marshal ApplicationLoadBalancer to YAML: %w", err)
			}

			fName := yamlFilename("applicationloadbalancer", result.ApplicationLoadBalancer.Namespace, result.ApplicationLoadBalancer.Name)
			if err := writer.write(fName, bytes); err != nil {
				return fmt.Errorf("failed to write ApplicationLoadBalancer resource: %w", err)
			}
		}

		if err := writeList(writer, "gateway", map[types.NamespacedName]*gatewayapi_v1.Gateway{
			k8snames.NamespacedName(result.Gateway): result.Gateway,
		}); err != nil {
			return err
		}

		if err := writeList(writer, "httproute", result.HTTPRoutes); err != nil {
			return err
		}

		if err := writeList(writer, "referencegrant", result.ReferenceGrants); err != nil {
			return err
		}

		if err := writeList(writer, "routepolicy", result.RoutePolicies); err != nil {
			return err
		}

		if err := writeList(writer, "healthcheckpolicy", result.HealthCheckPolicies); err != nil {
			return err
		}

		if err := writeList(writer, "wafpolicy", result.WAFPolicies); err != nil {
			return err
		}

		if err := writeList(writer, "backendtlspolicy", result.BackendTLSPolicies); err != nil {
			return err
		}

		if err := writeList(writer, "frontendtlspolicy", result.FrontendTLSPolicies); err != nil {
			return err
		}
	}

	// for now, the migration report is always written to stdout
	// TODO: this might be confusing, we should revisit it
	if opts.OutputDir == "" {
		fmt.Println("---")
	}

	fmt.Print(reportStr)

	return nil
}

func writeList(writer writer, kind string, objects any) error {
	items, err := convertMap(objects)
	if err != nil {
		return fmt.Errorf("failed to convert %s objects map: %w", kind, err)
	}

	for key, val := range items {
		str, err := marshalYAML(val)

		if err != nil {
			return fmt.Errorf("failed to marshal %s %s resource to YAML: %w", kind, key, err)
		}

		if err := writer.write(yamlFilename(kind, key.Namespace, key.Name), str); err != nil {
			return fmt.Errorf("failed to write %s %s resource: %w", kind, key, err)
		}
	}

	return nil
}

func yamlFilename(kind, namespace, name string) string {
	return fmt.Sprintf("%s-%s-%s.yaml", kind, namespace, name)
}

func marshalYAML(obj any) ([]byte, error) {
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	// delete the status, since it will obviously be empty
	delete(objMap, "status")

	// delete the (nil) creationTimestamp
	if objMeta, ok := getSubMap(objMap, "metadata"); ok {
		delete(objMeta, "creationTimestamp")
	}

	// delete nil sectionNames, which will fail validation on ingestion
	if objSpec, ok := getSubMap(objMap, "spec"); ok {
		if parentRefs, ok := getSubMap(objSpec, "targetRef"); ok {
			if sectionNames, ok := parentRefs["sectionNames"]; ok && sectionNames == nil {
				delete(parentRefs, "sectionNames")
			}
		}
	}

	if err := fixHealthCheckStatusCodeOrder(objMap); err != nil {
		return nil, fmt.Errorf("failed to fix health status code order: %w", err)
	}

	return yaml.Marshal(objMap)
}

// the k8s yaml encoding messes up the order that start/end fields appear in the final yaml
func fixHealthCheckStatusCodeOrder(unstructured map[string]any) error {
	if kind, ok := unstructured["kind"]; !ok || kind != k8snames.KindHealthCheckPolicy {
		return nil
	}

	// we do nothing, the act of converting it from a map to the struct and finally using the standard yaml encoder
	// will do the job of ensuring statusCodes are ordered as [start, end] instead of alphabetically.
	modify := func(statusCodes []*albcontrollerapi_v1.StatusCodes) []*albcontrollerapi_v1.StatusCodes {
		return statusCodes
	}

	return modifyUnstructuredPath(unstructured, modify, "spec", "default", "http", "match", "statusCodes")
}

func getSubMap(unstructured map[string]any, name string) (map[string]any, bool) {
	subMap, ok := unstructured[name]
	if !ok {
		return nil, false
	}

	if got, ok := (subMap).(map[string]any); ok {
		return got, true
	}

	return nil, false
}

func modifyUnstructuredPath[T any](unstructured map[string]any, modify func(out T) T, paths ...string) error {
	if len(paths) == 0 {
		return nil
	}

	for _, path := range paths[:len(paths)-1] {
		var ok bool
		unstructured, ok = getSubMap(unstructured, path)

		if !ok {
			return nil
		}
	}

	finalPath := paths[len(paths)-1]
	got, ok := unstructured[finalPath]

	if !ok {
		return nil
	}

	bytes, err := yaml.Marshal(got)
	if err != nil {
		return err
	}

	var t T
	if err := yaml.Unmarshal(bytes, &t); err != nil {
		return err
	}

	unstructured[finalPath] = modify(t)

	return nil
}

// convertMap converts map of type map[types.NamespacedName]T to map[types.NamespacedName]any, so that we can treat
// the different collections of resources (gateways, routes, etc) the same.
// This is a bit of evil to make the core logic easier to follow.
func convertMap(input any) (map[types.NamespacedName]any, error) {
	val := reflect.ValueOf(input)
	typ := val.Type()

	if typ.Kind() != reflect.Map {
		return nil, fmt.Errorf("input must be a map, got %v", typ.Kind())
	}

	expectedKeyType := reflect.TypeOf(types.NamespacedName{})
	if !typ.Key().AssignableTo(expectedKeyType) {
		return nil, fmt.Errorf("map key must be types.NamespacedName, got %v", typ.Key())
	}

	result := make(map[types.NamespacedName]any)

	for _, key := range val.MapKeys() {
		// Type assertion is safe here - we validated the key type above
		result[key.Interface().(types.NamespacedName)] = val.MapIndex(key).Interface() //nolint:errcheck
	}

	return result, nil
}

func reportToYAML(report reporting.MigrationReport) (string, error) {
	reportYAML, err := yaml.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("failed to marshal report to YAML: %w", err)
	}

	return MigrationReportHeader + string(reportYAML), nil
}
