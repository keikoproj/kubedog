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
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	util "github.com/keikoproj/kubedog/internal/utilities"
	log "github.com/sirupsen/logrus"
)

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
