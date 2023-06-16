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

package kube

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/keikoproj/kubedog/pkg/kube/common"
	"github.com/keikoproj/kubedog/pkg/kube/pod"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
)

type configuration struct {
	filesPath         string
	templateArguments interface{}
	waiterInterval    time.Duration
	waiterTries       int
}

func (kc *ClientSet) GetTimestamp(timestampName string) (time.Time, error) {
	commonErrorMessage := fmt.Sprintf("failed getting timestamp '%s'", timestampName)
	if kc.timestamps == nil {
		return time.Time{}, errors.Errorf("%s: 'ClientSet.Timestamps' is nil", commonErrorMessage)
	}
	timestamp, ok := kc.timestamps[timestampName]
	if !ok {
		return time.Time{}, errors.Errorf("%s: Timestamp not found", commonErrorMessage)
	}
	return timestamp, nil
}

func (kc *ClientSet) getResourcePath(resourceFileName string) string {
	templatesPath := kc.getTemplatesPath()
	return filepath.Join(templatesPath, resourceFileName)
}

func (kc *ClientSet) getTemplatesPath() string {
	defaultFilePath := "templates"
	if kc.config.filesPath != "" {
		return kc.config.filesPath
	}
	return defaultFilePath
}

func (kc *ClientSet) getWaiterInterval() time.Duration {
	defaultWaiterInterval := time.Second * 30
	if kc.config.waiterInterval > 0 {
		return kc.config.waiterInterval
	}
	return defaultWaiterInterval
}

func (kc *ClientSet) getWaiterTries() int {
	defaultWaiterTries := 40
	if kc.config.waiterTries > 0 {
		return kc.config.waiterTries
	}
	return defaultWaiterTries
}

func (kc *ClientSet) getWaiterConfig() common.WaiterConfig {
	return common.NewWaiterConfig(kc.getWaiterTries(), kc.getWaiterInterval())
}

func (kc *ClientSet) getExpBackoff() wait.Backoff {
	return pod.GetExpBackoff(kc.getWaiterTries())
}

func (kc *ClientSet) getDiscoveryClient() discovery.DiscoveryInterface {
	if kc.KubeInterface != nil {
		return kc.KubeInterface.Discovery()
	}
	return nil
}
