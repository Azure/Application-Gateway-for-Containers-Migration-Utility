package resources

type MigrationStatus string

// Possible values for MigrationStatus
const (
	MigrationStatusNotStarted   MigrationStatus = "NotStarted"
	MigrationStatusNotSupported MigrationStatus = "NotSupported"
	MigrationStatusIgnored      MigrationStatus = "Ignored"
	MigrationStatusCompleted    MigrationStatus = "Completed"
	MigrationStatusError        MigrationStatus = "Error"
	MigrationStatusWarning      MigrationStatus = "Warning"
)
