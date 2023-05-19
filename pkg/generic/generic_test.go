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

package generic

import (
	"testing"
)

func TestCommandExists(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Command SHOULD exist",
			args: args{
				command: "echo",
			},
			wantErr: false,
		},
		{
			name: "Command SHOULD NOT exist",
			args: args{
				command: "doesnotexist",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CommandExists(tt.args.command); (err != nil) != tt.wantErr {
				t.Errorf("CommandExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	type args struct {
		command       string
		args          string
		successOrFail string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Command FAILS and is EXPECTED TO",
			args: args{
				command:       "doesnotexist",
				args:          "not real",
				successOrFail: "fails",
			},
			wantErr: false,
		},
		{
			name: "Command FAILS and is NOT EXPECTED TO",
			args: args{
				command:       "doesnotexist",
				args:          "not real",
				successOrFail: "succeeds",
			},
			wantErr: true,
		},
		{
			name: "Command SUCCEEDS and is EXPECTED TO",
			args: args{
				command:       "echo",
				args:          "I want to succeed",
				successOrFail: "succeeds",
			},
			wantErr: false,
		},
		{
			name: "Command SUCCEEDS and is NOT EXPECTED TO",
			args: args{
				command:       "echo",
				args:          "what why would I succeed?",
				successOrFail: "fails",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RunCommand(tt.args.command, tt.args.args, tt.args.successOrFail); (err != nil) != tt.wantErr {
				t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
