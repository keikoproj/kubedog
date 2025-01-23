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
	"testing"
)

var (
	sampleResource = map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"containers": []map[string]any{
					{
						"name":    "someContainer",
						"image":   "someImage",
						"version": 1.4,
						"ports": []map[string]any{
							{"containerPort": 8080},
							{"containerPort": 8940},
						},
					},
				},
			},
		},
	}
)

func TestExtractField(t *testing.T) {
	type args struct {
		data          any
		path          []string
		expectedValue any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive Test",
			args: args{
				data:          sampleResource,
				path:          []string{"spec", "template", "containers[0]", "name"},
				expectedValue: "someContainer",
			},
		},
		{
			name: "Positive Test - multiple array",
			args: args{
				data:          sampleResource,
				path:          []string{"spec", "template", "containers[0]", "ports[1]", "containerPort"},
				expectedValue: 8940,
			},
		},
		{
			name: "Negative Test",
			args: args{
				data:          sampleResource,
				path:          []string{"spec", "path", "doesnt", "exist"},
				expectedValue: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if extractedValue, err := ExtractField(tt.args.data, tt.args.path); (err != nil) != tt.wantErr || extractedValue != tt.args.expectedValue {
				t.Errorf("ExtractField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
