// Package cmd is the command surface of Bytebase bb tool provided by bytebase.com.
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytebase/relay/hook"
	"github.com/bytebase/relay/sink"
	"github.com/flamego/flamego"
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

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	f := flamego.Classic()
	rootCmd := &cobra.Command{
		Use:   "relay",
		Short: "A webhook relay service provided by bytebase.com",
		Run: func(cmd *cobra.Command, args []string) {
			start(f)
		},
	}

	fs := rootCmd.PersistentFlags()
	github := hook.NewGitHub()
	hook.Register(fs, f, github, "/github")
	lark := sink.NewLark()
	sink.Register(fs, f, lark, "/github")

	return rootCmd
}

// Execute is the execute command for root command.
func Execute() error {
	return NewRootCmd().Execute()
}

func start(f *flamego.Flame) {
	if err := hook.Prepare(); err != nil {
		log.Fatalf("Failed to prepare hook: %v", err)
	}
	if err := sink.Prepare(); err != nil {
		log.Fatalf("Failed to prepare sink: %v", err)
	}
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
