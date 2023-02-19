package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"

	"github.com/bytebase/relay/payload"
	"github.com/bytebase/relay/util"
)

var (
	_ Sinker = (*larkSinker)(nil)
)

var (
	webhookURLs string
)

func init() {
	flag.StringVar(&webhookURLs, "lark-urls", "", "A comma separated list of Lark webhook URLs")
}

// NewLark creates a Lark sinker
func NewLark() Sinker {
	return &larkSinker{}
}

type larkSinker struct {
}

func (sinker *larkSinker) Mount() error {
	if webhookURLs == "" {
		fmt.Printf("--lark-urls is missing, Lark sinker will not be able to process any events.\n")
	}
	return nil
}

func (sinker *larkSinker) Process(c context.Context, path string, pi interface{}) error {
	if webhookURLs == "" {
		return fmt.Errorf("--lark-urls is required")
	}
	switch path {
	case "/github":
		p := pi.(payload.GitHubPushEvent)
		var text string
		if p.Deleted {
			text = fmt.Sprintf("%q has been deleted by %s", p.Ref, p.Sender.Login)
		} else {
			text = fmt.Sprintf(`New commits have been pushed to %q by %s(%s) at %s
Title: %s
Diff: %s`,
				p.Ref, p.HeadCommit.Author.Name, p.HeadCommit.Author.Email, p.HeadCommit.Timestamp,
				p.HeadCommit.Message,
				p.Compare,
			)
		}
		urlList := strings.Split(webhookURLs, ",")
		for _, url := range urlList {
			err := sendToLark(c, url, text)
			if err != nil {
				return fmt.Errorf("failed to send to Lark %q: %w", util.RedactLastN(url, 12), err)
			}
		}
	}
	return nil
}

type larkPayloadContent struct {
	Text string `json:"text"`
}

type larkPayload struct {
	MsgType string             `json:"msg_type"`
	Content larkPayloadContent `json:"content"`
}

func sendToLark(ctx context.Context, url, text string) error {
	payload := &larkPayload{
		MsgType: "text",
		Content: larkPayloadContent{Text: text},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "marshal payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "new request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
		return errors.Errorf("unexpected status code %d with body: %s", resp.StatusCode, body)
	}
	return nil
}
