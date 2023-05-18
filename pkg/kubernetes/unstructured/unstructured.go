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

package unstructured

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

func ResourceOperation(dynamicClient dynamic.Interface, resource unstructuredResource, operation string) error {
	return ResourceOperationInNamespace(dynamicClient, resource, operation, "")
}

func MultiResourceOperation(dynamicClient dynamic.Interface, resources []unstructuredResource, operation string) error {
	for _, resource := range resources {
		err := ResourceOperationInNamespace(dynamicClient, resource, operation, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func MultiResourceOperationInNamespace(dynamicClient dynamic.Interface, resources []unstructuredResource, operation, namespace string) error {
	for _, resource := range resources {
		err := ResourceOperationInNamespace(dynamicClient, resource, operation, namespace)
		if err != nil {
			return err
		}
	}
	return nil
}

func ResourceOperationInNamespace(dynamicClient dynamic.Interface, resource unstructuredResource, operation, namespace string) error {
	if err := validateDynamicClient(dynamicClient); err != nil {
		return err
	}

	gvr, unstruct := resource.GVR, resource.Resource

	if namespace == "" {
		namespace = unstruct.GetNamespace()
	}

	switch operation {
	case common.OperationCreate, common.OperationSubmit:
		_, err := dynamicClient.Resource(gvr.Resource).Namespace(namespace).Create(context.Background(), unstruct, metav1.CreateOptions{})
		if err != nil {
			if kerrors.IsAlreadyExists(err) {
				log.Infof("%s %s already created", unstruct.GetKind(), unstruct.GetName())
				break
			}
			return err
		}
		log.Infof("%s %s has been created in namespace %s", unstruct.GetKind(), unstruct.GetName(), namespace)
	case common.OperationUpdate:
		currentResourceVersion, err := dynamicClient.Resource(gvr.Resource).Namespace(namespace).Get(context.Background(), unstruct.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}

		unstruct.SetResourceVersion(currentResourceVersion.DeepCopy().GetResourceVersion())

		_, err = dynamicClient.Resource(gvr.Resource).Namespace(namespace).Update(context.Background(), unstruct, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		log.Infof("%s %s has been updated in namespace %s", unstruct.GetKind(), unstruct.GetName(), namespace)
	case common.OperationDelete:
		err := dynamicClient.Resource(gvr.Resource).Namespace(namespace).Delete(context.Background(), unstruct.GetName(), metav1.DeleteOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				log.Infof("%s %s already deleted", unstruct.GetKind(), unstruct.GetName())
				break
			}
			return err
		}
		log.Infof("%s %s has been deleted from namespace %s", unstruct.GetKind(), unstruct.GetName(), namespace)
	default:
		return fmt.Errorf("unsupported operation: %s", operation)
	}
	return nil
}

func ResourceOperationWithResult(dynamicClient dynamic.Interface, resource unstructuredResource, operation, expectedResult string) error {
	return ResourceOperationWithResultInNamespace(dynamicClient, resource, operation, "", expectedResult)
}

func ResourceOperationWithResultInNamespace(dynamicClient dynamic.Interface, resource unstructuredResource, operation, namespace, expectedResult string) error {
	var expectError = strings.EqualFold(expectedResult, "fail")
	err := ResourceOperationInNamespace(dynamicClient, resource, operation, namespace)
	if !expectError && err != nil {
		return fmt.Errorf("unexpected error when '%s' '%s': '%s'", operation, resource.Resource.GetName(), err.Error())
	} else if expectError && err == nil {
		return fmt.Errorf("expected error when '%s' '%s', but received none", operation, resource.Resource.GetName())
	}
	return nil
}

func ResourceShouldBe(dynamicClient dynamic.Interface, resource unstructuredResource, w common.WaiterConfig, state string) error {
	var (
		exists  bool
		counter int
	)

	if err := validateDynamicClient(dynamicClient); err != nil {
		return err
	}

	gvr, unstruct := resource.GVR, resource.Resource
	for {
		exists = true
		if counter >= w.GetTries() {
			return errors.New("waiter timed out waiting for resource state")
		}
		log.Infof("[KUBEDOG] waiting for resource %v/%v to become %v", unstruct.GetNamespace(), unstruct.GetName(), state)

		_, err := dynamicClient.Resource(gvr.Resource).Namespace(unstruct.GetNamespace()).Get(context.Background(), unstruct.GetName(), metav1.GetOptions{})
		if err != nil {
			if !kerrors.IsNotFound(err) {
				return err
			}
			log.Infof("[KUBEDOG] %v/%v is not found: %v", unstruct.GetNamespace(), unstruct.GetName(), err)
			exists = false
		}

		switch state {
		case common.StateDeleted:
			if !exists {
				log.Infof("[KUBEDOG] %v/%v is deleted", unstruct.GetNamespace(), unstruct.GetName())
				return nil
			}
		case common.StateCreated:
			if exists {
				log.Infof("[KUBEDOG] %v/%v is created", unstruct.GetNamespace(), unstruct.GetName())
				return nil
			}
		}
		counter++
		time.Sleep(w.GetInterval())
	}
}

func ResourceShouldConvergeToSelector(dynamicClient dynamic.Interface, resource unstructuredResource, w common.WaiterConfig, selector string) error {
	var counter int

	if err := validateDynamicClient(dynamicClient); err != nil {
		return err
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

	gvr, unstruct := resource.GVR, resource.Resource

	for {
		if counter >= w.GetTries() {
			return errors.New("waiter timed out waiting for resource")
		}
		//TODO: configure the logger to output "[KUBEDOG]" instead typing it in each log
		log.Infof("[KUBEDOG] waiting for resource %v/%v to converge to %v=%v", unstruct.GetNamespace(), unstruct.GetName(), key, value)
		cr, err := dynamicClient.Resource(gvr.Resource).Namespace(unstruct.GetNamespace()).Get(context.Background(), unstruct.GetName(), metav1.GetOptions{})
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
		time.Sleep(w.GetInterval())
	}

	return nil
}

func ResourceConditionShouldBe(dynamicClient dynamic.Interface, resource unstructuredResource, w common.WaiterConfig, conditionType, conditionValue string) error {
	var (
		counter        int
		expectedStatus = cases.Title(language.English).String(conditionValue)
	)

	if err := validateDynamicClient(dynamicClient); err != nil {
		return err
	}

	gvr, unstruct := resource.GVR, resource.Resource

	for {
		if counter >= w.GetTries() {
			return errors.New("waiter timed out waiting for resource state")
		}
		log.Infof("[KUBEDOG] waiting for resource %v/%v to meet condition %v=%v", unstruct.GetNamespace(), unstruct.GetName(), conditionType, expectedStatus)
		cr, err := dynamicClient.Resource(gvr.Resource).Namespace(unstruct.GetNamespace()).Get(context.Background(), unstruct.GetName(), metav1.GetOptions{})
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
				if condType == conditionType {
					status := condition["status"].(string)
					if corev1.ConditionStatus(status) == corev1.ConditionStatus(expectedStatus) {
						return nil
					}
				}
			}
		}
		counter++
		time.Sleep(w.GetInterval())
	}
}

func UpdateResourceWithField(dynamicClient dynamic.Interface, resource unstructuredResource, key string, value string) error {
	var (
		keySlice     = util.DeleteEmpty(strings.Split(key, "."))
		overrideType bool
		intValue     int64
		//err          error
	)

	if err := validateDynamicClient(dynamicClient); err != nil {
		return err
	}

	gvr, unstruct := resource.GVR, resource.Resource

	n, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		overrideType = true
		intValue = n
	}

	updateTarget, err := dynamicClient.Resource(gvr.Resource).Namespace(unstruct.GetNamespace()).Get(context.Background(), unstruct.GetName(), metav1.GetOptions{})
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

	_, err = dynamicClient.Resource(gvr.Resource).Namespace(unstruct.GetNamespace()).Update(context.Background(), updateTarget, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func DeleteResourcesAtPath(dynamicClient dynamic.Interface, dc discovery.DiscoveryInterface, TemplateArguments interface{}, w common.WaiterConfig, resourcesPath string) error {
	if err := validateDynamicClient(dynamicClient); err != nil {
		return err
	}

	var deleteFn = func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		resources, err := GetResources(dc, TemplateArguments, path)
		if err != nil {
			return err
		}
		for _, resource := range resources {
			gvr, unstruct := resource.GVR, resource.Resource
			err = dynamicClient.Resource(gvr.Resource).Namespace(unstruct.GetNamespace()).Delete(context.Background(), unstruct.GetName(), metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			log.Infof("[KUBEDOG] submitted deletion for %v/%v", unstruct.GetNamespace(), unstruct.GetName())
		}
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

		resources, err := GetResources(dc, TemplateArguments, path)
		if err != nil {
			return err
		}
		for _, resource := range resources {
			gvr, unstruct := resource.GVR, resource.Resource
			for {
				if counter >= w.GetTries() {
					return errors.New("waiter timed out waiting for deletion")
				}
				log.Infof("[KUBEDOG] waiting for resource deletion of %v/%v", unstruct.GetNamespace(), unstruct.GetName())
				_, err := dynamicClient.Resource(gvr.Resource).Namespace(unstruct.GetNamespace()).Get(context.Background(), unstruct.GetName(), metav1.GetOptions{})
				if err != nil {
					if kerrors.IsNotFound(err) {
						log.Infof("[KUBEDOG] resource %v/%v is deleted", unstruct.GetNamespace(), unstruct.GetName())
						break
					}
				}
				counter++
				time.Sleep(w.GetInterval())
			}
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

func VerifyInstanceGroups(dynamicClient dynamic.Interface) error {
	igs, err := ListInstanceGroups(dynamicClient)
	if err != nil {
		return err
	}

	for _, ig := range igs.Items {
		//currentStatus := getInstanceGroupStatus(&ig)
		var currentStatus string
		if val, ok, _ := unstructured.NestedString(ig.UnstructuredContent(), "status", "currentState"); ok {
			currentStatus = val
		}
		if !strings.EqualFold(currentStatus, common.StateReady) {
			return errors.Errorf("Expected Instance Group %s to be ready, but was '%s'", ig.GetName(), currentStatus)
		} else {
			log.Infof("Instance Group %s is ready", ig.GetName())
		}
	}

	return nil
}
