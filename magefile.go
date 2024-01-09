//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

const (
	planReporterApp   = "tf-plan-reporter"
	testPlanReaderApp = "test-plan-reader"
	prefix            = "./cmd"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = BuildMainApp

// A build step of the main App
func BuildMainApp() error {
	fmt.Printf("Building of %s ...", planReporterApp)

	envs := make(map[string]string)
	envs["CGO_ENABLED"] = "0"

	return sh.RunWith(envs, "go", "build", "-o", planReporterApp, fmt.Sprintf("%s/%s", prefix, planReporterApp))

}

// A build step of the test App
func BuildTestApp() error {
	fmt.Printf("Building of %s ...", testPlanReaderApp)

	envs := make(map[string]string)
	envs["CGO_ENABLED"] = "0"

	return sh.RunWith(envs, "go", "build", "-o", planReporterApp, fmt.Sprintf("%s/%s", prefix, planReporterApp))

}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(BuildMainApp)
	fmt.Println("Installing...")
	return os.Rename(fmt.Sprintf("./%s", planReporterApp), fmt.Sprintf("/usr/bin/%s", planReporterApp))
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll(planReporterApp)
	os.RemoveAll(testPlanReaderApp)
}
