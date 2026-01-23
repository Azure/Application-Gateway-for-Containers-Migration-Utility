package aggregation

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	agicHelmLabel               = "app=ingress-azure"
	agicAddonLabel              = "app=ingress-appgw"
	agicContainerName           = "ingress-azure"
	agicLogExistingConfigPrefix = "-- Existing App Gwy Config --"
	agicLogNewConfigPrefix      = "-- App Gwy config --"
)

// scrapeAGICLogsForWAFPolicyID attempts to extract a WAF Policy ID from the Ingress Controller logs.
// We could instead get it from ARM but that would require acquiring additional permissions from the user.
func scrapeAGICLogsForWAFPolicyID(ctx context.Context, agicLabel string, client kubernetes.Interface) (string, error) {
	stream, err := openAGICLogStream(ctx, agicLabel, client)
	if err != nil {
		return "", fmt.Errorf("failed to open log stream to ingress controller: %w", err)
	}

	defer func() { _ = stream.Close() }()

	configBlock, err := scrapeAGICConfigBlockFromLog(stream)
	if err != nil {
		return "", fmt.Errorf("failed to scrape config block from ingress-azure pod logs: %w", err)
	}

	wafID, err := extractWAFPolicyIDFromLoggedConfig(configBlock)
	if err != nil {
		return "", fmt.Errorf("failed to extract WAF policy ID from ingress-azure pod logs: %w", err)
	}

	return wafID, err
}

func openAGICLogStream(ctx context.Context, agicLabel string, client kubernetes.Interface) (io.ReadCloser, error) {
	pods, err := client.CoreV1().Pods("").List(ctx, meta_v1.ListOptions{
		LabelSelector: agicLabel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingress controller pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for label selector: %s", agicLabel)
	}

	pod := pods.Items[0]
	logs := client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &core_v1.PodLogOptions{
		Container: agicContainerName,
		Follow:    false,
	})

	stream, err := logs.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return stream, nil
}

func scrapeAGICConfigBlockFromLog(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)

	var current []string

	var lastBlock []string

	for scanner.Scan() {
		line := scanner.Text()

		if after, ok := strings.CutPrefix(line, agicLogExistingConfigPrefix); ok {
			current = append(current, after)
		} else if after, ok := strings.CutPrefix(line, agicLogNewConfigPrefix); ok {
			current = append(current, after)
		} else if len(current) > 0 {
			lastBlock = current
			current = nil
		}
	}

	if len(current) > 0 && lastBlock[len(lastBlock)-1] == "}" {
		lastBlock = current
	}

	if len(lastBlock) > 0 {
		if strings.HasPrefix(strings.TrimSpace(lastBlock[0]), "\"") {
			lastBlock = append([]string{"{"}, lastBlock...)
		}
	}

	return strings.Join(lastBlock, "\n"), scanner.Err()
}

func extractWAFPolicyIDFromLoggedConfig(configBlock string) (string, error) {
	var config struct {
		Properties struct {
			WAFPolicyID struct {
				ID string `json:"id"`
			} `json:"firewallPolicy"`
		}
	}

	if err := json.Unmarshal([]byte(configBlock), &config); err != nil {
		return "", err
	}

	return config.Properties.WAFPolicyID.ID, nil
}
