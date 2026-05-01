package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/urfave/cli/v3"
)

// parseCLI parses command lines arguments.
func parseCLI() error {
	appcmd := &cli.Command{
		Name:        "gocodegen",
		Usage:       "Command line that generate go code by analyzing go code comments from package files",
		UsageText:   "gocodegen [options]",
		Description: "Build: " + buildDate,
		Version:     version,
		CommandNotFound: func(c context.Context, cmd *cli.Command, name string) {
			fmt.Fprintf(os.Stderr, "Error. Unknown command: '%s'\n\n", name)
			cli.ShowAppHelpAndExit(cmd, 1)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "output",
				Value:       "",
				Usage:       "Output file name",
				Aliases:     []string{"o"},
				Required:    false,
				DefaultText: fmt.Sprintf("Default value is '%s'", "resources"),
			},
		},
	}

	cli.VersionPrinter = func(cmd *cli.Command) {
		fmt.Fprintln(os.Stdout, "Version:\t", cmd.Version)
	}

	appcmd.Action = action

	sort.Sort(cli.FlagsByName(appcmd.Flags))
	sort.Slice(appcmd.Commands, func(i, j int) bool {
		return appcmd.Commands[i].Name < appcmd.Commands[j].Name
	})

	if err := appcmd.Run(context.Background(), os.Args); err != nil {
		return fmt.Errorf("failed to parse command line arguments: %w", err)
	}

	return nil
}
