package pod

import (
	"bufio"
	"context"
	"strings"
	"time"

	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
