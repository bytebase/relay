package sink

import (
	"context"
	"fmt"
	"log"
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
	gerritUR               string
	gerritAccount          string
	gerritPassword         string
	gerritRepository       string
	gerritRepositoryBranch string
	bytebaseURL            string
	bytebaseServiceAccount string
	bytebaseServiceKey     string
	issueNameTemplate      string = "[%s] %s"
	filePathTemplate       string = "{{PROJECT_KEY}}/{{ENV_NAME}}/{{DB_NAME}}##{{VERSION}}##{{TYPE}}##{{DESCRIPTION}}.sql"
	placeholderRegexp      string = `[^\\/?%*:|"<>]+`
	placeholderList               = []string{
		"PROJECT_KEY",
		"ENV_NAME",
		"VERSION",
		"DB_NAME",
		"TYPE",
		"DESCRIPTION",
	}
)

func init() {
	flag.StringVar(&gerritUR, "gerrit-url", "https://gerrit.bytebase.com", "The Gerrit service URL")
	flag.StringVar(&gerritAccount, "gerrit-account", "", "The Gerrit service account name")
	flag.StringVar(&gerritPassword, "gerrit-password", "", "The Gerrit service account password")
	flag.StringVar(&gerritRepository, "gerrit-repository", "", "The Gerrit repository name")
	flag.StringVar(&gerritRepositoryBranch, "gerrit-branch", "main", "The branch name in Gerrit repository")
	flag.StringVar(&bytebaseURL, "bytebase-url", "http://localhost:8080", "The Bytebase service URL")
	flag.StringVar(&bytebaseServiceAccount, "bytebase-service-account", "", "The Bytebase service account name")
	flag.StringVar(&bytebaseServiceKey, "bytebase-service-key", "", "The Bytebase service account key")
}

// NewBytebase creates a Bytebase sinker
func NewBytebase() Sinker {
	return &bytebaseSinker{}
}

type bytebaseSinker struct {
	project         string
	branch          string
	gerritService   *service.GerritService
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
	if gerritUR == "" || gerritAccount == "" || gerritPassword == "" {
		return fmt.Errorf(`the "--gerrit-url, --gerrit-account and --gerrit-password" is required`)
	}
	if gerritRepository == "" || gerritRepositoryBranch == "" {
		return fmt.Errorf(`the "--gerrit-repository and --gerrit-branch" is required`)
	}
	if bytebaseURL == "" || bytebaseServiceAccount == "" || bytebaseServiceKey == "" {
		return fmt.Errorf(`the "--bytebase-url, --bytebase-service-account and --bytebase-service-key" is required`)
	}

	// hardcode for demo
	sinker.project = gerritRepository
	sinker.branch = gerritRepositoryBranch
	// sinker.gerritService = service.NewGerrit("https://gerrit.bytebase.com", "ed", "lhAqig2nnTL6gCNXIjDCRnCoi0y9nma48UfjcbsLzA")
	sinker.gerritService = service.NewGerrit(gerritUR, gerritAccount, gerritPassword)
	sinker.bytebaseService = service.NewBytebase(bytebaseURL, bytebaseServiceAccount, bytebaseServiceKey)
	return nil
}

func (sinker *bytebaseSinker) Process(c context.Context, _ string, pi interface{}) error {
	p := pi.(payload.GerritEvent)

	if p.Change.Branch != sinker.branch || p.Change.Project != sinker.project {
		log.Printf("ignore event as the branch or project doesn't match, expect %s:%s but got %s:%s", sinker.project, sinker.branch, p.Change.Project, p.Change.Branch)
		return nil
	}

	fileMap, err := sinker.gerritService.ListFilesInChange(c, p.Change.ID, p.PatchSet.Revision)
	if err != nil {
		return err
	}

	for fileName := range fileMap {
		log.Printf("processing file %s\n", fileName)

		if strings.HasPrefix(fileName, "/") {
			continue
		}
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		mi, err := parseMigrationInfo(fileName, filePathTemplate)
		if err != nil {
			return err
		}

		content, err := sinker.gerritService.GetFileContent(c, p.Change.ID, p.PatchSet.Revision, fileName)
		if err != nil {
			return err
		}

		issueName := fmt.Sprintf(issueNameTemplate, mi.Name, fileName)
		log.Printf("start create issue %s\n", issueName)
		issueCreate := &payload.IssueCreate{
			ProjectKey:    mi.Project,
			Database:      mi.Database,
			Environment:   mi.Environment,
			Name:          issueName,
			Description:   mi.Description,
			MigrationType: mi.Type,
			Statement:     content,
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
