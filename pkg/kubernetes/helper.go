package kube

import (
	"context"
	"fmt"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ListPodsWithLabelSelector lists pods with a label selector
func (kc *ClientSet) ListPodsWithLabelSelector(namespace, selector string) (*corev1.PodList, error) {
	pods, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: selector})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list pods")
	}

	return pods.(*corev1.PodList), nil
}

// ListNodes lists nodes
func (kc *ClientSet) ListNodes() (*corev1.NodeList, error) {
	nodes, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list nodes")
	}

	return nodes.(*corev1.NodeList), nil
}

// GetDaemonset gets a daemonset
func (kc *ClientSet) GetDaemonset(name, namespace string) (*appsv1.DaemonSet, error) {
	ds, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get daemonset")
	}
	return ds.(*appsv1.DaemonSet), nil
}

// GetDeployment gets a deployment
func (kc *ClientSet) GetDeployment(name, namespace string) (*appsv1.Deployment, error) {
	deploy, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployment")
	}
	return deploy.(*appsv1.Deployment), nil
}

// GetPersistentVolume gets a pv
func (kc *ClientSet) GetPersistentVolume(name string) (*corev1.PersistentVolume, error) {
	pvs, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get persistentvolume")
	}
	return pvs.(*corev1.PersistentVolume), nil
}

func (kc *ClientSet) ListInstanceGroups() (*unstructured.UnstructuredList, error) {
	const (
		instanceGroupNamespace   = "instance-manager"
		customResourceGroup      = "instancemgr"
		customResourceAPIVersion = "v1alpha1"
		customeResourceDomain    = "keikoproj.io"
		customResourceKind       = "instancegroups"
	)
	var (
		customResourceName    = fmt.Sprintf("%v.%v", customResourceGroup, customeResourceDomain)
		instanceGroupResource = schema.GroupVersionResource{Group: customResourceName, Version: customResourceAPIVersion, Resource: customResourceKind}
	)
	igs, err := kc.DynamicInterface.Resource(instanceGroupResource).Namespace(instanceGroupNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return igs, nil
}

// ListStatefulSets lists statefulsets
func (kc *ClientSet) ListStatefulSets(namespace string) (*appsv1.StatefulSetList, error) {
	sts, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list statefulsets")
	}
	return sts.(*appsv1.StatefulSetList), nil
}

// ListPersistentVolume lists pvs
func (kc *ClientSet) ListPersistentVolumes() (*corev1.PersistentVolumeList, error) {
	pvs, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list persistentvolumes")
	}
	return pvs.(*corev1.PersistentVolumeList), nil
}

func (kc *ClientSet) GetIngress(name, namespace string) (*networkingv1.Ingress, error) {
	ingress, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kc.KubeInterface.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clusterrolebinding '%v'", name)
	}
	return ingress.(*networkingv1.Ingress), nil
}
