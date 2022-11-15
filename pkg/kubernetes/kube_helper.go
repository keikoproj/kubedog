package kube

import (
	"context"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPodsWithLabelSelector lists pods with a label selector
func (kc *Client) ListPodsWithLabelSelector(namespace, selector string) (*corev1.PodList, error) {
	pods, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: selector})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	return pods.(*corev1.PodList), nil
}

// ListNodes lists nodes
func (kc *Client) ListNodes() (*corev1.NodeList, error) {
	nodes, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list nodes")
	}

	return nodes.(*corev1.NodeList), nil
}

// GetDaemonset gets a daemonset
func (kc *Client) GetDaemonset(name, namespace string) (*appsv1.DaemonSet, error) {
	ds, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get daemonset")
	}
	return ds.(*appsv1.DaemonSet), nil
}

// GetDeployment gets a deployment
func (kc *Client) GetDeployment(name, namespace string) (*appsv1.Deployment, error) {
	deploy, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployment")
	}
	return deploy.(*appsv1.Deployment), nil
}

// GetPersistentVolume gets a pv
func (kc *Client) GetPersistentVolume(name string) (*corev1.PersistentVolume, error) {
	pvs, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get persistentvolume")
	}
	return pvs.(*corev1.PersistentVolume), nil
}
