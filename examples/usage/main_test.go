package main

import (
	"log"
	"testing"

	"github.com/cucumber/godog"
	"github.com/keikoproj/kubedog"
)

var k kubedog.Test

func TestFeatures(t *testing.T) {
	status := godog.TestSuite{
		Name:                 "godogs",
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

func InitializeScenario(ctx *godog.ScenarioContext) {
	k.SetScenario(ctx)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		if err := k.KubeClientSet.DeleteAllTestResources(); err != nil {
			log.Printf("Failed deleting the test resources: %v\n\n", err)
		}
	})
	ctx.AfterSuite(func() {
		if err := k.KubeClientSet.DeleteAllTestResources(); err != nil {
			log.Printf("Failed deleting the test resources: %v\n\n", err)
		}
	})
	k.SetTestSuite(ctx)
}
