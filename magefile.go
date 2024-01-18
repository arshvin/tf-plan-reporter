//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

const (
	mainApp = "./cmd/tf-plan-reporter"
	auxApp  = "./cmd/test-plan-reader"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = BuildMainApp

// A build step of the main App
func BuildMainApp() error {
	fmt.Printf("Building of '%s' ...\n", path.Base(mainApp))

	envs := make(map[string]string)
	envs["CGO_ENABLED"] = "0"

	return build(envs, path.Base(mainApp), mainApp)

}

// A build step of the test App
func BuildTestApp() error {
	fmt.Printf("Building of '%s' ...\n", path.Base(auxApp))

	envs := make(map[string]string)
	envs["CGO_ENABLED"] = "0"

	return build(envs, path.Base(auxApp), auxApp)

}

// Executes golang build command
func build(envs map[string]string, binName, packPath string) error {
	return sh.RunWith(envs, "go", "build", "-o", binName, packPath)
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(BuildMainApp)
	fmt.Println("Installing...")
	return os.Rename(mainApp, fmt.Sprintf("/usr/bin/%s", mainApp))
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll(mainApp)
	os.RemoveAll(auxApp)
}
