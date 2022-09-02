package main

import (
	"fmt"
	"os"

	"github.com/bytebase/relay/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
