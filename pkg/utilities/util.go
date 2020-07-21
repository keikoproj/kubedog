package util

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

type TemplateArguments struct {
	ClusterName        string
	KeyPairName        string
	AmiID              string
	NodeRole           string
	NodeRoleArn        string
	NodeSecurityGroups []string
	Subnets            []string
}

func NewTemplateArguments() *TemplateArguments {
	return &TemplateArguments{
		ClusterName:        os.Getenv("EKS_CLUSTER"),
		KeyPairName:        os.Getenv("KEYPAIR_NAME"),
		AmiID:              os.Getenv("AMI_ID"),
		NodeRole:           os.Getenv("NODE_ROLE"),
		NodeRoleArn:        os.Getenv("NODE_ROLE_ARN"),
		NodeSecurityGroups: strings.Split(os.Getenv("SECURITY_GROUPS"), ","),
		Subnets:            strings.Split(os.Getenv("NODE_SUBNETS"), ","),
	}
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

	// DiscoveryClient queries API server about the resources
	/*dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}*/
	CachedDiscoveryInterface := memory.NewMemCacheClient(dc)
	DeferredDiscoveryRESTMapper := restmapper.NewDeferredDiscoveryRESTMapper(CachedDiscoveryInterface)
	RESTMapping, err := DeferredDiscoveryRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)

	if err != nil {
		return nil, err
	}

	return RESTMapping, nil
}

func GetResourceFromYaml(path string, dc discovery.DiscoveryInterface, args *TemplateArguments) (*meta.RESTMapping, *unstructured.Unstructured, error) {
	resource := &unstructured.Unstructured{}

	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, resource, err
	}

	template, err := template.New("ResourceFromYaml").Parse(string(d))
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
