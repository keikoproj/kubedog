package main

import (
	"log"
	"testing"

	"github.com/cucumber/godog"
	"github.com/keikoproj/kubedog"
	"github.com/keikoproj/kubedog/pkg/generic"
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
	k.KubeClientSet.SetTemplateArguments(getTemplateArguments(t))
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

func getTemplateArguments(t *testing.T) map[string]string {
	templateArguments := []generic.TemplateArgument{
		{
			Key:                 "Namespace",
			EnvironmentVariable: "KUBEDOG_EXAMPLE_NAMESPACE",
			Mandatory:           false,
			Default:             "kubedog-example",
		},
		{
			Key:                 "Image",
			EnvironmentVariable: "KUBEDOG_EXAMPLE_IMAGE",
			Mandatory:           false,
			Default:             "busybox:1.28",
		},
		{
			Key:                 "Message",
			EnvironmentVariable: "KUBEDOG_EXAMPLE_MESSAGE",
			Mandatory:           false,
			Default:             "Hello, Kubedog!",
		},
	}
	args, err := generic.TemplateArgumentsToMap(templateArguments...)
	if err != nil {
		t.Error(err)
	}
	return args
}
