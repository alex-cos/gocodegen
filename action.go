package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func action(c *cli.Context) error {
	output := c.String("output")

	pkg := loadPackage()

	pack := parsePackage(pkg)

	generated, err := generate(output, pack)
	if err != nil {
		panic(err)
	}
	for _, gen := range generated {
		file, err := os.Create(gen.Name)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		err = writeBuffer(file, gen.Buf)
		if err != nil {
			panic(err)
		}
	}

	return nil
}
