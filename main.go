package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytebase/relay/hook"
	"github.com/bytebase/relay/sink"
	"github.com/bytebase/relay/util"
	"github.com/flamego/flamego"
	flag "github.com/spf13/pflag"
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
	address string
)

func init() {
	flag.StringVar(&address, "address", os.Getenv("RELAY_ADDR"), "The host:port address where Relay runs, default to localhost:5678")
}

func main() {
	flag.Parse()

	h := "localhost"
	p := 5678
	var err error
	if address != "" {
		h, p, err = util.ParseHost(address)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			os.Exit(1)
		}
	}

	f := flamego.Classic()
	github := hook.NewGitHub()
	lark := sink.NewLark()
	hook.Mount(f, "/github", github, []sink.Sinker{lark})

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

	f.Run(h, p)

	fmt.Print(byeBanner)

	// Wait for CTRL-C.
	<-ctx.Done()
}
