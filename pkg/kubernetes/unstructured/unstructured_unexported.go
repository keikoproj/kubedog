package unstructured

import (
	"bytes"
	"html/template"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
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
