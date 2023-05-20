package main

import (
	"log"
	"testing"

	"github.com/cucumber/godog"
	"github.com/keikoproj/kubedog"
)

var k kubedog.Test

// Required for 'go test'
func TestFeatures(t *testing.T) {
	// Setting Godog and running test
	status := godog.TestSuite{
		Name:                 "kubedog-example",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}.Run()
	if status != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

// Required for godog
func InitializeScenario(ctx *godog.ScenarioContext) {
	// Required for Kubedog
	k.SetScenario(ctx)
}

// Required for Godog
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	// Optional: recommended hook
	ctx.BeforeSuite(func() {
		if err := k.KubeClientSet.DeleteAllTestResources(); err != nil {
			log.Printf("Failed deleting the test resources: %v\n\n", err)
		}
	})
	// Optional: recommended hook
	ctx.AfterSuite(func() {
		if err := k.KubeClientSet.DeleteAllTestResources(); err != nil {
			log.Printf("Failed deleting the test resources: %v\n\n", err)
		}
	})
	// Required for Kubedog
	k.SetTestSuite(ctx)
}
