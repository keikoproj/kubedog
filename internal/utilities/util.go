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

package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
)

type KubernetesResource struct {
	Gvr      *meta.RESTMapping
	Resource *unstructured.Unstructured
}

func IsNodeReady(n corev1.Node) bool {
	for _, condition := range n.Status.Conditions {
		if condition.Type == "Ready" {
			if condition.Status == "True" {
				return true
			}
		}
	}
	return false
}

func PathToOSFile(relativePath string) (*os.File, error) {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed generate absolute file path of %s", relativePath))
	}

	manifest, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to open file %s", path))
	}

	return manifest, nil
}

func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// find the corresponding GVR (available in *meta.RESTMapping) for gvk
func FindGVR(gvk *schema.GroupVersionKind, dc discovery.DiscoveryInterface) (*meta.RESTMapping, error) {

	CachedDiscoveryInterface := memory.NewMemCacheClient(dc)
	DeferredDiscoveryRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(CachedDiscoveryInterface)
	RESTMapping, err := DeferredDiscoveryRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)

	if err != nil {
		return nil, err
	}

	return RESTMapping, nil
}

func GetResourceFromYaml(path string, dc discovery.DiscoveryInterface, args interface{}) (*meta.RESTMapping, *unstructured.Unstructured, error) {
	resource := &unstructured.Unstructured{}

	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, resource, err
	}
	template, err := template.New("Resource").Parse(string(d))
	if err != nil {
		return nil, resource, err
	}
	var renderBuffer bytes.Buffer
	err = template.Execute(&renderBuffer, &args)
	if err != nil {
		return nil, resource, err
	}
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(renderBuffer.Bytes(), nil, resource)

	if err != nil {
		return nil, resource, err
	}

	gvr, err := FindGVR(gvk, dc)
	if err != nil {
		return nil, resource, err
	}

	return gvr, resource, nil
}

func GetMultipleResourcesFromYaml(path string, dc discovery.DiscoveryInterface, args interface{}) ([]KubernetesResource, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	manifests := bytes.Split(data, []byte("\n---"))
	resourceList := make([]KubernetesResource, 0)
	for _, manifest := range manifests {
		kubernetesResource, err := GetResourceFromString(string(manifest), dc, args)
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, kubernetesResource)
	}
	return resourceList, err
}
func GetResourceFromString(resourceString string, dc discovery.DiscoveryInterface, args interface{}) (KubernetesResource, error) {
	resource := &unstructured.Unstructured{}

	template, err := template.New("Resource").Parse(resourceString)
	if err != nil {
		return KubernetesResource{nil, resource}, err
	}
	var renderBuffer bytes.Buffer
	err = template.Execute(&renderBuffer, &args)
	if err != nil {
		return KubernetesResource{nil, resource}, err
	}
	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(renderBuffer.Bytes(), nil, resource)
	if err != nil {
		return KubernetesResource{nil, resource}, err
	}
	gvr, err := FindGVR(gvk, dc)
	if err != nil {
		return KubernetesResource{nil, resource}, err
	}
	return KubernetesResource{gvr, resource}, err
}
