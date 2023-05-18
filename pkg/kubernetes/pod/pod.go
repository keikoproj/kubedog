package pod

import (
	"context"
	"fmt"
	"strings"
	"time"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func Pods(kubeClientset kubernetes.Interface, namespace string) error {
	return PodsWithSelector(kubeClientset, namespace, "")
}

func PodsWithSelector(kubeClientset kubernetes.Interface, namespace, selector string) error {
	var readyCount = func(conditions []corev1.ContainerStatus) string {
		var readyCount = 0
		var containerCount = len(conditions)
		for _, condition := range conditions {
			if condition.Ready {
				readyCount++
			}
		}
		return fmt.Sprintf("%d/%d", readyCount, containerCount)
	}
	pods, err := ListPodsWithLabelSelector(kubeClientset, namespace, selector)
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return errors.Errorf("No pods matched selector '%s'", selector)
	}
	tableFormat := "%-64s%-12s%-24s"
	log.Infof(tableFormat, "NAME", "READY", "STATUS")
	for _, pod := range pods.Items {
		log.Infof(tableFormat, pod.Name, readyCount(pod.Status.ContainerStatuses), pod.Status.Phase)
	}
	return nil
}

func PodsWithSelectorHaveRestartCountLessThan(kubeClientset kubernetes.Interface, namespace string, selector string, restartCount int) error {
	pods, err := ListPodsWithLabelSelector(kubeClientset, namespace, selector)
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return errors.Errorf("No pods matched selector '%s'", selector)
	}

	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			log.Infof("Container '%s' of pod '%s' on node '%s' restarted %d times",
				containerStatus.Name, pod.Name, pod.Spec.NodeName, containerStatus.RestartCount)
			if int(containerStatus.RestartCount) >= restartCount {
				return errors.Errorf("Container '%s' of pod '%s' restarted %d times",
					containerStatus.Name, pod.Name, containerStatus.RestartCount)
			}
		}
	}

	return nil
}

func SomeOrAllPodsInNamespaceWithSelectorHaveStringInLogsSinceTime(kubeClientset kubernetes.Interface, expBackoff wait.Backoff, SomeOrAll, namespace, selector, searchKeyword string, since time.Time) error {
	return util.RetryOnAnyError(&expBackoff, func() error {
		pods, err := ListPodsWithLabelSelector(kubeClientset, namespace, selector)
		if err != nil {
			return err
		}
		if len(pods.Items) == 0 {
			return fmt.Errorf("no pods matched selector '%s'", selector)
		}

		const (
			somePodsKeyword = "some"
			allPodsKeyword  = "all"
		)
		var podsCount int
		for _, pod := range pods.Items {
			podCount, err := countStringInPodLogs(kubeClientset, pod, since, searchKeyword)
			if err != nil {
				return err
			}
			podsCount += podCount
			switch SomeOrAll {
			case somePodsKeyword:
				if podCount != 0 {
					log.Infof("'%s' pods required to have string in logs. pod '%s' has string '%s' in logs", somePodsKeyword, pod.Name, searchKeyword)
					return nil
				}
			case allPodsKeyword:
				if podCount == 0 {
					return fmt.Errorf("'%s' pods required to have string in logs. pod '%s' does not have string '%s' in logs", allPodsKeyword, pod.Name, searchKeyword)
				}
			default:
				return fmt.Errorf("wrong input as '%s', expected '(%s|%s)'", SomeOrAll, somePodsKeyword, allPodsKeyword)
			}
		}
		if podsCount == 0 {
			return fmt.Errorf("pods in namespace '%s' with selector '%s' do not have string '%s' in logs", namespace, selector, searchKeyword)
		}
		return nil
	})
}

func SomePodsInNamespaceWithSelectorDontHaveStringInLogsSinceTime(kubeClientset kubernetes.Interface, namespace, selector, searchkeyword string, since time.Time) error {
	pods, err := ListPodsWithLabelSelector(kubeClientset, namespace, selector)
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return errors.Errorf("No pods matched selector '%s'", selector)
	}
	for _, pod := range pods.Items {
		count, err := countStringInPodLogs(kubeClientset, pod, since, searchkeyword)
		if err != nil {
			return err
		}
		if count == 0 {
			return nil
		}
	}
	return fmt.Errorf("pod has '%s' message in the logs", searchkeyword)
}

func PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime(kubeClientset kubernetes.Interface, namespace string, selector string, since time.Time) error {
	pods, err := ListPodsWithLabelSelector(kubeClientset, namespace, selector)
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return errors.Errorf("No pods matched selector '%s'", selector)
	}

	for _, pod := range pods.Items {
		errorStrings := []string{`"level":"error"`, "level=error"}
		count, err := countStringInPodLogs(kubeClientset, pod, since, errorStrings...)
		if err != nil {
			return err
		}
		if count != 0 {
			return errors.Errorf("Pod %s has %d errors", pod.Name, count)
		}
	}

	return nil
}

func PodsInNamespaceWithSelectorHaveSomeErrorsInLogsSinceTime(kubeClientset kubernetes.Interface, namespace string, selector string, since time.Time) error {
	err := PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime(kubeClientset, namespace, selector, since)
	if err == nil {
		return fmt.Errorf("logs found from selector %q in namespace %q have errors", selector, namespace)
	}
	return nil
}

func PodInNamespaceShouldHaveLabels(kubeClientset kubernetes.Interface, name, namespace, labels string) error {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return err
	}

	pod, err := kubeClientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return errors.New("Error fetching pod: " + err.Error())
	}

	inputLabels := make(map[string]string)
	slc := strings.Split(labels, ",")
	for _, item := range slc {
		vals := strings.Split(item, "=")
		if len(vals) != 2 {
			continue
		}

		inputLabels[vals[0]] = vals[1]
	}

	for k, v := range inputLabels {
		pV, ok := pod.Labels[k]
		if !ok {
			return errors.New(fmt.Sprintf("Label %s missing in pod/namespace %s", k, name+"/"+namespace))
		}
		if v != pV {
			return errors.New(fmt.Sprintf("Label value %s doesn't match expected %s for key %s in pod/namespace %s", pV, v, k, name+"/"+namespace))
		}
	}

	return nil
}

func PodsInNamespaceWithSelectorShouldHaveLabels(kubeClientset kubernetes.Interface, namespace, selector, labels string) error {
	podList, err := ListPodsWithLabelSelector(kubeClientset, namespace, selector)
	if err != nil {
		return fmt.Errorf("error getting pods with selector %q: %v", selector, err)
	}

	if len(podList.Items) == 0 {
		return fmt.Errorf("no pods matched selector '%s'", selector)
	}

	for _, pod := range podList.Items {
		inputLabels := make(map[string]string)
		slc := strings.Split(labels, ",")
		for _, item := range slc {
			vals := strings.Split(item, "=")
			if len(vals) != 2 {
				continue
			}

			inputLabels[vals[0]] = vals[1]
		}

		for k, v := range inputLabels {
			pV, ok := pod.Labels[k]
			if !ok {
				return fmt.Errorf("label %s missing in pod/namespace %s", k, pod.Name+"/"+namespace)
			}
			if v != pV {
				return fmt.Errorf("label value %s doesn't match expected %s for key %s in pod/namespace %s", pV, v, k, pod.Name+"/"+namespace)
			}
		}
	}

	return nil
}
