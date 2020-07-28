package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

func GetResourceFromYaml(path string, dc discovery.DiscoveryInterface) (*meta.RESTMapping, *unstructured.Unstructured, error) {
	resource := &unstructured.Unstructured{}

	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, resource, err
	}

	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	_, gvk, err := dec.Decode(d, nil, resource)

	if err != nil {
		return nil, resource, err
	}

	gvr, err := FindGVR(gvk, dc)
	if err != nil {
		return nil, resource, err
	}

	return gvr, resource, nil
}
