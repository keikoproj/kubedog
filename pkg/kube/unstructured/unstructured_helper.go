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
	"context"
	"fmt"
	"html/template"
	"os"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

const (
	yamlSeparator = "\n---"
	trimTokens    = "\n "
)

type unstructuredResource struct {
	GVR      *meta.RESTMapping
	Resource *unstructured.Unstructured
}

func GetResource(dc discovery.DiscoveryInterface, TemplateArguments interface{}, resourceFilePath string) (unstructuredResource, error) {
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return unstructuredResource{nil, nil}, err
	}
	return getResourceFromString(string(data), dc, TemplateArguments)
}

func GetResources(dc discovery.DiscoveryInterface, TemplateArguments interface{}, resourcesFilePath string) ([]unstructuredResource, error) {
	data, err := os.ReadFile(resourcesFilePath)
	if err != nil {
		return nil, err
	}
	manifests := bytes.Split(data, []byte(yamlSeparator))
	resourceList := make([]unstructuredResource, 0)
	for _, manifest := range manifests {
		if len(bytes.Trim(manifest, trimTokens)) == 0 {
			continue
		}
		resource, err := getResourceFromString(string(manifest), dc, TemplateArguments)
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, resource)
	}
	return resourceList, err
}

func ListInstanceGroups(dynamicClient dynamic.Interface) (*unstructured.UnstructuredList, error) {
	const (
		instanceGroupNamespace   = "instance-manager"
		customResourceGroup      = "instancemgr"
		customResourceAPIVersion = "v1alpha1"
		customeResourceDomain    = "keikoproj.io"
		customResourceKind       = "instancegroups"
	)
	var (
		customResourceName    = fmt.Sprintf("%v.%v", customResourceGroup, customeResourceDomain)
		instanceGroupResource = schema.GroupVersionResource{Group: customResourceName, Version: customResourceAPIVersion, Resource: customResourceKind}
	)
	igs, err := dynamicClient.Resource(instanceGroupResource).Namespace(instanceGroupNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return igs, nil
}

func validateDynamicClient(dynamicClient dynamic.Interface) error {
	if dynamicClient == nil {
		return errors.Errorf("'k8s.io/client-go/dynamic.Interface' is nil.")
	}
	return nil
}

func getResourceFromString(resourceString string, dc discovery.DiscoveryInterface, args interface{}) (unstructuredResource, error) {
	resource := &unstructured.Unstructured{}
	var renderBuffer bytes.Buffer

	if args != nil {
		template, err := template.New("Resource").Parse(resourceString)
		if err != nil {
			return unstructuredResource{GVR: nil, Resource: resource}, err
		}

		err = template.Execute(&renderBuffer, &args)
		if err != nil {
			return unstructuredResource{GVR: nil, Resource: resource}, err
		}
	} else {
		renderBuffer.WriteString(resourceString)
	}

	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(renderBuffer.Bytes(), nil, resource)
	if err != nil {
		return unstructuredResource{GVR: nil, Resource: resource}, err
	}
	gvr, err := getGVR(gvk, dc)
	if err != nil {
		return unstructuredResource{GVR: nil, Resource: resource}, err
	}
	return unstructuredResource{GVR: gvr, Resource: resource}, err
}

func getGVR(gvk *schema.GroupVersionKind, dc discovery.DiscoveryInterface) (*meta.RESTMapping, error) {
	if dc == nil {
		return nil, errors.Errorf("'k8s.io/client-go/discovery.DiscoveryInterface' is nil.")
	}

	CachedDiscoveryInterface := memory.NewMemCacheClient(dc)
	DeferredDiscoveryRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(CachedDiscoveryInterface)
	RESTMapping, err := DeferredDiscoveryRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)

	if err != nil {
		return nil, err
	}

	return RESTMapping, nil
}
