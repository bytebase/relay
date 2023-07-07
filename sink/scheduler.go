package sink

import (
	"context"
	"time"

	"github.com/bytebase/relay/payload"
)

var (
	_ Sinker = (*schedulerSinker)(nil)
)

// NewScheduler creates a scheduler sinker
func NewScheduler(state *payload.GlobalState) Sinker {
	return &schedulerSinker{state: state}
}

type schedulerSinker struct {
	state *payload.GlobalState
}

func (sinker *schedulerSinker) Mount() error {
	return nil
}

func (sinker *schedulerSinker) Process(c context.Context, _ string, pi interface{}) error {
	issue := pi.(*payload.Issue)
	go func(id string) {
		issue, ok := sinker.state.IssueList[id]
		if ok {
			time.Sleep(time.Duration(5) * time.Second)
			issue.Status = "APPROVED"
		}
	}(issue.ID)
	return nil
}
