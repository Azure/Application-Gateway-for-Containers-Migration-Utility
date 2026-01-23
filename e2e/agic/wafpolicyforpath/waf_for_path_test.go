package wafpath

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/e2e/framework"
)

func TestWAFPolicyForPath_InputOutput(t *testing.T) {
	framework.NewTestCaseFromCurrentDir(t).Run(t)
}
