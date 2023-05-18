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

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	DefaultRetry = wait.Backoff{
		Steps:    6,
		Duration: 1000 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}
	retriableErrors = []string{
		"Unable to reach the kubernetes API",
		"Unable to connect to the server",
		"EOF",
		"transport is closing",
		"the object has been modified",
		"an error on the server",
	}
)

type FuncToRetryWithReturn func() (interface{}, error)
type FuncToRetry func() error

func PathToOSFile(relativePath string) (*os.File, error) {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed generate absolute file path of %s", relativePath))
	}

	manifest, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to open file %s", path))
	}

	return manifest, nil
}

func DeleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func IsRetriable(err error) bool {
	for _, msg := range retriableErrors {
		if strings.Contains(err.Error(), msg) {
			return true
		}
	}
	return false
}

func RetryOnError(backoff *wait.Backoff, retryExpected func(error) bool, fn FuncToRetryWithReturn) (interface{}, error) {
	var ex, lastErr error
	var out interface{}
	caller := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	err := wait.ExponentialBackoff(*backoff, func() (bool, error) {
		out, ex = fn()
		switch {
		case ex == nil:
			return true, nil
		case retryExpected(ex):
			lastErr = ex
			log.Warnf("A caller %v retried due to exception: %v", caller, ex)
			return false, nil
		default:
			return false, ex
		}
	})
	if err == wait.ErrWaitTimeout {
		err = lastErr
	}
	return out, err
}

func RetryOnAnyError(backoff *wait.Backoff, fn FuncToRetry) error {
	_, err := RetryOnError(backoff, func(err error) bool {
		return true
	}, func() (interface{}, error) {
		return nil, fn()
	})
	return err
}
