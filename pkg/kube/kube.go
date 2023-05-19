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

package kube

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/keikoproj/kubedog/pkg/kube/common"
	"github.com/keikoproj/kubedog/pkg/kube/pod"
	"github.com/keikoproj/kubedog/pkg/kube/structured"
	unstruct "github.com/keikoproj/kubedog/pkg/kube/unstructured"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientSet struct {
	KubeInterface    kubernetes.Interface
	DynamicInterface dynamic.Interface
	timestamps       map[string]time.Time
	config           configuration
}

// TODO: update docs to reflect the changes in unexported fields and rm methods from root package
func (kc *ClientSet) SetFilesPath(path string) {
	kc.config.filesPath = path
}

func (kc *ClientSet) SetTemplateArguments(args interface{}) {
	kc.config.templateArguments = args
}

func (kc *ClientSet) SetWaiterInterval(duration time.Duration) {
	kc.config.waiterInterval = duration
}

func (kc *ClientSet) SetWaiterTries(tries int) {
	kc.config.waiterTries = tries
}

func (kc *ClientSet) DiscoverClients() error {
	var (
		home, _        = os.UserHomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	)

	if exported := os.Getenv("KUBECONFIG"); exported != "" {
		kubeconfigPath = exported
	}
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return errors.Errorf("expected kubeconfig to exist for create operation, '%v'", kubeconfigPath)
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct dynamic client", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	_, err = client.Discovery().ServerVersion()
	if err != nil {
		return err
	}

	kc.DynamicInterface = dynClient
	kc.KubeInterface = client

	return nil
}

func (kc *ClientSet) SetTimestamp(timestampName string) error {
	now := time.Now()
	if kc.timestamps == nil {
		kc.timestamps = map[string]time.Time{}
	}
	kc.timestamps[timestampName] = now
	log.Infof("Set timestamp '%s' as '%v'", timestampName, now)
	return nil
}

func (kc *ClientSet) KubernetesClusterShouldBe(state string) error {
	switch state {
	case common.StateCreated, common.StateUpgraded:
		if err := pod.Pods(kc.KubeInterface, metav1.NamespaceSystem); err != nil {
			return errors.Errorf("failed validating cluster create/update, could not get pods: '%v'", err)
		}
		return nil
	case common.StateDeleted:
		if err := kc.DiscoverClients(); err == nil {
			return errors.New("failed validating cluster delete, cluster is still available")
		}
		return nil
	default:
		return fmt.Errorf("unsupported state: '%s'", state)
	}
}

func (kc *ClientSet) DeleteAllTestResources() error {
	return unstruct.DeleteResourcesAtPath(kc.DynamicInterface, kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getWaiterConfig(), kc.getTemplatesPath())
}

func (kc *ClientSet) ResourceOperation(operation, resourceFileName string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperation(kc.DynamicInterface, resource, operation)
}

func (kc *ClientSet) ResourceOperationInNamespace(operation, resourceFileName, namespace string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperationInNamespace(kc.DynamicInterface, resource, operation, namespace)
}

func (kc *ClientSet) ResourcesOperation(operation, resourcesFileName string) error {
	resources, err := unstruct.GetResources(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourcesFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourcesOperation(kc.DynamicInterface, resources, operation)
}

func (kc *ClientSet) ResourcesOperationInNamespace(operation, resourcesFileName, namespace string) error {
	resources, err := unstruct.GetResources(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourcesFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourcesOperationInNamespace(kc.DynamicInterface, resources, operation, namespace)
}

func (kc *ClientSet) ResourceOperationWithResult(operation, resourceFileName, expectedResult string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperationWithResult(kc.DynamicInterface, resource, operation, expectedResult)
}

func (kc *ClientSet) ResourceOperationWithResultInNamespace(operation, resourceFileName, namespace, expectedResult string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceOperationWithResultInNamespace(kc.DynamicInterface, resource, operation, namespace, expectedResult)
}

func (kc *ClientSet) ResourceShouldBe(resourceFileName, state string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceShouldBe(kc.DynamicInterface, resource, kc.getWaiterConfig(), state)
}

func (kc *ClientSet) ResourceShouldConvergeToSelector(resourceFileName, selector string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceShouldConvergeToSelector(kc.DynamicInterface, resource, kc.getWaiterConfig(), selector)
}

func (kc *ClientSet) ResourceConditionShouldBe(resourceFileName, conditionType, conditionValue string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.ResourceConditionShouldBe(kc.DynamicInterface, resource, kc.getWaiterConfig(), conditionType, conditionValue)
}

func (kc *ClientSet) UpdateResourceWithField(resourceFileName, key, value string) error {
	resource, err := unstruct.GetResource(kc.KubeInterface.Discovery(), kc.config.templateArguments, kc.getResourcePath(resourceFileName))
	if err != nil {
		return err
	}
	return unstruct.UpdateResourceWithField(kc.DynamicInterface, resource, key, value)
}

func (kc *ClientSet) VerifyInstanceGroups() error {
	return unstruct.VerifyInstanceGroups(kc.DynamicInterface)
}

func (kc *ClientSet) Pods(namespace string) error {
	return pod.Pods(kc.KubeInterface, namespace)
}

func (kc *ClientSet) PodsWithSelector(namespace, selector string) error {
	return pod.PodsWithSelector(kc.KubeInterface, namespace, selector)
}

func (kc *ClientSet) PodsWithSelectorHaveRestartCountLessThan(namespace, selector string, restartCount int) error {
	return pod.PodsWithSelectorHaveRestartCountLessThan(kc.KubeInterface, namespace, selector, restartCount)
}

func (kc *ClientSet) SomeOrAllPodsInNamespaceWithSelectorHaveStringInLogsSinceTime(someOrAll, namespace, selector, searchKeyword, sinceTime string) error {
	timestamp, err := kc.getTimestamp(sinceTime)
	if err != nil {
		return err
	}
	return pod.SomeOrAllPodsInNamespaceWithSelectorHaveStringInLogsSinceTime(kc.KubeInterface, kc.getExpBackoff(), someOrAll, namespace, selector, searchKeyword, timestamp)
}

func (kc *ClientSet) SomePodsInNamespaceWithSelectorDontHaveStringInLogsSinceTime(namespace, selector, searchKeyword, sinceTime string) error {
	timestamp, err := kc.getTimestamp(sinceTime)
	if err != nil {
		return err
	}
	return pod.SomePodsInNamespaceWithSelectorDontHaveStringInLogsSinceTime(kc.KubeInterface, namespace, selector, searchKeyword, timestamp)
}

func (kc *ClientSet) PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime(namespace, selector, sinceTime string) error {
	timestamp, err := kc.getTimestamp(sinceTime)
	if err != nil {
		return err
	}
	return pod.PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime(kc.KubeInterface, namespace, selector, timestamp)
}

func (kc *ClientSet) PodsInNamespaceWithSelectorHaveSomeErrorsInLogsSinceTime(namespace, selector, sinceTime string) error {
	timestamp, err := kc.getTimestamp(sinceTime)
	if err != nil {
		return err
	}
	return pod.PodsInNamespaceWithSelectorHaveSomeErrorsInLogsSinceTime(kc.KubeInterface, namespace, selector, timestamp)
}

func (kc *ClientSet) PodsInNamespaceWithSelectorShouldHaveLabels(namespace, selector, labels string) error {
	return pod.PodsInNamespaceWithSelectorShouldHaveLabels(kc.KubeInterface, namespace, selector, labels)
}

func (kc *ClientSet) PodInNamespaceShouldHaveLabels(name, namespace, labels string) error {
	return pod.PodInNamespaceShouldHaveLabels(kc.KubeInterface, name, namespace, labels)
}

func (kc *ClientSet) SecretOperationFromEnvironmentVariable(operation, name, namespace, environmentVariable string) error {
	return structured.SecretOperationFromEnvironmentVariable(kc.KubeInterface, operation, name, namespace, environmentVariable)
}

func (kc *ClientSet) SecretDelete(name, namespace string) error {
	return structured.SecretDelete(kc.KubeInterface, name, namespace)
}

func (kc *ClientSet) NodesWithSelectorShouldBe(expectedNodes int, selector, state string) error {
	return structured.NodesWithSelectorShouldBe(kc.KubeInterface, kc.getWaiterConfig(), expectedNodes, selector, state)
}

func (kc *ClientSet) ResourceInNamespace(resourceType, name, namespace string) error {
	return structured.ResourceInNamespace(kc.KubeInterface, resourceType, name, namespace)
}

func (kc *ClientSet) ScaleDeployment(name, namespace string, replicas int32) error {
	return structured.ScaleDeployment(kc.KubeInterface, name, namespace, replicas)
}

func (kc *ClientSet) ValidatePrometheusVolumeClaimTemplatesName(statefulsetName, namespace, volumeClaimTemplatesName string) error {
	return structured.ValidatePrometheusVolumeClaimTemplatesName(kc.KubeInterface, statefulsetName, namespace, volumeClaimTemplatesName)
}

func (kc *ClientSet) GetNodes() error {
	return structured.GetNodes(kc.KubeInterface)
}

func (kc *ClientSet) DaemonSetIsRunning(name, namespace string) error {
	return structured.DaemonSetIsRunning(kc.KubeInterface, kc.getExpBackoff(), name, namespace)
}

func (kc *ClientSet) DeploymentIsRunning(name, namespace string) error {
	return structured.DeploymentIsRunning(kc.KubeInterface, name, namespace)
}

func (kc *ClientSet) PersistentVolExists(name, expectedPhase string) error {
	return structured.PersistentVolExists(kc.KubeInterface, name, expectedPhase)
}

func (kc *ClientSet) ClusterRbacIsFound(resourceType, name string) error {
	return structured.ClusterRbacIsFound(kc.KubeInterface, resourceType, name)
}

func (kc *ClientSet) IngressAvailable(name, namespace string, port int, path string) error {
	return structured.IngressAvailable(kc.KubeInterface, kc.getWaiterConfig(), name, namespace, port, path)
}

func (kc *ClientSet) SendTrafficToIngress(tps int, name, namespace string, port int, path string, duration int, durationUnits string, expectedErrors int) error {
	return structured.SendTrafficToIngress(kc.KubeInterface, kc.getWaiterConfig(), tps, name, namespace, port, path, duration, durationUnits, expectedErrors)
}
