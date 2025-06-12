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

package pod

import (
	"testing"

	"github.com/keikoproj/kubedog/internal/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	fakeDiscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_PodsInNamespaceWithSelectorShouldHaveLabels(t *testing.T) {
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
	podWithLabels1 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-foo-xhhxj",
			Namespace: "foo",
			Labels: map[string]string{
				"app":   "foo",
				"label": "true",
			},
		},
	}
	podWithLabels2 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-foo-xhhzd",
			Namespace: "foo",
			Labels: map[string]string{
				"app":   "foo",
				"label": "true",
			},
		},
	}
	podMissingLabel := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-foo-xhhzr",
			Namespace: "foo",
			Labels: map[string]string{
				"app": "foo",
			},
		},
	}
	clientNoErr := fake.NewSimpleClientset(&ns, &podWithLabels1, &podWithLabels2)
	clientErr := fake.NewSimpleClientset(&ns, &podWithLabels1, &podWithLabels2, &podMissingLabel)
	dynScheme := runtime.NewScheme()
	fakeDynamicClient := fakeDynamic.NewSimpleDynamicClient(dynScheme)
	fakeDiscoveryClient := fakeDiscovery.FakeDiscovery{}

	type fields struct {
		KubeInterface      kubernetes.Interface
		DynamicInterface   dynamic.Interface
		DiscoveryInterface *fakeDiscovery.FakeDiscovery
	}
	type args struct {
		namespace string
		selector  string
		labels    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "No pods found",
			fields: fields{
				KubeInterface:      clientErr,
				DiscoveryInterface: &fakeDiscoveryClient,
				DynamicInterface:   fakeDynamicClient,
			},
			args: args{
				selector:  "app=doesnotexist",
				namespace: "foo",
				labels:    "app=foo,label=true",
			},
			wantErr: true,
		},
		{
			name: "Pods should have labels",
			fields: fields{
				KubeInterface:      clientNoErr,
				DiscoveryInterface: &fakeDiscoveryClient,
				DynamicInterface:   fakeDynamicClient,
			},
			args: args{
				selector:  "app=foo",
				namespace: "foo",
				labels:    "app=foo,label=true",
			},
			wantErr: false,
		},
		{
			name: "Error from pod missing label",
			fields: fields{
				KubeInterface:      clientErr,
				DiscoveryInterface: &fakeDiscoveryClient,
				DynamicInterface:   fakeDynamicClient,
			},
			args: args{
				selector:  "app=foo",
				namespace: "foo",
				labels:    "app=foo,label=true",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PodsInNamespaceWithSelectorShouldHaveLabels(tt.fields.KubeInterface, tt.args.namespace, tt.args.selector, tt.args.labels); (err != nil) != tt.wantErr {
				t.Errorf("ThePodsInNamespaceWithSelectorShouldHaveLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPodsInNamespaceWithLabelSelectorConvergeToFieldSelector(t *testing.T) {
	type args struct {
		kubeClientset kubernetes.Interface
		expBackoff    wait.Backoff
		namespace     string
		labelSelector string
		fieldSelector string
	}
	namespaceName := "test-ns"
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
	podSucceeded := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-xhhxj",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app": "test-service",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test",
			args: args{
				kubeClientset: fake.NewSimpleClientset(&ns, &podSucceeded),
				expBackoff:    util.DefaultRetry,
				namespace:     namespaceName,
				labelSelector: "app=test-service",
				fieldSelector: "status.phase=Succeeded",
			},
		},
		{
			name: "Negative Test: no pods with label selector",
			args: args{
				kubeClientset: fake.NewSimpleClientset(&ns),
				expBackoff:    util.DefaultRetry,
				namespace:     namespaceName,
				labelSelector: "app=test-service",
				fieldSelector: "status.phase=Succeeded",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PodsInNamespaceWithLabelSelectorConvergeToFieldSelector(tt.args.kubeClientset, tt.args.expBackoff, tt.args.namespace, tt.args.labelSelector, tt.args.fieldSelector); (err != nil) != tt.wantErr {
				t.Errorf("PodsInNamespaceWithLabelSelectorConvergeToFieldSelector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeletePodListWithLabelSelectorAndFieldSelector(t *testing.T) {
	namespaceName := "test-ns"
	ns := v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
	podSucceeded := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-xhhxj",
			Namespace: "test-ns",
			Labels: map[string]string{
				"app": "test-service",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}
	type args struct {
		kubeClientset kubernetes.Interface
		namespace     string
		labelSelector string
		fieldSelector string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test",
			args: args{
				kubeClientset: fake.NewSimpleClientset(&ns, &podSucceeded),
				namespace:     namespaceName,
				labelSelector: "app=test-service",
				fieldSelector: "status.phase=Succeeded",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeletePodListWithLabelSelectorAndFieldSelector(tt.args.kubeClientset, tt.args.namespace, tt.args.labelSelector, tt.args.fieldSelector); (err != nil) != tt.wantErr {
				t.Errorf("DeletePodListWithLabelSelectorAndFieldSelector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
