package main

import (
	"log"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/keikoproj/kubedog"
	"github.com/keikoproj/kubedog/examples/usage/generate"
)

var k kubedog.Test

func InitializeScenario(ctx *godog.ScenarioContext) {
	k.SetScenario(ctx)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		if err := k.KubeClientSet.DeleteAllTestResources(); err != nil {
			log.Printf("Failed deleting the test resources: %v", err)
		}
	})
	ctx.AfterSuite(func() {
		if err := k.KubeClientSet.DeleteAllTestResources(); err != nil {
			log.Printf("Failed deleting the test resources: %v", err)
		}
	})
	k.SetTestSuite(ctx)
}

func TestFeatures(t *testing.T) {
	if err := generate.Templates(); err != nil {
		t.Error(err)
	}
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
	os.Exit(status)
}
