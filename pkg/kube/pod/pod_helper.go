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

package pod

import (
	"bufio"
	"context"
	"strings"
	"time"

	"github.com/keikoproj/kubedog/internal/util"
	"github.com/keikoproj/kubedog/pkg/kube/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodListWithLabelSelector(kubeClientset kubernetes.Interface, namespace, labelSelector string) (*corev1.PodList, error) {
	return GetPodListWithLabelSelectorAndFieldSelector(kubeClientset, namespace, labelSelector, "")
}

func GetPodListWithLabelSelectorAndFieldSelector(kubeClientset kubernetes.Interface, namespace, labelSelector, fieldSelector string) (*corev1.PodList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	pods, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: labelSelector, FieldSelector: fieldSelector})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	return pods.(*corev1.PodList), nil
}

func DeletePodListWithLabelSelector(kubeClientset kubernetes.Interface, namespace, labelSelector string) error {
	return DeletePodListWithLabelSelectorAndFieldSelector(kubeClientset, namespace, labelSelector, "")
}

func DeletePodListWithLabelSelectorAndFieldSelector(kubeClientset kubernetes.Interface, namespace, labelSelector, fieldSelector string) error {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return err
	}

	err := kubeClientset.CoreV1().Pods(namespace).DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labelSelector,
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to delete pods with label selector %s and field selector %s in namespace %s", labelSelector, fieldSelector, namespace)
	}
	return nil
}

func countStringInPodLogs(kubeClientset kubernetes.Interface, pod corev1.Pod, since time.Time, stringsToFind ...string) (int, error) {
	foundCount := 0
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return foundCount, err
	}
	var sinceTime metav1.Time = metav1.NewTime(since)
	for _, container := range pod.Spec.Containers {
		podLogOpts := corev1.PodLogOptions{
			SinceTime: &sinceTime,
			Container: container.Name,
		}

		req := kubeClientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
		podLogs, err := req.Stream(context.Background())
		if err != nil {
			return 0, errors.Errorf("Error in opening stream for pod '%s', container '%s' : '%s'", pod.Name, container.Name, string(err.Error()))
		}

		scanner := bufio.NewScanner(podLogs)
		for scanner.Scan() {
			line := scanner.Text()
			for _, stringToFind := range stringsToFind {
				if strings.Contains(line, stringToFind) {
					foundCount += 1
					log.Infof("Found string '%s' in line '%s' in container '%s' of pod '%s'", stringToFind, line, container.Name, pod.Name)
				}
			}
		}
		podLogs.Close()
	}
	return foundCount, nil
}
