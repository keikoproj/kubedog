package pod

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func listPodsWithLabelSelector(KubeClientset kubernetes.Interface, namespace, selector string) (*corev1.PodList, error) {
	if err := validateClientset(KubeClientset); err != nil {
		return nil, err
	}
	pods, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return KubeClientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: selector})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	return pods.(*corev1.PodList), nil
}

func GetPods(KubeClientset kubernetes.Interface, namespace string) error {
	return GetPodsWithSelector(KubeClientset, namespace, "")
}

func GetPodsWithSelector(KubeClientset kubernetes.Interface, namespace, selector string) error {
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
	pods, err := listPodsWithLabelSelector(KubeClientset, namespace, selector)
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

func PodsWithSelectorHaveRestartCountLessThan(KubeClientset kubernetes.Interface, namespace string, selector string, restartCount int) error {
	pods, err := listPodsWithLabelSelector(KubeClientset, namespace, selector)
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

func SomeOrAllPodsInNamespaceWithSelectorHaveStringInLogsSinceTime(KubeClientset kubernetes.Interface, expBackoff wait.Backoff, SomeOrAll, namespace, selector, searchKeyword string, since time.Time) error {
	return util.RetryOnAnyError(&expBackoff, func() error {
		pods, err := listPodsWithLabelSelector(KubeClientset, namespace, selector)
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
			podCount, err := findStringInPodLogs(KubeClientset, pod, since, searchKeyword)
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

func GetExpBackoff(steps int) wait.Backoff {
	return wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2.0,
		Jitter:   0.5,
		Steps:    steps,
		Cap:      10 * time.Minute,
	}
}

func findStringInPodLogs(KubeClientset kubernetes.Interface, pod corev1.Pod, since time.Time, stringsToFind ...string) (int, error) {
	foundCount := 0
	if err := validateClientset(KubeClientset); err != nil {
		return foundCount, err
	}
	var sinceTime metav1.Time = metav1.NewTime(since)
	for _, container := range pod.Spec.Containers {
		podLogOpts := corev1.PodLogOptions{
			SinceTime: &sinceTime,
			Container: container.Name,
		}

		req := KubeClientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
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

func SomePodsInNamespaceWithSelectorDontHaveStringInLogsSinceTime(KubeClientset kubernetes.Interface, namespace, selector, searchkeyword string, since time.Time) error {
	pods, err := listPodsWithLabelSelector(KubeClientset, namespace, selector)
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return errors.Errorf("No pods matched selector '%s'", selector)
	}
	for _, pod := range pods.Items {
		count, err := findStringInPodLogs(KubeClientset, pod, since, searchkeyword)
		if err != nil {
			return err
		}
		if count == 0 {
			return nil
		}
	}
	return fmt.Errorf("pod has '%s' message in the logs", searchkeyword)
}

func PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime(KubeClientset kubernetes.Interface, namespace string, selector string, since time.Time) error {
	pods, err := listPodsWithLabelSelector(KubeClientset, namespace, selector)
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return errors.Errorf("No pods matched selector '%s'", selector)
	}

	for _, pod := range pods.Items {
		errorStrings := []string{`"level":"error"`, "level=error"}
		count, err := findStringInPodLogs(KubeClientset, pod, since, errorStrings...)
		if err != nil {
			return err
		}
		if count != 0 {
			return errors.Errorf("Pod %s has %d errors", pod.Name, count)
		}
	}

	return nil
}

func PodsInNamespaceWithSelectorHaveSomeErrorsInLogsSinceTime(KubeClientset kubernetes.Interface, namespace string, selector string, since time.Time) error {
	err := PodsInNamespaceWithSelectorHaveNoErrorsInLogsSinceTime(KubeClientset, namespace, selector, since)
	if err == nil {
		return fmt.Errorf("logs found from selector %q in namespace %q have errors", selector, namespace)
	}
	return nil
}

func PodInNamespaceShouldHaveLabels(KubeClientset kubernetes.Interface, name, namespace, labels string) error {
	if err := validateClientset(KubeClientset); err != nil {
		return err
	}

	pod, err := KubeClientset.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
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

func PodsInNamespaceWithSelectorShouldHaveLabels(KubeClientset kubernetes.Interface, namespace, selector, labels string) error {
	podList, err := listPodsWithLabelSelector(KubeClientset, namespace, selector)
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

func validateClientset(KubeClientset kubernetes.Interface) error {
	if KubeClientset == nil {
		return errors.Errorf("'k8s.io/client-go/kubernetes.Interface' is nil.")
	}
	return nil
}
