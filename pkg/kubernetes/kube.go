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

//Package kube provides steps implementations related to Kubernetes.
package kube

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/pkg/errors"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	KubeInterface      kubernetes.Interface
	DynamicInterface   dynamic.Interface
	DiscoveryInterface discovery.DiscoveryInterface
	FilesPath          string
	TemplateArguments  interface{}
}

const (
	OperationCreate = "create"
	OperationSubmit = "submit"
	OperationUpdate = "update"
	OperationDelete = "delete"

	ResourceStateCreated = "created"
	ResourceStateDeleted = "deleted"

	NodeStateReady = "ready"
	NodeStateFound = "found"

	DefaultWaiterInterval = time.Second * 30
	DefaultWaiterRetries  = 40
)

/*
AKubernetesCluster sets the Kubernetes clients given a valid kube config file in ~/.kube or the path set in the environment variable KUBECONFIG.
*/
func (kc *Client) AKubernetesCluster() error {
	var (
		home, _        = os.UserHomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	)

	if exported := os.Getenv("KUBECONFIG"); exported != "" {
		kubeconfigPath = exported
	}

	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return errors.Errorf("[KUBEDOG] expected kubeconfig to exist for create operation, '%v'", kubeconfigPath)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct dynamic client", err)
	}

	_, err = client.Discovery().ServerVersion()
	if err != nil {
		return err
	}

	kc.KubeInterface = client
	kc.DynamicInterface = dynClient
	kc.DiscoveryInterface = discoveryClient

	return nil
}

/*
ResourceOperation performs the given operation on the resource defined in resourceFileName. The operation could be “create”, “submit” or “delete”.
*/
func (kc *Client) ResourceOperation(operation, resourceFileName string) error {
	if kc.DynamicInterface == nil {
		return errors.Errorf("'Client.DynamicInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	} else if kc.DiscoveryInterface == nil {
		return errors.Errorf("'Client.DiscoveryInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	resourcePath := kc.getResourcePath(resourceFileName)
	gvr, resource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}

	switch operation {
	case OperationCreate, OperationSubmit:
		_, err = kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Create(resource, metav1.CreateOptions{})
		if err != nil {
			if kerrors.IsAlreadyExists(err) {
				// already created
				break
			}
			return err
		}
	case OperationDelete:
		err = kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Delete(resource.GetName(), &metav1.DeleteOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				// already deleted
				break
			}
			return err
		}
	}
	return nil
}

/*
ResourceShouldBe checks if the resource defined in resourceFileName is in the desired state. It retries every 30 seconds for a total of 40 times. The state could be “created” or “deleted”.
*/
func (kc *Client) ResourceShouldBe(resourceFileName, state string) error {
	var (
		exists  bool
		counter int
	)

	if kc.DynamicInterface == nil {
		return errors.Errorf("'Client.DynamicInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	} else if kc.DiscoveryInterface == nil {
		return errors.Errorf("'Client.DiscoveryInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	resourcePath := kc.getResourcePath(resourceFileName)

	gvr, resource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}

	for {
		exists = true
		if counter >= DefaultWaiterRetries {
			return errors.New("waiter timed out waiting for resource state")
		}
		log.Infof("[KUBEDOG] waiting for resource %v/%v to become %v", resource.GetNamespace(), resource.GetName(), state)

		_, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(resource.GetName(), metav1.GetOptions{})
		if err != nil {
			if !kerrors.IsNotFound(err) {
				return err
			}
			log.Infof("[KUBEDOG] %v/%v is not found: %v", resource.GetNamespace(), resource.GetName(), err)
			exists = false
		}

		switch state {
		case ResourceStateDeleted:
			if !exists {
				log.Infof("[KUBEDOG] %v/%v is deleted", resource.GetNamespace(), resource.GetName())
				return nil
			}
		case ResourceStateCreated:
			if exists {
				log.Infof("[KUBEDOG] %v/%v is created", resource.GetNamespace(), resource.GetName())
				return nil
			}
		}
		counter++
		time.Sleep(DefaultWaiterInterval)
	}
}

/*
ResourceShouldConvergeToSelector checks if the resource defined in resourceFileName has the desired selector. It retries every 30 seconds for a total of 40 times. Selector in the form <keys>=<value>.
*/
func (kc *Client) ResourceShouldConvergeToSelector(resourceFileName, selector string) error {
	var counter int

	if kc.DynamicInterface == nil {
		return errors.Errorf("'Client.DynamicInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	} else if kc.DiscoveryInterface == nil {
		return errors.Errorf("'Client.DiscoveryInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	split := util.DeleteEmpty(strings.Split(selector, "="))
	if len(split) != 2 {
		return errors.Errorf("Selector '%s' should meet format '<key>=<value>'", selector)
	}

	key := split[0]
	value := split[1]

	keySlice := util.DeleteEmpty(strings.Split(key, "."))
	if len(keySlice) < 1 {
		return errors.Errorf("Found empty 'key' in selector '%s' of form '<key>=<value>'", selector)
	}

	resourcePath := kc.getResourcePath(resourceFileName)

	gvr, resource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}

	for {
		if counter >= DefaultWaiterRetries {
			return errors.New("waiter timed out waiting for resource")
		}

		log.Infof("[KUBEDOG] waiting for resource %v/%v to converge to %v=%v", resource.GetNamespace(), resource.GetName(), key, value)
		cr, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(resource.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}

		if val, ok, err := unstructured.NestedString(cr.UnstructuredContent(), keySlice...); ok {
			if err != nil {
				return err
			}
			if strings.EqualFold(val, value) {
				break
			}
		}
		counter++
		time.Sleep(DefaultWaiterInterval)
	}

	return nil
}

/*
ResourceConditionShouldBe checks that the resource defined in resourceFileName has the condition of type cType in the desired status. It retries every 30 seconds for a total of 40 times.
*/
func (kc *Client) ResourceConditionShouldBe(resourceFileName, cType, status string) error {
	var (
		counter        int
		expectedStatus = strings.Title(status)
	)

	if kc.DynamicInterface == nil {
		return errors.Errorf("'Client.DynamicInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	} else if kc.DiscoveryInterface == nil {
		return errors.Errorf("'Client.DiscoveryInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	resourcePath := kc.getResourcePath(resourceFileName)
	gvr, resource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}

	for {
		if counter >= DefaultWaiterRetries {
			return errors.New("waiter timed out waiting for resource state")
		}
		log.Infof("[KUBEDOG] waiting for resource %v/%v to meet condition %v=%v", resource.GetNamespace(), resource.GetName(), cType, expectedStatus)
		cr, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(resource.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}

		if conditions, ok, err := unstructured.NestedSlice(cr.UnstructuredContent(), "status", "conditions"); ok {
			if err != nil {
				return err
			}

			for _, c := range conditions {
				condition, ok := c.(map[string]interface{})
				if !ok {
					continue
				}
				tp, found := condition["type"]
				if !found {
					continue
				}
				condType, ok := tp.(string)
				if !ok {
					continue
				}
				if condType == cType {
					status := condition["status"].(string)
					if corev1.ConditionStatus(status) == corev1.ConditionStatus(expectedStatus) {
						return nil
					}
				}
			}
		}
		counter++
		time.Sleep(DefaultWaiterInterval)
	}
}

/*
NodesWithSelectorShouldBe checks that n amount of nodes with the given selector are in the desired state. It retries every 30 seconds for a total of 40 times. Selector in the form <key>=<value>, the state can be "ready" or "found".
*/
func (kc *Client) NodesWithSelectorShouldBe(n int, selector, state string) error {
	var (
		counter int
		found   bool
	)

	if kc.KubeInterface == nil {
		return errors.Errorf("'Client.KubeInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	for {
		var (
			conditionNodes int
			opts           = metav1.ListOptions{
				LabelSelector: selector,
			}
		)

		if counter >= DefaultWaiterRetries {
			return errors.New("waiter timed out waiting for nodes")
		}

		log.Infof("[KUBEDOG] waiting for %v nodes to be %v with selector %v", n, state, selector)
		nodes, err := kc.KubeInterface.CoreV1().Nodes().List(opts)
		if err != nil {
			return err
		}

		switch state {
		case NodeStateFound:
			if len(nodes.Items) == n {
				log.Infof("[KUBEDOG] found %v nodes", n)
				found = true
			}
		case NodeStateReady:
			for _, node := range nodes.Items {
				if util.IsNodeReady(node) {
					conditionNodes++
				}
			}
			if conditionNodes == n {
				log.Infof("[KUBEDOG] found %v ready nodes", n)
				found = true
			}
		}

		if found {
			break
		}

		counter++
		time.Sleep(DefaultWaiterInterval)
	}
	return nil
}

/*
UpdateResourceWithField it updates the field found in the key of the resource defined in resourceFileName with value.
*/
func (kc *Client) UpdateResourceWithField(resourceFileName, key string, value string) error {
	var (
		keySlice     = util.DeleteEmpty(strings.Split(key, "."))
		overrideType bool
		intValue     int64
		//err          error
	)

	if kc.DynamicInterface == nil {
		return errors.Errorf("'Client.DynamicInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	} else if kc.DiscoveryInterface == nil {
		return errors.Errorf("'Client.DiscoveryInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	resourcePath := kc.getResourcePath(resourceFileName)
	gvr, resource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		overrideType = true
		intValue = n
	}

	updateTarget, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(resource.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	switch overrideType {
	case true:
		if err := unstructured.SetNestedField(updateTarget.UnstructuredContent(), intValue, keySlice...); err != nil {
			return err
		}
	case false:
		if err := unstructured.SetNestedField(updateTarget.UnstructuredContent(), value, keySlice...); err != nil {
			return err
		}
	}

	_, err = kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Update(updateTarget, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	return nil
}

/*
DeleteAllTestResources deletes all the resources defined by yaml files in the path given by FilesPath, if FilesPath is empty, it will look for the files in ./templates. Meant to be use in the before/after suite/scenario/step hooks
*/
func (kc *Client) DeleteAllTestResources() error {
	resourcesPath := kc.getTemplatesPath()

	// Getting context
	err := kc.AKubernetesCluster()
	if err != nil {
		return errors.Errorf("Failed getting the kubernetes client: %v", err)
	}

	var deleteFn = func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		gvr, resource, err := util.GetResourceFromYaml(path, kc.DiscoveryInterface, kc.TemplateArguments)
		if err != nil {
			return err
		}

		kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Delete(resource.GetName(), &metav1.DeleteOptions{})
		log.Infof("[KUBEDOG] submitted deletion for %v/%v", resource.GetNamespace(), resource.GetName())
		return nil
	}

	var waitFn = func(path string, info os.FileInfo, walkErr error) error {
		var (
			counter int
		)

		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		gvr, resource, err := util.GetResourceFromYaml(path, kc.DiscoveryInterface, kc.TemplateArguments)
		if err != nil {
			return err
		}

		for {
			if counter >= DefaultWaiterRetries {
				return errors.New("waiter timed out waiting for deletion")
			}
			log.Infof("[KUBEDOG] waiting for resource deletion of %v/%v", resource.GetNamespace(), resource.GetName())
			_, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(resource.GetName(), metav1.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					log.Infof("[KUBEDOG] resource %v/%v is deleted", resource.GetNamespace(), resource.GetName())
					break
				}
			}
			counter++
			time.Sleep(DefaultWaiterInterval)
		}
		return nil
	}

	if err := filepath.Walk(resourcesPath, deleteFn); err != nil {
		return err
	}
	if err := filepath.Walk(resourcesPath, waitFn); err != nil {
		return err
	}

	return nil
}

/*
DeploymentInNamespace check if deployment in the related namespace
*/
func (kc *Client) DeploymentInNamespace(name, ns string) error {
	if kc.KubeInterface == nil {
		return errors.Errorf("'Client.KubeInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	_, err := kc.KubeInterface.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

/*
ScaleDeployment scale up/down for the deployment
*/
func (kc *Client) ScaleDeployment(name, ns string, replica int32) error {
	if kc.KubeInterface == nil {
		return errors.Errorf("'Client.KubeInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replica,
		},
	}

	_, err := kc.KubeInterface.AppsV1().Deployments(ns).UpdateScale(name, scale)
	if err != nil {
		return err
	}
	return nil
}

func (kc *Client) getTemplatesPath() string {
	if kc.FilesPath != "" {
		return kc.FilesPath
	} else {
		return "templates"
	}
}
func (kc *Client) getResourcePath(resourceFileName string) string {
	templatesPath := kc.getTemplatesPath()
	return filepath.Join(templatesPath, resourceFileName)
}
