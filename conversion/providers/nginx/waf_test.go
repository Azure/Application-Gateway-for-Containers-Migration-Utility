package nginx

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/resources"
)

func TestNGINXHandleEnableModSecurity(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationEnableModSecurity, "true",
	)

	err := provider.handleEnableModSecurity(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleEnableModSecurity returned error: %v", err)
	}

	// Check that a WAF policy was created
	if len(graph.WAFPolicies) == 0 {
		t.Fatal("expected WAFPolicy to be created")
	}

	// Check that the annotation has a warning status
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleEnableModSecurityFalse(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationEnableModSecurity, "false",
	)

	err := provider.handleEnableModSecurity(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleEnableModSecurity returned error: %v", err)
	}

	// Check that no WAF policy was created
	if len(graph.WAFPolicies) != 0 {
		t.Error("expected no WAFPolicy to be created when modsecurity is false")
	}

	if annotationCtx.Status() != resources.MigrationStatusCompleted {
		t.Errorf("expected status Completed, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleEnableOWASPCoreRules(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationEnableOWASPCoreRules, "true",
	)

	err := provider.handleEnableOWASPCoreRules(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleEnableOwaspCoreRules returned error: %v", err)
	}

	// Check that a WAF policy was created
	if len(graph.WAFPolicies) == 0 {
		t.Fatal("expected WAFPolicy to be created")
	}

	// Check that the annotation has a warning status (needs manual WAF policy ID)
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleEnableOWASPCoreRulesFalse(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationEnableOWASPCoreRules, "false",
	)

	err := provider.handleEnableOWASPCoreRules(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleEnableOWASPCoreRules returned error: %v", err)
	}

	// Check that no WAF policy was created
	if len(graph.WAFPolicies) != 0 {
		t.Error("expected no WAFPolicy to be created when OWASP rules are false")
	}

	if annotationCtx.Status() != resources.MigrationStatusCompleted {
		t.Errorf("expected status Completed, got %s", annotationCtx.Status())
	}
}

func TestNGINXHandleModSecuritySnippet(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationModSecuritySnippet, "SecRule ...",
	)

	err := provider.handleModSecuritySnippet(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleModSecuritySnippet returned error: %v", err)
	}

	// ModSecurity snippets require manual conversion - expect warning
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning, got %s", annotationCtx.Status())
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issues to be registered")
	}
}

func TestNGINXHandleEnableModSecurityInvalid(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationEnableModSecurity, "invalid",
	)

	err := provider.handleEnableModSecurity(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err == nil {
		t.Error("expected error for invalid bool value")
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered for invalid value")
	}
}

func TestNGINXHandleEnableOWASPCoreRulesInvalid(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationEnableOWASPCoreRules, "invalid",
	)

	err := provider.handleEnableOWASPCoreRules(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err == nil {
		t.Error("expected error for invalid bool value")
	}

	if len(annotationCtx.Issues) == 0 {
		t.Error("expected issue to be registered for invalid value")
	}
}

func TestNGINXHandleModSecurityTransactionID(t *testing.T) {
	provider, graph, gwCtx, routeCtx, _, annotationCtx := setupAnnotationHandlerInputs(
		AnnotationModSecurityTransactionID, "$request_id",
	)

	err := provider.handleModSecurityTransactionID(graph, gwCtx, routeCtx, nil, annotationCtx)
	if err != nil {
		t.Errorf("handleModSecurityTransactionID returned error: %v", err)
	}

	// Should be marked as warning since it has an informational note about trackingId
	// The annotation IS implemented, but the warning provides useful context
	if annotationCtx.Status() != resources.MigrationStatusWarning {
		t.Errorf("expected status Warning (with informational note), got %s", annotationCtx.Status())
	}

	// Should have an informational issue about trackingId
	if len(annotationCtx.Issues) == 0 {
		t.Error("expected informational issue about trackingId to be registered")
	}

	// Verify the issue is the correct one
	found := false

	for _, issue := range annotationCtx.Issues {
		if issue.Code == resources.IssueNGINXModSecurityTransactionID {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected IssueNGINXModSecurityTransactionID issue to be registered")
	}
}
