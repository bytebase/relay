package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bytebase/relay/util"
	"github.com/flamego/flamego"
	"github.com/pkg/errors"
)

func main() {
	refPrefix := flag.String("ref-prefix", "refs/heads/", "The prefix for the ref")
	larkURLs := flag.String("lark-urls", os.Getenv("LARK_URLS"), "A comma separated list of Lark webhook URL")
	flag.Parse()

	if *larkURLs == "" {
		log.Fatal(`The "--lark-urls" or "LARK_URLS" is required`)
	}

	f := flamego.Classic()
	f.Post("/", func(r *http.Request) (int, string) {
		event := r.Header.Get("X-GitHub-Event")
		if event == "ping" {
			return http.StatusAccepted, "Pong"
		}

		if event != "push" {
			return http.StatusBadRequest, "Not a push event"
		}

		var payload struct {
			Ref     string `json:"ref"`
			Compare string `json:"compare"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			return http.StatusInternalServerError, fmt.Sprintf("Failed to decode request body: %v", err)
		}

		if !strings.HasPrefix(payload.Ref, *refPrefix) {
			// We don't want to fail the delivery entirely since it would make the webhook
			// look like not working on the GitHub interface.
			return http.StatusAccepted, fmt.Sprintf(`The ref %q does not have the required prefix %q`, payload.Ref, *refPrefix)
		}

		urlList := strings.Split(*larkURLs, ",")

		for _, url := range urlList {
			err = sendToLark(r.Context(), url, fmt.Sprintf("New commits have been pushed to %q: %s", payload.Ref, payload.Compare))
			if err != nil {
				return http.StatusInternalServerError, fmt.Sprintf("Failed to send to Lark %q: %v", util.RedactLastN(url, 12), err)
			}
		}

		return http.StatusOK, "OK"
	})
	f.Run()
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
