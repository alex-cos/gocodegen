package main

import (
	"fmt"
	"os"
)

var (
	version   = "unknown"
	buildDate = "unknown"
)

func main() {
	if err := parseCLI(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
