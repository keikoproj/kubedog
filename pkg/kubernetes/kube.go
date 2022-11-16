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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	util "github.com/keikoproj/kubedog/internal/utilities"

	"github.com/pkg/errors"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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
	WaiterInterval     time.Duration
	WaiterTries        int
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
	DefaultWaiterTries    = 40

	DefaultFilePath = "templates"
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
ResourceOperation performs the given operation on the resource defined in resourceFileName. The operation could be “create”, “submit”, “delete”, or "update".
*/
func (kc *Client) ResourceOperation(operation, resourceFileName string) error {
	return kc.ResourceOperationInNamespace(operation, resourceFileName, "")
}

/*
ResourceOperationInNamespace performs the given operation on the resource defined in resourceFileName in a specified namespace. The operation could be “create”, “submit”, “delete”, or "update".
*/
func (kc *Client) ResourceOperationInNamespace(operation, resourceFileName, ns string) error {
	unstructuredResource, err := kc.parseSingleResource(resourceFileName)
	if err != nil {
		return err
	}
	return kc.unstructuredResourceOperation(operation, ns, unstructuredResource)
}

func (kc *Client) parseSingleResource(resourceFileName string) (util.K8sUnstructuredResource, error) {
	if kc.DynamicInterface == nil {
		return util.K8sUnstructuredResource{}, errors.Errorf("'Client.DynamicInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	} else if kc.DiscoveryInterface == nil {
		return util.K8sUnstructuredResource{}, errors.Errorf("'Client.DiscoveryInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	resourcePath := kc.getResourcePath(resourceFileName)
	unstructuredResource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return util.K8sUnstructuredResource{}, err
	}

	return unstructuredResource, nil
}

/*
MultiResourceOperation performs the given operation on the resources defined in resourceFileName. The operation could be “create”, “submit” or “delete”.
Files created using this function cannot individually be addressed by filename.
*/
func (kc *Client) MultiResourceOperation(operation, resourceFileName string) error {
	resourceList, err := kc.parseMultipleResources(resourceFileName)
	if err != nil {
		return err
	}

	for _, unstructuredResource := range resourceList {
		err = kc.unstructuredResourceOperation(operation, "", unstructuredResource)
		if err != nil {
			return err
		}
	}

	return nil
}

func (kc *Client) MultiResourceOperationInNamespace(operation, resourceFileName, ns string) error {
	resourceList, err := kc.parseMultipleResources(resourceFileName)
	if err != nil {
		return err
	}

	for _, unstructuredResource := range resourceList {
		err = kc.unstructuredResourceOperation(operation, ns, unstructuredResource)
		if err != nil {
			return err
		}
	}

	return nil
}

func (kc *Client) parseMultipleResources(resourceFileName string) ([]util.K8sUnstructuredResource, error) {
	if kc.DynamicInterface == nil {
		return nil, errors.Errorf("'Client.DynamicInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	} else if kc.DiscoveryInterface == nil {
		return nil, errors.Errorf("'Client.DiscoveryInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	resourcePath := kc.getResourcePath(resourceFileName)

	resourceList, err := util.GetMultipleResourcesFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return nil, err
	}

	return resourceList, nil
}

func (kc *Client) unstructuredResourceOperation(operation, ns string, unstructuredResource util.K8sUnstructuredResource) error {
	gvr, resource := unstructuredResource.GVR, unstructuredResource.Resource

	if ns == "" {
		ns = resource.GetNamespace()
	}

	switch operation {
	case OperationCreate, OperationSubmit:
		_, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(ns).Create(context.Background(), resource, metav1.CreateOptions{})
		if err != nil {
			if kerrors.IsAlreadyExists(err) {
				log.Infof("%s %s already created", resource.GetKind(), resource.GetName())
				break
			}
			return err
		}
		log.Infof("%s %s has been created in namespace %s", resource.GetKind(), resource.GetName(), ns)
	case OperationUpdate:
		currentResourceVersion, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(ns).Get(context.Background(), resource.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}

		resource.SetResourceVersion(currentResourceVersion.DeepCopy().GetResourceVersion())

		_, err = kc.DynamicInterface.Resource(gvr.Resource).Namespace(ns).Update(context.Background(), resource, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		log.Infof("%s %s has been updated in namespace %s", resource.GetKind(), resource.GetName(), ns)
	case OperationDelete:
		err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(ns).Delete(context.Background(), resource.GetName(), metav1.DeleteOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				log.Infof("%s %s already deleted", resource.GetKind(), resource.GetName())
				break
			}
			return err
		}
		log.Infof("%s %s has been deleted from namespace %s", resource.GetKind(), resource.GetName(), ns)
	default:
		return fmt.Errorf("unsupported operation: %s", operation)
	}
	return nil
}

func (kc *Client) ResourceOperationWithResult(operation, resourceFileName, expectedResult string) error {
	return kc.ResourceOperationWithResultInNamespace(operation, resourceFileName, "", expectedResult)
}

func (kc *Client) ResourceOperationWithResultInNamespace(operation, resourceFileName, namespace, expectedResult string) error {
	var expectError = strings.EqualFold(expectedResult, "fail")
	err := kc.ResourceOperationInNamespace(operation, resourceFileName, "")
	if !expectError && err != nil {
		return fmt.Errorf("unexpected error when %s %s: %s", operation, resourceFileName, err.Error())
	} else if expectError && err == nil {
		return fmt.Errorf("expected error when %s %s, but received none", operation, resourceFileName)
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

	unstructuredResource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}
	gvr, resource := unstructuredResource.GVR, unstructuredResource.Resource
	for {
		exists = true
		if counter >= kc.getWaiterTries() {
			return errors.New("waiter timed out waiting for resource state")
		}
		log.Infof("[KUBEDOG] waiting for resource %v/%v to become %v", resource.GetNamespace(), resource.GetName(), state)

		_, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(context.Background(), resource.GetName(), metav1.GetOptions{})
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
		time.Sleep(kc.getWaiterInterval())
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

	unstructuredResource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}
	gvr, resource := unstructuredResource.GVR, unstructuredResource.Resource

	for {
		if counter >= kc.getWaiterTries() {
			return errors.New("waiter timed out waiting for resource")
		}

		log.Infof("[KUBEDOG] waiting for resource %v/%v to converge to %v=%v", resource.GetNamespace(), resource.GetName(), key, value)
		cr, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(context.Background(), resource.GetName(), metav1.GetOptions{})
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
		time.Sleep(kc.getWaiterInterval())
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
	unstructuredResource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}
	gvr, resource := unstructuredResource.GVR, unstructuredResource.Resource

	for {
		if counter >= kc.getWaiterTries() {
			return errors.New("waiter timed out waiting for resource state")
		}
		log.Infof("[KUBEDOG] waiting for resource %v/%v to meet condition %v=%v", resource.GetNamespace(), resource.GetName(), cType, expectedStatus)
		cr, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(context.Background(), resource.GetName(), metav1.GetOptions{})
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
		time.Sleep(kc.getWaiterInterval())
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
			nodesCount int
			opts       = metav1.ListOptions{
				LabelSelector: selector,
			}
		)

		if counter >= kc.getWaiterTries() {
			return errors.New("waiter timed out waiting for nodes")
		}

		nodes, err := kc.KubeInterface.CoreV1().Nodes().List(context.Background(), opts)
		if err != nil {
			return err
		}

		switch state {
		case NodeStateFound:
			nodesCount = len(nodes.Items)
			if nodesCount == n {
				log.Infof("[KUBEDOG] found %v nodes", n)
				found = true
			}
		case NodeStateReady:
			for _, node := range nodes.Items {
				if util.IsNodeReady(node) {
					nodesCount++
				}
			}
			if nodesCount == n {
				log.Infof("[KUBEDOG] found %v ready nodes", n)
				found = true
			}
		}

		if found {
			break
		}

		log.Infof("[KUBEDOG] found %v nodes, waiting for %v nodes to be %v with selector %v", nodesCount, n, state, selector)

		counter++
		time.Sleep(kc.getWaiterInterval())
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
	unstructuredResource, err := util.GetResourceFromYaml(resourcePath, kc.DiscoveryInterface, kc.TemplateArguments)
	if err != nil {
		return err
	}
	gvr, resource := unstructuredResource.GVR, unstructuredResource.Resource

	n, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		overrideType = true
		intValue = n
	}

	updateTarget, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(context.Background(), resource.GetName(), metav1.GetOptions{})
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

	_, err = kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Update(context.Background(), updateTarget, metav1.UpdateOptions{})
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

		unstructuredResource, err := util.GetResourceFromYaml(path, kc.DiscoveryInterface, kc.TemplateArguments)
		if err != nil {
			return err
		}
		gvr, resource := unstructuredResource.GVR, unstructuredResource.Resource

		kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Delete(context.Background(), resource.GetName(), metav1.DeleteOptions{})
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

		unstructuredResource, err := util.GetResourceFromYaml(path, kc.DiscoveryInterface, kc.TemplateArguments)
		if err != nil {
			return err
		}
		gvr, resource := unstructuredResource.GVR, unstructuredResource.Resource

		for {
			if counter >= kc.getWaiterTries() {
				return errors.New("waiter timed out waiting for deletion")
			}
			log.Infof("[KUBEDOG] waiting for resource deletion of %v/%v", resource.GetNamespace(), resource.GetName())
			_, err := kc.DynamicInterface.Resource(gvr.Resource).Namespace(resource.GetNamespace()).Get(context.Background(), resource.GetName(), metav1.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					log.Infof("[KUBEDOG] resource %v/%v is deleted", resource.GetNamespace(), resource.GetName())
					break
				}
			}
			counter++
			time.Sleep(kc.getWaiterInterval())
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
ResourceInNamespace check if (deployment|service) in the related namespace
*/
func (kc *Client) ResourceInNamespace(resource, name, ns string) error {
	var err error

	if kc.KubeInterface == nil {
		return errors.Errorf("'Client.KubeInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	switch resource {
	case "deployment":
		_, err = kc.KubeInterface.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})

	case "service":
		_, err = kc.KubeInterface.CoreV1().Services(ns).Get(context.Background(), name, metav1.GetOptions{})

	case "hpa", "horizontalpodautoscaler":
		_, err = kc.KubeInterface.AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(context.Background(), name, metav1.GetOptions{})

	case "pdb", "poddisruptionbudget":
		_, err = kc.KubeInterface.PolicyV1beta1().PodDisruptionBudgets(ns).Get(context.Background(), name, metav1.GetOptions{})
	case "sa", "serviceaccount":
		_, err = kc.KubeInterface.CoreV1().ServiceAccounts(ns).Get(context.Background(), name, metav1.GetOptions{})

	default:
		return errors.Errorf("Invalid resource type")
	}

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

	_, err := kc.KubeInterface.AppsV1().Deployments(ns).UpdateScale(context.Background(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (kc *Client) getTemplatesPath() string {
	if kc.FilesPath != "" {
		return kc.FilesPath
	}
	return DefaultFilePath
}

func (kc *Client) getResourcePath(resourceFileName string) string {
	templatesPath := kc.getTemplatesPath()
	return filepath.Join(templatesPath, resourceFileName)
}

/*
Cluster scoped role and bindings are found.
*/

func (kc *Client) ClusterRbacIsFound(resource, name string) error {
	var err error
	if kc.KubeInterface == nil {
		return errors.Errorf("'Client.KubeInterface' is nil. 'AKubernetesCluster' sets this interface, try calling it before using this method")
	}

	switch resource {
	case "clusterrole":
		_, err = kc.KubeInterface.RbacV1().ClusterRoles().Get(context.Background(), name, metav1.GetOptions{})
	case "clusterrolebinding":
		_, err = kc.KubeInterface.RbacV1().ClusterRoleBindings().Get(context.Background(), name, metav1.GetOptions{})
	default:
		return errors.Errorf("Invalid resource type")
	}

	if err != nil {
		return err
	}
	return nil
}

func (kc *Client) getWaiterInterval() time.Duration {
	if kc.WaiterInterval >= 0 {
		return kc.WaiterInterval
	}
	return DefaultWaiterInterval
}

func (kc *Client) getWaiterTries() int {
	if kc.WaiterTries > 0 {
		return kc.WaiterTries
	}
	return DefaultWaiterTries
}

func (kc *Client) PrintNodes() error {

	var readyStatus = func(conditions []v1.NodeCondition) string {
		var status = false
		var err error
		for _, condition := range conditions {
			if condition.Type == "Ready" {
				status, err = strconv.ParseBool(string(condition.Status))
				if err != nil {
					return "Unknown"
				}
				break
			}
		}
		if status {
			return "Ready"
		}
		return "NotReady"
	}
	// List nodes
	nodes, _ := kc.ListNodes()
	if nodes != nil {
		tableFormat := "%-64s%-12s%-24s%-16s"
		log.Infof(tableFormat, "NAME", "STATUS", "INSTANCEGROUP", "AZ")
		for _, node := range nodes.Items {
			log.Infof(tableFormat,
				node.Name,
				readyStatus(node.Status.Conditions),
				node.Labels["node.kubernetes.io/instancegroup"],
				node.Labels["failure-domain.beta.kubernetes.io/zone"])
		}
	}
	return nil
}

func (kc *Client) PrintPods(namespace string) error {
	return kc.PrintPodsWithSelector(namespace, "")
}

func (kc *Client) PrintPodsWithSelector(namespace, selector string) error {
	var readyCount = func(conditions []v1.ContainerStatus) string {
		var readyCount = 0
		var containerCount = len(conditions)
		for _, condition := range conditions {
			if condition.Ready {
				readyCount++
			}
		}
		return fmt.Sprintf("%d/%d", readyCount, containerCount)
	}
	pods, err := kc.ListPodsWithLabelSelector(namespace, selector)
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return errors.Errorf("No pods matched selector '%s'", selector)
	}
	tableFormat := "%-64s%-12s%-24s"
	log.Infof(tableFormat, "NAME", "READY", "STATUS")
	for _, pod := range pods.Items {
		log.Infof(tableFormat,
			pod.Name, readyCount(pod.Status.ContainerStatuses), pod.Status.Phase)
	}
	return nil
}

func (kc *Client) daemonsetIsRunning(dsName, namespace string) error {
	gomega.Eventually(func() error {
		ds, err := kc.GetDaemonset(dsName, namespace)
		if err != nil {
			return err
		}

		if ds.Status.DesiredNumberScheduled != ds.Status.CurrentNumberScheduled {
			return fmt.Errorf("daemonset %s/%s is not updated. status: %s", namespace, dsName, ds.Status.String())
		}

		return nil
	}, 10*time.Second).Should(gomega.Succeed(), func() string {
		// Print Pods after failure
		_ = kc.PrintPods(namespace)
		return fmt.Sprintf("daemonset %s/%s is not updated.", namespace, dsName)
	})

	return nil
}

func (kc *Client) deploymentIsRunning(deployName, namespace string) error {
	deploy, err := kc.GetDeployment(deployName, namespace)
	if err != nil {
		return err
	}
	if deploy.Status.ReadyReplicas != deploy.Status.Replicas {
		return fmt.Errorf("deployment %s/%s is not ready. status: %s", namespace, deployName, deploy.Status.String())
	}

	if deploy.Status.UpdatedReplicas != deploy.Status.Replicas {
		return fmt.Errorf("deploymemnt %s/%s is not updated. status: %s", namespace, deployName, deploy.Status.String())
	}

	return nil
}

func (kc *Client) ResourceIsRunning(kind, name, namespace string) error {
	kind = strings.ToLower(kind)
	switch kind {
	case "daemonset":
		return kc.daemonsetIsRunning(name, namespace)
	case "deployment":
		return kc.deploymentIsRunning(name, namespace)
	default:
		return fmt.Errorf("invalid resource type: %s", kind)
	}
}

func (kc *Client) PersistentVolExists(volName, expectedPhase string) error {
	vol, err := kc.GetPersistentVolume(volName)
	if err != nil {
		return err
	}
	phase := string(vol.Status.Phase)
	if phase != expectedPhase {
		return fmt.Errorf("persistentvolume had unexpected phase %v, expected phase %v", phase, expectedPhase)
	}
	return nil
}
