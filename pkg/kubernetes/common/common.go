package common

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

// TODO: seems not used, check and delete
const (
	OperationCreate = "create"
	OperationSubmit = "submit"
	OperationUpdate = "update"
	OperationDelete = "delete"

	StateCreated = "created"
	StateDeleted = "deleted"
	//stateUpgraded = "upgraded"
	StateReady = "ready"
	StateFound = "found"
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
