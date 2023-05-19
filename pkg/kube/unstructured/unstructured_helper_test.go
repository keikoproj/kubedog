package unstructured

import (
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
)

func TestGetResources(t *testing.T) {
	var (
		dynScheme           = runtime.NewScheme()
		fakeDynamicClient   = fakeDynamic.NewSimpleDynamicClient(dynScheme)
		fakeDiscovery       = fakeDiscovery.FakeDiscovery{}
		g                   = gomega.NewWithT(t)
		testTemplatePath, _ = filepath.Abs("../../../test/templates")
	)

	expectedResources := []*metav1.APIResourceList{
		newTestAPIResourceList("someGroup.apiVersion/SomeVersion", "someResource", "SomeKind"),
		newTestAPIResourceList("otherGroup.apiVersion/OtherVersion", "otherResource", "OtherKind"),
		newTestAPIResourceList("argoproj.io/v1alpha1", "AnalysisTemplate", "AnalysisTemplate"),
	}

	fakeDiscovery.Fake = &fakeDynamicClient.Fake
	fakeDiscovery.Resources = append(fakeDiscovery.Resources, expectedResources...)

	resourceToApiResourceList := func(resource *unstructured.Unstructured) *metav1.APIResourceList {
		return newTestAPIResourceList(
			resource.GetAPIVersion(),
			resource.GetName(),
			resource.GetKind(),
		)
	}

	tests := []struct {
		testResourcePath  string
		numResources      int
		expectError       bool
		expectedResources []*metav1.APIResourceList
	}{
		{ // PositiveTest
			testResourcePath: testTemplatePath + "/test-multi-resourcefile.yaml",
			numResources:     2,
			expectError:      false,
			expectedResources: []*metav1.APIResourceList{
				newTestAPIResourceList("someGroup.apiVersion/SomeVersion", "someResource", "SomeKind"),
				newTestAPIResourceList("otherGroup.apiVersion/OtherVersion", "otherResource", "OtherKind"),
			},
		},
		{ // NegativeTest: file doesn't exist
			testResourcePath:  testTemplatePath + "/wrongName_manifest.yaml",
			numResources:      0,
			expectError:       true,
			expectedResources: []*metav1.APIResourceList{},
		},
		{ // Avoid text/template no function found error when working with AnalysisTemplate/no template args
			testResourcePath: testTemplatePath + "/analysis-template.yaml",
			numResources:     1,
			expectError:      false,
			expectedResources: []*metav1.APIResourceList{
				newTestAPIResourceList("argoproj.io/v1alpha1", "args-test", "AnalysisTemplate"),
			},
		},
	}

	for _, test := range tests {
		resourceList, err := GetResources(&fakeDiscovery, nil, test.testResourcePath)

		g.Expect(len(resourceList)).To(gomega.Equal(test.numResources))
		if test.expectError {
			g.Expect(err).Should(gomega.HaveOccurred())
		} else {
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			for i, resource := range resourceList {
				g.Expect(resourceToApiResourceList(resource.Resource)).To(gomega.Equal(test.expectedResources[i]))
			}
		}
	}
}
