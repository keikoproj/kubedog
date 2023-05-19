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

// TODO: restruct this like kube and aws?
package generic

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	util "github.com/keikoproj/kubedog/internal/utilities"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type TemplateArgument struct {
	Key                 string
	EnvironmentVariable string
	Default             string
	Mandatory           bool
}

var (
	KubernetesClusterTagKey = "KubernetesCluster"
)

// GetValue returns the value of the Environment Variable defined by 'TemplateArgument.EnvironmentVariable'.
// If 'TemplateArgument.EnvironmentVariable' is empty or the ENV. VAR. it defines is unset, 'TemplateArgument.Default' is returned.
// That is, if 'TemplateArgument.Mandatory' is not 'true', in which case, an error is returned.
func (ta TemplateArgument) GetValue() (string, error) {
	if ta.Key == "" {
		return "", errors.Errorf("'TemplateArgument.Key' can not be empty.")
	} else if value, ok := os.LookupEnv(ta.EnvironmentVariable); ok {
		return value, nil
	} else if ta.Mandatory {
		return "", errors.Errorf("'TemplateArgument.Mandatory'='true' but the Environment Variable '%s' defined by 'TemplateArgument.EnvironmentVariable' is not set", ta.EnvironmentVariable)
	} else {
		return ta.Default, nil
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
	generated, err := os.ReadFile(generatedFilePath)
	if err != nil {
		return "", errors.Errorf("Error reading generated file '%s': %v", generatedFilePath, err)
	}

	log.Infof("Generated file '%s': \n %s", generatedFilePath, string(generated))

	return generatedFilePath, nil
}

func WaitFor(duration int, durationUnits string) error {
	switch durationUnits {
	case util.DurationMinutes:
		increment := 1
		d := increment
		for d <= duration {
			time.Sleep(time.Duration(increment) * time.Minute)
			log.Infof("waited '%d' out of '%d' '%s'", d, duration, durationUnits)
			d += increment
		}
		return nil
	case util.DurationSeconds:
		increment := 30
		d := increment
		for d <= duration {
			time.Sleep(time.Duration(increment) * time.Second)
			log.Infof("waited '%d' out of '%d' '%s'", d, duration, durationUnits)
			d += increment
		}
		lastIncrement := duration - d + increment
		if lastIncrement > 0 {
			time.Sleep(time.Duration(lastIncrement) * time.Second)
			d += lastIncrement - increment
			log.Infof("waited '%d' out of '%d' '%s'", d, duration, durationUnits)
		}
		return nil
	default:
		return fmt.Errorf("unsupported duration units: '%s'", durationUnits)
	}
}

func CommandExists(command string) error {
	if _, err := exec.LookPath(command); err != nil {
		return err
	}

	return nil
}

func RunCommand(command string, args, successOrFail string) error {
	// split to support args being passed from .feature file.
	// slice param type not supported by godog.
	splitArgs := strings.Split(args, " ")
	toRun := exec.Command(command, splitArgs...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	toRun.Stdout = &stdout
	toRun.Stderr = &stderr

	cmdStr := toRun.String()
	log.Infof("Running command: %s", cmdStr)
	err := toRun.Run()
	if successOrFail == "succeeds" && err != nil {
		return fmt.Errorf("command %s did not succeed: %s", cmdStr, stderr.String())
	}

	if successOrFail == "fails" && err == nil {
		return fmt.Errorf("command %s succeeded but was expected to fail: %s", cmdStr, stdout.String())
	}

	return nil
}
