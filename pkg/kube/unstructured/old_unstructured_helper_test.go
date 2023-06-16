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

// func TestGetResources(t *testing.T) {
// 	var (
// 		dynScheme         = runtime.NewScheme()
// 		fakeDynamicClient = fakeDynamic.NewSimpleDynamicClient(dynScheme)
// 		fakeDiscovery     = fakeDiscovery.FakeDiscovery{}
// 		g                 = gomega.NewWithT(t)
// 		testTemplatePath  = "./test-fixtures"
// 	)

// 	expectedResources := []*metav1.APIResourceList{
// 		newTestAPIResourceList("someGroup.apiVersion/SomeVersion", "someResource", "SomeKind"),
// 		newTestAPIResourceList("otherGroup.apiVersion/OtherVersion", "otherResource", "OtherKind"),
// 		newTestAPIResourceList("argoproj.io/v1alpha1", "AnalysisTemplate", "AnalysisTemplate"),
// 	}

// 	fakeDiscovery.Fake = &fakeDynamicClient.Fake
// 	fakeDiscovery.Resources = append(fakeDiscovery.Resources, expectedResources...)

// 	resourceToApiResourceList := func(resource *unstructured.Unstructured) *metav1.APIResourceList {
// 		return newTestAPIResourceList(
// 			resource.GetAPIVersion(),
// 			resource.GetName(),
// 			resource.GetKind(),
// 		)
// 	}

// 	tests := []struct {
// 		testResourcePath  string
// 		numResources      int
// 		expectError       bool
// 		expectedResources []*metav1.APIResourceList
// 	}{
// 		{ // PositiveTest
// 			testResourcePath: testTemplatePath + "/multi-resource.yaml",
// 			numResources:     2,
// 			expectError:      false,
// 			expectedResources: []*metav1.APIResourceList{
// 				newTestAPIResourceList("someGroup.apiVersion/SomeVersion", "someResource", "SomeKind"),
// 				newTestAPIResourceList("otherGroup.apiVersion/OtherVersion", "otherResource", "OtherKind"),
// 			},
// 		},
// 		{ // NegativeTest: file doesn't exist
// 			testResourcePath:  testTemplatePath + "/wrongName_manifest.yaml",
// 			numResources:      0,
// 			expectError:       true,
// 			expectedResources: []*metav1.APIResourceList{},
// 		},
// 		{ // Avoid text/template no function found error when working with AnalysisTemplate/no template args
// 			testResourcePath: testTemplatePath + "/analysis-template.yaml",
// 			numResources:     1,
// 			expectError:      false,
// 			expectedResources: []*metav1.APIResourceList{
// 				newTestAPIResourceList("argoproj.io/v1alpha1", "args-test", "AnalysisTemplate"),
// 			},
// 		},
// 	}

// 	for _, test := range tests {
// 		resourceList, err := GetResources(&fakeDiscovery, nil, test.testResourcePath)

// 		g.Expect(len(resourceList)).To(gomega.Equal(test.numResources))
// 		if test.expectError {
// 			g.Expect(err).Should(gomega.HaveOccurred())
// 		} else {
// 			g.Expect(err).ShouldNot(gomega.HaveOccurred())
// 			for i, resource := range resourceList {
// 				g.Expect(resourceToApiResourceList(resource.Resource)).To(gomega.Equal(test.expectedResources[i]))
// 			}
// 		}
// 	}
// }

// func newTestAPIResourceList(apiVersion, name, kind string) *metav1.APIResourceList {
// 	return &metav1.APIResourceList{
// 		GroupVersion: apiVersion,
// 		APIResources: []metav1.APIResource{
// 			{
// 				Name:       name,
// 				Kind:       kind,
// 				Namespaced: true,
// 			},
// 		},
// 	}
// }
