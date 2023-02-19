package sink

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bytebase/relay/payload"
	"github.com/bytebase/relay/service"
	flag "github.com/spf13/pflag"
)

var (
	_ Sinker = (*bytebaseSinker)(nil)
)

var (
	bytebaseURL            string
	bytebaseServiceAccount string
	bytebaseServiceKey     string
	// hard code for demo
	issueNameTemplate string = "[%s] %s"
	filePathTemplate  string = "{{PROJECT_KEY}}/{{ENV_NAME}}/{{DB_NAME}}##{{VERSION}}##{{TYPE}}##{{DESCRIPTION}}.sql"
	placeholderRegexp string = `[^\\/?%*:|"<>]+`
	placeholderList          = []string{
		"PROJECT_KEY",
		"ENV_NAME",
		"VERSION",
		"DB_NAME",
		"TYPE",
		"DESCRIPTION",
	}
)

func init() {
	flag.StringVar(&bytebaseURL, "bytebase-url", "http://localhost:8080", "The Bytebase service URL")
	flag.StringVar(&bytebaseServiceAccount, "bytebase-service-account", "", "The Bytebase service account name")
	flag.StringVar(&bytebaseServiceKey, "bytebase-service-key", "", "The Bytebase service account key")
}

// NewBytebase creates a Bytebase sinker
func NewBytebase() Sinker {
	return &bytebaseSinker{}
}

type bytebaseSinker struct {
	bytebaseService *service.BytebaseService
}

type migrationInfo struct {
	Type        payload.MigrationType
	Version     string
	Database    string
	Environment string
	Description string
	Project     string
	Name        string
}

func (sinker *bytebaseSinker) Mount() error {
	if bytebaseURL == "" {
		fmt.Printf("--bytebase-url is missing, Bytebase sinker will not be able to process any events.\n")
		return nil
	}
	if bytebaseServiceAccount == "" {
		fmt.Printf("--bytebase-service-account is missing, Bytebase sinker will not be able to process any events.\n")
		return nil
	}
	if bytebaseServiceKey == "" {
		fmt.Printf("---bytebase-service-key is missing, Bytebase sinker will not be able to process any events.\n")
		return nil
	}

	sinker.bytebaseService = service.NewBytebase(bytebaseURL, bytebaseServiceAccount, bytebaseServiceKey)
	return nil
}

func (sinker *bytebaseSinker) Process(c context.Context, _ string, pi interface{}) error {
	if bytebaseURL == "" {
		return fmt.Errorf("--bytebase-url is required")
	}
	if bytebaseServiceAccount == "" {
		return fmt.Errorf("--bytebase-service-account is required")
	}
	if bytebaseServiceKey == "" {
		return fmt.Errorf("---bytebase-service-key is required")
	}

	change := pi.(payload.GerritFileChangeMessage)

	for _, file := range change.Files {
		mi, err := parseMigrationInfo(file.FileName, filePathTemplate)
		if err != nil {
			return err
		}

		issueName := fmt.Sprintf(issueNameTemplate, mi.Name, file.FileName)
		issueCreate := &payload.IssueCreate{
			ProjectKey:    mi.Project,
			Database:      mi.Database,
			Environment:   mi.Environment,
			Name:          issueName,
			Description:   mi.Description,
			MigrationType: mi.Type,
			Statement:     file.Content,
			SchemaVersion: mi.Version,
		}
		if err := sinker.bytebaseService.CreateIssue(c, issueCreate); err != nil {
			return err
		}
	}

	return nil
}

// parseMigrationInfo matches filePath against filePathTemplate
func parseMigrationInfo(filePath, filePathTemplate string) (*migrationInfo, error) {
	// Escape "." characters to match literals instead of using it as a wildcard.
	filePathRegex := strings.ReplaceAll(filePathTemplate, `.`, `\.`)

	filePathRegex = strings.ReplaceAll(filePathRegex, `/*/`, `/[^/]+/`)
	filePathRegex = strings.ReplaceAll(filePathRegex, `**`, `.*`)

	for _, placeholder := range placeholderList {
		filePathRegex = strings.ReplaceAll(filePathRegex, fmt.Sprintf("{{%s}}", placeholder), fmt.Sprintf(`(?P<%s>%s)`, placeholder, placeholderRegexp))
	}
	myRegex, err := regexp.Compile(filePathRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid file path template: %q", filePathTemplate)
	}
	if !myRegex.MatchString(filePath) {
		// File path does not match file path template.
		return nil, nil
	}

	mi := &migrationInfo{
		Type: payload.Migrate,
	}
	matchList := myRegex.FindStringSubmatch(filePath)
	for _, placeholder := range placeholderList {
		index := myRegex.SubexpIndex(placeholder)
		if index >= 0 {
			switch placeholder {
			case "PROJECT_KEY":
				mi.Project = matchList[index]
			case "ENV_NAME":
				mi.Environment = matchList[index]
			case "VERSION":
				mi.Version = matchList[index]
			case "DB_NAME":
				mi.Database = matchList[index]
			case "TYPE":
				switch matchList[index] {
				case "data", "dml":
					mi.Type = payload.Data
					mi.Type = payload.Data
					mi.Name = "Change data"
				case "migrate", "ddl":
					mi.Type = payload.Migrate
					mi.Type = payload.Migrate
					mi.Name = "Alter schema"
				default:
					return nil, fmt.Errorf("file path %q contains invalid migration type %q, must be 'migrate'('ddl') or 'data'('dml')", filePath, matchList[index])
				}
			case "DESCRIPTION":
				mi.Description = matchList[index]
			}
		}
	}

	if mi.Version == "" {
		return nil, fmt.Errorf("file path %q does not contain {{VERSION}}, configured file path template %q", filePath, filePathTemplate)
	}
	if mi.Database == "" {
		return nil, fmt.Errorf("file path %q does not contain {{DB_NAME}}, configured file path template %q", filePath, filePathTemplate)
	}

	if mi.Description == "" {
		switch mi.Type {
		case payload.Data:
			mi.Description = fmt.Sprintf("Create %s data change", mi.Database)
		default:
			mi.Description = fmt.Sprintf("Create %s schema migration", mi.Database)
		}
	} else {
		// Replace _ with space
		mi.Description = strings.ReplaceAll(mi.Description, "_", " ")
		// Capitalize first letter
		mi.Description = strings.ToUpper(mi.Description[:1]) + mi.Description[1:]
	}

	return mi, nil
}
