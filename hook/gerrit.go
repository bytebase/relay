package hook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bytebase/relay/payload"
)

var (
	_ Hooker = (*gerritHooker)(nil)
)

// NewGerrit creates a Gerrit hooker
func NewGerrit() Hooker {
	return &gerritHooker{}
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

		return Response{
			httpCode: http.StatusOK,
			payload:  message,
		}
	}, nil
}
