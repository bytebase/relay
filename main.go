package main

import (
	"os"

	"github.com/bytebase/relay/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
