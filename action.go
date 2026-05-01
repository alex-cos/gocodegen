package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

func action(c context.Context, cmd *cli.Command) error {
	output := cmd.String("output")

	pkg, err := loadPackage()
	if err != nil {
		return err
	}

	pack := parsePackage(pkg)

	generated, err := generate(output, pack)
	if err != nil {
		return err
	}
	for _, gen := range generated {
		file, err := os.Create(gen.Name)
		if err != nil {
			return err
		}
		defer file.Close()
		err = writeBuffer(file, gen.Buf)
		if err != nil {
			return err
		}
	}

	return nil
}
