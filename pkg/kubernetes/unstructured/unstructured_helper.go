package unstructured

import (
	"context"
	"fmt"

	util "github.com/keikoproj/kubedog/internal/utilities"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

func GetResource(dc discovery.DiscoveryInterface, TemplateArguments interface{}, resourceFilePath string) (util.K8sUnstructuredResource, error) {
	unstructuredResource, err := util.GetResourceFromYaml(resourceFilePath, dc, TemplateArguments)
	if err != nil {
		return util.K8sUnstructuredResource{}, err
	}
	return unstructuredResource, nil
}

func GetResources(dc discovery.DiscoveryInterface, TemplateArguments interface{}, resourcesFilePath string) ([]util.K8sUnstructuredResource, error) {
	resourceList, err := util.GetResourcesFromYaml(resourcesFilePath, dc, TemplateArguments)
	if err != nil {
		return nil, err
	}
	return resourceList, nil
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
