package common

import (
	"os"
	"os/user"
	"reflect"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	retriableErrors = []string{
		"Unable to reach the kubernetes API",
		"Unable to connect to the server",
		"EOF",
		"transport is closing",
		"the object has been modified",
		"an error on the server",
	}

	DefaultRetry = wait.Backoff{
		Steps:    6,
		Duration: 1000 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}
)

func IsRetriable(err error) bool {
	for _, msg := range retriableErrors {
		if strings.Contains(err.Error(), msg) {
			return true
		}
	}
	return false
}

func RetryOnError(backoff *wait.Backoff, retryExpected func(error) bool, fn condFunc) (interface{}, error) {
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

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetUsernamePrefix() string {
	currUser, err := user.Current()
	if err != nil || currUser.Username == "root" {
		return ""
	}
	return currUser.Username + "-"
}
