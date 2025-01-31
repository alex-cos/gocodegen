//nolint:testpackage
package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePackage(t *testing.T) {
	t.Parallel()

	pkg := loadPackage()

	pack := parsePackage(pkg)
	fmt.Printf("pack = %v\n", pack)

	generated, err := generate("", pack)
	if err != nil {
		panic(err)
	}
	fmt.Printf("generated = %v\n", generated)
	assert.NoError(t, err)

	generated, err = generate("test", pack)
	if err != nil {
		panic(err)
	}
	fmt.Printf("generated = %v\n", generated)
	assert.NoError(t, err)
}
