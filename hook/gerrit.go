package hook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bytebase/relay/payload"
	flag "github.com/spf13/pflag"
)

var (
	_                   Hooker = (*gerritHooker)(nil)
	gerritProject       string
	gerritProjectBranch string
)

// NewGerrit creates a Gerrit hooker
func NewGerrit() Hooker {
	return &gerritHooker{}
}

func init() {
	// For demo we only supports monitor one branch in one project.
	flag.StringVar(&gerritProject, "gerrit-project", "", "The Gerrit repository name")
	flag.StringVar(&gerritProjectBranch, "gerrit-branch", "main", "The branch name in Gerrit repository")
}

type gerritHooker struct {
}

func (hooker *gerritHooker) handler() (func(r *http.Request) Response, error) {
	return func(r *http.Request) Response {
		var message payload.GerritEvent
		err := json.NewDecoder(r.Body).Decode(&message)
		if err != nil {
			return Response{
				httpCode: http.StatusInternalServerError,
				detail:   fmt.Sprintf("Failed to decode request body: %q", err),
			}
		}

		log.Printf("gerrit event received: %s\n", message.Type)

		if message.Type != payload.GerritEventChangeMerged {
			return Response{
				httpCode: http.StatusAccepted,
				detail:   fmt.Sprintf("Skip %s event", message.Type),
			}
		}

		if message.Change.Project != gerritProject || message.Change.Branch != gerritProjectBranch {
			log.Printf("ignore event as the branch or project doesn't match, expect %s:%s but got %s:%s", gerritProject, gerritProjectBranch, message.Change.Project, message.Change.Branch)
			return Response{
				httpCode: http.StatusAccepted,
				detail:   fmt.Sprintf("Skip the message for %s branch in %s project", message.Change.Branch, message.Change.Project),
			}
		}

		return Response{
			httpCode: http.StatusOK,
			payload:  message,
		}
	}, nil
}
