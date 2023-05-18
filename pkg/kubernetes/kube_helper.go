package kube

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/keikoproj/kubedog/pkg/kubernetes/common"
	"github.com/keikoproj/kubedog/pkg/kubernetes/pod"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (kc *ClientSet) getTimestamp(timestampName string) (time.Time, error) {
	commonErrorMessage := fmt.Sprintf("failed getting timestamp '%s'", timestampName)
	if kc.Timestamps == nil {
		return time.Time{}, errors.Errorf("%s: 'ClientSet.Timestamps' is nil", commonErrorMessage)
	}
	timestamp, ok := kc.Timestamps[timestampName]
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
	if kc.FilesPath != "" {
		return kc.FilesPath
	}
	return defaultFilePath
}

func (kc *ClientSet) getWaiterInterval() time.Duration {
	defaultWaiterInterval := time.Second * 30
	if kc.WaiterInterval > 0 {
		return kc.WaiterInterval
	}
	return defaultWaiterInterval
}

func (kc *ClientSet) getWaiterTries() int {
	defaultWaiterTries := 40
	if kc.WaiterTries > 0 {
		return kc.WaiterTries
	}
	return defaultWaiterTries
}

func (kc *ClientSet) getWaiterConfig() common.WaiterConfig {
	return common.NewWaiterConfig(kc.getWaiterTries(), kc.getWaiterInterval())
}

func (kc *ClientSet) getExpBackoff() wait.Backoff {
	return pod.GetExpBackoff(kc.getWaiterTries())
}
