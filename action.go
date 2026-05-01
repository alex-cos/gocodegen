package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func action(c *cli.Context) error {
	output := c.String("output")

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
