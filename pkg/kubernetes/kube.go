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

// Package kube provides steps implementations related to Kubernetes.
package kube

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/keikoproj/kubedog/pkg/common"
	"github.com/pkg/errors"
	vegeta "github.com/tsenart/vegeta/v12/lib"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (kc *ClientSet) KubernetesClusterShouldBe(state string) error {
	if err := kc.Validate(); err != nil {
		return err
	}
	switch state {
	case stateCreated, stateUpgraded:
		if _, err := kc.KubeInterface.CoreV1().Pods(metav1.NamespaceSystem).List(context.TODO(), metav1.ListOptions{}); err != nil {
			return err
		}
		return nil
	case stateDeleted:
		if err := kc.DiscoverClients(); err == nil {
			return errors.New("failed validating cluster delete, cluster is still available")
		}
		return nil
	default:
		return fmt.Errorf("unsupported state: '%s'", state)
	}
}

func (kc *ClientSet) NodesWithSelectorShouldBe(n int, selector, state string) error {
	var (
		counter int
		found   bool
	)

	if err := kc.Validate(); err != nil {
		return err
	}

	for {
		var (
			nodesCount int
			opts       = metav1.ListOptions{
				LabelSelector: selector,
			}
		)

		if counter >= kc.getWaiterTries() {
			return errors.New("waiter timed out waiting for nodes")
		}

		nodes, err := kc.KubeInterface.CoreV1().Nodes().List(context.Background(), opts)
		if err != nil {
			return err
		}

		switch state {
		case stateFound:
			nodesCount = len(nodes.Items)
			if nodesCount == n {
				log.Infof("[KUBEDOG] found %v nodes", n)
				found = true
			}
		case stateReady:
			for _, node := range nodes.Items {
				if util.IsNodeReady(node) {
					nodesCount++
				}
			}
			if nodesCount == n {
				log.Infof("[KUBEDOG] found %v ready nodes", n)
				found = true
			}
		}

		if found {
			break
		}

		log.Infof("[KUBEDOG] found %v nodes, waiting for %v nodes to be %v with selector %v", nodesCount, n, state, selector)

		counter++
		time.Sleep(kc.getWaiterInterval())
	}
	return nil
}

func (kc *ClientSet) ScaleDeployment(name, ns string, replica int32) error {
	if err := kc.Validate(); err != nil {
		return err
	}

	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replica,
		},
	}

	_, err := kc.KubeInterface.AppsV1().Deployments(ns).UpdateScale(context.Background(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (kc *ClientSet) ClusterRbacIsFound(resource, name string) error {
	var err error
	if err := kc.Validate(); err != nil {
		return err
	}

	switch resource {
	case "clusterrole":
		_, err = kc.KubeInterface.RbacV1().ClusterRoles().Get(context.Background(), name, metav1.GetOptions{})
	case "clusterrolebinding":
		_, err = kc.KubeInterface.RbacV1().ClusterRoleBindings().Get(context.Background(), name, metav1.GetOptions{})
	default:
		return errors.Errorf("Invalid resource type")
	}

	if err != nil {
		return err
	}
	return nil
}

func (kc *ClientSet) GetNodes() error {

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
	nodes, _ := kc.ListNodes()
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

// TODO: export it and use it instead of ResourceIsRunning
func (kc *ClientSet) daemonsetIsRunning(dsName, namespace string) error {
	// TODO: implement this differently, remove gomega
	gomega.Eventually(func() error {
		ds, err := kc.GetDaemonset(dsName, namespace)
		if err != nil {
			return err
		}

		if ds.Status.DesiredNumberScheduled != ds.Status.CurrentNumberScheduled {
			return fmt.Errorf("daemonset %s/%s is not updated. status: %s", namespace, dsName, ds.Status.String())
		}

		return nil
	}, 10*time.Second).Should(gomega.Succeed(), func() string {
		// Print Pods after failure
		_ = kc.GetPods(namespace)
		return fmt.Sprintf("daemonset %s/%s is not updated.", namespace, dsName)
	})

	return nil
}

// TODO: export it and use it instead of ResourceIsRunning
func (kc *ClientSet) deploymentIsRunning(deployName, namespace string) error {
	deploy, err := kc.GetDeployment(deployName, namespace)
	if err != nil {
		return err
	}
	if deploy.Status.ReadyReplicas != deploy.Status.Replicas {
		return fmt.Errorf("deployment %s/%s is not ready. status: %s", namespace, deployName, deploy.Status.String())
	}

	if deploy.Status.UpdatedReplicas != deploy.Status.Replicas {
		return fmt.Errorf("deploymemnt %s/%s is not updated. status: %s", namespace, deployName, deploy.Status.String())
	}

	return nil
}

func (kc *ClientSet) ResourceIsRunning(kind, name, namespace string) error {
	kind = strings.ToLower(kind)
	switch kind {
	case "daemonset":
		return kc.daemonsetIsRunning(name, namespace)
	case "deployment":
		return kc.deploymentIsRunning(name, namespace)
	default:
		return fmt.Errorf("invalid resource type: %s", kind)
	}
}

func (kc *ClientSet) PersistentVolExists(volName, expectedPhase string) error {
	vol, err := kc.GetPersistentVolume(volName)
	if err != nil {
		return err
	}
	phase := string(vol.Status.Phase)
	if phase != expectedPhase {
		return fmt.Errorf("persistentvolume had unexpected phase %v, expected phase %v", phase, expectedPhase)
	}
	return nil
}

func (kc *ClientSet) VerifyInstanceGroups() error {
	igs, err := kc.ListInstanceGroups()
	if err != nil {
		return err
	}

	for _, ig := range igs.Items {
		currentStatus := getInstanceGroupStatus(&ig)
		if !strings.EqualFold(currentStatus, stateReady) {
			return errors.Errorf("Expected Instance Group %s to be ready, but was '%s'", ig.GetName(), currentStatus)
		} else {
			log.Infof("Instance Group %s is ready", ig.GetName())
		}
	}

	return nil
}

// TODO: delete function, use code directly in VerifyInstanceGroups
func getInstanceGroupStatus(instanceGroup *unstructured.Unstructured) string {
	if val, ok, _ := unstructured.NestedString(instanceGroup.UnstructuredContent(), "status", "currentState"); ok {
		return val
	}
	return ""
}

func (kc *ClientSet) ValidatePrometheusVolumeClaimTemplatesName(statefulsetName string, namespace string, volumeClaimTemplatesName string) error {
	var sfsvolumeClaimTemplatesName string
	// Prometheus StatefulSets deployed, then validate volumeClaimTemplate name.
	// Validation required:
	// 	- To retain existing persistent volumes and not to loose any data.
	//	- And avoid creating new name persistent volumes.
	sfs, err := kc.ListStatefulSets(namespace)
	if err != nil {
		return err
	}
	for _, sfsItem := range sfs.Items {
		if sfsItem.Name == statefulsetName {
			pvcClaimRef := sfsItem.Spec.VolumeClaimTemplates
			sfsvolumeClaimTemplatesName = pvcClaimRef[0].Name
		}
	}
	if sfsvolumeClaimTemplatesName == "" {
		return errors.Errorf("prometheus statefulset not deployed, name given: %v", volumeClaimTemplatesName)
	} else if sfsvolumeClaimTemplatesName != volumeClaimTemplatesName {
		return errors.Errorf("Prometheus volumeClaimTemplate name changed', got: %v", sfsvolumeClaimTemplatesName)
	}
	// Validate Persistent Volume label
	err = kc.validatePrometheusPVLabels(volumeClaimTemplatesName)
	if err != nil {
		return err
	}

	return nil
}

func (kc *ClientSet) validatePrometheusPVLabels(volumeClaimTemplatesName string) error {
	// Get prometheus PersistentVolume list
	pv, err := kc.ListPersistentVolumes()
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

func (kc *ClientSet) SecretDelete(secretName, namespace string) error {
	return kc.SecretOperationFromEnvironmentVariable(operationDelete, secretName, namespace, "")
}

func (kc *ClientSet) SecretOperationFromEnvironmentVariable(operation, secretName, namespace, environmentVariable string) error {
	var (
		secretValue string
		ok          bool
	)
	if err := kc.Validate(); err != nil {
		return err
	}
	if operation != operationDelete {
		secretValue, ok = os.LookupEnv(environmentVariable)
		if !ok {
			return errors.Errorf("couldn't lookup environment variable '%s'", environmentVariable)
		}
	}
	switch operation {
	case operationCreate, operationSubmit:
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				environmentVariable: []byte(secretValue),
			},
		}
		_, err := kc.KubeInterface.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
		if kerrors.IsAlreadyExists(err) {
			log.Infof("secret '%s' already created", secretName)
			return nil
		}
		return err
	case operationUpdate:
		currentSecret, err := kc.KubeInterface.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		secret := currentSecret.DeepCopy()
		if len(secret.Data) == 0 {
			secret.Data = map[string][]byte{}
		}
		secret.Data[environmentVariable] = []byte(secretValue)
		_, err = kc.KubeInterface.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
		return err
	case operationDelete:
		err := kc.KubeInterface.CoreV1().Secrets(namespace).Delete(context.TODO(), secretName, metav1.DeleteOptions{})
		if kerrors.IsNotFound(err) {
			log.Infof("secret '%s' already deleted", secretName)
			return nil
		}
		return err
	default:
		return fmt.Errorf("unsupported operation: '%s'", operation)
	}
}

func (kc *ClientSet) GetIngressEndpoint(name, namespace string, port int, path string) (string, error) {
	var (
		counter int
	)
	for {
		log.Info("waiting for ingress availability")
		if counter >= kc.getWaiterTries() {
			return "", errors.New("waiter timed out waiting for resource state")
		}
		ingress, err := kc.GetIngress(name, namespace)
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
		time.Sleep(kc.getWaiterInterval())
	}
}

func (kc *ClientSet) IngressAvailable(name, namespace string, port int, path string) error {
	var (
		counter int
	)
	endpoint, err := kc.GetIngressEndpoint(name, namespace, port, path)
	if err != nil {
		return err
	}
	for {
		log.Info("waiting for ingress availability")
		if counter >= kc.getWaiterTries() {
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
				time.Sleep(kc.getWaiterInterval())
				return nil
			}
		} else {
			log.Infof("endpoint %v is not available yet: %v", endpoint, err)
		}
		counter++
		time.Sleep(kc.getWaiterInterval())
	}
}

func (kc *ClientSet) SendTrafficToIngress(tps int, name, namespace string, port int, path string, duration int, durationUnits string, expectedErrors int) error {
	endpoint, err := kc.GetIngressEndpoint(name, namespace, port, path)
	if err != nil {
		return err
	}
	log.Infof("sending traffic to %v with rate of %v tps for %v %s...", endpoint, tps, duration, durationUnits)
	rate := vegeta.Rate{Freq: tps, Per: time.Second}
	var d time.Duration
	switch durationUnits {
	case common.DurationMinutes:
		d = time.Minute * time.Duration(duration)
	case common.DurationSeconds:
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

func init() {

	// Register ginkgo.Fail as the Fail handler. This handler panics
	// and subsequently auto recovers from the panic, which is what we need
	// for gracefully exiting failures.
	// https://github.com/onsi/ginkgo/blob/v1.16.5/ginkgo_dsl.go#L283-L303
	gomega.RegisterFailHandler(ginkgo.Fail)
}

func (kc *ClientSet) ResourceInNamespace(resource, name, ns string) error {
	var err error

	if err := kc.Validate(); err != nil {
		return err
	}

	switch resource {
	case "deployment":
		_, err = kc.KubeInterface.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})

	case "service":
		_, err = kc.KubeInterface.CoreV1().Services(ns).Get(context.Background(), name, metav1.GetOptions{})

	case "hpa", "horizontalpodautoscaler":
		_, err = kc.KubeInterface.AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(context.Background(), name, metav1.GetOptions{})

	case "pdb", "poddisruptionbudget":
		_, err = kc.KubeInterface.PolicyV1beta1().PodDisruptionBudgets(ns).Get(context.Background(), name, metav1.GetOptions{})
	case "sa", "serviceaccount":
		_, err = kc.KubeInterface.CoreV1().ServiceAccounts(ns).Get(context.Background(), name, metav1.GetOptions{})

	default:
		return errors.Errorf("Invalid resource type")
	}

	if err != nil {
		return err
	}
	return nil
}
