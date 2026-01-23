package aggregation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/testutil"

	"k8s.io/client-go/kubernetes/fake"
)

func TestOpenAGICLogStream(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// Create a fake pod with the AGIC label
		pod := testutil.MakeAGICPod(strings.Split(agicHelmLabel, "=")...)
		fakeClient := fake.NewClientset(&pod)
		ctx := context.Background()
		stream, err := openAGICLogStream(ctx, agicHelmLabel, fakeClient)

		// The fake client doesn't support log streaming so we only fail on certain errors
		if err != nil {
			if !strings.Contains(err.Error(), "stream") {
				t.Fatalf("Unexpected error: %v", err)
			}
		} else {
			t.Cleanup(func() { _ = stream.Close() })
		}
	})

	t.Run("no pods found", func(t *testing.T) {
		fakeClient := fake.NewClientset()
		ctx := context.Background()

		_, err := openAGICLogStream(ctx, agicHelmLabel, fakeClient)
		if err == nil {
			t.Fatal("expected an error when no pods found, got none")
		}

		if !strings.Contains(err.Error(), "no pods found") {
			t.Fatalf("expected 'no pods found' error, got: %v", err)
		}
	})

	t.Run("pod list fails", func(t *testing.T) {
		fakeClient := fake.NewClientset()
		ctx := context.Background()

		// Simulate a listing failure by using a cancelled context
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		_, err := openAGICLogStream(cancelledCtx, agicHelmLabel, fakeClient)
		if err == nil {
			t.Fatal("expected an error, got none")
		}
	})
}

func TestScrapeAGICConfigBlockFromLog(t *testing.T) {
	t.Run("config not present", func(t *testing.T) {
		input := `some random log output
another line without config
yet another line
`
		result, err := scrapeAGICConfigBlockFromLog(strings.NewReader(input))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "" {
			t.Fatalf("expected empty result, got: %q", result)
		}
	})

	t.Run("existing config", func(t *testing.T) {
		input := `some log output
-- Existing App Gwy Config --{
-- Existing App Gwy Config --  "properties": {
-- Existing App Gwy Config --    "sku": {
-- Existing App Gwy Config --      "name": "Standard_v2"
-- Existing App Gwy Config --    }
-- Existing App Gwy Config --  }
-- Existing App Gwy Config --}
more log output after
`
		expectedConfig := `{
  "properties": {
    "sku": {
      "name": "Standard_v2"
    }
  }
}`
		result, err := scrapeAGICConfigBlockFromLog(strings.NewReader(input))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != expectedConfig {
			t.Fatalf("expected:\n%s\n\ngot:\n%s", expectedConfig, result)
		}
	})

	t.Run("new config", func(t *testing.T) {
		input := `some log output
-- App Gwy config --{
-- App Gwy config --  "properties": {
-- App Gwy config --    "firewallPolicy": {
-- App Gwy config --      "id": "/subscriptions/test/resourceGroups/rg/providers/Microsoft.Network/wafPolicies/test-waf"
-- App Gwy config --    }
-- App Gwy config --  }
-- App Gwy config --}
more log output after
`
		expectedConfig := `{
  "properties": {
    "firewallPolicy": {
      "id": "/subscriptions/test/resourceGroups/rg/providers/Microsoft.Network/wafPolicies/test-waf"
    }
  }
}`
		result, err := scrapeAGICConfigBlockFromLog(strings.NewReader(input))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != expectedConfig {
			t.Fatalf("expected:\n%s\n\ngot:\n%s", expectedConfig, result)
		}
	})

	t.Run("reading from the stream fails", func(t *testing.T) {
		expectedErr := errors.New("simulated read error")
		failingReader := &failingReader{err: expectedErr}

		_, err := scrapeAGICConfigBlockFromLog(failingReader)
		if err == nil {
			t.Fatal("expected an error but got none")
		}

		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

func TestExtractWAFPolicyIDFromLoggedConfig(t *testing.T) {
	t.Run("config block has a WAF Policy ID", func(t *testing.T) {
		expectID := "test-waf-id"
		configBlock := fmt.Sprintf(`
		{
			"properties": {
				"firewallPolicy": {
					"id": "%s"
				}
			}
		}`, expectID)
		gotID, err := extractWAFPolicyIDFromLoggedConfig(configBlock)

		if err != nil {
			t.Fatalf("got unexpected err: %v", err)
		}

		if gotID != expectID {
			t.Fatalf("got ID: %q, expected: %q", gotID, expectID)
		}
	})
	t.Run("config block does not have a WAF Policy ID", func(t *testing.T) {
		gotID, err := extractWAFPolicyIDFromLoggedConfig("{}")
		if err != nil {
			t.Fatalf("got unexpected err: %v", err)
		}

		if gotID != "" {
			t.Fatalf("got unexpected id: %q", gotID)
		}
	})

	t.Run("config block is not valid JSON", func(t *testing.T) {
		_, err := extractWAFPolicyIDFromLoggedConfig("{")
		if err == nil {
			t.Fatal("expected an error but got none")
		}
	})
}

// failingReader in a mocked io.Reader that returns an error
type failingReader struct {
	err error
}

func (f *failingReader) Read([]byte) (n int, err error) {
	return 0, f.err
}
