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

package unstructured

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
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
	kTesting "k8s.io/client-go/testing"
)

func TestPositiveResourceOperation(t *testing.T) {
	var (
		err               error
		g                 = gomega.NewWithT(t)
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		testResource      *unstructured.Unstructured
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		//fakeClient        *fake.Clientset
	)

	const fileName = "test-resourcefile.yaml"

	testResource, err = resourceFromYaml(fileName)
	if err != nil {
		t.Errorf("Failed getting the test resource from the file %v: %v", fileName, err)
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, newTestAPIResourceList(testResource.GetAPIVersion(), testResource.GetName(), testResource.GetKind()))

	// kc := ClientSet{
	// 	KubeInterface:      fakeClient,
	// 	DynamicInterface:   fakeDynamicClient,
	// 	DiscoveryInterface: &fakeDiscovery,
	// 	FilesPath:          "../../test/templates",
	// }
	resource, err := GetResource(&fakeDiscovery, nil, resourcePath(fileName))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = ResourceOperation(fakeDynamicClient, resource, common.OperationCreate)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = ResourceOperation(fakeDynamicClient, resource, common.OperationDelete)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPositiveResourceShouldBe(t *testing.T) {
	var (
		err               error
		g                 = gomega.NewWithT(t)
		dynScheme         = runtime.NewScheme()
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		//fakeClient          *fake.Clientset
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

	// kc := ClientSet{
	// 	DynamicInterface:   fakeDynamicClient,
	// 	DiscoveryInterface: &fakeDiscovery,
	// 	KubeInterface:      fakeClient,
	// 	FilesPath:          "../../test/templates",
	// }
	resource, err := GetResource(&fakeDiscovery, nil, resourcePath(fileName))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = ResourceShouldBe(fakeDynamicClient, resource, common.WaiterConfig{}, common.StateCreated)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	fakeDiscovery.ReactionChain[0] = &kTesting.SimpleReactor{
		Verb:     "get",
		Resource: "someResource",
		Reaction: deletedReactionFunc,
	}

	err = ResourceShouldBe(fakeDynamicClient, resource, common.WaiterConfig{}, common.StateDeleted)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPositiveResourceShouldConvergeToSelector(t *testing.T) {

	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		//fakeClient        *fake.Clientset
		testResource    *unstructured.Unstructured
		getReactionFunc = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
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

	// kc := ClientSet{
	// 	DynamicInterface:   fakeDynamicClient,
	// 	DiscoveryInterface: &fakeDiscovery,
	// 	KubeInterface:      fakeClient,
	// 	FilesPath:          "../../test/templates",
	// }
	resource, err := GetResource(&fakeDiscovery, nil, resourcePath(fileName))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = ResourceShouldConvergeToSelector(fakeDynamicClient, resource, common.WaiterConfig{}, selector)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
}

func TestPositiveResourceConditionShouldBe(t *testing.T) {

	var (
		err               error
		g                 = gomega.NewWithT(t)
		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
		//fakeClient        *fake.Clientset
		testResource    *unstructured.Unstructured
		getReactionFunc = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
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

	// kc := ClientSet{
	// 	DynamicInterface:   fakeDynamicClient,
	// 	DiscoveryInterface: &fakeDiscovery,
	// 	KubeInterface:      fakeClient,
	// 	FilesPath:          "../../test/templates",
	// }
	resource, err := GetResource(&fakeDiscovery, nil, resourcePath(fileName))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = ResourceConditionShouldBe(fakeDynamicClient, resource, common.WaiterConfig{}, testConditionType, testConditionStatus)
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
		//fakeClient        *fake.Clientset
		testResource    *unstructured.Unstructured
		getReactionFunc = func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
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

	// kc := ClientSet{
	// 	DynamicInterface:   fakeDynamicClient,
	// 	DiscoveryInterface: &fakeDiscovery,
	// 	KubeInterface:      fakeClient,
	// 	FilesPath:          "../../test/templates",
	// }
	resource, err := GetResource(&fakeDiscovery, nil, resourcePath(fileName))
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	err = UpdateResourceWithField(fakeDynamicClient, resource, testUpdateKeyChain, testUpdateValue)
	g.Expect(err).ShouldNot(gomega.HaveOccurred())

	expectedLabelValue, found, err := unstructured.NestedString(testResource.UnstructuredContent(), "metadata", "labels", "testUpdateKey")
	g.Expect(err).ShouldNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(expectedLabelValue).To(gomega.Equal(testUpdateValue))
}

func TestResourceOperationInNamespace(t *testing.T) {
	type clientFields struct {
		DynamicInterface dynamic.Interface
	}
	type funcArgs struct {
		operation string
		namespace string
		resource  unstructuredResource
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
				namespace: "test-namespace",
				resource: unstructuredResource{
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
				namespace: "test-namespace",
				resource: unstructuredResource{
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
				namespace: "test-namespace",
				resource: unstructuredResource{
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
				resource: unstructuredResource{
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
				resource: unstructuredResource{
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
				namespace: "override-namespace",
				resource: unstructuredResource{
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
				resource: unstructuredResource{
					GVR:      &meta.RESTMapping{},
					Resource: resourceNs,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// kc := &ClientSet{
			// 	DynamicInterface: tt.clientFields.DynamicInterface,
			// }
			if err := ResourceOperationInNamespace(tt.clientFields.DynamicInterface, tt.funcArgs.resource, tt.funcArgs.operation, tt.funcArgs.namespace); (err != nil) != tt.wantErr {
				t.Errorf("ResourceOperationInNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func resourcePath(resourceFileName string) string {
	return filepath.Join("../../../test/templates", resourceFileName)
}

// TODO: do not return error, receive t *testing.T and fail instead
func resourceFromYaml(resourceFileName string) (*unstructured.Unstructured, error) {
	resourcePath := resourcePath(resourceFileName)
	d, err := os.ReadFile(resourcePath)
	if err != nil {
		return nil, err
	}
	return resourceFromBytes(d)
}

// TODO: do not return error, receive t *testing.T and fail instead
func resourceFromBytes(bytes []byte) (*unstructured.Unstructured, error) {
	resource := &unstructured.Unstructured{}
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err := dec.Decode(bytes, nil, resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// TODO: this is implemented twice, maybe have a test helper pkg?
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

// TODO: do not log error, receive t *testing.T and fail instead
func addLabel(in *unstructured.Unstructured, key, value string) {
	labels, _, _ := unstructured.NestedMap(in.Object, "metadata", "labels")

	labels[key] = value

	err := unstructured.SetNestedMap(in.Object, labels, "metadata", "labels")
	if err != nil {
		log.Errorf("Failed adding label %v=%v to the resource %v: %v", key, value, in.GetName(), err)
	}
}
