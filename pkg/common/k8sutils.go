package common

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListPodsWithLabelSelector lists pods with a label selector
func ListPodsWithLabelSelector(client kubernetes.Interface, namespace, selector string) (*corev1.PodList, error) {
	pods, err := RetryOnError(&DefaultRetry, IsRetriable, func() (interface{}, error) {
		return client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: selector})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	return pods.(*corev1.PodList), nil
}

// ListNodes lists nodes
func ListNodes(client kubernetes.Interface) (*corev1.NodeList, error) {
	nodes, err := RetryOnError(&DefaultRetry, IsRetriable, func() (interface{}, error) {
		return client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list nodes")
	}

	return nodes.(*corev1.NodeList), nil
}
