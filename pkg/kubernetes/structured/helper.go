package structured

import (
	"context"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNodes(kubeClientset kubernetes.Interface) (*corev1.NodeList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	nodes, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list nodes")
	}

	return nodes.(*corev1.NodeList), nil
}

func GetDaemonSet(kubeClientset kubernetes.Interface, name, namespace string) (*appsv1.DaemonSet, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	ds, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get daemonset")
	}
	return ds.(*appsv1.DaemonSet), nil
}

func GetDeployment(kubeClientset kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	deploy, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployment")
	}
	return deploy.(*appsv1.Deployment), nil
}

func GetPersistentVolume(kubeClientset kubernetes.Interface, name string) (*corev1.PersistentVolume, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	pvs, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get persistentvolume")
	}
	return pvs.(*corev1.PersistentVolume), nil
}

func ListStatefulSets(kubeClientset kubernetes.Interface, namespace string) (*appsv1.StatefulSetList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	sts, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list statefulsets")
	}
	return sts.(*appsv1.StatefulSetList), nil
}

func ListPersistentVolumes(kubeClientset kubernetes.Interface) (*corev1.PersistentVolumeList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	pvs, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list persistentvolumes")
	}
	return pvs.(*corev1.PersistentVolumeList), nil
}

func GetIngress(kubeClientset kubernetes.Interface, name, namespace string) (*networkingv1.Ingress, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	ingress, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get clusterrolebinding '%v'", name)
	}
	return ingress.(*networkingv1.Ingress), nil
}
