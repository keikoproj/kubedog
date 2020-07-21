package main

import (
	"github.com/cucumber/godog"
	kdog "github.com/keikoproj/kubedog/examples"
)

var t kdog.Test

// godog v0.10.0 (latest)
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		t.Godogs = &Godogs
		Godogs = 0
	})
	t.SetTestSuite(ctx)
}
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(func(*godog.Scenario) {
		Godogs = 0
	})
	ctx.Step(`^I buy (\d+) more$`, iBuyMore)
	t.SetScenario(ctx)
	t.Run()
}

func iBuyMore(arg1 int) error {
	*t.Godogs += arg1
	return nil
}
