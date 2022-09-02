package hook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytebase/relay/payload"
	"github.com/bytebase/relay/sink"
	"github.com/flamego/flamego"
	flag "github.com/spf13/pflag"
)

var (
	_ Hooker = (*hooker)(nil)
)

var (
	refPrefix string
)

// NewGitHub creates a GitHub hooker
func NewGitHub() Hooker {
	return &hooker{}
}

type hooker struct {
}

func (hooker *hooker) register(fs *flag.FlagSet, f *flamego.Flame, path string) {
	fs.StringVar(&refPrefix, "github-ref-prefix", "refs/heads/", "The prefix for the GitHub ref")

	f.Post(path, func(r *http.Request) (int, string) {
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

		if err := sink.Process(r.Context(), path, payload); err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Encountered error send to sink %q: %v", path, err)
		}

		return http.StatusOK, "OK"
	})
}

func (hooker *hooker) prepare() error {
	return nil
}
