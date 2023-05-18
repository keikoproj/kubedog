package unstructured

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
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
