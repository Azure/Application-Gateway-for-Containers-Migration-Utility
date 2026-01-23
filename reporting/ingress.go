package reporting

import (
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources/k8snames"
)

type Ingress struct {
	types.NamespacedName     `yaml:",inline"`
	MigratedIngressResources struct {
		HTTPRoutes      []types.NamespacedName `yaml:"httpRoutes"` // List of HTTPRoutes created from this Ingress
		ReferenceGrants []types.NamespacedName `yaml:"referenceGrants,omitempty"`

		// We will want to consciously decide what to include in the report, these may not make sense
		// e.g. we might only want to return a count
		HealthCheckPolicies           []Policy `yaml:"healthCheckPolicies,omitempty"`
		RoutePolicies                 []Policy `yaml:"routePolicies,omitempty"`
		WebApplicationFirewalPolicies []Policy `yaml:"webApplicationFirewallPolicies,omitempty"`
		FrontendTLSPolicies           []Policy `yaml:"frontendTLSPolicies,omitempty"`
		BackendTLSPolicies            []Policy `yaml:"backendTLSPolicies,omitempty"`
	} `yaml:"migratedResources"`
	Annotations Annotations               `yaml:"annotations"`
	Result      resources.MigrationStatus `yaml:"-"`
}

type Policy struct {
	types.NamespacedName `yaml:",inline"`
}

type Annotations struct {
	Completed   []CompletedAnnotation `yaml:"completed,omitempty"`
	Warnings    []FailedAnnotation    `yaml:"warnings,omitempty"`
	Failed      []FailedAnnotation    `yaml:"failed,omitempty"`
	Unsupported []FailedAnnotation    `yaml:"unsupported,omitempty"`
}

type CompletedAnnotation struct {
	KeyValue string `yaml:"annotation"`
}

type FailedAnnotation struct {
	KeyValue string  `yaml:"annotation"`
	Issues   []Issue `yaml:"issues,omitempty"`
}

type AnnotationResult struct {
	KeyValue string                    `yaml:"annotation"`
	Status   resources.MigrationStatus `yaml:"status"`
	Issues   []Issue                   `yaml:"issues,omitempty"`
}

func newIngress(ingress *resources.IngressContext) Ingress {
	reportIngress := Ingress{
		NamespacedName: types.NamespacedName{
			Name:      ingress.Ingress.Name,
			Namespace: ingress.Ingress.Namespace,
		},

		Result: ingress.Status,
	}

	reportIngress.MigratedIngressResources.HTTPRoutes = append(reportIngress.MigratedIngressResources.HTTPRoutes, ingress.HTTPRoutes...)

	healthCheckPolicies := sets.New[types.NamespacedName]()
	backendTLSPolicies := sets.New[types.NamespacedName]()
	frontendTLSPolicies := sets.New[types.NamespacedName]()
	routePolicies := sets.New[types.NamespacedName]()
	httpRoutes := sets.New[types.NamespacedName]()
	webApplicationFirewallPolicies := sets.New[types.NamespacedName]()

	orderedAnnotations := make([]string, 0, len(ingress.Annotations))
	for key := range ingress.Annotations {
		orderedAnnotations = append(orderedAnnotations, key)
	}

	slices.Sort(orderedAnnotations)

	for _, key := range orderedAnnotations {
		annotation := ingress.Annotations[key]
		for resource := range annotation.DestinationResources {
			policy := Policy{
				NamespacedName: types.NamespacedName{Name: resource.Name, Namespace: resource.Namespace},
			}

			switch resource.Kind {
			case k8snames.KindRoutePolicy:
				if _, ok := routePolicies[policy.NamespacedName]; !ok {
					reportIngress.MigratedIngressResources.RoutePolicies = append(reportIngress.MigratedIngressResources.RoutePolicies, policy)
					routePolicies.Insert(policy.NamespacedName)
				}
			case k8snames.KindHealthCheckPolicy:
				if _, ok := healthCheckPolicies[policy.NamespacedName]; !ok {
					reportIngress.MigratedIngressResources.HealthCheckPolicies = append(reportIngress.MigratedIngressResources.HealthCheckPolicies, policy)
					healthCheckPolicies.Insert(policy.NamespacedName)
				}
			case k8snames.KindWebApplicationFirewallPolicy:
				if _, ok := webApplicationFirewallPolicies[policy.NamespacedName]; !ok {
					reportIngress.MigratedIngressResources.WebApplicationFirewalPolicies = append(reportIngress.MigratedIngressResources.WebApplicationFirewalPolicies, policy)
					webApplicationFirewallPolicies.Insert(policy.NamespacedName)
				}
			case k8snames.KindHTTPRoute:
				if _, ok := httpRoutes[policy.NamespacedName]; !ok {
					reportIngress.MigratedIngressResources.HTTPRoutes = append(reportIngress.MigratedIngressResources.HTTPRoutes, policy.NamespacedName)
					httpRoutes.Insert(policy.NamespacedName)
				}
			case k8snames.KindBackendTLSPolicy:
				if _, ok := backendTLSPolicies[policy.NamespacedName]; !ok {
					reportIngress.MigratedIngressResources.BackendTLSPolicies = append(reportIngress.MigratedIngressResources.BackendTLSPolicies, policy)
					backendTLSPolicies.Insert(policy.NamespacedName)
				}
			case k8snames.KindFrontendTLSPolicy:
				if _, ok := frontendTLSPolicies[policy.NamespacedName]; !ok {
					reportIngress.MigratedIngressResources.FrontendTLSPolicies = append(reportIngress.MigratedIngressResources.FrontendTLSPolicies, policy)
					frontendTLSPolicies.Insert(policy.NamespacedName)
				}
			}
		}

		kv := fmt.Sprintf("%s=%s", annotation.Key, annotation.Value)

		switch annotation.Status() {
		case resources.MigrationStatusCompleted:
			reportIngress.Annotations.Completed = append(reportIngress.Annotations.Completed, CompletedAnnotation{KeyValue: kv})
		case resources.MigrationStatusWarning:
			reportIngress.Annotations.Warnings = append(reportIngress.Annotations.Warnings, newFailedAnnotation(*annotation))
		case resources.MigrationStatusError:
			reportIngress.Annotations.Failed = append(reportIngress.Annotations.Failed, newFailedAnnotation(*annotation))
		case resources.MigrationStatusNotSupported:
			reportIngress.Annotations.Unsupported = append(reportIngress.Annotations.Unsupported, newFailedAnnotation(*annotation))
		}
	}

	return reportIngress
}

func newFailedAnnotation(annoCtx resources.IngressAnnotationContext) FailedAnnotation {
	result := FailedAnnotation{
		KeyValue: fmt.Sprintf("%s=%s", annoCtx.Key, annoCtx.Value),
		Issues:   []Issue{},
	}

	codes := sets.New[resources.IssueCode]()
	for _, issue := range annoCtx.Issues {
		// filter out duplicates, which can occur when an ingresses get chopped up inter overlapping routes
		if codes.Has(issue.Code) {
			continue
		}

		codes.Insert(issue.Code)
		libraryEntry := resources.IssueLibrary[issue.Code]
		// since here the error appears alongside the recommendation, omit the generic "please review the error" recommendation
		if libraryEntry.Recommendation == resources.RecommendationPleaseReviewTheErrorMessage {
			libraryEntry.Recommendation = ""
		}

		msg := ""
		if issue.Error != nil {
			msg = issue.Error.Error()
		}

		result.Issues = append(result.Issues, Issue{
			Description:    libraryEntry.Description,
			Recommendation: libraryEntry.Recommendation,
			Message:        msg,
		})
	}

	return result
}

func (i Ingress) sortContents() {
	for _, s := range []*[]Policy{
		&i.MigratedIngressResources.HealthCheckPolicies, &i.MigratedIngressResources.RoutePolicies, &i.MigratedIngressResources.WebApplicationFirewalPolicies,
		&i.MigratedIngressResources.FrontendTLSPolicies, &i.MigratedIngressResources.BackendTLSPolicies,
	} {
		slices.SortFunc(*s, func(a, b Policy) int {
			return namespacedNameCmp(a.NamespacedName, b.NamespacedName)
		})
	}

	slices.SortFunc(i.MigratedIngressResources.HTTPRoutes, namespacedNameCmp)
}
