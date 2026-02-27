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

package structured

import (
	"context"
	"fmt"
	"time"

	"github.com/keikoproj/kubedog/internal/util"
	"github.com/keikoproj/kubedog/pkg/kube/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetNodeList(kubeClientset kubernetes.Interface) (*corev1.NodeList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	nodes, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list nodes")
	}

	result, ok := nodes.(*corev1.NodeList)
	if !ok {
		return nil, errors.Errorf("failed to list nodes: unexpected type '%T'", nodes)
	}
	return result, nil
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
	result, ok := ds.(*appsv1.DaemonSet)
	if !ok {
		return nil, errors.Errorf("failed to get daemonset: unexpected type '%T'", ds)
	}
	return result, nil
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
	result, ok := deploy.(*appsv1.Deployment)
	if !ok {
		return nil, errors.Errorf("failed to get deployment: unexpected type '%T'", deploy)
	}
	return result, nil
}

func GetConfigMap(kubeClientset kubernetes.Interface, name, namespace string) (*corev1.ConfigMap, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	result, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get configmap")
	}
	configmap, ok := result.(*corev1.ConfigMap)
	if !ok {
		return nil, errors.Errorf("failed to get configmap: expected type '*corev1.ConfigMap' but got '%T'", result)
	}
	if configmap.Name != name {
		return nil, errors.Errorf("failed to get configmap: expected name '%s' but got '%s'", name, configmap.Name)
	}
	return configmap, nil
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
	result, ok := pvs.(*corev1.PersistentVolume)
	if !ok {
		return nil, errors.Errorf("failed to get persistentvolume: unexpected type '%T'", pvs)
	}
	return result, nil
}

func GetPersistentVolumeClaim(kubeClientset kubernetes.Interface, name string, namespace string) (*corev1.PersistentVolumeClaim, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	pvc, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get persistentvolumeclaim")
	}
	result, ok := pvc.(*corev1.PersistentVolumeClaim)
	if !ok {
		return nil, errors.Errorf("failed to get persistentvolumeclaim: unexpected type '%T'", pvc)
	}
	return result, nil
}

func GetStatefulSetList(kubeClientset kubernetes.Interface, namespace string) (*appsv1.StatefulSetList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	sts, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.AppsV1().StatefulSets(namespace).List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list statefulsets")
	}
	result, ok := sts.(*appsv1.StatefulSetList)
	if !ok {
		return nil, errors.Errorf("failed to list statefulsets: unexpected type '%T'", sts)
	}
	return result, nil
}

func GetPersistentVolumeList(kubeClientset kubernetes.Interface) (*corev1.PersistentVolumeList, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	pvs, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list persistentvolumes")
	}
	result, ok := pvs.(*corev1.PersistentVolumeList)
	if !ok {
		return nil, errors.Errorf("failed to list persistentvolumes: unexpected type '%T'", pvs)
	}
	return result, nil
}

func GetIngress(kubeClientset kubernetes.Interface, name, namespace string) (*networkingv1.Ingress, error) {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return nil, err
	}

	ingress, err := util.RetryOnError(&util.DefaultRetry, util.IsRetriable, func() (interface{}, error) {
		return kubeClientset.NetworkingV1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get ingress '%v'", name)
	}
	result, ok := ingress.(*networkingv1.Ingress)
	if !ok {
		return nil, errors.Errorf("failed to get ingress '%v': unexpected type '%T'", name, ingress)
	}
	return result, nil
}

// TODO: remove use of service.beta.kubernetes.io/aws-load-balancer-subnets or make generic
func GetIngressEndpoint(kubeClientset kubernetes.Interface, w common.WaiterConfig, name, namespace string, port int, path string) (string, error) {
	var (
		counter int
	)
	for {
		log.Info("waiting for ingress availability")
		if counter >= w.GetTries() {
			return "", errors.New("waiter timed out waiting for resource state")
		}
		ingress, err := GetIngress(kubeClientset, name, namespace)
		if err != nil {
			return "", err
		}
		annotations := ingress.GetAnnotations()
		albSubnets := annotations["service.beta.kubernetes.io/aws-load-balancer-subnets"]
		log.Infof("Alb IngressSubnets associated are: %v", albSubnets)
		var ingressReconciled bool
		ingressStatus := ingress.Status.LoadBalancer.Ingress
		if ingressStatus == nil {
			log.Infof("ingress %v/%v is not ready yet", namespace, name)
		} else {
			ingressReconciled = true
		}
		if ingressReconciled {
			hostname := ingressStatus[0].Hostname
			endpoint := fmt.Sprintf("http://%v:%v%v", hostname, port, path)
			return endpoint, nil
		}
		counter++
		time.Sleep(w.GetInterval())
	}
}

// TODO: This is hardcoded based on prometheus names in IKS clusters. Might be worth making it more generic in the future
func validatePrometheusPVLabels(kubeClientset kubernetes.Interface, volumeClaimTemplatesName string) error {
	// Get prometheus PersistentVolume list
	pv, err := GetPersistentVolumeList(kubeClientset)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range pv.Items {
		pvcname := item.Spec.ClaimRef.Name
		if pvcname == volumeClaimTemplatesName+"-prometheus-k8s-prometheus-0" || pvcname == volumeClaimTemplatesName+"-prometheus-k8s-prometheus-1" {
			if k1, k2 := item.Labels["failure-domain.beta.kubernetes.io/zone"], item.Labels["topology.kubernetes.io/zone"]; k1 == "" && k2 == "" {
				return errors.Errorf("Prometheus volumes does not exist label - kubernetes.io/zone")
			}
		}
	}
	return nil
}

func isNodeReady(n corev1.Node) bool {
	for _, condition := range n.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				return true
			}
		}
	}
	return false
}
