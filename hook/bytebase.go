package hook

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/bytebase/relay/payload"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// NewBytebase creates a Bytebase hooker
func NewBytebase(state *payload.GlobalState) Hooker {
	return &bytebaseHooker{
		state: state,
	}
}

type bytebaseHooker struct {
	state *payload.GlobalState
}

func getIssueID(path string) string {
	pathes := strings.Split(path, "/approval/")
	if len(pathes) != 2 {
		return ""
	}
	return pathes[1]
}

func (hooker *bytebaseHooker) handler() (func(r *http.Request) Response, error) {
	return func(r *http.Request) Response {
		switch r.Method {
		case http.MethodPost:
			issue := &payload.BytebaseIssue{}
			err := json.NewDecoder(r.Body).Decode(issue)
			if err != nil {
				return Response{
					httpCode: http.StatusInternalServerError,
					detail:   fmt.Sprintf("Failed to decode request body: %q", err),
				}
			}

			id := randStringRunes(10)
			hooker.state.IssueList[id] = &payload.Issue{
				ID:            id,
				Status:        "PENDING",
				BytebaseIssue: issue,
			}
			return Response{
				httpCode: http.StatusOK,
				payload:  hooker.state.IssueList[id],
			}
		case http.MethodPatch:
			id := getIssueID(r.RequestURI)
			existed, ok := hooker.state.IssueList[id]
			if !ok {
				return Response{
					httpCode: http.StatusNotFound,
					detail:   fmt.Sprintf("Cannot found issue %s", id),
				}
			}
			issue := &payload.BytebaseIssue{}
			err := json.NewDecoder(r.Body).Decode(issue)
			if err != nil {
				return Response{
					httpCode: http.StatusInternalServerError,
					detail:   fmt.Sprintf("Failed to decode request body: %q", err),
				}
			}
			hooker.state.IssueList[id] = &payload.Issue{
				ID:            id,
				Status:        existed.Status,
				BytebaseIssue: issue,
			}
			return Response{
				httpCode: http.StatusOK,
				payload:  hooker.state.IssueList[id],
			}
		case http.MethodGet:
			id := getIssueID(r.RequestURI)
			issue, ok := hooker.state.IssueList[id]
			if !ok {
				return Response{
					httpCode: http.StatusNotFound,
					detail:   fmt.Sprintf("Cannot found issue %s", id),
				}
			}
			return Response{
				httpCode: http.StatusOK,
				payload:  issue,
			}
		default:
			return Response{
				httpCode: http.StatusBadRequest,
				detail:   "Unaccept request",
			}
		}
	}, nil
}
