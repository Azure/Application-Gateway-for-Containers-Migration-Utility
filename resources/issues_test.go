package resources

import "testing"

func TestIssueLevelMigrationStatus(t *testing.T) {
	tests := []struct {
		name           string
		level          IssueLevel
		expectedStatus MigrationStatus
	}{
		{
			name:           "NotSupported level returns NotSupported status",
			level:          IssueLevelNotSupported,
			expectedStatus: MigrationStatusNotSupported,
		},
		{
			name:           "Warning level returns Warning status",
			level:          IssueLevelWarning,
			expectedStatus: MigrationStatusWarning,
		},
		{
			name:           "Error level returns Error status",
			level:          IssueLevelError,
			expectedStatus: MigrationStatusError,
		},
		{
			name:           "Unknown level returns Error status",
			level:          IssueLevel("unknown"),
			expectedStatus: MigrationStatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.MigrationStatus()
			if got != tt.expectedStatus {
				t.Errorf("IssueLevel.MigrationStatus() = %v, want %v", got, tt.expectedStatus)
			}
		})
	}
}
