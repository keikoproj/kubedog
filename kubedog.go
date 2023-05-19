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

	"github.com/cucumber/godog"
	aws "github.com/keikoproj/kubedog/pkg/aws"
	"github.com/keikoproj/kubedog/pkg/generic"
	"github.com/keikoproj/kubedog/pkg/kube"
	log "github.com/sirupsen/logrus"
)

//go:generate go run generate/syntax/main.go

// TODO: are these the right names for this struct and its fields?
// TODO: make struct unexported and add NewTest func?
type Test struct {
	suiteContext    *godog.TestSuiteContext
	scenarioContext *godog.ScenarioContext
	KubeClient      kube.ClientSet
	AwsClient       aws.Client
}

const (
	testFailedStatus int = 1
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
	//syntax-generation:begin
	//syntax-generation:title-0:Generic steps
	kdt.scenarioContext.Step(`^(?:I )?wait (?:for )?(\d+) (minutes|seconds)$`, generic.WaitFor)
	kdt.scenarioContext.Step(`^the (\S+) command is available$`, generic.CommandExists)
	kdt.scenarioContext.Step(`^I run the (\S+) command with the ([^"]*) args and the command (fails|succeeds)$`, generic.RunCommand)
	//syntax-generation:title-0:Kubernetes steps
	kdt.scenarioContext.Step(`^((?:a )?Kubernetes cluster|(?:there are )?(?:valid )?Kubernetes Credentials)$`, kdt.KubeClient.DiscoverClients)
	kdt.scenarioContext.Step(`^(?:the )?Kubernetes cluster should be (created|deleted|upgraded)$`, kdt.KubeClient.KubernetesClusterShouldBe)
	kdt.scenarioContext.Step(`^(?:I )?store (?:the )?current time as ([^"]*)$`, kdt.KubeClient.SetTimestamp)
	//syntax-generation:title-1:Unstructured Resources
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+)$`, kdt.KubeClient.ResourceOperation)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+) in (?:the )?([^"]*) namespace$`, kdt.KubeClient.ResourceOperationInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+)$`, kdt.KubeClient.ResourcesOperation)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+) in (?:the )?([^"]*) namespace$`, kdt.KubeClient.ResourcesOperationInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+), the operation should (succeed|fail)$`, kdt.KubeClient.ResourceOperationWithResult)
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+) in (?:the )?([^"]*) namespace, the operation should (succeed|fail)$`, kdt.KubeClient.ResourceOperationWithResultInNamespace)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) should be (created|deleted)$`, kdt.KubeClient.ResourceShouldBe)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) (?:should )?converge to selector (\S+)$`, kdt.KubeClient.ResourceShouldConvergeToSelector)
	kdt.scenarioContext.Step(`^(?:the )?resource ([^"]*) condition ([^"]*) should be (true|false)$`, kdt.KubeClient.ResourceConditionShouldBe)
	kdt.scenarioContext.Step(`^(?:I )?update (?:the )?resource ([^"]*) with ([^"]*) set to ([^"]*)$`, kdt.KubeClient.UpdateResourceWithField)
	kdt.scenarioContext.Step(`^(?:I )?verify InstanceGroups (?:are )?in "ready" state$`, kdt.KubeClient.VerifyInstanceGroups)
	//syntax-generation:title-1:Structured Resources
	//syntax-generation:title-2:Pods
	kdt.scenarioContext.Step(`^(?:I )?get (?:the )?pods in namespace ([^"]*)$`, kdt.KubeClient.Pods)
	kdt.scenarioContext.Step(`^(?:I )?get (?:the )?pods in namespace ([^"]*) with selector (\S+)$`, kdt.KubeClient.PodsWithSelector)
	kdt.scenarioContext.Step(`^(?:the )?pods in namespace ([^"]*) with selector (\S+) have restart count less than (\d+)$`, kdt.KubeClient.PodsWithSelectorHaveRestartCountLessThan)
	kdt.scenarioContext.Step(`^(some|all) pods in namespace (\S+) with selector (\S+) have "([^"]*)" in logs since ([^"]*) time$`, kdt.KubeClient.SomeOrAllPodsInNamespaceWithSelectorHaveStringInLogsSinceTime)
	kdt.scenarioContext.Step(`^some pods in namespace (\S+) with selector (\S+) don't have "([^"]*)" in logs since ([^"]*) time$`, kdt.KubeClient.SomePodsInNamespaceWithSelectorDontHaveStringInLogsSinceTime)
	kdt.scenarioContext.Step(`^(?:the )?pods in namespace (\S+) with selector (\S+) have no errors in logs since ([^"]*) time$`, kdt.KubeClient.PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime)
	kdt.scenarioContext.Step(`^(?:the )?pods in namespace (\S+) with selector (\S+) have some errors in logs since ([^"]*) time$`, kdt.KubeClient.PodsInNamespaceWithSelectorHaveSomeErrorsInLogsSinceTime)
	kdt.scenarioContext.Step(`^(?:the )?pods in namespace (\S+) with selector (\S+) should have labels (\S+)$`, kdt.KubeClient.PodsInNamespaceWithSelectorShouldHaveLabels)
	kdt.scenarioContext.Step(`^(?:the )?pod (\S+) in namespace (\S+) should have labels (\S+)$`, kdt.KubeClient.PodInNamespaceShouldHaveLabels)
	//syntax-generation:title-2:Others
	kdt.scenarioContext.Step(`^(?:I )?(create|submit|update) (?:the )?secret (\S+) in namespace (\S+) from (?:environment variable )?(\S+)$`, kdt.KubeClient.SecretOperationFromEnvironmentVariable)
	kdt.scenarioContext.Step(`^(?:I )?delete (?:the )?secret (\S+) in namespace (\S+)$`, kdt.KubeClient.SecretDelete)
	kdt.scenarioContext.Step(`^(\d+) node(?:s)? with selector (\S+) should be (found|ready)$`, kdt.KubeClient.NodesWithSelectorShouldBe)
	kdt.scenarioContext.Step(`^(?:the )?(deployment|hpa|horizontalpodautoscaler|service|pdb|poddisruptionbudget|sa|serviceaccount) ([^"]*) is in namespace ([^"]*)$`, kdt.KubeClient.ResourceInNamespace)
	kdt.scenarioContext.Step(`^(?:I )?scale (?:the )?deployment ([^"]*) in namespace ([^"]*) to (\d+)$`, kdt.KubeClient.ScaleDeployment)
	kdt.scenarioContext.Step(`^(?:I )?validate Prometheus Statefulset ([^"]*) in namespace ([^"]*) has volumeClaimTemplates name ([^"]*)$`, kdt.KubeClient.ValidatePrometheusVolumeClaimTemplatesName)
	kdt.scenarioContext.Step(`^(?:I )?get (?:the )?nodes list$`, kdt.KubeClient.GetNodes)
	kdt.scenarioContext.Step(`^(?:the )?daemonset ([^"]*) is running in namespace ([^"]*)$`, kdt.KubeClient.DaemonSetIsRunning)
	kdt.scenarioContext.Step(`^(?:the )?deployment ([^"]*) is running in namespace ([^"]*)$`, kdt.KubeClient.DeploymentIsRunning)
	kdt.scenarioContext.Step(`^(?:the )?persistentvolume ([^"]*) exists with status (Available|Bound|Released|Failed|Pending)$`, kdt.KubeClient.PersistentVolExists)
	kdt.scenarioContext.Step(`^(?:the )?(clusterrole|clusterrolebinding) with name ([^"]*) should be found$`, kdt.KubeClient.ClusterRbacIsFound)
	kdt.scenarioContext.Step(`^(?:the )?ingress (\S+) in (?:the )?namespace (\S+) (?:is )?(?:available )?on port (\d+) and path ([^"]*)$`, kdt.KubeClient.IngressAvailable)
	kdt.scenarioContext.Step(`^(?:I )?send (\d+) tps to ingress (\S+) in (?:the )?namespace (\S+) (?:available )?on port (\d+) and path ([^"]*) for (\d+) (minutes|seconds) expecting up to (\d+) error(?:s)?$`, kdt.KubeClient.SendTrafficToIngress)
	//syntax-generation:title-0:AWS steps
	kdt.scenarioContext.Step(`^(?:there are )?(?:valid )?AWS Credentials$`, kdt.AwsClient.GetAWSCredsAndClients)
	kdt.scenarioContext.Step(`^an Auto Scaling Group named ([^"]*)$`, kdt.AwsClient.AnASGNamed)
	kdt.scenarioContext.Step(`^(?:I )?update (?:the )?current Auto Scaling Group with ([^"]*) set to ([^"]*)$`, kdt.AwsClient.UpdateFieldOfCurrentASG)
	kdt.scenarioContext.Step(`^(?:the )?current Auto Scaling Group (?:is )?scaled to \(min, max\) = \((\d+), (\d+)\)$`, kdt.AwsClient.ScaleCurrentASG)
	kdt.scenarioContext.Step(`^(?:the )?DNS name (\S+) (should|should not) be created in hostedZoneID (\S+)$`, kdt.AwsClient.DnsNameShouldOrNotInHostedZoneID)
	kdt.scenarioContext.Step(`^(?:I )?(add|remove) (?:the )?(\S+) role as trusted entity to iam role ([^"]*)$`, kdt.AwsClient.IamRoleTrust)
	kdt.scenarioContext.Step(`^(?:I )?(add|remove) cluster shared iam role$`, kdt.AwsClient.ClusterSharedIamOperation)
	//syntax-generation:end
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
