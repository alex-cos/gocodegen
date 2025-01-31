package main

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

func loadPackage() *packages.Package {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes |
			packages.NeedTypesSizes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		panic(err)
	}

	if len(pkgs) != 1 {
		panic(fmt.Errorf("unexpected number of packages '%d'", len(pkgs)))
	}
	return pkgs[0]
}
