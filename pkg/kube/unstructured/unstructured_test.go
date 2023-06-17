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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/keikoproj/kubedog/internal/util"
	"github.com/keikoproj/kubedog/pkg/generic"
	"github.com/keikoproj/kubedog/pkg/kube/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	kTesting "k8s.io/client-go/testing"
)

func TestResourceOperationInNamespace(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
		resource      unstructuredResource
		operation     string
		namespace     string
	}
	resource := getResourceFromYaml(t, getFilePath("resource.yaml"))
	resourceNoNs := getResourceFromYaml(t, getFilePath("resource-no-ns.yaml"))
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: Create/Submit resource, ns in file",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				operation:     common.OperationCreate,
				namespace:     "",
			},
		},
		{
			name: "Negative Test: invalid client",
			args: args{
				dynamicClient: nil,
			},
			wantErr: true,
		},
		{
			name: "Positive Test: Create/Submit resource already created",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				operation:     common.OperationSubmit,
				namespace:     "",
			},
		},
		{
			name: "Positive Test: Update resource, ns in file",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				operation:     common.OperationUpdate,
				namespace:     "",
			},
		},
		{
			name: "Negative Test: Update resource, 'Get' call fails",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				operation:     common.OperationUpdate,
				namespace:     "",
			},
			wantErr: true,
		},
		{
			name: "Negative Test: Update resource, 'Update' call fails",
			args: args{
				dynamicClient: newFakeDynamicClientWithReactors(
					&kTesting.SimpleReactor{
						Verb:     "get",
						Resource: resource.Resource.GetName(),
						Reaction: newReactionFunc(),
					},
					&kTesting.SimpleReactor{
						Verb:     "update",
						Resource: resource.Resource.GetName(),
						Reaction: newReactionFuncWithError(errors.New("an error")),
					},
				),
				resource:  resource,
				operation: common.OperationUpdate,
				namespace: "",
			},
			wantErr: true,
		},
		{
			name: "Positive Test: Delete resource, ns in file",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				operation:     common.OperationDelete,
				namespace:     "",
			},
		},
		{
			name: "Positive Test: Delete resource, 'Not Found'",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				operation:     common.OperationDelete,
				namespace:     "",
			},
		},
		{
			name: "Negative Test: invalid operation",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				operation:     "invalid-operation",
				namespace:     "",
			},
			wantErr: true,
		},
		{
			name: "Positive Test: Operation on resource, ns as parameter",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resourceNoNs,
				operation:     common.OperationCreate,
				namespace:     "any-namespace",
			},
		},
		{
			name: "Negative Test: Operation on resource, ns in file and as parameter",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				operation:     common.OperationCreate,
				namespace:     "any-namespace",
			},
			wantErr: true,
		},
		{
			name: "Positive Test: Operation on resource, no ns",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resourceNoNs,
				operation:     common.OperationCreate,
				namespace:     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResourceOperationInNamespace(tt.args.dynamicClient, tt.args.resource, tt.args.operation, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("ResourceOperationInNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TODO: why have Resources* and Resource* when the multi-resource func can handle a single?
func TestResourcesOperation(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
		resources     []unstructuredResource
		operation     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: Operation on multi-resource file",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resources:     getResourcesFromYaml(t, getFilePath("multi-resource.yaml")),
				operation:     common.OperationCreate,
			},
		},
		{
			name: "Negative Test: 'ResourceOperationInNamespace' fails",
			args: args{
				dynamicClient: nil,
				resources:     getResourcesFromYaml(t, getFilePath("multi-resource.yaml")),
				operation:     common.OperationCreate,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResourcesOperation(tt.args.dynamicClient, tt.args.resources, tt.args.operation); (err != nil) != tt.wantErr {
				t.Errorf("ResourcesOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourcesOperationInNamespace(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
		resources     []unstructuredResource
		operation     string
		namespace     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: Operation on multi-resource with ns in file ",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resources:     getResourcesFromYaml(t, getFilePath("multi-resource.yaml")),
				operation:     common.OperationCreate,
				namespace:     "",
			},
		},
		{
			name: "Positive Test: Operation on multi-resource file with ns as parameter",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resources:     getResourcesFromYaml(t, getFilePath("multi-resource-no-ns.yaml")),
				operation:     common.OperationCreate,
				namespace:     "any-namespace",
			},
		},
		{
			name: "Negative Test: Operation on multi-resource with ns in file and as parameter",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resources:     getResourcesFromYaml(t, getFilePath("multi-resource.yaml")),
				operation:     common.OperationCreate,
				namespace:     "any-namespace",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResourcesOperationInNamespace(tt.args.dynamicClient, tt.args.resources, tt.args.operation, tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("ResourcesOperationInNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceOperationWithResultInNamespace(t *testing.T) {
	type args struct {
		dynamicClient  dynamic.Interface
		resource       unstructuredResource
		operation      string
		namespace      string
		expectedResult string
	}
	const (
		expectedResultSucceed = "succeed"
		expectedResultFail    = "fail"
	)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: got what expected",
			args: args{
				dynamicClient:  newFakeDynamicClient(),
				resource:       getResourceFromYaml(t, getFilePath("resource.yaml")),
				operation:      common.OperationCreate,
				namespace:      "",
				expectedResult: expectedResultSucceed,
			},
		},
		{
			name: "Negative Test: expectedResultSucceed but failed",
			args: args{
				dynamicClient:  newFakeDynamicClient(),
				resource:       getResourceFromYaml(t, getFilePath("resource.yaml")),
				operation:      common.OperationCreate,
				namespace:      "any-namespace",
				expectedResult: expectedResultSucceed,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: expectedResultFail but succeeded",
			args: args{
				dynamicClient:  newFakeDynamicClient(),
				resource:       getResourceFromYaml(t, getFilePath("resource.yaml")),
				operation:      common.OperationCreate,
				namespace:      "",
				expectedResult: expectedResultFail,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResourceOperationWithResultInNamespace(tt.args.dynamicClient, tt.args.resource, tt.args.operation, tt.args.namespace, tt.args.expectedResult); (err != nil) != tt.wantErr {
				t.Errorf("ResourceOperationWithResultInNamespace() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceShouldBe(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
		resource      unstructuredResource
		w             common.WaiterConfig
		state         string
	}
	resource := getResourceFromYaml(t, getFilePath("resource.yaml"))
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: StateCreated",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				state:         common.StateCreated,
			},
		},
		{
			name: "Negative Test: Error 'Not Found'",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				state:         common.StateCreated,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: Error different than 'Not Found'",
			args: args{
				dynamicClient: newFakeDynamicClientWithReaction(
					"get",
					resource.Resource.GetName(),
					newReactionFuncWithError(errors.New("an error")),
				),
				resource: resource,
				state:    common.StateCreated,
			},
			wantErr: true,
		},
		{
			name: "Positive Test: StateDeleted",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				state:         common.StateDeleted,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.w = common.NewWaiterConfig(1, time.Second)
			if err := ResourceShouldBe(tt.args.dynamicClient, tt.args.resource, tt.args.w, tt.args.state); (err != nil) != tt.wantErr {
				t.Errorf("ResourceShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceShouldConvergeToSelector(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
		resource      unstructuredResource
		w             common.WaiterConfig
		selector      string
	}
	resource := getResourceFromYaml(t, getFilePath("resource.yaml"))
	labelKey, labelValue := getOneLabel(t, *resource.Resource)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				selector:      ".metadata.labels." + labelKey + "=" + labelValue,
			},
		},
		{
			name: "Negative Test: invalid client",
			args: args{
				dynamicClient: nil,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: invalid selector",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				selector:      ".invalid.selector.",
			},
			wantErr: true,
		},
		{
			name: "Negative Test: invalid key",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				selector:      ".=invalid-key",
			},
			wantErr: true,
		},
		{
			name: "Negative Test: 'Get' call fails",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				selector:      ".metadata.labels." + labelKey + "=" + labelValue,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: waiter timed out, selector not found",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				selector:      ".metadata.labels.someNotFoundKey=someNotFoundValue",
			},
			wantErr: true,
		},
		{
			name: "Negative Test: waiter timed out, value does not match",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				selector:      ".metadata.labels." + labelKey + "=someNotFoundValue",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.w = common.NewWaiterConfig(1, time.Second)
			if err := ResourceShouldConvergeToSelector(tt.args.dynamicClient, tt.args.resource, tt.args.w, tt.args.selector); (err != nil) != tt.wantErr {
				t.Errorf("ResourceShouldConvergeToSelector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceConditionShouldBe(t *testing.T) {
	type args struct {
		dynamicClient  dynamic.Interface
		resource       unstructuredResource
		w              common.WaiterConfig
		conditionType  string
		conditionValue string
	}
	resource := getResourceFromYaml(t, getFilePath("resource.yaml"))
	conditionType, conditionStatus := getOneCondition(t, *resource.Resource)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test",
			args: args{
				dynamicClient:  newFakeDynamicClientWithResource(resource),
				resource:       resource,
				conditionType:  conditionType,
				conditionValue: conditionStatus,
			},
		},
		{
			name: "Negative Test: invalid client",
			args: args{
				dynamicClient: nil,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: 'Get' call fails",
			args: args{
				dynamicClient:  newFakeDynamicClient(),
				resource:       resource,
				conditionType:  conditionType,
				conditionValue: conditionStatus,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: waiter timed out, condition type does not match",
			args: args{
				dynamicClient:  newFakeDynamicClientWithResource(resource),
				resource:       resource,
				conditionType:  "conditionTypeNotFound",
				conditionValue: conditionStatus,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.w = common.NewWaiterConfig(1, time.Second)
			if err := ResourceConditionShouldBe(tt.args.dynamicClient, tt.args.resource, tt.args.w, tt.args.conditionType, tt.args.conditionValue); (err != nil) != tt.wantErr {
				t.Errorf("ResourceConditionShouldBe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateResourceWithField(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
		resource      unstructuredResource
		key           string
		value         string
	}
	resource := getResourceFromYaml(t, getFilePath("resource.yaml"))
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: string value",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				key:           ".metadata.labels.someNewLabelKey",
				value:         "someNewLabelValue",
			},
		},
		{
			name: "Positive Test: integer value",
			args: args{
				dynamicClient: newFakeDynamicClientWithResource(resource),
				resource:      resource,
				key:           ".metadata.labels.someNewIntegerLabelKey",
				value:         "1",
			},
		},
		{
			name: "Negative Test: invalid client",
			args: args{
				dynamicClient: nil,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: 'Get' call fails",
			args: args{
				dynamicClient: newFakeDynamicClient(),
				resource:      resource,
				key:           ".metadata.labels.someNewLabelKey",
				value:         "someNewLabelValue",
			},
			wantErr: true,
		},
		{
			name: "Negative Test: 'Update' call fails",
			args: args{
				dynamicClient: newFakeDynamicClientWithReactors(
					&kTesting.SimpleReactor{
						Verb:     "get",
						Resource: resource.Resource.GetName(),
						Reaction: newReactionFunc(),
					},
					&kTesting.SimpleReactor{
						Verb:     "update",
						Resource: resource.Resource.GetName(),
						Reaction: newReactionFuncWithError(errors.New("an error")),
					},
				),
				resource: resource,
				key:      ".metadata.labels.someNewLabelKey",
				value:    "someNewLabelValue",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateResourceWithField(tt.args.dynamicClient, tt.args.resource, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("UpdateResourceWithField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteResourcesAtPath(t *testing.T) {
	type args struct {
		dynamicClient     dynamic.Interface
		dc                discovery.DiscoveryInterface
		TemplateArguments interface{}
		w                 common.WaiterConfig
		resourcesPath     string
	}

	resourcePath := getFilePath("resource.yaml")
	resource := getResourceFromYaml(t, resourcePath)
	client := newFakeDynamicClientWithResourceList(resource)
	clientWithResource := newFakeDynamicClientWithResourcesAndResourcesLists(resource)

	resourcesPath := getFilePath("multi-resource.yaml")
	resources := getResourcesFromYaml(t, resourcesPath)
	clientWithResources := newFakeDynamicClientWithResourcesAndResourcesLists(resources...)
	clientWithResourcesInTestDir := newFakeDynamicClientWithResourcesInDir(t, getTestFilesDirPath())

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: delete resource in directory",
			args: args{
				dynamicClient: clientWithResourcesInTestDir,
				dc:            newFakeDiscoveryClient(&clientWithResourcesInTestDir.Fake),
				resourcesPath: getTestFilesDirPath(),
			},
		},
		{
			name: "Positive Test: delete resource",
			args: args{
				dynamicClient: clientWithResource,
				dc:            newFakeDiscoveryClient(&clientWithResource.Fake),
				resourcesPath: resourcePath,
			},
		},
		{
			name: "Positive Test: already deleted",
			args: args{
				dynamicClient: client,
				dc:            newFakeDiscoveryClient(&client.Fake),
				resourcesPath: resourcePath,
			},
		},
		{
			name: "Positive Test: delete multiple resources",
			args: args{
				dynamicClient: clientWithResources,
				dc:            newFakeDiscoveryClient(&clientWithResources.Fake),
				resourcesPath: resourcesPath,
			},
		},
		{
			name: "Negative Test: invalid client",
			args: args{
				dynamicClient: nil,
			},
			wantErr: true,
		},
		{
			name: "Negative Test: 'GetResources' fails",
			args: args{
				dynamicClient: client,
				dc:            nil,
				resourcesPath: resourcePath,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.w = common.NewWaiterConfig(1, time.Second)
			if err := DeleteResourcesAtPath(tt.args.dynamicClient, tt.args.dc, tt.args.TemplateArguments, tt.args.w, tt.args.resourcesPath); (err != nil) != tt.wantErr {
				t.Errorf("DeleteResourcesAtPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyInstanceGroups(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test: .status.currentState=Ready",
			args: args{
				dynamicClient: newFakeDynamicClientWithCustomListKinds(
					getInstanceGroupFromYaml(t,
						getFilePath("instance-group.yaml"),
					),
				),
			},
		},
		{
			name: "Negative Test: .status.currentState=NotReady",
			args: args{
				dynamicClient: newFakeDynamicClientWithCustomListKinds(
					getInstanceGroupFromYaml(t,
						getFilePath("instance-group-not-ready.yaml"),
					),
				),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyInstanceGroups(tt.args.dynamicClient); (err != nil) != tt.wantErr {
				t.Errorf("VerifyInstanceGroups() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetResource(t *testing.T) {
	type args struct {
		dc                discovery.DiscoveryInterface
		TemplateArguments interface{}
		resourceFilePath  string
	}
	resourcePath := getFilePath("resource.yaml")
	resource := getResourceFromYaml(t, resourcePath)
	templatedPath := getTemplatedFilePath("templated.yaml")
	templateArgs := []generic.TemplateArgument{
		{
			Key:     "Kind",
			Default: "myKind",
		},
		{
			Key:     "ApiVersion",
			Default: "myApiVersion",
		},
		{
			Key:     "Name",
			Default: "myName",
		},
	}
	templateMap := templateArgsToMap(t, templateArgs...)
	generatedPath := generateFileFromTemplate(t, templatedPath, templateMap)
	generatedResource := getResourceFromYaml(t, generatedPath)
	analysisTemplatePath := getFilePath("analysis-template.yaml")
	analysisTemplateResource := getResourceFromYaml(t, analysisTemplatePath)
	tests := []struct {
		name    string
		args    args
		want    unstructuredResource
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test",
			args: args{
				dc:                newFakeDiscoveryClient(&newFakeDynamicClientWithResourceList(resource).Fake),
				TemplateArguments: nil,
				resourceFilePath:  resourcePath,
			},
			want: resource,
		},
		{
			name: "Positive Test: templated",
			args: args{
				dc:                newFakeDiscoveryClient(&newFakeDynamicClientWithResourceList(generatedResource).Fake),
				TemplateArguments: templateMap,
				resourceFilePath:  templatedPath,
			},
			want: generatedResource,
		},
		{
			name: "Positive Test: templated file BUT no template arguments",
			args: args{
				dc:                newFakeDiscoveryClient(&newFakeDynamicClientWithResourceList(analysisTemplateResource).Fake),
				TemplateArguments: nil,
				resourceFilePath:  analysisTemplatePath,
			},
			want: analysisTemplateResource,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResource(tt.args.dc, tt.args.TemplateArguments, tt.args.resourceFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResource() = %s, want %s", util.StructToPrettyString(got), util.StructToPrettyString(tt.want))

			}
		})
	}
}

func TestGetResources(t *testing.T) {
	type args struct {
		dc                discovery.DiscoveryInterface
		TemplateArguments interface{}
		resourcesFilePath string
	}
	resourcesPath := getFilePath("multi-resource.yaml")
	resources := getResourcesFromYaml(t, resourcesPath)
	tests := []struct {
		name    string
		args    args
		want    []unstructuredResource
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test",
			args: args{
				dc:                newFakeDiscoveryClient(&newFakeDynamicClientWithResourcesLists(resources...).Fake),
				TemplateArguments: nil,
				resourcesFilePath: resourcesPath,
			},
			want:    resources,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResources(tt.args.dc, tt.args.TemplateArguments, tt.args.resourcesFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListInstanceGroups(t *testing.T) {
	type args struct {
		dynamicClient dynamic.Interface
	}
	resource := getInstanceGroupFromYaml(t, getFilePath("instance-group.yaml"))
	tests := []struct {
		name    string
		args    args
		want    *unstructured.UnstructuredList
		wantErr bool
	}{
		// TODO: add negative tests
		{
			name: "Positive Test",
			args: args{
				dynamicClient: newFakeDynamicClientWithCustomListKinds(resource),
			},
			want: resourceToList(t, resource),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListInstanceGroups(tt.args.dynamicClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListInstanceGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListInstanceGroups() = %s, want %s", util.StructToPrettyString(got), util.StructToPrettyString(tt.want))
			}
		})
	}
}

func templateArgsToMap(t *testing.T, args ...generic.TemplateArgument) map[string]string {
	argsMap, err := generic.TemplateArgumentsToMap(args...)
	if err != nil {
		t.Error(err)
	}
	return argsMap
}

func generateFileFromTemplate(t *testing.T, templatedFilePath string, templateArgs interface{}) string {
	generatedPath, err := generic.GenerateFileFromTemplate(templatedFilePath, templateArgs)
	if err != nil {
		t.Error(err)
	}
	return generatedPath
}

func resourceToList(t *testing.T, resource unstructuredResource) *unstructured.UnstructuredList {
	unstructList := unstructured.Unstructured{}
	unstructList.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": resource.Resource.GetAPIVersion(),
		"kind":       resource.Resource.GetKind() + "List",
		"metadata": map[string]interface{}{
			"resourceVersion": "",
		},
	})
	list, err := unstructList.ToList()
	if err != nil {
		t.Error(err)
	}
	list.Items = append(list.Items, *resource.Resource)
	return list
}

func getTestDirPath() string {
	return "./test"
}

func getTestFilesDirPath() string {
	return filepath.Join(getTestDirPath(), "files")
}

func getFilePath(testFileName string) string {
	return filepath.Join(getTestFilesDirPath(), testFileName)
}
func getTemplatedFilePath(testFileName string) string {
	return filepath.Join(getTestDirPath(), "templates", testFileName)
}

func getInstanceGroupFromYaml(t *testing.T, resourceFilePath string) unstructuredResource {
	rawResource, err := os.ReadFile(resourceFilePath)
	if err != nil {
		t.Error(err)
	}
	resource := &unstructured.Unstructured{}
	decoder := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := decoder.Decode(rawResource, nil, resource)
	if err != nil {
		t.Error(err)
	}
	gvr := &meta.RESTMapping{
		Resource: schema.GroupVersionResource{
			Group:   gvk.Group,
			Version: gvk.Version,
			// required for 'fakeDynamic' client verb 'List' to work properly
			Resource: "instancegroups",
		},
		GroupVersionKind: *gvk,
	}
	return unstructuredResource{GVR: gvr, Resource: resource}
}

func getResourceFromBytes(t *testing.T, rawResource []byte) unstructuredResource {
	resource := &unstructured.Unstructured{}
	decoder := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := decoder.Decode(rawResource, nil, resource)
	if err != nil {
		t.Error(err)
	}
	scope := meta.RESTScopeNamespace
	if resource.GetNamespace() == "" {
		scope = meta.RESTScopeRoot
	}
	gvr := &meta.RESTMapping{
		Resource: schema.GroupVersionResource{
			Group:   gvk.Group,
			Version: gvk.Version,
			// TODO: fix this, Resource is not the name
			Resource: resource.GetName(),
		},
		GroupVersionKind: *gvk,
		Scope:            scope,
	}
	return unstructuredResource{GVR: gvr, Resource: resource}
}

func getResourceFromYaml(t *testing.T, resourceFilePath string) unstructuredResource {
	rawResource, err := os.ReadFile(resourceFilePath)
	if err != nil {
		t.Error(err)
	}
	return getResourceFromBytes(t, rawResource)
}

func getResourcesFromYaml(t *testing.T, resourcesFilePath string) []unstructuredResource {
	rawResources, err := os.ReadFile(resourcesFilePath)
	if err != nil {
		t.Error(err)
	}
	rawResourcesSplit := bytes.Split(rawResources, []byte(yamlSeparator))
	resources := make([]unstructuredResource, 0)
	for _, rawResource := range rawResourcesSplit {
		if len(bytes.Trim(rawResource, trimTokens)) == 0 {
			continue
		}
		resource := getResourceFromBytes(t, rawResource)
		resources = append(resources, resource)
	}
	return resources
}

func newFakeDiscoveryClient(client *kTesting.Fake) *fakeDiscovery.FakeDiscovery {
	return &fakeDiscovery.FakeDiscovery{Fake: client}
}

func newFakeDynamicClient() *fakeDynamic.FakeDynamicClient {
	return fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
}

func newFakeDynamicClientWithCustomListKinds(resource unstructuredResource) *fakeDynamic.FakeDynamicClient {
	return fakeDynamic.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			// must end in 'List': https://github.com/kubernetes/client-go/blob/1309f64d6648411b4a36a2f7fa84dd8df31884b6/dynamic/fake/simple.go#L92
			resource.GVR.Resource: resource.Resource.GetKind() + "List",
		},
		resource.Resource,
	)
}

func newFakeDynamicClientWithResourceList(resource unstructuredResource) *fakeDynamic.FakeDynamicClient {
	client := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())

	namespaced := true
	if resource.Resource.GetNamespace() == "" {
		namespaced = false
	}
	client.Resources = append(client.Resources, newAPIResourceList(
		resource.GVR.GroupVersionKind.GroupVersion(),
		resource.Resource.GetName(),
		resource.Resource.GetKind(),
		namespaced,
	))
	return client
}

func newFakeDynamicClientWithResourcesLists(resources ...unstructuredResource) *fakeDynamic.FakeDynamicClient {
	client := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	for _, resource := range resources {
		namespaced := true
		if resource.Resource.GetNamespace() == "" {
			namespaced = false
		}
		client.Resources = append(client.Resources, newAPIResourceList(
			resource.GVR.GroupVersionKind.GroupVersion(),
			resource.Resource.GetName(),
			resource.Resource.GetKind(),
			namespaced,
		))
	}
	return client
}

func newFakeDynamicClientWithResourcesInDir(t *testing.T, dirPath string) *fakeDynamic.FakeDynamicClient {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		t.Error(err)
	}
	allResources := []unstructuredResource{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yaml") {
			resources := getResourcesFromYaml(t, filepath.Join(dirPath, file.Name()))
			allResources = append(allResources, resources...)
		}
	}
	if len(allResources) == 0 {
		t.Errorf("No resources found in '%s'", dirPath)
	}
	return newFakeDynamicClientWithResourcesAndResourcesLists(allResources...)
}

func newFakeDynamicClientWithResourcesAndResourcesLists(resources ...unstructuredResource) *fakeDynamic.FakeDynamicClient {
	uniqueResources := map[string]unstructuredResource{}
	for _, resource := range resources {
		key := /*resource.Resource.GetNamespace() + "/" +*/ resource.Resource.GetKind() + "/" + resource.Resource.GetName()
		if _, ok := uniqueResources[key]; !ok {
			uniqueResources[key] = resource
		}
	}

	client := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	for _, resource := range uniqueResources {
		_ = client.Tracker().Create(resource.GVR.Resource, resource.Resource, resource.Resource.GetNamespace())
		namespaced := true
		if resource.Resource.GetNamespace() == "" {
			namespaced = false
		}
		client.Resources = append(client.Resources, newAPIResourceList(
			resource.GVR.GroupVersionKind.GroupVersion(),
			resource.Resource.GetName(),
			resource.Resource.GetKind(),
			namespaced,
		))
	}
	return client
}

func newFakeDynamicClientWithReaction(verb, resource string, fn kTesting.ReactionFunc) *fakeDynamic.FakeDynamicClient {
	client := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	client.PrependReactor(verb, resource, fn)
	return client
}
func newFakeDynamicClientWithReactors(reactors ...kTesting.Reactor) *fakeDynamic.FakeDynamicClient {
	client := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	client.ReactionChain = append(reactors, client.ReactionChain...)
	return client
}

func newFakeDynamicClientWithResource(resource unstructuredResource) *fakeDynamic.FakeDynamicClient {
	client := fakeDynamic.NewSimpleDynamicClient(runtime.NewScheme())
	_ = client.Tracker().Create(resource.GVR.Resource, resource.Resource, resource.Resource.GetNamespace())
	return client
}

func newReactionFunc() kTesting.ReactionFunc {
	return func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &unstructured.Unstructured{}, nil
	}
}

func newReactionFuncWithError(retErr error) kTesting.ReactionFunc {
	return func(action kTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, retErr
	}
}

func newAPIResourceList(groupVersion schema.GroupVersion, name, kind string, namespaced bool) *metav1.APIResourceList {
	return &metav1.APIResourceList{
		GroupVersion: groupVersion.String(),
		APIResources: []metav1.APIResource{
			{
				Name:       name,
				Kind:       kind,
				Namespaced: namespaced,
			},
		},
	}
}

func getOneLabel(t *testing.T, resource unstructured.Unstructured) (string, string) {
	labelsMap := resource.GetLabels()
	for key, value := range labelsMap {
		if key == "" || value == "" {
			continue
		}
		return key, value
	}
	t.Errorf(("No labels found in resource: '%v'"), resource)
	return "", ""
}

func getOneCondition(t *testing.T, resource unstructured.Unstructured) (string, string) {
	if conditions, ok, err := unstructured.NestedSlice(resource.UnstructuredContent(), "status", "conditions"); ok {
		if err != nil {
			t.Error(err)
		}
		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			tp, ok := condition["type"].(string)
			if !ok {
				continue
			}
			status, ok := condition["status"].(string)
			if !ok {
				continue
			}
			return tp, status
		}
	}
	t.Errorf(("No conditions found in resource: '%v'"), resource)
	return "", ""
}
