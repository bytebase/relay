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

func (hooker *githubHooker) handler() (func(r *http.Request) (int, interface{}), error) {
	return func(r *http.Request) (int, interface{}) {
		event := r.Header.Get("X-GitHub-Event")
		if event == "ping" {
			return http.StatusAccepted, "Pong"
		}

		if event != "push" {
			return http.StatusBadRequest, "Not a push event"
		}

		var payload payload.GitHubPushEvent
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Failed to decode request body: %v", err)
		}

		if !strings.HasPrefix(payload.Ref, refPrefix) {
			// We don't want to fail the delivery entirely since it would make the webhook
			// look like not working on the GitHub interface.
			return http.StatusAccepted, fmt.Sprintf(`The ref %q does not have the required prefix %q`, payload.Ref, refPrefix)
		}

		return http.StatusOK, payload
	}, nil
}
