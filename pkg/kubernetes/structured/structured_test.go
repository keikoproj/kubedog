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
	"testing"

	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	hpa "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPositiveNodesWithSelectorShouldBe(t *testing.T) {

	var (
		g                 = gomega.NewWithT(t)
		testReadySelector = "testing-ShouldBeReady=some-value"
		testFoundSelector = "testing-ShouldBeFound=some-value"
		testReadyLabel    = map[string]string{"testing-ShouldBeReady": "some-value"}
		testFoundLabel    = map[string]string{"testing-ShouldBeFound": "some-value"}
		fakeClient        *fake.Clientset
		// dynScheme         = runtime.NewScheme()
		// fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		// fakeDiscovery     = &fakeDiscovery.FakeDiscovery{}
	)

	fakeClient = fake.NewSimpleClientset(&v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "SomeReady-Node",
			Labels: testReadyLabel,
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:   v1.NodeReady,
					Status: v1.ConditionTrue,
				},
			},
		},
	}, &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "SomeFound-name",
			Labels: testFoundLabel,
		},
	})

	// kc := ClientSet{
	// 	KubeInterface:      fakeClient,
	// 	DiscoveryInterface: fakeDiscovery,
	// 	DynamicInterface:   fakeDynamicClient,
	// 	FilesPath:          "../../test/templates",
	// }

	err := NodesWithSelectorShouldBe(fakeClient, common.WaiterConfig{}, 1, testReadySelector, common.StateReady)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = NodesWithSelectorShouldBe(fakeClient, common.WaiterConfig{}, 1, testFoundSelector, common.StateFound)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestResourceInNamespace(t *testing.T) {
	var (
		err        error
		g          = gomega.NewWithT(t)
		fakeClient = fake.NewSimpleClientset()
		namespace  = "test_ns"
		// fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		// fakeDiscovery     = &fakeDiscovery.FakeDiscovery{}
	)

	tests := []struct {
		resource string
		name     string
	}{
		{
			resource: "deployment",
			name:     "test_deploy",
		},
		{
			resource: "service",
			name:     "test_service",
		},
		{
			resource: "hpa",
			name:     "test_hpa",
		},
		{
			resource: "horizontalpodautoscaler",
			name:     "test_hpa",
		},
		{
			resource: "pdb",
			name:     "test_pdb",
		},
		{
			resource: "poddisruptionbudget",
			name:     "test_pdb",
		},
		{
			resource: "serviceaccount",
			name:     "mock_service_account",
		},
	}

	// kc := ClientSet{
	// 	KubeInterface:      fakeKubeClient,
	// 	DynamicInterface:   fakeDynamicClient,
	// 	DiscoveryInterface: fakeDiscovery,
	// }

	_, _ = fakeClient.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Status: v1.NamespaceStatus{Phase: v1.NamespaceActive},
	}, metav1.CreateOptions{})

	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			meta := metav1.ObjectMeta{
				Name: tt.name,
			}

			switch tt.resource {
			case "deployment":
				_, _ = fakeClient.AppsV1().Deployments(namespace).Create(context.Background(), &appsv1.Deployment{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "service":
				_, _ = fakeClient.CoreV1().Services(namespace).Create(context.Background(), &v1.Service{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "hpa":
				_, _ = fakeClient.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Create(context.Background(), &hpa.HorizontalPodAutoscaler{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "pdb":
				_, _ = fakeClient.PolicyV1beta1().PodDisruptionBudgets(namespace).Create(context.Background(), &policy.PodDisruptionBudget{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "serviceaccount":
				_, _ = fakeClient.CoreV1().ServiceAccounts(namespace).Create(context.Background(), &v1.ServiceAccount{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			}
			err = ResourceInNamespace(fakeClient, tt.resource, tt.name, namespace)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		})
	}
}

func TestScaleDeployment(t *testing.T) {
	var (
		err          error
		g            = gomega.NewWithT(t)
		fakeClient   = fake.NewSimpleClientset()
		namespace    = "test_ns"
		deployName   = "test_deploy"
		replicaCount = int32(1)
	)

	// kc := ClientSet{
	// 	KubeInterface: fakeKubeClient,
	// }

	_, _ = fakeClient.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Status: v1.NamespaceStatus{Phase: v1.NamespaceActive},
	}, metav1.CreateOptions{})

	_, _ = fakeClient.AppsV1().Deployments(namespace).Create(context.Background(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
		},
	}, metav1.CreateOptions{})
	err = ScaleDeployment(fakeClient, deployName, namespace, 2)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	s, _ := fakeClient.AppsV1().Deployments(namespace).GetScale(context.Background(), deployName, metav1.GetOptions{})
	g.Expect(s.Spec.Replicas).To(gomega.Equal(int32(2)))
}

// TODO: implemented twice. maybe have a test helper pkg?
func newTestAPIResourceList(apiVersion, name, kind string) *metav1.APIResourceList {
	return &metav1.APIResourceList{
		GroupVersion: apiVersion,
		APIResources: []metav1.APIResource{
			{
				Name:       name,
				Kind:       kind,
				Namespaced: true,
			},
		},
	}
}

func TestClusterRoleAndBindingIsFound(t *testing.T) {
	var (
		err        error
		g          = gomega.NewWithT(t)
		fakeClient = fake.NewSimpleClientset()
	)

	// kc := ClientSet{
	// 	KubeInterface: fakeKubeClient,
	// }

	tests := []struct {
		resource string
		name     string
	}{
		{
			resource: "clusterrole",
			name:     "mock_cluster_role",
		},
		{
			resource: "clusterrolebinding",
			name:     "mock_cluster_role_binding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			meta := metav1.ObjectMeta{
				Name: tt.name,
			}

			switch tt.resource {
			case "clusterrole":
				_, _ = fakeClient.RbacV1().ClusterRoles().Create(context.Background(), &rbacv1.ClusterRole{

					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "clusterrolebinding":
				_, _ = fakeClient.RbacV1().ClusterRoleBindings().Create(context.Background(), &rbacv1.ClusterRoleBinding{

					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			}
			err = ClusterRbacIsFound(fakeClient, tt.resource, tt.name)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		})
	}
}
