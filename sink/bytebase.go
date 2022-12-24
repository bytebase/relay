package sink

import (
	"context"
	"strings"

	"github.com/bytebase/relay/payload"
	"github.com/bytebase/relay/service"
)

var (
	_ Sinker = (*bytebaseSinker)(nil)
)

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

func (sinker *bytebaseSinker) Mount() error {
	// hardcode for demo
	sinker.project = "demo"
	sinker.branch = "main"
	sinker.gerritService = service.NewGerrit("https://gerrit.bytebase.com", "ed", "lhAqig2nnTL6gCNXIjDCRnCoi0y9nma48UfjcbsLzA")
	sinker.bytebaseService = service.NewBytebase("http://localhost:8080", "demo@service.bytebase.com", "")
	return nil
}

func (sinker *bytebaseSinker) Process(c context.Context, _ string, pi interface{}) error {
	p := pi.(payload.GerritEvent)

	if p.Change.Branch != sinker.branch || p.Change.Project != sinker.project {
		return nil
	}

	fileMap, err := sinker.gerritService.ListFilesInChange(c, p.ChangeKey.Key, p.PatchSet.Revision)
	if err != nil {
		return err
	}

	for fileName := range fileMap {
		if strings.HasPrefix(fileName, "/") {
			continue
		}
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		// we just consume the file pattern must be {{PROJECT_KEY}}/{{ENV_NAME}}/{{DB_NAME}}.sql in the demo
		sections := strings.Split(fileName, "/")
		if len(sections) != 3 {
			continue
		}

		content, err := sinker.gerritService.GetFileContent(c, p.ChangeKey.Key, p.PatchSet.Revision, fileName)
		if err != nil {
			return err
		}

		projectKey := sections[0]
		envName := sections[1]
		dbName := strings.Split(sections[2], ".sql")[0]

		issueCreate := &payload.IssueCreate{
			ProjectKey: projectKey,
			Name:       "Alter schema",
			Type:       payload.IssueDatabaseSchemaUpdate,
			MigrationList: []*payload.MigrationDetail{
				{
					DatabaseName:    dbName,
					EnvironmentName: envName,
					Statement:       content,
					MigrationType:   payload.Migrate,
				},
			},
		}
		if err := sinker.bytebaseService.CreateIssue(c, issueCreate); err != nil {
			return err
		}
	}

	return nil
}
