package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/bytebase/relay/payload"
	"github.com/bytebase/relay/util"
	"github.com/pkg/errors"
	flag "github.com/spf13/pflag"
)

var (
	_ Sinker = (*sinker)(nil)
)

var (
	webhookURLs string
)

// NewLark creates a Lark sinker
func NewLark() Sinker {
	return &sinker{}
}

type sinker struct {
}

func (sinker *sinker) register(fs *flag.FlagSet) {
	fs.StringVar(&webhookURLs, "lark-urls", os.Getenv("LARK_URLS"), "A comma separated list of Lark webhook URLs")
}

func (sink *sinker) prepare() error {
	if webhookURLs == "" {
		return fmt.Errorf(`the "--lark-urls" or "LARK_URLS" is required`)
	}
	return nil
}

func (sinker *sinker) process(c context.Context, path string, pi interface{}) error {
	switch path {
	case "/github":
		p := pi.(payload.GitHubPushEvent)
		urlList := strings.Split(webhookURLs, ",")
		for _, url := range urlList {
			err := sendToLark(c, url, fmt.Sprintf("New commits have been pushed to %q: %s", p.Ref, p.Compare))
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
