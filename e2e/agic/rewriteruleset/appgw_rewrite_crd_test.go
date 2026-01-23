package healthcheckpolicy

import (
	"testing"

	"github.com/Azure/Application-Gateway-for-Containers-Migration-Utility/e2e/framework"
)

func TestAppgwRewriteCRD_InputOutput(t *testing.T) {
	framework.NewTestCaseFromCurrentDir(t).Run(t)
}
