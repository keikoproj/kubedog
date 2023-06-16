Generated TestGetResource
Generated TestGetResources
Generated TestListInstanceGroups
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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

func TestGetResource(t *testing.T) {
	type args struct {
		dc                discovery.DiscoveryInterface
		TemplateArguments interface{}
		resourceFilePath  string
	}
	tests := []struct {
		name    string
		args    args
		want    unstructuredResource
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResource(tt.args.dc, tt.args.TemplateArguments, tt.args.resourceFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResource() = %v, want %v", got, tt.want)
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
	tests := []struct {
		name    string
		args    args
		want    []unstructuredResource
		wantErr bool
	}{
		// TODO: Add test cases.
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
	tests := []struct {
		name    string
		args    args
		want    *unstructured.UnstructuredList
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListInstanceGroups(tt.args.dynamicClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListInstanceGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListInstanceGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
