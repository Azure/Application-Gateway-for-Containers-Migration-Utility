// Package agicmigration provides utils for the agic migration tool
package agicmigration

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var version string

// GetVersion returns the current version of the agicmigration tool
func GetVersion() string {
	return strings.TrimSpace(version)
}
