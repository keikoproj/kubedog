package kube

import (
	"testing"

	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	kTesting "k8s.io/client-go/testing"
)

func newUnstructured(version, kind, namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": version,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
		},
	}
}

func newTestBaseUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "someGroup.apiVersion/SomeVersion",
			"kind":       "SomeKind",
			"metadata": map[string]interface{}{
				"namespace": "someTestNamespace",
				"name":      "someResource",
			},
		},
	}
}

func newTestBaseAPIResourceList() []*metav1.APIResourceList {
	return []*metav1.APIResourceList{
		{
			GroupVersion: "someGroup.apiVersion/SomeVersion",
			APIResources: []metav1.APIResource{
				{
					Name: "someResource",
					Kind: "SomeKind",
				},
			},
		},
	}
}

func addCondition(in *unstructured.Unstructured, name, status string) {
	conditions, _, _ := unstructured.NestedSlice(in.Object, "status", "conditions")
	conditions = append(conditions, map[string]interface{}{
		"type":   name,
		"status": status,
	})
	unstructured.SetNestedSlice(in.Object, conditions, "status", "conditions")
}

func addLabel(in *unstructured.Unstructured, key, value string) {
	labels, _, _ := unstructured.NestedSlice(in.Object, "metadata", "labels")
	labels = append(labels, map[string]interface{}{
		key: value,
	})

	err := unstructured.SetNestedSlice(in.Object, labels, "metadata", "labels")

	if err != nil {
		log.Errorf("Failed adding label %v=%v to the resource %v: %v", key, value, in.GetName(), err)
	}
}

func TestPositiveNodesWithSelectorShouldBe(t *testing.T) {

	var (
		g                 = gomega.NewWithT(t)
		testReadySelector = "testing-ShouldBeReady=some-value"
		testFoundSelector = "testing-ShouldBeFound=some-value"

		testReadyLabel = map[string]string{"testing-ShouldBeReady": "some-value"}
		testFoundLabel = map[string]string{"testing-ShouldBeFound": "some-value"}

		FakeClient = fake.NewSimpleClientset(&v1.Node{
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
	)

	KC := Client{KubeInterface: FakeClient}

	err := KC.NodesWithSelectorShouldBe(1, testReadySelector, NodeStateReady)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = KC.NodesWithSelectorShouldBe(1, testFoundSelector, NodeStateFound)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

/*func TestNegativeNodesWithSelectorShouldBe(t *testing.T) {
	// error timed out
	// error getting nodes
	var (
		g           = gomega.NewWithT(t)
		EmptyClient = fake.NewSimpleClientset()
	)

	KC := Client{KubeInterface: EmptyClient}

	err := KC.NodesWithSelectorShouldBe(1, "Some-key=Some-value", NodeStateReady)
	g.Expect(err).Should(gomega.HaveOccurred())
}//*/

func TestPostitiveResourceOperation(t *testing.T) {
	var (
		g                 = gomega.NewWithT(t)
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{
			Fake: &fakeDynamicClient.Fake,
		}
	)

	const fileName = "test-resourcefile.yaml"

	fakeDiscovery.Resources = newTestBaseAPIResourceList()

	KC := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
	}

	// TODO: test resource already created and already deleted
	err := KC.ResourceOperation(OperationCreate, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = KC.ResourceOperation(OperationDelete, fileName)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	// TODO: negative test
}

func TestPostitiveResourceShouldBe(t *testing.T) {
	var (
		g                 = gomega.NewWithT(t)
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{
			Fake: &fakeDynamicClient.Fake,
		}
	)

	const fileName = "test-resourcefile.yaml"

	fakeDiscovery.Resources = newTestBaseAPIResourceList()

	KC := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
	}

	KC.ResourceOperation(OperationCreate, fileName)
	err := KC.ResourceShouldBe(fileName, ResourceStateCreated)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	KC.ResourceOperation(OperationDelete, fileName)
	err = KC.ResourceShouldBe(fileName, ResourceStateDeleted)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	// TODO: negative test
}

func TestPostitiveResourceShouldConvergeToSelector(t *testing.T) {
	var (
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{
			Fake: &fakeDynamicClient.Fake,
		}
		testResource = newTestBaseUnstructured()
		ReactionFunc = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, testResource, nil
		}
	)

	const (
		fileName = "test-resourcefile.yaml"
		selector = "someTestKey=someTestValue"
	)

	fakeDiscovery.Resources = newTestBaseAPIResourceList()
	addLabel(testResource, "someTestKey", "someTestValue")

	fakeDynamicClient.PrependReactor("get", testResource.GetName(), ReactionFunc)

	KC := Client{
		DynamicInterface:   fakeDynamicClient,
		DiscoveryInterface: &fakeDiscovery,
	}

	err := KC.ResourceShouldConvergeToSelector(fileName, selector)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	// TODO: negative test
}
