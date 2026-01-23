package wafpath

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/e2e/framework"
)

const WAFID = "/subscriptions/11111111-1111-1111-bbbb-bbbbbbbbbbbb/resourceGroups/onebox-rg/providers/Microsoft.Network/applicationGatewayWebApplicationFirewallPolicies/my-policy"

func TestWAFPolicyManual_InputOutput(t *testing.T) {
	framework.NewTestCaseFromCurrentDir(t).Run(t, "--byo-resource-id", framework.BYOAGC, "--waf-id", WAFID)
}
