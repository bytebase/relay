// Package cmd is the command surface of Bytebase bb tool provided by bytebase.com.
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bytebase/relay/util"
	"github.com/flamego/flamego"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (

	// greetingBanner is the greeting banner.
	// http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=Relay
	greetingBanner = `
██████╗ ███████╗██╗      █████╗ ██╗   ██╗
██╔══██╗██╔════╝██║     ██╔══██╗╚██╗ ██╔╝
██████╔╝█████╗  ██║     ███████║ ╚████╔╝
██╔══██╗██╔══╝  ██║     ██╔══██║  ╚██╔╝
██║  ██║███████╗███████╗██║  ██║   ██║
╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝   ╚═╝
___________________________________________________________________________________________

`
	// byeBanner is the bye banner.
	// http://patorjk.com/software/taag/#p=display&f=ANSI%20Shadow&t=BYE
	byeBanner = `
██████╗ ██╗   ██╗███████╗
██╔══██╗╚██╗ ██╔╝██╔════╝
██████╔╝ ╚████╔╝ █████╗
██╔══██╗  ╚██╔╝  ██╔══╝
██████╔╝   ██║   ███████╗
╚═════╝    ╚═╝   ╚══════╝

`
)

var (
	flags struct {
		// GitHub
		githubRefPrefix string

		// Lark
		larkWebbookURLs string
	}
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "relay",
		Short: "A webhook relay service provided by bytebase.com",
		Run: func(cmd *cobra.Command, args []string) {
			start()
		},
	}

	rootCmd.PersistentFlags().StringVar(&flags.githubRefPrefix, "github-ref-prefix", "refs/heads/", "The prefix for the GitHub ref")
	rootCmd.PersistentFlags().StringVar(&flags.larkWebbookURLs, "lark-urls", os.Getenv("LARK_URLS"), "A comma separated list of Lark webhook URLs")

	return rootCmd
}

// Execute is the execute command for root command.
func Execute() error {
	return NewRootCmd().Execute()
}

func start() {
	if flags.larkWebbookURLs == "" {
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

		if !strings.HasPrefix(payload.Ref, flags.githubRefPrefix) {
			// We don't want to fail the delivery entirely since it would make the webhook
			// look like not working on the GitHub interface.
			return http.StatusAccepted, fmt.Sprintf(`The ref %q does not have the required prefix %q`, payload.Ref, flags.githubRefPrefix)
		}

		urlList := strings.Split(flags.larkWebbookURLs, ",")

		for _, url := range urlList {
			err = sendToLark(r.Context(), url, fmt.Sprintf("New commits have been pushed to %q: %s", payload.Ref, payload.Compare))
			if err != nil {
				return http.StatusInternalServerError, fmt.Sprintf("Failed to send to Lark %q: %v", util.RedactLastN(url, 12), err)
			}
		}

		return http.StatusOK, "OK"
	})

	// Setup signal handlers.
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	// Trigger graceful shutdown on SIGINT or SIGTERM.
	// The default signal sent by the `kill` command is SIGTERM,
	// which is taken as the graceful shutdown signal for many systems, eg., Kubernetes, Gunicorn.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		f.Stop()
		cancel()
	}()

	fmt.Print(greetingBanner)

	f.Run()

	fmt.Print(byeBanner)

	// Wait for CTRL-C.
	<-ctx.Done()
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
