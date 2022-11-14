package common

import (
	"context"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
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

// GetDaemonset gets a daemonset
func GetDaemonset(client kubernetes.Interface, name, namespace string) (*appsv1.DaemonSet, error) {
	ds, err := RetryOnError(&DefaultRetry, IsRetriable, func() (interface{}, error) {
		return client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get daemonset")
	}
	return ds.(*appsv1.DaemonSet), nil
}

// GetDeployment gets a deployment
func GetDeployment(client kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	deploy, err := RetryOnError(&DefaultRetry, IsRetriable, func() (interface{}, error) {
		return client.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployment")
	}
	return deploy.(*appsv1.Deployment), nil
}

// GetPersistentVolume gets a pv
func GetPersistentVolume(client kubernetes.Interface, name string) (*corev1.PersistentVolume, error) {
	pvs, err := RetryOnError(&DefaultRetry, IsRetriable, func() (interface{}, error) {
		return client.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get persistentvolume")
	}
	return pvs.(*corev1.PersistentVolume), nil
}
