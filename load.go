package main

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

func loadPackage() (*packages.Package, error) {
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
		return nil, err
	}

	if len(pkgs) != 1 {
		err := fmt.Errorf("unexpected number of packages '%d'", len(pkgs))
		return nil, err
	}
	return pkgs[0], nil
}
