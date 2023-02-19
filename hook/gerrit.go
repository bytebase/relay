package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/relay/payload"
	"github.com/bytebase/relay/service"
	flag "github.com/spf13/pflag"
)

var (
	_                   Hooker = (*gerritHooker)(nil)
	gerritProject       string
	gerritProjectBranch string
	gerritURL           string
	gerritAccount       string
	gerritPassword      string
)

// NewGerrit creates a Gerrit hooker
func NewGerrit() Hooker {
	return &gerritHooker{
		gerritService: service.NewGerrit(gerritURL, gerritAccount, gerritPassword),
	}
}

func init() {
	// For demo we only supports monitor one branch in one project.
	flag.StringVar(&gerritProject, "gerrit-repository", "", "The Gerrit repository name")
	flag.StringVar(&gerritProjectBranch, "gerrit-branch", "main", "The branch name in Gerrit repository")

	flag.StringVar(&gerritURL, "gerrit-url", "https://gerrit.bytebase.com", "The Gerrit service URL")
	flag.StringVar(&gerritAccount, "gerrit-account", "", "The Gerrit service account name")
	flag.StringVar(&gerritPassword, "gerrit-password", "", "The Gerrit service account password")
}

type gerritHooker struct {
	gerritService *service.GerritService
}

func (hooker *gerritHooker) handler() (func(r *http.Request) Response, error) {
	return func(r *http.Request) Response {
		if gerritURL == "" {
			return Response{
				httpCode: http.StatusAccepted,
				detail:   "Skip, --gerrit-url is not set",
			}
		}
		if gerritAccount == "" {
			return Response{
				httpCode: http.StatusAccepted,
				detail:   "Skip, --gerrit-account is not set",
			}
		}
		if gerritPassword == "" {
			return Response{
				httpCode: http.StatusAccepted,
				detail:   "Skip, --gerrit-password is not set",
			}
		}

		var message payload.GerritEvent
		err := json.NewDecoder(r.Body).Decode(&message)
		if err != nil {
			return Response{
				httpCode: http.StatusBadRequest,
				detail:   fmt.Sprintf("Failed to decode request body: %q", err),
			}
		}

		if message.Type != payload.GerritEventChangeMerged {
			return Response{
				httpCode: http.StatusAccepted,
				detail:   fmt.Sprintf("Skip %s event", message.Type),
			}
		}

		if message.Change.Project != gerritProject || message.Change.Branch != gerritProjectBranch {
			return Response{
				httpCode: http.StatusAccepted,
				detail:   fmt.Sprintf("Skip the message for %s branch in %s project", message.Change.Branch, message.Change.Project),
			}
		}

		ctx := context.Background()
		fileMap, err := hooker.gerritService.ListFilesInChange(ctx, message.Change.ID, message.PatchSet.Revision)
		if err != nil {
			return Response{
				httpCode: http.StatusInternalServerError,
				payload:  err.Error(),
			}
		}

		changedFileList := []*payload.GerritChangedFile{}
		for fileName := range fileMap {
			if strings.HasPrefix(fileName, "/") {
				continue
			}
			if !strings.HasSuffix(fileName, ".sql") {
				continue
			}
			content, err := hooker.gerritService.GetFileContent(ctx, message.Change.ID, message.PatchSet.Revision, fileName)
			if err != nil {
				return Response{
					httpCode: http.StatusInternalServerError,
					payload:  err.Error(),
				}
			}

			changedFileList = append(changedFileList, &payload.GerritChangedFile{
				FileName: fileName,
				Content:  content,
			})
		}

		return Response{
			httpCode: http.StatusOK,
			payload: payload.GerritFileChangeMessage{
				Files: changedFileList,
			},
		}
	}, nil
}
