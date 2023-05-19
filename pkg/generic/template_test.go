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
