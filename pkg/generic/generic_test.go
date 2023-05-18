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
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
)

func TestGetValue(t *testing.T) {
	var (
		g     = gomega.NewWithT(t)
		tests = []struct {
			templateArgument TemplateArgument
			setup            func()
			expectedValue    string
			expectError      bool
		}{
			// PositiveTests:
			{ // Mandatory, EnvironmentVariable set
				templateArgument: TemplateArgument{
					Key:                 "key1",
					EnvironmentVariable: "VAR1",
					Mandatory:           true,
					Default:             "fallback1",
				},
				setup: func() {
					os.Setenv("VAR1", "value1")
				},
				expectedValue: "value1",
				expectError:   false,
			},
			{ // not Mandatory, EnvironmentVariable unset, Default not empty
				templateArgument: TemplateArgument{
					Key:                 "key2",
					EnvironmentVariable: "VAR2",
					Mandatory:           false,
					Default:             "fallback2",
				},
				setup: func() {
					os.Unsetenv("VAR2")
				},
				expectedValue: "fallback2",
				expectError:   false,
			},
			{ // not Mandatory, EnvironmentVariable set empty
				templateArgument: TemplateArgument{
					Key:                 "key3",
					EnvironmentVariable: "VAR3",
					Mandatory:           false,
					Default:             "fallback3",
				},
				setup: func() {
					os.Setenv("VAR3", "")
				},
				expectedValue: "",
				expectError:   false,
			},
			{ // not Mandatory, EnvironmentVariable unset, Default empty
				templateArgument: TemplateArgument{
					Key:                 "key4",
					EnvironmentVariable: "VAR4",
					Mandatory:           false,
					Default:             "",
				},
				setup: func() {
					os.Unsetenv("VAR4")
				},
				expectedValue: "",
				expectError:   false,
			},
			{ // not Mandatory, EnvironmentVariable empty
				templateArgument: TemplateArgument{
					Key:                 "key5",
					EnvironmentVariable: "",
					Mandatory:           false,
					Default:             "fallback5",
				},
				setup:         func() {},
				expectedValue: "fallback5",
				expectError:   false,
			},
			// NegativeTests:
			{ // Mandatory, EnvironmentVariable unset
				templateArgument: TemplateArgument{
					Key:                 "key",
					EnvironmentVariable: "VAR",
					Mandatory:           true,
					Default:             "fallback",
				},
				setup: func() {
					os.Unsetenv("VAR")
				},
				expectedValue: "",
				expectError:   true,
			},
			{ // Key empty
				templateArgument: TemplateArgument{
					Key:                 "",
					EnvironmentVariable: "VAR",
					Mandatory:           true,
					Default:             "fallback",
				},
				setup: func() {
					os.Setenv("VAR", "value")
				},
				expectedValue: "",
				expectError:   true,
			},
			{ // Mandatory, EnvironmentVariable empty
				templateArgument: TemplateArgument{
					Key:                 "key",
					EnvironmentVariable: "",
					Mandatory:           true,
					Default:             "fallback",
				},
				setup: func() {
					os.Setenv("VAR", "value")
				},
				expectedValue: "",
				expectError:   true,
			},
		}
	)

	for _, test := range tests {
		test.setup()
		value, err := test.templateArgument.GetValue()
		if test.expectError {
			g.Expect(err).Should(gomega.HaveOccurred())
		} else {
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		}
		g.Expect(value).To(gomega.Equal(test.expectedValue))
	}
}

func TestTemplateArgumentsToMap(t *testing.T) {
	var (
		g     = gomega.NewWithT(t)
		tests = []struct {
			templateArguments []TemplateArgument
			setup             func()
			expectedArgs      map[string]string
			expectError       bool
		}{
			{ // PositiveTest
				templateArguments: []TemplateArgument{
					{ // Mandatory, EnvironmentVariable set
						Key:                 "key1",
						EnvironmentVariable: "VAR1",
						Mandatory:           true,
						Default:             "fallback1",
					},
					{ // not Mandatory, EnvironmentVariable unset, Default not empty
						Key:                 "key2",
						EnvironmentVariable: "VAR2",
						Mandatory:           false,
						Default:             "fallback2",
					},
					{ // not Mandatory, EnvironmentVariable set empty
						Key:                 "key3",
						EnvironmentVariable: "VAR3",
						Mandatory:           false,
						Default:             "fallback3",
					},
					{ // not Mandatory, EnvironmentVariable unset, Default empty
						Key:                 "key4",
						EnvironmentVariable: "VAR4",
						Mandatory:           false,
						Default:             "",
					},
					{ // not Mandatory, EnvironmentVariable empty
						Key:                 "key5",
						EnvironmentVariable: "",
						Mandatory:           false,
						Default:             "fallback5",
					},
				},
				setup: func() {
					os.Setenv("VAR1", "value1")
					os.Unsetenv("VAR2")
					os.Setenv("VAR3", "")
					os.Unsetenv("VAR4")
				},
				expectedArgs: map[string]string{
					"key1": "value1",
					"key2": "fallback2",
					"key3": "",
					"key4": "",
					"key5": "fallback5",
				},
				expectError: false,
			},
		}
	)

	for _, test := range tests {
		test.setup()
		args, err := TemplateArgumentsToMap(test.templateArguments...)
		if test.expectError {
			g.Expect(err).Should(gomega.HaveOccurred())
		} else {
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		}
		g.Expect(args).To(gomega.Equal(test.expectedArgs))
	}
}

func TestGenerateFileFromTemplate(t *testing.T) {
	type templateArgs struct {
		Kind       string
		ApiVersion string
		Name       string
	}

	fileToString := func(filePath string) string {
		file, _ := os.ReadFile(filePath)
		return string(file)
	}

	templateArgsToExpectedFileAsString := func(args templateArgs) string {
		return `kind: ` + args.Kind + `
apiVersion: ` + args.ApiVersion + `
metadata:
  name: ` + args.Name
	}

	var (
		g                    = gomega.NewWithT(t)
		testTemplatesPath, _ = filepath.Abs("../../test/templates")
		tests                = []struct {
			templatedFilePath string
			args              templateArgs
			expectedFilePath  string
			expectError       bool
		}{
			{ // PositiveTest
				templatedFilePath: testTemplatesPath + "/manifest.yaml",
				args: templateArgs{
					Kind:       "myKind",
					ApiVersion: "myApiVersion",
					Name:       "myName",
				},
				expectedFilePath: testTemplatesPath + "/generated_manifest.yaml",
				expectError:      false,
			},
			{ // NegativeTest: template.ParseFiles fails
				templatedFilePath: testTemplatesPath + "/wrongName_manifest.yaml",
				args: templateArgs{
					Kind:       "myKind",
					ApiVersion: "myApiVersion",
					Name:       "myName",
				},
				expectedFilePath: "",
				expectError:      true,
			},
			{ // NegativeTest: template.Execute fails
				templatedFilePath: testTemplatesPath + "/badKind_manifest.yaml",
				args: templateArgs{
					Kind:       "myKind",
					ApiVersion: "myApiVersion",
					Name:       "myName",
				},
				expectedFilePath: "",
				expectError:      true,
			},
		}
	)

	for _, test := range tests {
		generatedFilePath, err := GenerateFileFromTemplate(test.templatedFilePath, test.args)

		g.Expect(generatedFilePath).To(gomega.Equal(test.expectedFilePath))

		if test.expectError {
			g.Expect(err).Should(gomega.HaveOccurred())
		} else {
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			generatedFileString := fileToString(generatedFilePath)
			expectedFileString := templateArgsToExpectedFileAsString(test.args)
			g.Expect(generatedFileString).To(gomega.Equal(expectedFileString))
		}
	}
}

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
