/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubedog

import (
	"os"
	"time"

	"github.com/cucumber/godog"
	aws "github.com/keikoproj/kubedog/pkg/aws"
	kube "github.com/keikoproj/kubedog/pkg/kubernetes"
	log "github.com/sirupsen/logrus"
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

/*
Run contains the steps definition, should be called in the InitializeScenario function required by godog.
Check https://github.com/keikoproj/kubedog/blob/master/docs/syntax.md for steps syntax details.
*/
func (kdt *Test) Run() {
	if kdt.scenarioContext == nil {
		log.Fatalln("kubedog.Test.scenarioContext was not set, use kubedog.Test.InitScenario")
		os.Exit(testFailedStatus)
	}

	// Kubernetes related steps
	kdt.scenarioContext.Step(`^a Kubernetes cluster$`, kdt.KubeContext.AKubernetesCluster)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+)$`, kdt.KubeContext.ResourceOperation)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+) in (?:the )?([^"]*) namespace`, kdt.KubeContext.ResourceOperationInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+)$`, kdt.KubeContext.MultiResourceOperation)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+) in (?:the )?([^"]*) namespace`, kdt.KubeContext.MultiResourceOperationInNamespace)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) should be (created|deleted)$`, kdt.KubeContext.ResourceShouldBe)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) converged to selector ([^"]*)$`, kdt.KubeContext.ResourceShouldConvergeToSelector)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) should converge to selector ([^"]*)$`, kdt.KubeContext.ResourceShouldConvergeToSelector)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) condition ([^"]*) should be (true|false)$`, kdt.KubeContext.ResourceConditionShouldBe)
	kdt.scenarioContext.Step(`^(?:I )?update (?:a )?resource ([^"]*) with ([^"]*) set to ([^"]*)$`, kdt.KubeContext.UpdateResourceWithField)
	kdt.scenarioContext.Step(`^(\d+) node\(s\) with selector ([^"]*) should be (found|ready)$`, kdt.KubeContext.NodesWithSelectorShouldBe)
	kdt.scenarioContext.Step(`^(?:the )?(deployment|hpa|horizontalpodautoscaler|service|pdb|poddisruptionbudget|sa|serviceaccount) ([^"]*) is in namespace ([^"]*)$`, kdt.KubeContext.ResourceInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?scale (?:the )?deployment ([^"]*) in namespace ([^"]*) to (\d+)$`, kdt.KubeContext.ScaleDeployment)
	kdt.scenarioContext.Step(`^(?:the )?(clusterrole|clusterrolebinding) with name ([^"]*) should be found`, kdt.KubeContext.ClusterRbacIsFound)
	// AWS related steps
	kdt.scenarioContext.Step(`^valid AWS Credentials$`, kdt.AwsContext.GetAWSCredsAndClients)
	kdt.scenarioContext.Step(`^an Auto Scaling Group named ([^"]*)$`, kdt.AwsContext.AnASGNamed)
	kdt.scenarioContext.Step(`^(?:I )?update (?:the )current Auto Scaling Group with ([^"]*) set to ([^"]*)$`, kdt.AwsContext.UpdateFieldOfCurrentASG)
	kdt.scenarioContext.Step(`(?:the )?current Auto Scaling Group (?:is )scaled to \(min, max\) = \((\d+), (\d+)\)$`, kdt.AwsContext.ScaleCurrentASG)
	kdt.scenarioContext.Step(`^(?:I )?print (?:the )?nodes list$`, kdt.KubeContext.PrintNodes)
	kdt.scenarioContext.Step(`^(?:I )?print (?:the )?pods in namespace ([^"]*) with selector ([^"]*)$`, kdt.KubeContext.PrintPodsWithSelector)
	kdt.scenarioContext.Step(`^(?:I )?print (?:the )?pods in namespace ([^"]*)$`, kdt.KubeContext.PrintPods)
	kdt.scenarioContext.Step(`^(?:the )?deployment ([^"]*) is running in namespace ([^"]*)$`, kdt.KubeContext.DeploymentIsRunning)
	kdt.scenarioContext.Step(`^(?:the )?daemonset ([^"]*) is running in namespace ([^"]*)$`, kdt.KubeContext.DaemonsetIsRunning)
	kdt.scenarioContext.Step(`^(?:the )?persistentvolume ([^"]*) exists with status (Available|Bound|Released|Failed|Pending)$`, kdt.KubeContext.PersistentVolExists)
}

/*
SetTestSuite sets the TestSuiteContext, should be use in the InitializeTestSuite function required by godog.
*/
func (kdt *Test) SetTestSuite(testSuite *godog.TestSuiteContext) {

	kdt.suiteContext = testSuite
}

/*
SetScenario sets the ScenarioContext, should be use in the InitializeScenario function required by godog.
*/
func (kdt *Test) SetScenario(scenario *godog.ScenarioContext) {

	kdt.scenarioContext = scenario
}

/*
SetTestFilesPath sets the path for the test files. If SetTestFilesPath was not used, the methods that operate with/on files will look for them in ./templates by default.
*/
func (kdt *Test) SetTestFilesPath(testFilesPath string) {
	kdt.KubeContext.FilesPath = testFilesPath
}

func (kdt *Test) SetWaiterInterval(duration time.Duration) {
	kdt.KubeContext.WaiterInterval = duration
}

func (kdt *Test) SetWaiterTries(tries int) {
	kdt.KubeContext.WaiterTries = tries
}
