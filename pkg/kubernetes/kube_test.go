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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	decoder "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	kTesting "k8s.io/client-go/testing"
)

type KubernetesResource struct {
	Gvr      *meta.RESTMapping
	Resource *unstructured.Unstructured
	Err      error
}

func TestPositiveNodesWithSelectorShouldBe(t *testing.T) {

	var (
		g                 = gomega.NewWithT(t)
		testReadySelector = "testing-ShouldBeReady=some-value"
		testFoundSelector = "testing-ShouldBeFound=some-value"
		testReadyLabel    = map[string]string{"testing-ShouldBeReady": "some-value"}
		testFoundLabel    = map[string]string{"testing-ShouldBeFound": "some-value"}
		fakeClient        *fake.Clientset
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
		KubeInterface: fakeClient,
		FilesPath:     "../../test/templates",
	}

	err := kc.NodesWithSelectorShouldBe(1, testReadySelector, NodeStateReady)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = kc.NodesWithSelectorShouldBe(1, testFoundSelector, NodeStateFound)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPostitiveResourceOperation(t *testing.T) {
	var (
		err               error
		g                 = gomega.NewWithT(t)
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		testResource      *unstructured.Unstructured
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
	)

	const fileName = "test-resourcefile.yaml"

	testResource, err = resourceFromYaml(fileName)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	kc := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceOperation(OperationCreate, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = kc.ResourceOperation(OperationDelete, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
func TestPostitiveMultipleResourcesOperation(t *testing.T) {
	var (
		err               error
		g                 = gomega.NewWithT(t)
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		testResource      *unstructured.Unstructured
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
	)
	kc := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
		FilesPath:          "../../test/templates",
	}
	const fileName = "test-resourcefile.yaml"

	testResourceList, err := multipleResourcesFromYaml(fileName, kc.DiscoveryInterface)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}
	testResource = testResourceList[0]

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	err = kc.ResourceOperation(OperationCreate, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = kc.ResourceOperation(OperationDelete, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}
func TestPostitiveResourceShouldBe(t *testing.T) {
	var (
		err                 error
		g                   = gomega.NewWithT(t)
		dynScheme           = runtime.NewScheme()
		fakeDynamicClient   = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery       = fakeDiscovery.FakeDiscovery{}
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
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceShouldBe(fileName, ResourceStateCreated)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	fakeDiscovery.ReactionChain[0] = &kTesting.SimpleReactor{
		Verb:     "get",
		Resource: "someResource",
		Reaction: deletedReactionFunc,
	}

	err = kc.ResourceShouldBe(fileName, ResourceStateDeleted)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPostitiveResourceShouldConvergeToSelector(t *testing.T) {

	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
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
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceShouldConvergeToSelector(fileName, selector)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPostitiveResourceConditionShouldBe(t *testing.T) {

	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
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
		FilesPath:          "../../test/templates",
	}

	err = kc.ResourceConditionShouldBe(fileName, testConditionType, testConditionStatus)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPostitiveUpdateResourceWithField(t *testing.T) {

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
		FilesPath:          "../../test/templates",
	}

	err = kc.UpdateResourceWithField(fileName, testUpdateKeyChain, testUpdateValue)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	expectedLabelValue, found, err := unstructured.NestedString(testResource.UnstructuredContent(), "metadata", "labels", "testUpdateKey")
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(expectedLabelValue).To(gomega.Equal(testUpdateValue))
}

func TestDeploymentInNamespace(t *testing.T) {
	var (
		err            error
		g              = gomega.NewWithT(t)
		fakeKubeClient = fake.NewSimpleClientset()
		namespace      = "test_ns"
		deployName     = "test_deploy"
	)

	kc := Client{
		KubeInterface: fakeKubeClient,
	}

	_, _ = kc.KubeInterface.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Status: v1.NamespaceStatus{Phase: v1.NamespaceActive},
	})

	_, _ = kc.KubeInterface.AppsV1().Deployments(namespace).Create(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
	})

	err = kc.DeploymentInNamespace(deployName, namespace)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
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

	_, _ = kc.KubeInterface.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
		Status: v1.NamespaceStatus{Phase: v1.NamespaceActive},
	})

	_, _ = kc.KubeInterface.AppsV1().Deployments(namespace).Create(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
		},
	})
	err = kc.ScaleDeployment(deployName, namespace, 2)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	s, _ := kc.KubeInterface.AppsV1().Deployments(namespace).GetScale(deployName, metav1.GetOptions{})
	g.Expect(s.Spec.Replicas).To(gomega.Equal(int32(2)))
}

func resourceFromYaml(resourceFileName string) (*unstructured.Unstructured, error) {

	resourcePath := filepath.Join("../../test/templates", resourceFileName)
	d, err := ioutil.ReadFile(resourcePath)
	if err != nil {
		return nil, err
	}
	log.Infof(string(d))
	println(len(d))
	resource := &unstructured.Unstructured{}
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = dec.Decode(d, nil, resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
func multipleResourcesFromYaml(resourceFileName string, dc discovery.DiscoveryInterface) ([]*unstructured.Unstructured, error) {
	resource := &unstructured.Unstructured{}
	resourcePath := filepath.Join("../../test/templates", resourceFileName)
	file, err := os.Open(resourcePath)
	if err != nil {
		return nil, err
	}
	yamlDecoder := decoder.NewDocumentDecoder(io.NopCloser(file))
	buf := make([]byte, 4*1024) //same size as internal buffer in yamlDecoder
	chunk := make([]byte, 0)
	resourceList := make([]*unstructured.Unstructured, 0)
	for {
		chunk = chunk[:0]
		for { //Keep adding to the chunk until: a separator is hit (err==nil), EOF, or other error
			length, err := yamlDecoder.Read(buf)
			chunk = append(chunk, buf[:length]...)
			if err != io.ErrShortBuffer {
				break
			}
		}
		log.Infof("Chunk")
		log.Infof(string(chunk))
		println(len(chunk))
		if (err != nil && err != io.EOF) || len(chunk) == 0 {
			//If we are in error state, abort
			break
		}
		dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
		_, gvk, err := dec.Decode(chunk, nil, resource)
		println(gvk)
		if err != nil {
			//If any resource has an error, abort
			println(err.Error())
			println("ERROR")
			break
		}
		resourceList = append(resourceList, resource)
		if err == io.EOF {
			err = nil
			break
		}
	}
	yamlDecoder.Close()
	return resourceList, err
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
