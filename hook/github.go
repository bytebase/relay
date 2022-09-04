package hook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/relay/payload"
	flag "github.com/spf13/pflag"
)

var (
	_ Hooker = (*githubHooker)(nil)
)

var (
	refPrefix string
)

func init() {
	flag.StringVar(&refPrefix, "github-ref-prefix", "refs/heads/", "The prefix for the GitHub ref")
}

// NewGitHub creates a GitHub hooker
func NewGitHub() Hooker {
	return &githubHooker{}
}

type githubHooker struct {
}

func (hooker *githubHooker) handler() (func(r *http.Request) Response, error) {
	return func(r *http.Request) Response {
		event := r.Header.Get("X-GitHub-Event")
		if event == "ping" {
			return Response{
				httpCode: http.StatusAccepted,
				detail:   "Pong",
			}
		}

		if event != "push" {
			return Response{
				httpCode: http.StatusBadRequest,
				detail:   "Not a push event",
			}
		}

		var payload payload.GitHubPushEvent
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return Response{
				httpCode: http.StatusInternalServerError,
				err:      fmt.Errorf("failed to decode request body: %w", err),
			}
		}

		if !strings.HasPrefix(payload.Ref, refPrefix) {
			// We don't want to fail the delivery entirely since it would make the webhook
			// look like not working on the GitHub interface.
			return Response{
				httpCode: http.StatusAccepted,
				detail:   fmt.Sprintf(`The ref %q does not have the required prefix %q`, payload.Ref, refPrefix),
			}
		}

		return Response{
			httpCode: http.StatusOK,
			payload:  payload,
		}
	}, nil
}
