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
	"net/http"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/keikoproj/kubedog/internal/util"
	"github.com/keikoproj/kubedog/pkg/kube/common"
	"github.com/keikoproj/kubedog/pkg/kube/pod"
	"github.com/pkg/errors"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

func NodesWithSelectorShouldBe(kubeClientset kubernetes.Interface, w common.WaiterConfig, expectedNodes int, labelSelector, state string) error {
	var (
		counter int
		found   bool
	)

	if err := common.ValidateClientset(kubeClientset); err != nil {
		return err
	}

	for {
		var (
			nodesCount int
			opts       = metav1.ListOptions{
				LabelSelector: labelSelector,
			}
		)

		if counter >= w.GetTries() {
			return errors.New("waiter timed out waiting for nodes")
		}

		nodes, err := kubeClientset.CoreV1().Nodes().List(context.Background(), opts)
		if err != nil {
			return err
		}

		switch state {
		case common.StateFound:
			nodesCount = len(nodes.Items)
			if nodesCount == expectedNodes {
				log.Infof("found %v nodes", expectedNodes)
				found = true
			}
		case common.StateReady:
			for _, node := range nodes.Items {
				if isNodeReady(node) {
					nodesCount++
				}
			}
			if nodesCount == expectedNodes {
				log.Infof("found %v ready nodes", expectedNodes)
				found = true
			}
		}

		if found {
			break
		}

		log.Infof("found %v nodes, waiting for %v nodes to be %v with selector %v", nodesCount, expectedNodes, state, labelSelector)

		counter++
		time.Sleep(w.GetInterval())
	}
	return nil
}

func ScaleDeployment(kubeClientset kubernetes.Interface, name, namespace string, replicas int32) error {
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return err
	}

	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replicas,
		},
	}

	_, err := kubeClientset.AppsV1().Deployments(namespace).UpdateScale(context.Background(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func ClusterRbacIsFound(kubeClientset kubernetes.Interface, resourceType, name string) error {
	var err error
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return err
	}

	switch resourceType {
	case "clusterrole":
		_, err = kubeClientset.RbacV1().ClusterRoles().Get(context.Background(), name, metav1.GetOptions{})
	case "clusterrolebinding":
		_, err = kubeClientset.RbacV1().ClusterRoleBindings().Get(context.Background(), name, metav1.GetOptions{})
	default:
		return errors.Errorf("Invalid resource type")
	}

	if err != nil {
		return err
	}
	return nil
}

func ListNodes(kubeClientset kubernetes.Interface) error {

	var readyStatus = func(conditions []corev1.NodeCondition) string {
		var status = false
		var err error
		for _, condition := range conditions {
			if condition.Type == "Ready" {
				status, err = strconv.ParseBool(string(condition.Status))
				if err != nil {
					return "Unknown"
				}
				break
			}
		}
		if status {
			return "Ready"
		}
		return "NotReady"
	}
	// List nodes
	nodes, _ := GetNodeList(kubeClientset)
	if nodes != nil {
		tableFormat := "%-64s%-12s%-24s%-16s"
		log.Infof(tableFormat, "NAME", "STATUS", "INSTANCEGROUP", "AZ")
		for _, node := range nodes.Items {
			log.Infof(tableFormat,
				node.Name,
				readyStatus(node.Status.Conditions),
				node.Labels["node.kubernetes.io/instancegroup"],
				node.Labels["failure-domain.beta.kubernetes.io/zone"])
		}
	}
	return nil
}

func DaemonSetIsRunning(kubeClientset kubernetes.Interface, expBackoff wait.Backoff, name, namespace string) error {
	err := util.RetryOnAnyError(&expBackoff, func() error {
		ds, err := GetDaemonSet(kubeClientset, name, namespace)
		if err != nil {
			return err
		}

		if ds.Status.DesiredNumberScheduled != ds.Status.CurrentNumberScheduled {
			return fmt.Errorf("daemonset '%s/%s' not updated. status: '%s'", namespace, name, ds.Status.String())
		}

		return nil
	})
	if err != nil {
		// Print Pods after failure
		_ = pod.ListPods(kubeClientset, namespace)
		return fmt.Errorf("daemonset '%s/%s' not updated: '%v'", namespace, name, err)
	}
	return nil
}

func DeploymentIsRunning(kubeClientset kubernetes.Interface, name, namespace string) error {
	deploy, err := GetDeployment(kubeClientset, name, namespace)
	if err != nil {
		return err
	}
	if deploy.Status.ReadyReplicas != deploy.Status.Replicas {
		return fmt.Errorf("deployment %s/%s is not ready. status: %s", namespace, name, deploy.Status.String())
	}

	if deploy.Status.UpdatedReplicas != deploy.Status.Replicas {
		return fmt.Errorf("deployment %s/%s is not updated. status: %s", namespace, name, deploy.Status.String())
	}

	return nil
}

func PersistentVolExists(kubeClientset kubernetes.Interface, name, expectedPhase string) error {
	vol, err := GetPersistentVolume(kubeClientset, name)
	if err != nil {
		return err
	}
	phase := string(vol.Status.Phase)
	if phase != expectedPhase {
		return fmt.Errorf("persistentvolume had unexpected phase %v, expected phase %v", phase, expectedPhase)
	}
	return nil
}

func ValidatePrometheusVolumeClaimTemplatesName(kubeClientset kubernetes.Interface, statefulsetName, namespace, volumeClaimTemplatesName string) error {
	// Prometheus StatefulSets deployed, then validate volumeClaimTemplate name.
	// Validation required:
	// 	- To retain existing persistent volumes and not to loose any data.
	//	- And avoid creating new name persistent volumes.
	sfs, err := GetStatefulSetList(kubeClientset, namespace)
	if err != nil {
		return err
	}

	var sfsvolumeClaimTemplatesNames []string
	for _, sfsItem := range sfs.Items {
		if sfsItem.Name == statefulsetName {
			pvcClaimRefs := sfsItem.Spec.VolumeClaimTemplates
			for _, pvcClaimRef := range pvcClaimRefs {
				if pvcClaimRef.Name != "" {
					sfsvolumeClaimTemplatesNames = append(sfsvolumeClaimTemplatesNames, pvcClaimRef.Name)
				}
			}
		}
	}
	if len(sfsvolumeClaimTemplatesNames) == 0 {
		return errors.Errorf("StatefulSet '%s' had no VolumeClaimTemplates with non empty name", statefulsetName)
	}

	found := false
	for _, sfsvolumeClaimTemplatesName := range sfsvolumeClaimTemplatesNames {
		if sfsvolumeClaimTemplatesName == volumeClaimTemplatesName {
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf("StatefulSet '%s' had no VolumeClaimTemplates with name '%s'", statefulsetName, volumeClaimTemplatesName)
	}

	// Validate Persistent Volume label
	err = validatePrometheusPVLabels(kubeClientset, volumeClaimTemplatesName)
	if err != nil {
		return err
	}
	return nil
}

func SecretDelete(kubeClientset kubernetes.Interface, name, namespace string) error {
	return SecretOperationFromEnvironmentVariable(kubeClientset, common.OperationDelete, name, namespace, "")
}

func SecretOperationFromEnvironmentVariable(kubeClientset kubernetes.Interface, operation, name, namespace, environmentVariable string) error {
	var (
		secretValue string
		ok          bool
	)
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return err
	}
	if operation != common.OperationDelete {
		secretValue, ok = os.LookupEnv(environmentVariable)
		if !ok {
			return errors.Errorf("couldn't lookup environment variable '%s'", environmentVariable)
		}
	}
	switch operation {
	case common.OperationCreate, common.OperationSubmit:
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Data: map[string][]byte{
				environmentVariable: []byte(secretValue),
			},
		}
		_, err := kubeClientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if kerrors.IsAlreadyExists(err) {
			return fmt.Errorf("secret '%s' already created", name)
		}
		return err
	case common.OperationUpdate:
		currentSecret, err := kubeClientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		secret := currentSecret.DeepCopy()
		if len(secret.Data) == 0 {
			secret.Data = map[string][]byte{}
		}
		secret.Data[environmentVariable] = []byte(secretValue)
		_, err = kubeClientset.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
		return err
	case common.OperationDelete:
		err := kubeClientset.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if kerrors.IsNotFound(err) {
			log.Infof("secret '%s' was not found", name)
			return nil
		}
		return err
	default:
		return fmt.Errorf("unsupported operation: '%s'", operation)
	}
}

func IngressAvailable(kubeClientset kubernetes.Interface, w common.WaiterConfig, name, namespace string, port int, path string) error {
	var (
		counter int
	)
	endpoint, err := GetIngressEndpoint(kubeClientset, w, name, namespace, port, path)
	if err != nil {
		return err
	}
	for {
		log.Info("waiting for ingress availability")
		if counter >= w.GetTries() {
			return errors.New("waiter timed out waiting for resource state")
		}
		log.Infof("waiting for endpoint %v to become available", endpoint)
		client := http.Client{
			Timeout: 10 * time.Second,
		}
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		if resp, err := client.Do(req); resp != nil {
			if resp.StatusCode == 200 {
				log.Infof("endpoint %v is available", endpoint)
				time.Sleep(w.GetInterval())
				return nil
			}
		} else {
			log.Infof("endpoint %v is not available yet: %v", endpoint, err)
		}
		counter++
		time.Sleep(w.GetInterval())
	}
}

func SendTrafficToIngress(kubeClientset kubernetes.Interface, w common.WaiterConfig, tps int, name, namespace string, port int, path string, duration int, durationUnits string, expectedErrors int) error {
	endpoint, err := GetIngressEndpoint(kubeClientset, w, name, namespace, port, path)
	if err != nil {
		return err
	}
	log.Infof("sending traffic to %v with rate of %v tps for %v %s...", endpoint, tps, duration, durationUnits)
	rate := vegeta.Rate{Freq: tps, Per: time.Second}
	var d time.Duration
	switch durationUnits {
	case util.DurationMinutes:
		d = time.Minute * time.Duration(duration)
	case util.DurationSeconds:
		d = time.Second * time.Duration(duration)
	default:
		return fmt.Errorf("unsupported duration units: '%s'", durationUnits)
	}
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    endpoint,
	})
	attacker := vegeta.NewAttacker()
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, d, namespace+"/"+name) {
		metrics.Add(res)
	}
	metrics.Close()
	if len(metrics.Errors) > expectedErrors {
		return errors.Errorf("traffic test had '%v' errors but expected '%d'", metrics.Errors, expectedErrors)
	}
	return nil
}

func ResourceInNamespace(kubeClientset kubernetes.Interface, resourceType, name, namespace string) error {
	var err error
	if err := common.ValidateClientset(kubeClientset); err != nil {
		return err
	}

	switch resourceType {
	case "deployment":
		_, err = kubeClientset.AppsV1().Deployments(namespace).Get(context.Background(), name, metav1.GetOptions{})
	case "service":
		_, err = kubeClientset.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
	case "hpa", "horizontalpodautoscaler":
		_, err = kubeClientset.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(context.Background(), name, metav1.GetOptions{})
	case "pdb", "poddisruptionbudget":
		_, err = kubeClientset.PolicyV1beta1().PodDisruptionBudgets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	case "sa", "serviceaccount":
		_, err = kubeClientset.CoreV1().ServiceAccounts(namespace).Get(context.Background(), name, metav1.GetOptions{})
	default:
		return errors.Errorf("Invalid resource type")
	}
	return err

}

func ResourceNotInNamespace(kubeClientset kubernetes.Interface, resourceType, name, namespace string) error {
	err := ResourceInNamespace(kubeClientset, resourceType, name, namespace)
	if err == nil {
		return errors.Errorf("expected resource '%s/%s' to not be found in ns '%s'", resourceType, name, namespace)
	}
	if !kerrors.IsNotFound(err) {
		return err
	}
	return nil
}
