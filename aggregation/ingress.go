package aggregation

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
	"github.com/kubernetes-sigs/ingress2gateway/pkg/i2gw/providers/common"
	network_v1 "k8s.io/api/networking/v1"
	ctrl_client "sigs.k8s.io/controller-runtime/pkg/client"
)

// Functions that implement ingressFilterFunc should return true if the given ingress passes the filter.
type ingressFilterFunc func(network_v1.Ingress) bool

func (a Aggregator) ingressClassNameFilter(ingressClass string) ingressFilterFunc {
	return func(ingress network_v1.Ingress) bool {
		if ingressClass == "" {
			return true
		}

		gotClass := common.GetIngressClass(ingress)
		if gotClass != ingressClass {
			a.log.Info(
				"Ignoring Ingress due to wrong class",
				"ingress",
				k8snames.NamespacedName(&ingress),
				"ingress-class",
				gotClass,
				"filter-class",
				ingressClass,
			)

			return false
		}

		return true
	}
}

func (Aggregator) collectIngressesFromCluster(
	ctx context.Context,
	client ctrl_client.Client,
	filter ingressFilterFunc,
) (map[types.NamespacedName]*network_v1.Ingress, error) {
	var ingressList network_v1.IngressList
	err := client.List(ctx, &ingressList)

	if err != nil {
		return nil, fmt.Errorf("failed to list Ingresses from cluster: %w", err)
	}

	ingresses := map[types.NamespacedName]*network_v1.Ingress{}

	for _, ingress := range ingressList.Items {
		if filter(ingress) {
			ingresses[types.NamespacedName{Namespace: ingress.Namespace, Name: ingress.Name}] = &ingress
		}
	}

	return ingresses, nil
}

func (a Aggregator) collectIngressesFromFiles(
	filter ingressFilterFunc,
	filepath ...string,
) (map[types.NamespacedName]*network_v1.Ingress, error) {
	ingresses := make(map[types.NamespacedName]*network_v1.Ingress)

	for _, filename := range filepath {
		unstructuredObjects, err := a.readUnstructuredFromFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to process file %q: %w", filename, err)
		}

		for _, obj := range unstructuredObjects {
			if obj.GroupVersionKind().Kind != k8snames.KindIngress {
				continue
			}

			var ingress network_v1.Ingress
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(
				obj.UnstructuredContent(),
				&ingress,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"failed to process object to ingress in file %q, %w",
					filename,
					err,
				)
			}

			if filter(ingress) {
				ingressNN := k8snames.NamespacedName(&ingress)
				if _, ok := ingresses[ingressNN]; ok {
					a.log.Warn(
						"Duplicate Ingress with being overwritten",
						"ingress",
						ingressNN,
						"overwritten-by",
						filename,
					)
				}

				ingresses[ingressNN] = &ingress
			}
		}
	}

	return ingresses, nil
}

func (a Aggregator) readUnstructuredFromFile(
	filename string,
) ([]*unstructured.Unstructured, error) {
	file, err := os.Open(filename) //nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			a.log.Error("failed to close file", "file", filename, "error", err)
		}
	}()

	unstructuredObjects, err := common.ExtractObjectsFromReader(file, "")
	if err != nil {
		return nil, fmt.Errorf("failed to extract objects: %w", err)
	}

	return unstructuredObjects, nil
}
