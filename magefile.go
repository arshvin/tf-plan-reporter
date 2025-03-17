//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

const (
	mainAppPackagePath = "./cmd/tf-plan-reporter"
	auxAppPackagePath  = "./cmd/test-plan-reader"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = BuildMainApp

// A build target of the Main App (tf-plan-reporter)
func BuildMainApp(ctx context.Context) error {
	mg.Deps(TestMainApp)

	fmt.Printf("Building of '%s' ...\n", path.Base(mainAppPackagePath))

	envs := make(map[string]string)
	if os.Getenv("CGO_ENABLED") != "" {
		envs["CGO_ENABLED"] = os.Getenv("CGO_ENABLED")
		fmt.Printf("CGO_ENABLED environment variable was set up to: %s\n", os.Getenv("CGO_ENABLED"))
	}

	return build(envs, path.Base(mainAppPackagePath), mainAppPackagePath)

}

// A test target of the Main App (tf-plan-reporter)
func TestMainApp(ctx context.Context) error {
	fmt.Printf("Testing of all packages...\n")

	envs := make(map[string]string)
	if os.Getenv("CGO_ENABLED") != "" {
		envs["CGO_ENABLED"] = os.Getenv("CGO_ENABLED")
		fmt.Printf("CGO_ENABLED environment variable was set up to: %s\n", os.Getenv("CGO_ENABLED"))
	}

	return runTest(envs)

}


// A build target of the Auxiliary Test App (test-plan-reader)
func BuildTestApp() error {
	fmt.Printf("Building of '%s' ...\n", path.Base(auxAppPackagePath))

	envs := make(map[string]string)
	if os.Getenv("CGO_ENABLED") != "" {
		envs["CGO_ENABLED"] = os.Getenv("CGO_ENABLED")
		fmt.Printf("CGO_ENABLED environment variable was set up to: %s\n", os.Getenv("CGO_ENABLED"))
	}

	return build(envs, path.Base(auxAppPackagePath), auxAppPackagePath)

}

// Executes golang build command
func build(envs map[string]string, binName, packPath string) error {
	return sh.RunWith(envs, "go", "build", "-o", binName, packPath)
}

// Executes golang test command
func runTest(envs map[string]string) error {
	return sh.RunWith(envs, "go", "test", "./...")
}

// A custom install step to target folder
func Install(targetFolder string) error {
	mg.Deps(BuildMainApp)

	if targetFolder == "" {
		fmt.Errorf("Target folder must not be an empty string")
	}

	fmt.Printf("Installing to %s...\n", targetFolder)
	return os.Rename(path.Base(mainAppPackagePath), fmt.Sprintf("%s/%s", targetFolder,path.Base(mainAppPackagePath)))
}

// Clean up after yourself
func Clean() {
	fmt.Printf("Cleaning of %s...\n", path.Base(mainAppPackagePath))
	os.RemoveAll(path.Base(mainAppPackagePath))
	fmt.Printf("Cleaning of %s...\n", path.Base(auxAppPackagePath))
	os.RemoveAll(path.Base(auxAppPackagePath))
}
