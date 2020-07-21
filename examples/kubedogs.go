package example

import (
	"fmt"
	"os"

	"github.com/cucumber/godog"
)

type Test struct {
	Godogs          *int // For godogs example
	suiteContext    *godog.TestSuiteContext
	scenarioContext *godog.ScenarioContext
}

const (
	testSucceededStatus int = 0
	testFailedStatus    int = 1
)

func (kdt Test) Run() {

	if kdt.scenarioContext == nil {
		fmt.Println("FATAL: kubedog.Test.scenarioContext was not set, use kubedog.Test.InitScenario")
		os.Exit(testFailedStatus)
	}

	kdt.scenarioContext.Step(`^there are (\d+) godogs$`, kdt.thereAreGodogs)
	kdt.scenarioContext.Step(`^I eat (\d+)$`, kdt.iEat)
	kdt.scenarioContext.Step(`^there should be (\d+) remaining$`, kdt.thereShouldBeRemaining)
}

func (kdt *Test) SetTestSuite(testSuite *godog.TestSuiteContext) {

	kdt.suiteContext = testSuite
}

func (kdt *Test) SetScenario(scenario *godog.ScenarioContext) {

	kdt.scenarioContext = scenario
}

func (kdt *Test) thereAreGodogs(available int) error {

	if kdt.Godogs == nil {
		return fmt.Errorf("The godogs pointer was not set, set it in the before suit hook")
	}

	*kdt.Godogs = available
	return nil
}

func (kdt *Test) iEat(num int) error {

	if kdt.Godogs == nil {
		return fmt.Errorf("The godogs pointer was not set, set it in the before suit hook")
	}

	if *kdt.Godogs < num {
		return fmt.Errorf("you cannot eat %d godogs, there are %d available", num, *kdt.Godogs)
	}
	*kdt.Godogs -= num
	return nil
}

func (kdt *Test) thereShouldBeRemaining(remaining int) error {

	if kdt.Godogs == nil {
		return fmt.Errorf("The godogs pointer was not set, set it in the before suit hook")
	}

	if *kdt.Godogs != remaining {
		return fmt.Errorf("expected %d godogs to be remaining, but there is %d", remaining, *kdt.Godogs)
	}
	return nil
}
