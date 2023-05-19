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

//go:generate go run generate/syntax/main.go
import (
	"github.com/cucumber/godog"
	aws "github.com/keikoproj/kubedog/pkg/aws"
	"github.com/keikoproj/kubedog/pkg/generic"
	"github.com/keikoproj/kubedog/pkg/kube"
)

type Test struct {
	suite         *godog.TestSuiteContext
	scenario      *godog.ScenarioContext
	KubeClientSet kube.ClientSet
	AwsClientSet  aws.ClientSet
}

/*
SetScenario sets the ScenarioContext and contains the steps definition, should be called in the InitializeScenario function required by godog.
Check https://github.com/keikoproj/kubedog/blob/master/docs/syntax.md for steps syntax details.
*/
func (kdt *Test) SetScenario(scenario *godog.ScenarioContext) {
	kdt.scenario = scenario
	//syntax-generation:begin
	//syntax-generation:title-0:Generic steps
	kdt.scenario.Step(`^(?:I )?wait (?:for )?(\d+) (minutes|seconds)$`, generic.WaitFor)
	kdt.scenario.Step(`^the (\S+) command is available$`, generic.CommandExists)
	kdt.scenario.Step(`^I run the (\S+) command with the ([^"]*) args and the command (fails|succeeds)$`, generic.RunCommand)
	//syntax-generation:title-0:Kubernetes steps
	kdt.scenario.Step(`^((?:a )?Kubernetes cluster|(?:there are )?(?:valid )?Kubernetes Credentials)$`, kdt.KubeClientSet.DiscoverClients)
	kdt.scenario.Step(`^(?:the )?Kubernetes cluster should be (created|deleted|upgraded)$`, kdt.KubeClientSet.KubernetesClusterShouldBe)
	kdt.scenario.Step(`^(?:I )?store (?:the )?current time as ([^"]*)$`, kdt.KubeClientSet.SetTimestamp)
	//syntax-generation:title-1:Unstructured Resources
	kdt.scenario.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+)$`, kdt.KubeClientSet.ResourceOperation)
	kdt.scenario.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+) in (?:the )?([^"]*) namespace$`, kdt.KubeClientSet.ResourceOperationInNamespace)
	kdt.scenario.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+)$`, kdt.KubeClientSet.ResourcesOperation)
	kdt.scenario.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resources in (\S+) in (?:the )?([^"]*) namespace$`, kdt.KubeClientSet.ResourcesOperationInNamespace)
	kdt.scenario.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+), the operation should (succeed|fail)$`, kdt.KubeClientSet.ResourceOperationWithResult)
	kdt.scenario.Step(`^(?:I )?(create|submit|delete|update) (?:the )?resource (\S+) in (?:the )?([^"]*) namespace, the operation should (succeed|fail)$`, kdt.KubeClientSet.ResourceOperationWithResultInNamespace)
	kdt.scenario.Step(`^(?:the )?resource ([^"]*) should be (created|deleted)$`, kdt.KubeClientSet.ResourceShouldBe)
	kdt.scenario.Step(`^(?:the )?resource ([^"]*) (?:should )?converge to selector (\S+)$`, kdt.KubeClientSet.ResourceShouldConvergeToSelector)
	kdt.scenario.Step(`^(?:the )?resource ([^"]*) condition ([^"]*) should be (true|false)$`, kdt.KubeClientSet.ResourceConditionShouldBe)
	kdt.scenario.Step(`^(?:I )?update (?:the )?resource ([^"]*) with ([^"]*) set to ([^"]*)$`, kdt.KubeClientSet.UpdateResourceWithField)
	kdt.scenario.Step(`^(?:I )?verify InstanceGroups (?:are )?in "ready" state$`, kdt.KubeClientSet.VerifyInstanceGroups)
	//syntax-generation:title-1:Structured Resources
	//syntax-generation:title-2:Pods
	kdt.scenario.Step(`^(?:I )?get (?:the )?pods in namespace ([^"]*)$`, kdt.KubeClientSet.Pods)
	kdt.scenario.Step(`^(?:I )?get (?:the )?pods in namespace ([^"]*) with selector (\S+)$`, kdt.KubeClientSet.PodsWithSelector)
	kdt.scenario.Step(`^(?:the )?pods in namespace ([^"]*) with selector (\S+) have restart count less than (\d+)$`, kdt.KubeClientSet.PodsWithSelectorHaveRestartCountLessThan)
	kdt.scenario.Step(`^(some|all) pods in namespace (\S+) with selector (\S+) have "([^"]*)" in logs since ([^"]*) time$`, kdt.KubeClientSet.SomeOrAllPodsInNamespaceWithSelectorHaveStringInLogsSinceTime)
	kdt.scenario.Step(`^some pods in namespace (\S+) with selector (\S+) don't have "([^"]*)" in logs since ([^"]*) time$`, kdt.KubeClientSet.SomePodsInNamespaceWithSelectorDontHaveStringInLogsSinceTime)
	kdt.scenario.Step(`^(?:the )?pods in namespace (\S+) with selector (\S+) have no errors in logs since ([^"]*) time$`, kdt.KubeClientSet.PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime)
	kdt.scenario.Step(`^(?:the )?pods in namespace (\S+) with selector (\S+) have some errors in logs since ([^"]*) time$`, kdt.KubeClientSet.PodsInNamespaceWithSelectorHaveSomeErrorsInLogsSinceTime)
	kdt.scenario.Step(`^(?:the )?pods in namespace (\S+) with selector (\S+) should have labels (\S+)$`, kdt.KubeClientSet.PodsInNamespaceWithSelectorShouldHaveLabels)
	kdt.scenario.Step(`^(?:the )?pod (\S+) in namespace (\S+) should have labels (\S+)$`, kdt.KubeClientSet.PodInNamespaceShouldHaveLabels)
	//syntax-generation:title-2:Others
	kdt.scenario.Step(`^(?:I )?(create|submit|update) (?:the )?secret (\S+) in namespace (\S+) from (?:environment variable )?(\S+)$`, kdt.KubeClientSet.SecretOperationFromEnvironmentVariable)
	kdt.scenario.Step(`^(?:I )?delete (?:the )?secret (\S+) in namespace (\S+)$`, kdt.KubeClientSet.SecretDelete)
	kdt.scenario.Step(`^(\d+) node(?:s)? with selector (\S+) should be (found|ready)$`, kdt.KubeClientSet.NodesWithSelectorShouldBe)
	kdt.scenario.Step(`^(?:the )?(deployment|hpa|horizontalpodautoscaler|service|pdb|poddisruptionbudget|sa|serviceaccount) ([^"]*) is in namespace ([^"]*)$`, kdt.KubeClientSet.ResourceInNamespace)
	kdt.scenario.Step(`^(?:I )?scale (?:the )?deployment ([^"]*) in namespace ([^"]*) to (\d+)$`, kdt.KubeClientSet.ScaleDeployment)
	kdt.scenario.Step(`^(?:I )?validate Prometheus Statefulset ([^"]*) in namespace ([^"]*) has volumeClaimTemplates name ([^"]*)$`, kdt.KubeClientSet.ValidatePrometheusVolumeClaimTemplatesName)
	kdt.scenario.Step(`^(?:I )?get (?:the )?nodes list$`, kdt.KubeClientSet.GetNodes)
	kdt.scenario.Step(`^(?:the )?daemonset ([^"]*) is running in namespace ([^"]*)$`, kdt.KubeClientSet.DaemonSetIsRunning)
	kdt.scenario.Step(`^(?:the )?deployment ([^"]*) is running in namespace ([^"]*)$`, kdt.KubeClientSet.DeploymentIsRunning)
	kdt.scenario.Step(`^(?:the )?persistentvolume ([^"]*) exists with status (Available|Bound|Released|Failed|Pending)$`, kdt.KubeClientSet.PersistentVolExists)
	kdt.scenario.Step(`^(?:the )?(clusterrole|clusterrolebinding) with name ([^"]*) should be found$`, kdt.KubeClientSet.ClusterRbacIsFound)
	kdt.scenario.Step(`^(?:the )?ingress (\S+) in (?:the )?namespace (\S+) (?:is )?(?:available )?on port (\d+) and path ([^"]*)$`, kdt.KubeClientSet.IngressAvailable)
	kdt.scenario.Step(`^(?:I )?send (\d+) tps to ingress (\S+) in (?:the )?namespace (\S+) (?:available )?on port (\d+) and path ([^"]*) for (\d+) (minutes|seconds) expecting up to (\d+) error(?:s)?$`, kdt.KubeClientSet.SendTrafficToIngress)
	//syntax-generation:title-0:AWS steps
	kdt.scenario.Step(`^(?:there are )?(?:valid )?AWS Credentials$`, kdt.AwsClientSet.DiscoverClients)
	kdt.scenario.Step(`^an Auto Scaling Group named ([^"]*)$`, kdt.AwsClientSet.AnASGNamed)
	kdt.scenario.Step(`^(?:I )?update (?:the )?current Auto Scaling Group with ([^"]*) set to ([^"]*)$`, kdt.AwsClientSet.UpdateFieldOfCurrentASG)
	kdt.scenario.Step(`^(?:the )?current Auto Scaling Group (?:is )?scaled to \(min, max\) = \((\d+), (\d+)\)$`, kdt.AwsClientSet.ScaleCurrentASG)
	kdt.scenario.Step(`^(?:the )?DNS name (\S+) (should|should not) be created in hostedZoneID (\S+)$`, kdt.AwsClientSet.DnsNameShouldOrNotInHostedZoneID)
	kdt.scenario.Step(`^(?:I )?(add|remove) (?:the )?(\S+) role as trusted entity to iam role ([^"]*)$`, kdt.AwsClientSet.IamRoleTrust)
	kdt.scenario.Step(`^(?:I )?(add|remove) cluster shared iam role$`, kdt.AwsClientSet.ClusterSharedIamOperation)
	//syntax-generation:end
}

/*
SetTestSuite sets the TestSuiteContext, should be use in the InitializeTestSuite function required by godog.
*/
func (kdt *Test) SetTestSuite(testSuite *godog.TestSuiteContext) {
	kdt.suite = testSuite
}
