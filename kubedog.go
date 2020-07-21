package kubedog

import (
	"fmt"
	"os"

	"github.com/cucumber/godog"
	aws "github.com/keikoproj/kubedog/pkg/aws"
	kube "github.com/keikoproj/kubedog/pkg/kubernetes"
)

type Test struct {
	suiteContext    *godog.TestSuiteContext
	scenarioContext *godog.ScenarioContext
	KubeContext     kube.Client
	AwsContext      aws.Client
}

const (
	testSucceededStatus int = 0
	testFailedStatus    int = 1
)

func (kdt *Test) Run() {

	// TODO: define default suite hooks if any, check that the suite context was set

	if kdt.scenarioContext == nil {
		fmt.Println("FATAL: kubedog.Test.scenarioContext was not set, use kubedog.Test.InitScenario")
		os.Exit(testFailedStatus)
	}

	// TODO: define default scenario hooks if any
	// TODO: define default step hooks if any

	// Kubernetes related steps
	kdt.scenarioContext.Step(`^a Kubernetes cluster$`, kdt.KubeContext.AKubernetesCluster)
	kdt.scenarioContext.Step(`^I (create|submit|delete) the custom resource ([^"]*)$`, kdt.KubeContext.ResourceOperation)
	kdt.scenarioContext.Step(`^the custom resource ([^"]*) should be (created|deleted)$`, kdt.KubeContext.ResourceShouldBe)
	kdt.scenarioContext.Step(`^the custom resource ([^"]*) should converge to selector ([^"]*)$`, kdt.KubeContext.ResourceShouldConvergeToSelector)
	kdt.scenarioContext.Step(`^the custom resource ([^"]*) converge to selector ([^"]*)$`, kdt.KubeContext.ResourceShouldConvergeToSelector)
	kdt.scenarioContext.Step(`^the custom resource ([^"]*) condition ([^"]*) should be (true|false)$`, kdt.KubeContext.ResourceConditionShouldBe)
	kdt.scenarioContext.Step(`^I update a custom resource ([^"]*) with ([^"]*) set to ([^"]*)$`, kdt.KubeContext.UpdateResourceWithField)
	kdt.scenarioContext.Step(`^(\d+) nodes with selector ([^"]*) should be (found|ready)$`, kdt.KubeContext.NodesWithSelectorShouldBe)
	// AWS related steps
	kdt.scenarioContext.Step(`^valid AWS Credentials$`, kdt.AwsContext.GetAWSCredsAndClients)
	kdt.scenarioContext.Step(`^an Auto Scaling Group named ([^"]*)$`, kdt.AwsContext.AnASGNamed)
	kdt.scenarioContext.Step(`^I update the current Auto Scaling Group with ([^"]*) set to ([^"]*)$`, kdt.AwsContext.UpdateFieldOfCurrentASG)
	kdt.scenarioContext.Step(`the current Auto Scaling Group is scaled to \(min, max\) = \((\d+), (\d+)\)$`, kdt.AwsContext.ScaleCurrentASG)
}

func (kdt *Test) SetTestSuite(testSuite *godog.TestSuiteContext) {

	kdt.suiteContext = testSuite
}

func (kdt *Test) SetScenario(scenario *godog.ScenarioContext) {

	kdt.scenarioContext = scenario
}
