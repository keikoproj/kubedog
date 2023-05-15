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
	"reflect"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
)

// TODO: most of this needs to be moved to unstructured and to the packages where they are used

const (
	YamlSeparator                  = "\n---"
	TrimTokens                     = "\n "
	clusterNameEnvironmentVariable = "CLUSTER_NAME"
)

var (
	DefaultRetry = wait.Backoff{
		Steps:    6,
		Duration: 1000 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}
	retriableErrors = []string{
		"Unable to reach the kubernetes API",
		"Unable to connect to the server",
		"EOF",
		"transport is closing",
		"the object has been modified",
		"an error on the server",
	}
)

type FuncToRetryWithReturn func() (interface{}, error)
type FuncToRetry func() error

type K8sUnstructuredResource struct {
	GVR      *meta.RESTMapping
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

func GetResourceFromYaml(path string, dc discovery.DiscoveryInterface, args interface{}) (K8sUnstructuredResource, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return K8sUnstructuredResource{nil, nil}, err
	}
	return GetResourceFromString(string(data), dc, args)
}

func GetMultipleResourcesFromYaml(path string, dc discovery.DiscoveryInterface, args interface{}) ([]K8sUnstructuredResource, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	manifests := bytes.Split(data, []byte(YamlSeparator))
	resourceList := make([]K8sUnstructuredResource, 0)
	for _, manifest := range manifests {
		if len(bytes.Trim(manifest, TrimTokens)) == 0 {
			continue
		}
		unstructuredResource, err := GetResourceFromString(string(manifest), dc, args)
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, unstructuredResource)
	}
	return resourceList, err
}

func GetResourceFromString(resourceString string, dc discovery.DiscoveryInterface, args interface{}) (K8sUnstructuredResource, error) {
	resource := &unstructured.Unstructured{}
	var renderBuffer bytes.Buffer

	if args != nil {
		template, err := template.New("Resource").Parse(resourceString)
		if err != nil {
			return K8sUnstructuredResource{nil, resource}, err
		}

		err = template.Execute(&renderBuffer, &args)
		if err != nil {
			return K8sUnstructuredResource{nil, resource}, err
		}
	} else {
		renderBuffer.WriteString(resourceString)
	}

	dec := serializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(renderBuffer.Bytes(), nil, resource)
	if err != nil {
		return K8sUnstructuredResource{nil, resource}, err
	}
	gvr, err := FindGVR(gvk, dc)
	if err != nil {
		return K8sUnstructuredResource{nil, resource}, err
	}
	return K8sUnstructuredResource{gvr, resource}, err
}

func IsRetriable(err error) bool {
	for _, msg := range retriableErrors {
		if strings.Contains(err.Error(), msg) {
			return true
		}
	}
	return false
}

func RetryOnError(backoff *wait.Backoff, retryExpected func(error) bool, fn FuncToRetryWithReturn) (interface{}, error) {
	var ex, lastErr error
	var out interface{}
	caller := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	err := wait.ExponentialBackoff(*backoff, func() (bool, error) {
		out, ex = fn()
		switch {
		case ex == nil:
			return true, nil
		case retryExpected(ex):
			lastErr = ex
			log.Warnf("A caller %v retried due to exception: %v", caller, ex)
			return false, nil
		default:
			return false, ex
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastErr
	}
	return out, err
}

func RetryOnAnyError(backoff *wait.Backoff, fn FuncToRetry) error {
	_, err := RetryOnError(backoff, func(err error) bool {
		return true
	}, func() (interface{}, error) {
		return nil, fn()
	})
	return err
}

func GetClusterName() (string, error) {
	return GetEnvironmentVariable(clusterNameEnvironmentVariable)
}

func GetEnvironmentVariable(envName string) (string, error) {
	if envValue, ok := os.LookupEnv(envName); ok {
		return envValue, nil
	}
	return "", fmt.Errorf("could not get environment variable '%s'", envName)
}
