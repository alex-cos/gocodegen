package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
)

// parseCLI parses command lines arguments.
func parseCLI() error {
	cliapp := cli.NewApp()
	cliapp.Name = "gocodegen"
	cliapp.Usage = "Command line that generate go code by analyzing go code comments from package files"
	cliapp.UsageText = "gocodegen [options]"
	cliapp.Description = "Build: " + buildDate
	cliapp.Version = version

	cliapp.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(os.Stderr, "Error. Unknown command: '%s'\n\n", command)
		cli.ShowAppHelpAndExit(c, 1)
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintln(os.Stdout, "Version:\t", c.App.Version)
	}

	cliapp.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Value:       "",
			Usage:       "Output file name",
			Aliases:     []string{"o"},
			Required:    false,
			DefaultText: fmt.Sprintf("Default value is '%s'", "resources"),
		},
	}

	cliapp.Action = action

	sort.Sort(cli.FlagsByName(cliapp.Flags))
	sort.Sort(cli.CommandsByName(cliapp.Commands))

	if err := cliapp.Run(os.Args); err != nil {
		return fmt.Errorf("failed to parse command line arguments: %w", err)
	}
	return nil
}
