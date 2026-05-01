//nolint:testpackage
package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePackage(t *testing.T) {
	t.Parallel()

	pkg, err := loadPackage()
	assert.NoError(t, err)

	pack := parsePackage(pkg)
	fmt.Printf("pack = %v\n", pack)

	generated, err := generate("", pack)
	assert.NoError(t, err)

	fmt.Printf("generated = %v\n", generated)
	assert.NoError(t, err)

	generated, err = generate("test", pack)
	assert.NoError(t, err)

	fmt.Printf("generated = %v\n", generated)
	assert.NoError(t, err)
}
