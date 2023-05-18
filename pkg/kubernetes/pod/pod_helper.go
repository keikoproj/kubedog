package pod

import (
	"context"
	"time"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func ListPodsWithLabelSelector(kubeClientset kubernetes.Interface, namespace, selector string) (*corev1.PodList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	pods, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: selector})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	return pods.(*corev1.PodList), nil
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
