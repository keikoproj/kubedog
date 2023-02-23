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

package kube

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	hpa "k8s.io/api/autoscaling/v2beta2"
	v1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	kTesting "k8s.io/client-go/testing"
)

func TestPositiveNodesWithSelectorShouldBe(t *testing.T) {

	var (
		g                 = gomega.NewWithT(t)
		testReadySelector = "testing-ShouldBeReady=some-value"
		testFoundSelector = "testing-ShouldBeFound=some-value"
		testReadyLabel    = map[string]string{"testing-ShouldBeReady": "some-value"}
		testFoundLabel    = map[string]string{"testing-ShouldBeFound": "some-value"}
		fakeClient        *fake.Clientset
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery     = &fakeDiscovery.FakeDiscovery{}
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

	kc := Client{
		KubeInterface:      fakeClient,
		DiscoveryInterface: fakeDiscovery,
		DynamicInterface:   fakeDynamicClient,
		FilesPath:          "../../test/templates",
	}

	err := kc.NodesWithSelectorShouldBe(1, testReadySelector, NodeStateReady)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = kc.NodesWithSelectorShouldBe(1, testFoundSelector, NodeStateFound)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPositiveResourceOperation(t *testing.T) {
	var (
		err               error
		g                 = gomega.NewWithT(t)
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		testResource      *unstructured.Unstructured
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		fakeClient        *fake.Clientset
	)

	const fileName = "test-resourcefile.yaml"

	testResource, err = resourceFromYaml(fileName)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	kc := Client{
		KubeInterface:      fakeClient,
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceOperation(OperationCreate, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = kc.ResourceOperation(OperationDelete, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
func TestMultipleResourcesOperation(t *testing.T) {

	var (
		dynScheme           = runtime.NewScheme()
		fakeDynamicClient   = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery       = fakeDiscovery.FakeDiscovery{}
		g                   = gomega.NewWithT(t)
		testTemplatePath, _ = filepath.Abs("../../test/templates")
	)

	expectedResources := []*metav1.APIResourceList{
		newTestAPIResourceList("someGroup.apiVersion/SomeVersion", "someResource", "SomeKind"),
		newTestAPIResourceList("otherGroup.apiVersion/OtherVersion", "otherResource", "OtherKind"),
		newTestAPIResourceList("argoproj.io/v1alpha1", "AnalysisTemplate", "AnalysisTemplate"),
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, expectedResources...)

	resourceToApiResourceList := func(resource *unstructured.Unstructured) *metav1.APIResourceList {
		return newTestAPIResourceList(
			resource.GetAPIVersion(),
			resource.GetName(),
			resource.GetKind(),
		)
	}

	tests := []struct {
		testResourcePath  string
		numResources      int
		expectError       bool
		expectedResources []*metav1.APIResourceList
	}{
		{ // PositiveTest
			testResourcePath: testTemplatePath + "/test-multi-resourcefile.yaml",
			numResources:     2,
			expectError:      false,
			expectedResources: []*metav1.APIResourceList{
				newTestAPIResourceList("someGroup.apiVersion/SomeVersion", "someResource", "SomeKind"),
				newTestAPIResourceList("otherGroup.apiVersion/OtherVersion", "otherResource", "OtherKind"),
			},
		},
		{ // NegativeTest: file doesn't exist
			testResourcePath:  testTemplatePath + "/wrongName_manifest.yaml",
			numResources:      0,
			expectError:       true,
			expectedResources: []*metav1.APIResourceList{},
		},
		{ // Avoid text/template no function found error when working with AnalysisTemplate/no template args
			testResourcePath: testTemplatePath + "/analysis-template.yaml",
			numResources:     1,
			expectError:      false,
			expectedResources: []*metav1.APIResourceList{
				newTestAPIResourceList("argoproj.io/v1alpha1", "args-test", "AnalysisTemplate"),
			},
		},
	}

	for _, test := range tests {
		resourceList, err := util.GetMultipleResourcesFromYaml(test.testResourcePath, &fakeDiscovery, nil)

		g.Expect(len(resourceList)).To(gomega.Equal(test.numResources))
		if test.expectError {
			g.Expect(err).Should(gomega.HaveOccurred())
		} else {
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			for i, resource := range resourceList {
				g.Expect(resourceToApiResourceList(resource.Resource)).To(gomega.Equal(test.expectedResources[i]))
			}
		}
	}
}
func TestPositiveResourceShouldBe(t *testing.T) {
	var (
		err                 error
		g                   = gomega.NewWithT(t)
		dynScheme           = runtime.NewScheme()
		fakeDynamicClient   = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery       = fakeDiscovery.FakeDiscovery{}
		fakeClient          *fake.Clientset
		testResource        *unstructured.Unstructured
		createdReactionFunc = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, testResource, nil
		}
		deletedReactionFunc = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, kerrors.NewNotFound(schema.GroupResource{}, testResource.GetName())
		}
	)

	const fileName = "test-resourcefile.yaml"

	testResource, err = resourceFromYaml(fileName)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	fakeDynamicClient.PrependReactor("get", "someResource", createdReactionFunc)

	kc := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
		KubeInterface:      fakeClient,
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceShouldBe(fileName, StateCreated)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	fakeDiscovery.ReactionChain[0] = &kTesting.SimpleReactor{
		Verb:     "get",
		Resource: "someResource",
		Reaction: deletedReactionFunc,
	}

	err = kc.ResourceShouldBe(fileName, StateDeleted)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPositiveResourceShouldConvergeToSelector(t *testing.T) {

	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		fakeClient        *fake.Clientset
		testResource      *unstructured.Unstructured
		getReactionFunc   = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, testResource, nil
		}
	)

	const (
		fileName = "test-resourcefile.yaml"
		selector = ".metadata.labels.someTestKey=someTestValue"
	)

	testResource, err = resourceFromYaml(fileName)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	fakeDynamicClient.PrependReactor("get", "someResource", getReactionFunc)

	kc := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
		KubeInterface:      fakeClient,
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceShouldConvergeToSelector(fileName, selector)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPositiveResourceConditionShouldBe(t *testing.T) {

	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		fakeClient        *fake.Clientset
		testResource      *unstructured.Unstructured
		getReactionFunc   = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, testResource, nil
		}
	)

	const (
		fileName            = "test-resourcefile.yaml"
		testConditionType   = "someConditionType"
		testConditionStatus = "true"
	)

	testResource, err = resourceFromYaml(fileName)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	fakeDynamicClient.PrependReactor("get", "someResource", getReactionFunc)

	kc := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
		KubeInterface:      fakeClient,
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceConditionShouldBe(fileName, testConditionType, testConditionStatus)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPositiveUpdateResourceWithField(t *testing.T) {

	const (
		fileName           = "test-resourcefile.yaml"
		testUpdateKeyChain = ".metadata.labels.testUpdateKey"
		testUpdateKey      = "testUpdateKey"
		testUpdateValue    = "testUpdateValue"
	)

	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		fakeClient        *fake.Clientset
		testResource      *unstructured.Unstructured
		getReactionFunc   = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, testResource, nil
		}
		updateReactionFunc = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
			addLabel(testResource, testUpdateKey, testUpdateValue)
			return true, testResource, nil
		}
	)

	testResource, err = resourceFromYaml(fileName)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	fakeDynamicClient.PrependReactor("get", "someResource", getReactionFunc)
	fakeDynamicClient.PrependReactor("update", "someResource", updateReactionFunc)

	kc := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
		KubeInterface:      fakeClient,
		FilesPath:          "../../test/templates",
	}

	err = kc.UpdateResourceWithField(fileName, testUpdateKeyChain, testUpdateValue)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	expectedLabelValue, found, err := unstructured.NestedString(testResource.UnstructuredContent(), "metadata", "labels", "testUpdateKey")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(expectedLabelValue).To(gomega.Equal(testUpdateValue))
}

func TestResourceInNamespace(t *testing.T) {
	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeKubeClient    = fake.NewSimpleClientset()
		namespace         = "test_ns"
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = &fakeDiscovery.FakeDiscovery{}
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

	kc := Client{
		KubeInterface:      fakeKubeClient,
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: fakeDiscovery,
	}

	_, _ = kc.KubeInterface.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
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
				_, _ = kc.KubeInterface.AppsV1().Deployments(namespace).Create(context.Background(), &appsv1.Deployment{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "service":
				_, _ = kc.KubeInterface.CoreV1().Services(namespace).Create(context.Background(), &v1.Service{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "hpa":
				_, _ = kc.KubeInterface.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Create(context.Background(), &hpa.HorizontalPodAutoscaler{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "pdb":
				_, _ = kc.KubeInterface.PolicyV1beta1().PodDisruptionBudgets(namespace).Create(context.Background(), &policy.PodDisruptionBudget{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "serviceaccount":
				_, _ = kc.KubeInterface.CoreV1().ServiceAccounts(namespace).Create(context.Background(), &v1.ServiceAccount{
					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			}
			err = kc.ResourceInNamespace(tt.resource, tt.name, namespace)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		})
	}
}

func TestScaleDeployment(t *testing.T) {
	var (
		err            error
		g              = gomega.NewWithT(t)
		fakeKubeClient = fake.NewSimpleClientset()
		namespace      = "test_ns"
		deployName     = "test_deploy"
		replicaCount   = int32(1)
	)

	kc := Client{
		KubeInterface: fakeKubeClient,
	}

	_, _ = kc.KubeInterface.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Status: v1.NamespaceStatus{Phase: v1.NamespaceActive},
	}, metav1.CreateOptions{})

	_, _ = kc.KubeInterface.AppsV1().Deployments(namespace).Create(context.Background(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
		},
	}, metav1.CreateOptions{})
	err = kc.ScaleDeployment(deployName, namespace, 2)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	s, _ := kc.KubeInterface.AppsV1().Deployments(namespace).GetScale(context.Background(), deployName, metav1.GetOptions{})
	g.Expect(s.Spec.Replicas).To(gomega.Equal(int32(2)))
}

func resourceFromYaml(resourceFileName string) (*unstructured.Unstructured, error) {

	resourcePath := filepath.Join("../../test/templates", resourceFileName)
	d, err := ioutil.ReadFile(resourcePath)
	if err != nil {
		return nil, err
	}
	return resourceFromBytes(d)
}
func resourceFromBytes(bytes []byte) (*unstructured.Unstructured, error) {
	resource := &unstructured.Unstructured{}
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode(bytes, nil, resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
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

func addLabel(in *unstructured.Unstructured, key, value string) {
	labels, _, _ := unstructured.NestedMap(in.Object, "metadata", "labels")

	labels[key] = value

	err := unstructured.SetNestedMap(in.Object, labels, "metadata", "labels")
	if err != nil {
		log.Errorf("Failed adding label %v=%v to the resource %v: %v", key, value, in.GetName(), err)
	}
}

func TestClusterRoleAndBindingIsFound(t *testing.T) {
	var (
		err            error
		g              = gomega.NewWithT(t)
		fakeKubeClient = fake.NewSimpleClientset()
	)

	kc := Client{
		KubeInterface: fakeKubeClient,
	}

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
				_, _ = kc.KubeInterface.RbacV1().ClusterRoles().Create(context.Background(), &rbacv1.ClusterRole{

					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			case "clusterrolebinding":
				_, _ = kc.KubeInterface.RbacV1().ClusterRoleBindings().Create(context.Background(), &rbacv1.ClusterRoleBinding{

					ObjectMeta: meta,
				}, metav1.CreateOptions{})
			}
			err = kc.ClusterRbacIsFound(tt.resource, tt.name)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		})
	}
}

func Test_unstructuredResourceOperation(t *testing.T) {
	type clientFields struct {
		DynamicInterface dynamic.Interface
	}
	type funcArgs struct {
		operation            string
		ns                   string
		unstructuredResource util.K8sUnstructuredResource
	}

	resourceNoNs, err := resourceFromYaml("../../test/templates/resource-without-namespace.yaml")
	if err != nil {
		t.Errorf(err.Error())
	}

	resourceNs, err := resourceFromYaml("../../test/templates/resource-with-namespace.yaml")
	if err != nil {
		t.Errorf(err.Error())
	}

	resourceNoNsUpdate, err := resourceFromYaml("../../test/templates/resource-without-namespace-update.yaml")
	if err != nil {
		t.Errorf(err.Error())
	}

	resourceNsUpdate, err := resourceFromYaml("../../test/templates/resource-with-namespace-update.yaml")
	if err != nil {
		t.Errorf(err.Error())
	}

	dynScheme := runtime.NewScheme()
	fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(dynScheme)

	tests := []struct {
		name         string
		clientFields clientFields
		funcArgs     funcArgs
		wantErr      bool
	}{
		{
			name: "Resource create succeeds when namespace is configurable",
			clientFields: clientFields{
				DynamicInterface: fakeDynamicClient,
			},
			funcArgs: funcArgs{
				operation: "create",
				ns:        "test-namespace",
				unstructuredResource: util.K8sUnstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNoNs,
				},
			},
			wantErr: false,
		},
		{
			name: "Resource update succeeds when namespace is configurable",
			clientFields: clientFields{
				DynamicInterface: fakeDynamicClient,
			},
			funcArgs: funcArgs{
				operation: "update",
				ns:        "test-namespace",
				unstructuredResource: util.K8sUnstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNoNsUpdate,
				},
			},
			wantErr: false,
		},
		{
			name: "Resource delete succeeds when namespace is configurable",
			clientFields: clientFields{
				DynamicInterface: fakeDynamicClient,
			},
			funcArgs: funcArgs{
				operation: "delete",
				ns:        "test-namespace",
				unstructuredResource: util.K8sUnstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNoNs,
				},
			},
			wantErr: false,
		},
		{
			name: "Resource create succeeds when namespace in YAML",
			clientFields: clientFields{
				DynamicInterface: fakeDynamicClient,
			},
			funcArgs: funcArgs{
				operation: "create",
				unstructuredResource: util.K8sUnstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNs,
				},
			},
			wantErr: false,
		},
		{
			name: "Resource update succeeds when namespace is in YAML",
			clientFields: clientFields{
				DynamicInterface: fakeDynamicClient,
			},
			funcArgs: funcArgs{
				operation: "update",
				unstructuredResource: util.K8sUnstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNsUpdate,
				},
			},
			wantErr: false,
		},
		{
			name: "Resource create fails when namespace configured and in YAML",
			clientFields: clientFields{
				DynamicInterface: fakeDynamicClient,
			},
			funcArgs: funcArgs{
				operation: "create",
				ns:        "override-ns",
				unstructuredResource: util.K8sUnstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNs,
				},
			},
			wantErr: true,
		},
		{
			name: "Unsupported operation produces error",
			clientFields: clientFields{
				DynamicInterface: fakeDynamicClient,
			},
			funcArgs: funcArgs{
				operation: "invalid",
				unstructuredResource: util.K8sUnstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNs,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := &Client{
				DynamicInterface: tt.clientFields.DynamicInterface,
			}
			if err := kc.unstructuredResourceOperation(tt.funcArgs.operation, tt.funcArgs.ns, tt.funcArgs.unstructuredResource); (err != nil) != tt.wantErr {
				t.Errorf("Client.unstructuredResourceOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ThePodsInNamespaceWithSelectorShouldHaveLabels(t *testing.T) {
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
	podWithLabels1 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-foo-xhhxj",
			Namespace: "foo",
			Labels: map[string]string{
				"app":   "foo",
				"label": "true",
			},
		},
	}
	podWithLabels2 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-foo-xhhzd",
			Namespace: "foo",
			Labels: map[string]string{
				"app":   "foo",
				"label": "true",
			},
		},
	}
	podMissingLabel := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-foo-xhhzr",
			Namespace: "foo",
			Labels: map[string]string{
				"app": "foo",
			},
		},
	}
	clientNoErr := fake.NewSimpleClientset(&ns, &podWithLabels1, &podWithLabels2)
	clientErr := fake.NewSimpleClientset(&ns, &podWithLabels1, &podWithLabels2, &podMissingLabel)
	dynScheme := runtime.NewScheme()
	fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(dynScheme)
	fakeDiscoveryClient := fakeDiscovery.FakeDiscovery{}

	type fields struct {
		KubeInterface      kubernetes.Interface
		DynamicInterface   dynamic.Interface
		DiscoveryInterface *fakeDiscovery.FakeDiscovery
	}
	type args struct {
		namespace string
		selector  string
		labels    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "No pods found",
			fields: fields{
				KubeInterface:      clientErr,
				DiscoveryInterface: &fakeDiscoveryClient,
				DynamicInterface:   fakeDynamicClient,
			},
			args: args{
				selector:  "app=doesnotexist",
				namespace: "foo",
				labels:    "app=foo,label=true",
			},
			wantErr: true,
		},
		{
			name: "Pods should have labels",
			fields: fields{
				KubeInterface:      clientNoErr,
				DiscoveryInterface: &fakeDiscoveryClient,
				DynamicInterface:   fakeDynamicClient,
			},
			args: args{
				selector:  "app=foo",
				namespace: "foo",
				labels:    "app=foo,label=true",
			},
			wantErr: false,
		},
		{
			name: "Error from pod missing label",
			fields: fields{
				KubeInterface:      clientErr,
				DiscoveryInterface: &fakeDiscoveryClient,
				DynamicInterface:   fakeDynamicClient,
			},
			args: args{
				selector:  "app=foo",
				namespace: "foo",
				labels:    "app=foo,label=true",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := &Client{
				KubeInterface:      tt.fields.KubeInterface,
				DynamicInterface:   tt.fields.DynamicInterface,
				DiscoveryInterface: tt.fields.DiscoveryInterface,
			}
			if err := kc.ThePodsInNamespaceWithSelectorShouldHaveLabels(tt.args.namespace, tt.args.selector, tt.args.labels); (err != nil) != tt.wantErr {
				t.Errorf("ThePodsInNamespaceWithSelectorShouldHaveLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
