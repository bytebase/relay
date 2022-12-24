package payload

// MigrationType is the type of a migration.
type MigrationType string

// IssueType is the type of an issue.
type IssueType string

const (
	// Used for DDL change including CREATE DATABASE.
	Migrate MigrationType = "MIGRATE"
	// Used for DML change.
	Data MigrationType = "DATA"

	// IssueDatabaseSchemaUpdate is the issue type for updating database schemas (DDL).
	IssueDatabaseSchemaUpdate IssueType = "bb.issue.database.schema.update"
	// IssueDatabaseDataUpdate is the issue type for updating database data (DML).
	IssueDatabaseDataUpdate IssueType = "bb.issue.database.data.update"
)

// MigrationDetail is the detail for database migration such as Migrate, Data.
type MigrationDetail struct {
	MigrationType   MigrationType `json:"migrationType"`
	DatabaseName    string        `json:"databaseName"`
	EnvironmentName string        `json:"environmentName"`
	Statement       string        `json:"statement"`
	SchemaVersion   string        `json:"schemaVersion"`
}

type IssueCreate struct {
	ProjectKey    string             `json:"projectKey"`
	Name          string             `json:"name"`
	Type          IssueType          `json:"type"`
	Description   string             `json:"description"`
	MigrationList []*MigrationDetail `json:"migrationList"`
}
