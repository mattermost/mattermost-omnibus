// +build mage

package main

import (
	"github.com/magefile/mage/sh"
)

// Builds the test binary
func Build() error {
	return sh.RunV("go", "test", "-c")
}

// Updates the vendor folder
func Vendor() error {
	err := sh.RunV("go", "mod", "vendor")
	if err != nil {
		return err
	}

	return sh.RunV("go", "mod", "tidy")
}

// Format the golang source
func Fmt() error {
	return sh.RunV("go", "fmt", "./...")
}
