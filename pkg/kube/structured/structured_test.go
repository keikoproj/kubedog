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
	"strings"
	"testing"
	"time"

	"github.com/keikoproj/kubedog/internal/util"
	"github.com/keikoproj/kubedog/pkg/kube/common"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	deploymentType         = "deployment"
	serviceType            = "service"
	hpaType                = "horizontalpodautoscaler"
	pdbType                = "poddisruptionbudget"
	saType                 = "serviceaccount"
	clusterRoleType        = "clusterrole"
	clusterRoleBindingType = "clusterrolebinding"
	nodeType               = "node"
	daemonsetType          = "daemonset"
)

func TestNodesWithSelectorShouldBe(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		w             common.WaiterConfig
		expectedNodes int
		labelSelector string
		state         string
	}
	label := "some-label-key=some-label-value"
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test: state found",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, nodeType, "node1", "", label)),
				expectedNodes: 1,
				labelSelector: label,
				state:         common.StateFound,
			},
		},
		{
			name: "Positive Test: state ready",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getNodeWithStatus(t, "node1", label, corev1.NodeReady, corev1.ConditionTrue)),
				expectedNodes: 1,
				labelSelector: label,
				state:         common.StateReady,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.w = common.NewWaiterConfig(1, time.Second)
			if err := NodesWithSelectorShouldBe(tt.args.kubeClientset, tt.args.w, tt.args.expectedNodes, tt.args.labelSelector, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("NodesWithSelectorShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceInNamespace(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		resourceType  string
		name          string
		namespace     string
	}

	deploymentName := "deployment1"
	serviceName := "service1"
	hpaName := "horizontalpodautoscaler1"
	pdbName := "poddisruptionbudget1"
	saName := "serviceaccount1"

	namespace := "namespace1"
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test: deployment",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, deploymentType, deploymentName, namespace, "")),
				resourceType:  deploymentType,
				name:          deploymentName,
				namespace:     namespace,
			},
		},
		{
			name: "Positive Test: service",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, serviceType, serviceName, namespace, "")),
				resourceType:  serviceType,
				name:          serviceName,
				namespace:     namespace,
			},
		},
		{
			name: "Positive Test: horizontalpodautoscaler",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, hpaType, hpaName, namespace, "")),
				resourceType:  hpaType,
				name:          hpaName,
				namespace:     namespace,
			},
		},
		{
			name: "Positive Test: poddisruptionbudget",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, pdbType, pdbName, namespace, "")),
				resourceType:  pdbType,
				name:          pdbName,
				namespace:     namespace,
			},
		},
		{
			name: "Positive Test: serviceaccount",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, saType, saName, namespace, "")),
				resourceType:  saType,
				name:          saName,
				namespace:     namespace,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResourceInNamespace(tt.args.kubeClientset, tt.args.resourceType, tt.args.name, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("ResourceInNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScaleDeployment(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		name          string
		namespace     string
		replicas      int32
	}
	deploymentName := "deployment1"
	namespace := "namespace1"
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, deploymentType, deploymentName, namespace, "")),
				name:          deploymentName,
				namespace:     namespace,
				replicas:      0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ScaleDeployment(tt.args.kubeClientset, tt.args.name, tt.args.namespace, tt.args.replicas); (err != nil) != tt.wantErr {
				t.Errorf("ScaleDeployment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClusterRbacIsFound(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		resourceType  string
		name          string
	}
	clusterRoleName := "clusterrole1"
	clusterRoleBindingName := "clusterrolebinding1"
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test: ClusterRole",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, clusterRoleType, clusterRoleName, "", "")),
				resourceType:  clusterRoleType,
				name:          clusterRoleName,
			},
		},
		{
			name: "Positive Test: ClusterRoleBinding",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, clusterRoleBindingType, clusterRoleBindingName, "", "")),
				resourceType:  clusterRoleBindingType,
				name:          clusterRoleBindingName,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ClusterRbacIsFound(tt.args.kubeClientset, tt.args.resourceType, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("ClusterRbacIsFound() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetNodes(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, nodeType, "node1", "", "")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GetNodes(tt.args.kubeClientset); (err != nil) != tt.wantErr {
				t.Errorf("GetNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDaemonSetIsRunning(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		expBackoff    wait.Backoff
		name          string
		namespace     string
	}
	daemonsetName := "daemonset1"
	namespace := "namespace1"
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test",
			args: args{
				kubeClientset: fake.NewSimpleClientset(getResource(t, daemonsetType, daemonsetName, namespace, "")),
				expBackoff:    util.DefaultRetry,
				name:          daemonsetName,
				namespace:     namespace,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DaemonSetIsRunning(tt.args.kubeClientset, tt.args.expBackoff, tt.args.name, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("DaemonSetIsRunning() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement
func TestDeploymentIsRunning(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		name          string
		namespace     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeploymentIsRunning(tt.args.kubeClientset, tt.args.name, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("DeploymentIsRunning() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement
func TestPersistentVolExists(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		name          string
		expectedPhase string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PersistentVolExists(tt.args.kubeClientset, tt.args.name, tt.args.expectedPhase); (err != nil) != tt.wantErr {
				t.Errorf("PersistentVolExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement
func TestValidatePrometheusVolumeClaimTemplatesName(t *testing.T) {
	type args struct {
		kubeClientset            kubernetes.Interface
		statefulsetName          string
		namespace                string
		volumeClaimTemplatesName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePrometheusVolumeClaimTemplatesName(tt.args.kubeClientset, tt.args.statefulsetName, tt.args.namespace, tt.args.volumeClaimTemplatesName); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrometheusVolumeClaimTemplatesName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement
func TestSecretDelete(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		name          string
		namespace     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SecretDelete(tt.args.kubeClientset, tt.args.name, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("SecretDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement
func TestSecretOperationFromEnvironmentVariable(t *testing.T) {
	type args struct {
		kubeClientset       kubernetes.Interface
		operation           string
		name                string
		namespace           string
		environmentVariable string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SecretOperationFromEnvironmentVariable(tt.args.kubeClientset, tt.args.operation, tt.args.name, tt.args.namespace, tt.args.environmentVariable); (err != nil) != tt.wantErr {
				t.Errorf("SecretOperationFromEnvironmentVariable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement
func TestIngressAvailable(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		w             common.WaiterConfig
		name          string
		namespace     string
		port          int
		path          string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IngressAvailable(tt.args.kubeClientset, tt.args.w, tt.args.name, tt.args.namespace, tt.args.port, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("IngressAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: implement
func TestSendTrafficToIngress(t *testing.T) {
	type args struct {
		kubeClientset  kubernetes.Interface
		w              common.WaiterConfig
		tps            int
		name           string
		namespace      string
		port           int
		path           string
		duration       int
		durationUnits  string
		expectedErrors int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SendTrafficToIngress(tt.args.kubeClientset, tt.args.w, tt.args.tps, tt.args.name, tt.args.namespace, tt.args.port, tt.args.path, tt.args.duration, tt.args.durationUnits, tt.args.expectedErrors); (err != nil) != tt.wantErr {
				t.Errorf("SendTrafficToIngress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getNodeWithStatus(t *testing.T, name, label string, statusType corev1.NodeConditionType, status corev1.ConditionStatus) runtime.Object {
	nodeInterface := getResource(t, nodeType, name, "", label)
	node, ok := nodeInterface.(*corev1.Node)
	if !ok {
		t.Errorf("'runtime.Object' could not be cast to '*corev1.Node': %v", nodeInterface)
	}
	node.Status = corev1.NodeStatus{
		Conditions: []corev1.NodeCondition{
			{
				Type:   statusType,
				Status: status,
			},
		},
	}
	return node
}

func getLabelParts(t *testing.T, label string) (string, string) {
	labelSplit := strings.Split(label, "=")
	if len(labelSplit) != 2 {
		t.Errorf("expected label format '<key>=<value>', got '%s'", label)
	}
	return labelSplit[0], labelSplit[1]
}

func getResource(t *testing.T, resourceType, name, namespace, label string) runtime.Object {
	labels := map[string]string{}
	if label != "" {
		key, value := getLabelParts(t, label)
		labels[key] = value
	}

	switch resourceType {
	case deploymentType:
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	case serviceType:
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	case "hpa", hpaType:
		return &v2beta2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	case "pdb", pdbType:
		return &v1beta1.PodDisruptionBudget{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	case "sa", saType:
		return &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	case clusterRoleType:
		errorIfNamespaceNotEmpty(t, namespace)
		return &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
		}
	case clusterRoleBindingType:
		errorIfNamespaceNotEmpty(t, namespace)
		return &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
		}
	case nodeType:
		errorIfNamespaceNotEmpty(t, namespace)
		return &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
		}
	case daemonsetType:
		return &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	default:
		t.Errorf("Invalid resource type: %s", resourceType)
	}
	return nil
}

func errorIfNamespaceNotEmpty(t *testing.T, ns string) {
	if ns != "" {
		t.Errorf("Namespace should be empty, but is: %s", ns)
	}
}
