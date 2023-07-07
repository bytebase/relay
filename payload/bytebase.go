package payload

import "time"

// MigrationType is the type of a migration.
type MigrationType string

// IssueType is the type of an issue.
type IssueType string

const (
	// Used for DDL change including CREATE DATABASE.
	Migrate MigrationType = "MIGRATE"
	// Used for DML change.
	Data MigrationType = "DATA"
)

// IssueCreate is the API message for creating a issue.
type IssueCreate struct {
	// Related fields
	ProjectKey  string `json:"projectKey"`
	Database    string `json:"database"`
	Environment string `json:"environment"`

	// Domain specific fields
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	MigrationType MigrationType `json:"migrationType"`
	Statement     string        `json:"statement"`
	SchemaVersion string        `json:"schemaVersion"`
}

type GlobalState struct {
	IssueList map[string]*Issue
}

type Issue struct {
	ID            string         `json:"id"`
	Status        string         `json:"status"`
	BytebaseIssue *BytebaseIssue `json:"issue"`
}

type BytebaseIssue struct {
	IssueID     string    `json:"issueId"`
	Creator     string    `json:"creator"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Project     string    `json:"project"`
	Assignee    string    `json:"assignee"`
	Statement   string    `json:"statement"`
	CreateTime  time.Time `json:"createTime"`
}
