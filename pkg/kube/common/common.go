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

package common

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	OperationCreate = "create"
	OperationSubmit = "submit"
	OperationUpdate = "update"
	OperationDelete = "delete"

	StateCreated  = "created"
	StateDeleted  = "deleted"
	StateUpgraded = "upgraded"
	StateReady    = "ready"
	StateFound    = "found"
)

type WaiterConfig struct {
	tries    int
	interval time.Duration
}

func NewWaiterConfig(tries int, interval time.Duration) WaiterConfig {
	return WaiterConfig{tries: tries, interval: interval}
}

func (w WaiterConfig) GetInterval() time.Duration {
	defaultWaiterInterval := time.Second * 30
	if w.interval > 0 {
		return w.interval
	}
	return defaultWaiterInterval
}

func (w WaiterConfig) GetTries() int {
	defaultWaiterTries := 40
	if w.tries > 0 {
		return w.tries
	}
	return defaultWaiterTries
}

func ValidateClientset(kubeClientset kubernetes.Interface) error {
	if kubeClientset == nil {
		return errors.Errorf("'k8s.io/client-go/kubernetes.Interface' is nil.")
	}
	return nil
}
