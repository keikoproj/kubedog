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

// Package generic provides functions and steps implementations not specifically related to Kubernetes nor AWS.
package generic

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type TemplateArgument struct {
	Key                 string
	EnvironmentVariable string
	Fallback            string
	Mandatory           bool
}

// GetValue returns the value of the Environment Variable defined by 'TemplateArgument.EnvironmentVariable'.
// If 'TemplateArgument.EnvironmentVariable' is empty or the ENV. VAR. it defines is unset, 'TemplateArgument.Fallback' is returned.
// That is, if 'TemplateArgument.Mandatory' is not 'true', in which case, an error is returned.
func (ta TemplateArgument) GetValue() (string, error) {
	if ta.Key == "" {
		return "", errors.Errorf("'TemplateArgument.Key' can not be empty.")
	} else if value, ok := os.LookupEnv(ta.EnvironmentVariable); ok {
		return value, nil
	} else if ta.Mandatory {
		return "", errors.Errorf("'TemplateArgument.Mandatory'='true' but the Environment Variable '%s' defined by 'TemplateArgument.EnvironmentVariable' is not set", ta.EnvironmentVariable)
	} else {
		return ta.Fallback, nil
	}
}

// TemplateArgumentsToMap uses the elements of 'templateArguments' to populate the key:value pairs of the returned map.
// The key is the '.Key' variable of the corresponding element, and the value is the string returned by the 'GetValue' method of said element.
func TemplateArgumentsToMap(templateArguments ...TemplateArgument) (map[string]string, error) {
	args := map[string]string{}
	for i, ta := range templateArguments {
		value, err := ta.GetValue()
		if err != nil {
			return args, errors.Errorf("'templateArguments[%d].GetValue()' failed. 'templateArguments[%d]'='%v'. error: '%v'", i, i, ta, err)
		}
		args[ta.Key] = value
	}
	return args, nil
}

// GenerateFileFromTemplate applies the template defined in templatedFilePath to templateArgs.
// The generated file will be named 'generated_<templated-file-base>' and it will be created in the same directory of the template.
func GenerateFileFromTemplate(templatedFilePath string, templateArgs interface{}) (string, error) {
	t, err := template.ParseFiles(templatedFilePath)
	if err != nil {
		return "", errors.Errorf("Error parsing templated file '%s': %v", templatedFilePath, err)
	}

	templatedFileDir := filepath.Dir(templatedFilePath)
	templatedFileName := filepath.Base(templatedFilePath)
	generatedFilePath := filepath.Join(templatedFileDir, "generated_"+templatedFileName)
	f, err := os.Create(generatedFilePath)
	if err != nil {
		return "", errors.Errorf("Error creating generated file '%s': %v", generatedFilePath, err)
	}
	defer f.Close()

	err = t.Execute(f, templateArgs)
	if err != nil {
		return "", errors.Errorf("Error executing template '%v' against '%s': %v", templateArgs, templatedFilePath, err)
	}

	generated, err := ioutil.ReadFile(generatedFilePath)
	if err != nil {
		return "", errors.Errorf("Error reading generated file '%s': %v", generatedFilePath, err)
	}

	log.Infof("Generated file '%s': \n %s", generatedFilePath, string(generated))

	return generatedFilePath, nil
}
